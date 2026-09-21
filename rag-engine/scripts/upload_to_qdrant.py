import json
import sys
import uuid
from pathlib import Path
from sentence_transformers import SentenceTransformer
from qdrant_client import QdrantClient
from qdrant_client.models import Distance, PayloadSchemaType, PointStruct, VectorParams

# 1. Calcular dinámicamente la raíz del proyecto (2 niveles arriba de 'rag-engine/scripts/')
BASE_DIR = Path(__file__).resolve().parent.parent.parent

CHUNKS_INPUT = BASE_DIR / "output" / "manual_chunks.json"
COLLECTION_NAME = "manuales_tecnicos"
POINT_NAMESPACE = uuid.UUID("b8d8c5bb-0af3-4e76-88f6-7fa0c6e5c2f4")

# 2. Verificar existencia del archivo de chunks
if not CHUNKS_INPUT.exists():
    print(f"\n[ERROR] No existe el archivo '{CHUNKS_INPUT}'.")
    print("Asegúrate de ejecutar primero 'index_manual.py'.")
    sys.exit(1)

# 3. Conectar a Qdrant local
print("Conectando a Qdrant en localhost:6333...")
try:
    client = QdrantClient(host="localhost", port=6333)
    client.get_collections()
except Exception as e:
    print(f"\n[ERROR] No se pudo conectar a Qdrant: {e}")
    print("Asegúrate de tener corriendo la instancia de Qdrant en Docker.")
    sys.exit(1)

# 4. Cargar modelo de embeddings (BGE-M3 genera vectores de 1024 dimensiones)
print("Cargando modelo BGE-M3...")
model = SentenceTransformer("BAAI/bge-m3")

# 5. Crear colección si no existe
if not client.collection_exists(COLLECTION_NAME):
    client.create_collection(
        collection_name=COLLECTION_NAME,
        vectors_config=VectorParams(
            size=1024,  # Tamaño del vector para BGE-M3
            distance=Distance.COSINE
        )
    )
    print(f"Colección '{COLLECTION_NAME}' creada.")

# 6. Cargar chunks
with open(CHUNKS_INPUT, "r", encoding="utf-8") as f:
    chunks = json.load(f)

print(f"Generando embeddings e insertando {len(chunks)} chunks en Qdrant...")

points = []
for idx, item in enumerate(chunks):
    # Generar vector denso
    vector = model.encode(item["text"]).tolist()
    
    # Preparar el punto para Qdrant
    metadata = item.get("metadata", {})
    document_id = metadata.get("document_id", "")
    part = metadata.get("part", 1)
    page = metadata.get("page", 0)
    point_key = f"{document_id}:{part}:{page}:{idx}"
    payload = {
        "text": item["text"],
        "page": page,
        "source": metadata.get("source", "Desconocido"),
        "document_id": document_id,
        "part": part,
    }
    for field in ("chapter", "section"):
        if metadata.get(field):
            payload[field] = metadata[field]

    points.append(
        PointStruct(
            id=str(uuid.uuid5(POINT_NAMESPACE, point_key)),
            vector=vector,
            payload=payload,
        )
    )
    completed = idx + 1
    width = 30
    filled = int(width * completed / len(chunks)) if chunks else width
    bar = "=" * filled + ">" + " " * max(width - filled - 1, 0)
    print(f"\rEmbeddings [{bar}] {completed}/{len(chunks)}", end="", flush=True)

print()

# Keyword indexes keep metadata filters efficient as the collection grows.
for field in ("document_id", "chapter", "section"):
    client.create_payload_index(
        collection_name=COLLECTION_NAME,
        field_name=field,
        field_schema=PayloadSchemaType.KEYWORD,
    )

# 7. Subir a Qdrant
client.upsert(
    collection_name=COLLECTION_NAME,
    points=points
)

print(f"\n ¡Listo! {len(points)} puntos subidos correctamente a Qdrant.")