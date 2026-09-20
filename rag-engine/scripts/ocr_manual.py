import json
import os
from pathlib import Path
from PIL import Image
from pdf2image import convert_from_path
import pytesseract

# Desactivar límite de píxeles
Image.MAX_IMAGE_PIXELS = None

pytesseract.pytesseract.tesseract_cmd = (
    r"C:\Program Files\Tesseract-OCR\tesseract.exe"
)

# Obtener la raíz del proyecto (dos niveles arriba de scripts/)
BASE_DIR = Path(__file__).resolve().parent.parent.parent

PDF_FILE = os.path.join(BASE_DIR, "documents", "0-lubricacion-mantenimiento.pdf")
OUTPUT_FILE = os.path.join(BASE_DIR, "output", "manual_pages.json")

print("Converting PDF pages to images (DPI 200)...")

# Reducimos a DPI 200 para evitar que Tesseract se cuelgue
pages = convert_from_path(
    PDF_FILE,
    dpi=200
)

results = []
total_pages = len(pages)

for index, page in enumerate(pages, start=1):
    print(f"Processing page {index}/{total_pages}...", flush=True)

    text = pytesseract.image_to_string(
        page,
        lang="spa"
    )

    results.append(
        {
            "page": index,
            "text": text.strip()
        }
    )

with open(OUTPUT_FILE, "w", encoding="utf-8") as file:
    json.dump(results, file, ensure_ascii=False, indent=2)

print(f"\nDone. Output saved to:\n{OUTPUT_FILE}")