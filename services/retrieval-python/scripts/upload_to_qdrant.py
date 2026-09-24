import argparse
import json
import os
import re
import sys
import uuid
from datetime import datetime, timezone
from pathlib import Path
from sentence_transformers import SentenceTransformer
from qdrant_client import QdrantClient
from qdrant_client.models import Distance, PayloadSchemaType, PointStruct, VectorParams

SERVICE_DIR = Path(__file__).resolve().parent.parent
BASE_DIR = SERVICE_DIR.parent.parent
CHUNKS_INPUT = BASE_DIR / "data" / "artifacts" / "manual_chunks.json"
POINT_NAMESPACE = uuid.UUID("b8d8c5bb-0af3-4e76-88f6-7fa0c6e5c2f4")
VALID_COLLECTION_RE = re.compile(r"^[A-Za-z0-9_-]+$")


def validate_collection_name(value: str | None, default: str = "generic_manuals") -> str:
    candidate = (value or default).strip()
    if not candidate or candidate == "default_collection":
        return default
    if not VALID_COLLECTION_RE.fullmatch(candidate):
        raise ValueError(
            f"Colección inválida '{candidate}'. Usa solo letras, números, guion y guion bajo."
        )
    return candidate


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Carga chunks en una colección Qdrant específica.")
    parser.add_argument("--collection", default="generic_manuals", help="Nombre de la colección Qdrant a usar.")
    parser.add_argument("--chunks-input", type=Path, default=CHUNKS_INPUT)
    parser.add_argument("--job-id")
    parser.add_argument("--object-key", default="")
    parser.add_argument("--file-sha256", default="")
    parser.add_argument("--ingested-at", default=None)
    return parser.parse_args()


arguments = parse_arguments()
COLLECTION_NAME = validate_collection_name(arguments.collection)
CHUNKS_INPUT = arguments.chunks_input
INGESTED_AT = arguments.ingested_at or datetime.now(timezone.utc).isoformat()

if not CHUNKS_INPUT.exists():
    print(f"\n[ERROR] No existe el archivo '{CHUNKS_INPUT}'.")
    print("Asegúrate de ejecutar primero 'index_manual.py'.")
    sys.exit(1)

qdrant_host = os.getenv("QDRANT_HOST", "localhost")
qdrant_port = int(os.getenv("QDRANT_PORT", "6333"))
print(f"Conectando a Qdrant en {qdrant_host}:{qdrant_port}...")
try:
    client = QdrantClient(host=qdrant_host, port=qdrant_port)
    client.get_collections()
except Exception as e:
    print(f"\n[ERROR] No se pudo conectar a Qdrant: {e}")
    print("Asegúrate de tener corriendo la instancia de Qdrant en Docker.")
    sys.exit(1)

print("Cargando modelo BGE-M3...")
model = SentenceTransformer("BAAI/bge-m3")

if not client.collection_exists(COLLECTION_NAME):
    client.create_collection(
        collection_name=COLLECTION_NAME,
        vectors_config=VectorParams(
            size=1024,
            distance=Distance.COSINE,
        )
    )
    print(f"Colección '{COLLECTION_NAME}' creada.")

with open(CHUNKS_INPUT, "r", encoding="utf-8") as f:
    chunks = json.load(f)

print(f"Generando embeddings e insertando {len(chunks)} chunks en Qdrant...")

points = []
for idx, item in enumerate(chunks):
    vector = model.encode(item["text"]).tolist()
    metadata = item.get("metadata", {})
    collection_name = COLLECTION_NAME
    document_id = metadata.get("document_id", "")
    part = metadata.get("part", 1)
    page = metadata.get("page", 0)
    point_key = f"{collection_name}:{document_id}:{part}:{page}:{idx}"
    payload = {
        "text": item["text"],
        "page": page,
        "source": metadata.get("source", "Desconocido"),
        "document_id": document_id,
        "part": part,
        "collection": collection_name,
        "job_id": arguments.job_id or metadata.get("job_id", ""),
        "minio_object_key": arguments.object_key or metadata.get("minio_object_key", ""),
        "file_sha256": arguments.file_sha256 or metadata.get("file_sha256", ""),
        "ingested_at": metadata.get("ingested_at", INGESTED_AT),
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

for field in ("document_id", "chapter", "section"):
    client.create_payload_index(
        collection_name=COLLECTION_NAME,
        field_name=field,
        field_schema=PayloadSchemaType.KEYWORD,
    )

client.upsert(
    collection_name=COLLECTION_NAME,
    points=points,
)

print(f"\n ¡Listo! {len(points)} puntos subidos correctamente a Qdrant en '{COLLECTION_NAME}'.", flush=True)
os._exit(0)
