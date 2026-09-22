import argparse
import json
import os
import shutil
import sys
from concurrent.futures import ProcessPoolExecutor
from pathlib import Path
from typing import Dict, Any, List

from PIL import Image, ImageFilter, ImageOps
from pdf2image import convert_from_path, pdfinfo_from_path
import pytesseract

Image.MAX_IMAGE_PIXELS = None

SERVICE_DIR = Path(__file__).resolve().parent.parent
BASE_DIR = SERVICE_DIR.parent.parent
RAG_ENGINE_DIR = SERVICE_DIR

sys.path.insert(0, str(RAG_ENGINE_DIR / "src"))

try:
    from manual_rag.application.ocr_processing import choose_ocr_text
except ImportError:
    # Fallback por si se ejecuta de forma aislada
    def choose_ocr_text(proc: str, raw: str):
        return (proc, "processed", 1.0) if len(proc) >= len(raw) else (raw, "raw", 1.0)


def resolve_tesseract_command() -> str:
    configured = os.getenv("TESSERACT_CMD")
    if configured:
        return configured

    discovered = shutil.which("tesseract")
    if discovered:
        return discovered

    windows_path = Path(r"C:\Program Files\Tesseract-OCR\tesseract.exe")
    if windows_path.exists():
        return str(windows_path)

    raise RuntimeError("No se encontró Tesseract. Configura TESSERACT_CMD o instala tesseract-ocr.")


def preprocess_page(page: Image.Image) -> Image.Image:
    grayscale = ImageOps.grayscale(page)
    contrasted = ImageOps.autocontrast(grayscale)
    return contrasted.filter(ImageFilter.MedianFilter(size=3))


def process_single_page_ocr(args_tuple: tuple) -> Dict[str, Any]:
    """Función de trabajo para ejecutar en paralelo."""
    page, index, language, document_id, pdf_name, part, collection, fast_mode = args_tuple
    
    # 1. Asegurar que el proceso hijo encuentre Tesseract
    pytesseract.pytesseract.tesseract_cmd = resolve_tesseract_command()
    
    if fast_mode:
        # Una sola pasada de OCR
        text = pytesseract.image_to_string(page, lang=language)
        mode = "raw_fast"
        quality = 1.0
    else:
        # Doble pasada original
        processed_text = pytesseract.image_to_string(preprocess_page(page), lang=language)
        raw_text = pytesseract.image_to_string(page, lang=language)
        text, mode, quality = choose_ocr_text(processed_text, raw_text)

    page.close()

    return {
        "page": index,
        "text": text,
        "source": pdf_name,
        "document_id": document_id,
        "part": part,
        "collection": collection,
        "ocr_mode": mode,
        "ocr_quality": quality,
    }


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Extrae texto OCR de un manual PDF.")
    parser.add_argument("--pdf", type=Path, default=BASE_DIR / "data" / "documents" / "0-lubricacion-mantenimiento.pdf")
    parser.add_argument("--output", type=Path, default=BASE_DIR / "data" / "artifacts" / "manual_pages.json")
    parser.add_argument("--document-id", default=None)
    parser.add_argument("--part", type=int, default=1)
    parser.add_argument("--collection", default="generic_manuals")
    parser.add_argument("--pages", type=int, default=None, help="Número máximo de páginas a procesar. Si no se indica, se procesan todas.")
    parser.add_argument("--dpi", type=int, default=150)
    parser.add_argument("--language", default="spa")
    
    # Nuevas opciones de optimización
    parser.add_argument(
        "--memory-mode",
        choices=["disk", "memory"],
        default="disk",
        help="'disk' procesa página a página controlando la RAM. 'memory' carga todo el PDF en RAM de una vez."
    )
    parser.add_argument(
        "--workers",
        type=int,
        default=3,
        help="Número de procesos en paralelo (3 es ideal para 8GB de RAM en i7-2670QM)."
    )
    parser.add_argument(
        "--fast-ocr",
        action="store_true",
        help="Desactiva el preprocesamiento con filtro de mediana y hace una sola pasada de OCR (duplica velocidad)."
    )
    return parser.parse_args()


def main() -> None:
    arguments = parse_arguments()
    if not arguments.pdf.exists():
        raise FileNotFoundError(f"No existe el PDF: {arguments.pdf}")

    pytesseract.pytesseract.tesseract_cmd = resolve_tesseract_command()
    collection_name = arguments.collection or "generic_manuals"
    if collection_name == "default_collection":
        collection_name = "generic_manuals"
    document_id = arguments.document_id or arguments.pdf.stem
    
    total_pages = int(pdfinfo_from_path(str(arguments.pdf))["Pages"])
    max_pages = total_pages if arguments.pages is None else min(int(arguments.pages), total_pages)
    if arguments.pages is not None and arguments.pages <= 0:
        raise ValueError("El argumento --pages debe ser mayor que 0.")

    print(f"Procesando '{arguments.pdf.name}' ({max_pages}/{total_pages} páginas) en modo '{arguments.memory_mode}' con {arguments.workers} workers...")

    results: List[Dict[str, Any]] = []

    if arguments.memory_mode == "memory":
        all_pages = convert_from_path(arguments.pdf, dpi=arguments.dpi)
        tasks = [
            (page, idx + 1, arguments.language, document_id, arguments.pdf.name, arguments.part, collection_name, arguments.fast_ocr)
            for idx, page in enumerate(all_pages[:max_pages])
        ]
        with ProcessPoolExecutor(max_workers=arguments.workers) as executor:
            results = list(executor.map(process_single_page_ocr, tasks))
    else:
        with ProcessPoolExecutor(max_workers=arguments.workers) as executor:
            for index in range(1, max_pages + 1):
                pages = convert_from_path(
                    arguments.pdf,
                    dpi=arguments.dpi,
                    first_page=index,
                    last_page=index,
                )
                if not pages:
                    continue
                
                task = (pages[0], index, arguments.language, document_id, arguments.pdf.name, arguments.part, collection_name, arguments.fast_ocr)
                future = executor.submit(process_single_page_ocr, task)
                results.append(future.result())
                print(f"\rPágina {index}/{max_pages} completada", end="", flush=True)
            print()

    # Ordenar por número de página
    results.sort(key=lambda x: x["page"])

    arguments.output.parent.mkdir(parents=True, exist_ok=True)
    with arguments.output.open("w", encoding="utf-8") as file:
        json.dump(results, file, ensure_ascii=False, indent=2)

    print(f"\nProceso finalizado. Guardado en:\n{arguments.output}")


if __name__ == "__main__":
    main()