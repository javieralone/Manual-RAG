import argparse
import json
import os
import shutil
import sys
from pathlib import Path

from PIL import Image, ImageFilter, ImageOps
from pdf2image import convert_from_path, pdfinfo_from_path
import pytesseract

Image.MAX_IMAGE_PIXELS = None

SERVICE_DIR = Path(__file__).resolve().parent.parent
BASE_DIR = SERVICE_DIR.parent.parent
RAG_ENGINE_DIR = SERVICE_DIR

sys.path.insert(0, str(RAG_ENGINE_DIR))

from app.services.ocr_processing import choose_ocr_text


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


def extract_page_text(page: Image.Image, language: str) -> tuple[str, str, float]:
    processed_text = pytesseract.image_to_string(preprocess_page(page), lang=language)
    raw_text = pytesseract.image_to_string(page, lang=language)
    return choose_ocr_text(processed_text, raw_text)


def show_progress(current: int, total: int, label: str) -> None:
    width = 30
    completed = int(width * current / total) if total else width
    bar = "=" * completed + ">" + " " * max(width - completed - 1, 0)
    print(f"\r{label} [{bar}] {current}/{total}", end="", flush=True)


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Extrae texto OCR de un manual PDF.")
    parser.add_argument("--pdf", type=Path, default=BASE_DIR / "data" / "documents" / "0-lubricacion-mantenimiento.pdf")
    parser.add_argument("--output", type=Path, default=BASE_DIR / "data" / "artifacts" / "manual_pages.json")
    parser.add_argument("--document-id", default=None)
    parser.add_argument("--part", type=int, default=1)
    parser.add_argument("--dpi", type=int, default=200)
    parser.add_argument("--language", default="spa")
    return parser.parse_args()

def main() -> None:
    arguments = parse_arguments()
    if not arguments.pdf.exists():
        raise FileNotFoundError(f"No existe el PDF: {arguments.pdf}")

    pytesseract.pytesseract.tesseract_cmd = resolve_tesseract_command()
    document_id = arguments.document_id or arguments.pdf.stem
    print(f"Converting PDF pages to images (DPI {arguments.dpi})...")
    total_pages = int(pdfinfo_from_path(str(arguments.pdf))["Pages"])
    results = []

    for index in range(1, total_pages + 1):
        show_progress(index - 1, total_pages, "OCR")
        pages = convert_from_path(
            arguments.pdf,
            dpi=arguments.dpi,
            first_page=index,
            last_page=index,
        )
        if not pages:
            raise RuntimeError(f"No se pudo convertir la pagina {index}")

        page = pages[0]
        try:
            text, mode, quality = extract_page_text(page, arguments.language)
        finally:
            page.close()
            del pages

        results.append({
            "page": index,
            "text": text,
            "source": arguments.pdf.name,
            "document_id": document_id,
            "part": arguments.part,
            "ocr_mode": mode,
            "ocr_quality": quality,
        })
        show_progress(index, total_pages, "OCR")

    arguments.output.parent.mkdir(parents=True, exist_ok=True)
    with arguments.output.open("w", encoding="utf-8") as file:
        json.dump(results, file, ensure_ascii=False, indent=2)

    print(f"\nDone. Output saved to:\n{arguments.output}")


if __name__ == "__main__":
    main()