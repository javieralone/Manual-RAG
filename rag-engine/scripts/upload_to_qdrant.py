import json
import sys
from pathlib import Path
from sentence_transformers import SentenceTransformer
from qdrant_client import QdrantClient
from qdrant_client.models import VectorParams, Distance, PointStruct

# 1. Calcular dinámicamente la raíz del proyecto (2 niveles arriba de 'rag-engine/scripts/')
BASE_DIR = Path(__file__).resolve().parent.parent.parent

CHUNKS_INPUT = BASE_DIR / "output" / "manual_chunks.json"
COLLECTION_NAME = "manuales_tecnicos"

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
    points.append(
        PointStruct(
            id=idx,
            vector=vector,
            payload={
                "text": item["text"],
                "page": item["metadata"]["page"],
                "source": item["metadata"]["source"]
            }
        )
    )

# 7. Subir a Qdrant
client.upsert(
    collection_name=COLLECTION_NAME,
    points=points
)

print(f"\n ¡Listo! {len(points)} puntos subidos correctamente a Qdrant.")