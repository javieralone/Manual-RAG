import json
import sys
from pathlib import Path
from langchain_text_splitters import RecursiveCharacterTextSplitter

# 1. Calcular dinámicamente la raíz del proyecto (2 niveles arriba de 'rag-engine/scripts/')
# Si el script está en C:\Dev\AI\Manual-RAG\rag-engine\scripts\index_manual.py,
# BASE_DIR será C:\Dev\AI\Manual-RAG
BASE_DIR = Path(__file__).resolve().parent.parent.parent

# 2. Construir rutas relativas a la raíz
JSON_INPUT = BASE_DIR / "output" / "manual_pages.json"
CHUNKS_OUTPUT = BASE_DIR / "output" / "manual_chunks.json"

print(f"Buscando archivo de entrada en: {JSON_INPUT}")

if not JSON_INPUT.exists():
    print(f"\n[ERROR] No existe el archivo '{JSON_INPUT}'.")
    print("Asegúrate de ejecutar primero 'ocr_manual.py'.")
    sys.exit(1)

print("Cargando páginas procesadas...")
with open(JSON_INPUT, "r", encoding="utf-8") as f:
    pages_data = json.load(f)

# Configurar el separador recursivo
text_splitter = RecursiveCharacterTextSplitter(
    chunk_size=700,
    chunk_overlap=150,
    separators=["\n\n", "\n", ". ", " ", ""]
)

all_chunks = []

print("Creando chunks con metadatos...")
for item in pages_data:
    page_num = item.get("page", 0)
    text = item.get("text", "")
    source = item.get("source", "0-lubricacion-mantenimiento.pdf")
    document_id = item.get("document_id", Path(source).stem)
    chapter = item.get("chapter")
    section = item.get("section")
    
    if not text.strip():
        continue
        
    page_chunks = text_splitter.split_text(text)
    
    for idx, chunk_text in enumerate(page_chunks):
        metadata = {
            "page": page_num,
            "source": source,
            "document_id": document_id,
        }
        if chapter:
            metadata["chapter"] = chapter
        if section:
            metadata["section"] = section

        all_chunks.append({
            "id": f"page_{page_num}_chunk_{idx}",
            "text": chunk_text,
            "metadata": metadata
        })

# Asegurar que la carpeta output exista antes de guardar
CHUNKS_OUTPUT.parent.mkdir(parents=True, exist_ok=True)

with open(CHUNKS_OUTPUT, "w", encoding="utf-8") as f:
    json.dump(all_chunks, f, ensure_ascii=False, indent=2)

print(f"\n Proceso completado: {len(all_chunks)} chunks creados en:\n{CHUNKS_OUTPUT}")