import argparse
import json
import logging
import re
import subprocess
import sys
from contextlib import contextmanager
from pathlib import Path


BASE_DIR = Path(__file__).resolve().parent.parent.parent
SCRIPTS_DIR = BASE_DIR / "rag-engine" / "scripts"
DOCUMENTS_DIR = BASE_DIR / "documents"
NEW_DIR = DOCUMENTS_DIR / "new"
READING_DIR = DOCUMENTS_DIR / "reading"
COMPLETED_DIR = DOCUMENTS_DIR / "completed"
LOG_DIR = BASE_DIR / "logs"
LOG_FILE = LOG_DIR / "ingestion.log"
LOCK_FILE = LOG_DIR / "ingestion.lock"
PART_PATTERN = re.compile(r"^(?P<document>.+)__parte-(?P<part>\d+)\.pdf$", re.IGNORECASE)


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Procesa los PDFs nuevos y los carga en Qdrant.")
    parser.add_argument("--pdf", type=Path, help="Procesa un PDF concreto sin usar la cola.")
    parser.add_argument("--dpi", type=int, default=200)
    parser.add_argument("--language", default="spa")
    parser.add_argument("--python", default=sys.executable)
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


def process_pdf(pdf_path: Path, arguments: argparse.Namespace) -> None:
    pages_path = BASE_DIR / "output" / "manual_pages.json"
    chunks_path = BASE_DIR / "output" / "manual_chunks.json"
    document_id, part = document_identity(pdf_path)

    run_step("1/3 OCR", [
        arguments.python, str(SCRIPTS_DIR / "ocr_manual.py"),
        "--pdf", str(pdf_path), "--output", str(pages_path),
        "--document-id", document_id, "--part", str(part),
        "--dpi", str(arguments.dpi), "--language", arguments.language,
    ])
    require_non_empty_json(pages_path, "text")

    run_step("2/3 indexacion", [arguments.python, str(SCRIPTS_DIR / "index_manual.py")])
    require_non_empty_json(chunks_path, "text")

    run_step("3/3 carga en Qdrant", [arguments.python, str(SCRIPTS_DIR / "upload_to_qdrant.py")])


def process_queued_pdf(source: Path, arguments: argparse.Namespace, logger: logging.Logger) -> bool | None:
    reading_path = READING_DIR / source.name
    try:
        source.rename(reading_path)
        logger.info("file=%s stage=reading", source.name)
        process_pdf(reading_path, arguments)
        reading_path.rename(COMPLETED_DIR / source.name)
        logger.info("file=%s stage=completed", source.name)
        print(f"Completado: {source.name}", flush=True)
        return True
    except KeyboardInterrupt:
        logger.exception("file=%s stage=interrupted", source.name)
        if reading_path.exists():
            reading_path.rename(NEW_DIR / source.name)
        print(f"Interrumpido: {source.name}. Volvio a new.", flush=True)
        return None
    except Exception:
        logger.exception("file=%s stage=failed", source.name)
        if reading_path.exists():
            reading_path.rename(NEW_DIR / source.name)
        print(f"Fallo: {source.name}. Volvio a new. Revisa {LOG_FILE}", flush=True)
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
            process_pdf(pdf_path, arguments)
            return 0

        queued_files = sorted(
            path for path in NEW_DIR.iterdir()
            if path.is_file() and path.suffix.lower() == ".pdf"
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