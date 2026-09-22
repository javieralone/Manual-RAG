import argparse
import json
import logging
import re
import subprocess
import sys
from contextlib import contextmanager
from pathlib import Path

SERVICE_DIR = Path(__file__).resolve().parent.parent
BASE_DIR = SERVICE_DIR.parent.parent
SCRIPTS_DIR = SERVICE_DIR / "scripts"
DOCUMENTS_DIR = BASE_DIR / "data" / "documents"
NEW_DIR = DOCUMENTS_DIR / "new"
READING_DIR = DOCUMENTS_DIR / "reading"
COMPLETED_DIR = DOCUMENTS_DIR / "completed"
LOG_DIR = BASE_DIR / "logs"
LOG_FILE = LOG_DIR / "ingestion.log"
LOCK_FILE = LOG_DIR / "ingestion.lock"
DEFAULT_COLLECTION = "generic_manuals"
VALID_COLLECTION_RE = re.compile(r"^[A-Za-z0-9_-]+$")
PART_PATTERN = re.compile(r"^(?P<document>.+)__parte-(?P<part>\d+)\.pdf$", re.IGNORECASE)


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Procesa los PDFs nuevos y los carga en Qdrant.")
    parser.add_argument("--pdf", type=Path, help="Procesa un PDF concreto sin usar la cola.")
    parser.add_argument("--collection", default=None, help="Colección Qdrant destino. Si no se indica, se infiere desde la carpeta data/documents/new/<colección>.")
    parser.add_argument("--pages", type=int, default=None, help="Número máximo de páginas a procesar por PDF. Si no se indica, se procesan todas.")
    parser.add_argument("--dpi", type=int, default=150)
    parser.add_argument("--language", default="spa")
    parser.add_argument("--python", default=sys.executable)

    # Parámetros de optimización para ocr_manual.py
    parser.add_argument(
        "--memory-mode",
        choices=["disk", "memory"],
        default="disk",
        help="Modo de gestión de RAM en el OCR ('disk' procesa página a página, 'memory' carga todo en RAM)."
    )
    parser.add_argument(
        "--workers",
        type=int,
        default=3,
        help="Número de procesos en paralelo para el OCR (3 es ideal para tu i7 y 8GB RAM)."
    )
    parser.add_argument(
        "--fast-ocr",
        action="store_true",
        help="Desactiva el preprocesamiento y realiza una sola pasada de OCR."
    )
    return parser.parse_args()


def configure_logging() -> logging.Logger:
    LOG_DIR.mkdir(parents=True, exist_ok=True)
    logging.basicConfig(
        filename=LOG_FILE,
        level=logging.INFO,
        format="%(asctime)s %(levelname)s file=%(filename)s message=%(message)s",
    )
    return logging.getLogger("manual-ingestion")


@contextmanager
def ingestion_lock():
    LOG_DIR.mkdir(parents=True, exist_ok=True)
    try:
        handle = LOCK_FILE.open("x", encoding="utf-8")
    except FileExistsError as error:
        raise RuntimeError(f"Ya hay una ingesta en curso: {LOCK_FILE}") from error

    try:
        handle.write(str(Path.cwd()))
        handle.close()
        yield
    finally:
        LOCK_FILE.unlink(missing_ok=True)


def validate_collection_name(value: str | None, default: str = DEFAULT_COLLECTION) -> str:
    candidate = (value or default).strip()
    if not candidate or candidate == "default_collection":
        return default
    if not VALID_COLLECTION_RE.fullmatch(candidate):
        raise ValueError(
            f"Colección inválida '{candidate}'. Usa solo letras, números, guion y guion bajo."
        )
    return candidate


def infer_collection_name(pdf_path: Path, explicit: str | None = None) -> str:
    if explicit:
        return validate_collection_name(explicit)

    resolved = pdf_path.resolve()
    for parent in reversed(resolved.parents):
        if parent.name == "new" and parent.parent == DOCUMENTS_DIR:
            continue
        if parent == NEW_DIR:
            continue
        if parent.parent == NEW_DIR:
            return validate_collection_name(parent.name)
    return DEFAULT_COLLECTION


def document_identity(pdf_path: Path) -> tuple[str, int]:
    match = PART_PATTERN.match(pdf_path.name)
    if match:
        return match.group("document"), int(match.group("part"))
    return pdf_path.stem, 1


def run_step(name: str, command: list[str]) -> None:
    print(f"\n=== {name} ===", flush=True)
    print(" ".join(command), flush=True)
    subprocess.run(command, cwd=BASE_DIR, check=True)


def require_non_empty_json(path: Path, key: str) -> None:
    if not path.exists():
        raise RuntimeError(f"El paso anterior no genero {path}")

    try:
        with path.open(encoding="utf-8") as file:
            data = json.load(file)
    except (OSError, json.JSONDecodeError) as error:
        raise RuntimeError(f"No se pudo validar {path}: {error}") from error

    if not isinstance(data, list) or not any(
        isinstance(item, dict) and str(item.get(key, "")).strip() for item in data
    ):
        raise RuntimeError(f"El archivo {path} no contiene datos utilizables")


def process_pdf(pdf_path: Path, arguments: argparse.Namespace, collection_name: str | None = None) -> str:
    pages_path = BASE_DIR / "data" / "artifacts" / "manual_pages.json"
    chunks_path = BASE_DIR / "data" / "artifacts" / "manual_chunks.json"
    document_id, part = document_identity(pdf_path)
    collection_name = validate_collection_name(collection_name or infer_collection_name(pdf_path, arguments.collection))

    ocr_command = [
        arguments.python, str(SCRIPTS_DIR / "ocr_manual_opt.py"),
        "--pdf", str(pdf_path),
        "--output", str(pages_path),
        "--document-id", document_id,
        "--part", str(part),
        "--collection", collection_name,
        "--dpi", str(arguments.dpi),
        "--language", arguments.language,
        "--memory-mode", arguments.memory_mode,
        "--workers", str(arguments.workers),
    ]

    if arguments.fast_ocr:
        ocr_command.append("--fast-ocr")
    if arguments.pages is not None:
        ocr_command.extend(["--pages", str(arguments.pages)])

    run_step("1/3 OCR", ocr_command)
    require_non_empty_json(pages_path, "text")

    run_step("2/3 indexacion", [arguments.python, str(SCRIPTS_DIR / "index_manual.py")])
    require_non_empty_json(chunks_path, "text")

    run_step("3/3 carga en Qdrant", [arguments.python, str(SCRIPTS_DIR / "upload_to_qdrant.py"), "--collection", collection_name])
    return collection_name


def process_queued_pdf(source: Path, arguments: argparse.Namespace, logger: logging.Logger) -> bool | None:
    collection_name = infer_collection_name(source, arguments.collection)
    reading_path = READING_DIR / collection_name / source.name
    completed_path = COMPLETED_DIR / collection_name / source.name
    original_path = source
    reading_path.parent.mkdir(parents=True, exist_ok=True)
    completed_path.parent.mkdir(parents=True, exist_ok=True)
    try:
        source.rename(reading_path)
        logger.info("file=%s collection=%s stage=reading", source.name, collection_name)
        process_pdf(reading_path, arguments, collection_name)
        reading_path.rename(completed_path)
        logger.info("file=%s collection=%s stage=completed", source.name, collection_name)
        print(f"Completado: {source.name} [{collection_name}]", flush=True)
        return True
    except KeyboardInterrupt:
        logger.exception("file=%s collection=%s stage=interrupted", source.name, collection_name)
        if reading_path.exists():
            reading_path.rename(original_path)
        print(f"Interrumpido: {source.name}. Volvio a {original_path.parent}", flush=True)
        return None
    except Exception:
        logger.exception("file=%s collection=%s stage=failed", source.name, collection_name)
        if reading_path.exists():
            reading_path.rename(original_path)
        print(f"Fallo: {source.name}. Volvio a {original_path.parent}. Revisa {LOG_FILE}", flush=True)
        return False


def main() -> int:
    arguments = parse_arguments()
    logger = configure_logging()
    for directory in (NEW_DIR, READING_DIR, COMPLETED_DIR):
        directory.mkdir(parents=True, exist_ok=True)

    with ingestion_lock():
        if arguments.pdf:
            pdf_path = arguments.pdf.resolve()
            if not pdf_path.exists() or pdf_path.suffix.lower() != ".pdf":
                raise FileNotFoundError(f"El archivo no existe o no es PDF: {pdf_path}")
            process_pdf(pdf_path, arguments, validate_collection_name(arguments.collection) if arguments.collection else infer_collection_name(pdf_path))
            return 0

        queued_files = []
        for collection_dir in sorted(NEW_DIR.iterdir(), key=lambda item: item.name):
            if not collection_dir.is_dir():
                continue
            queued_files.extend(
                sorted(
                    path for path in collection_dir.rglob("*.pdf") if path.is_file()
                )
            )
        if not queued_files:
            print(f"No hay PDFs nuevos en {NEW_DIR}")
            return 0

        failed = 0
        for pdf_path in queued_files:
            result = process_queued_pdf(pdf_path, arguments, logger)
            if result is None:
                return 130
            if not result:
                failed += 1
        if failed:
            print(f"Proceso terminado con {failed} archivo(s) fallido(s).", flush=True)
            return 1

    print("Todos los PDFs nuevos fueron procesados correctamente.", flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())