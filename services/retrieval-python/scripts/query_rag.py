import sys
from pathlib import Path
from sentence_transformers import SentenceTransformer
from qdrant_client import QdrantClient
import ollama

# 1. Definir la raíz del proyecto dinámicamente
BASE_DIR = Path(__file__).resolve().parent.parent.parent

COLLECTION_NAME = "generic_manuals"
OLLAMA_MODEL = "qwen2.5:1.5b"

# 2. Conectar a Qdrant y cargar el modelo de embeddings
print("Conectando a Qdrant y cargando BGE-M3...")
try:
    client = QdrantClient(host="localhost", port=6333)
    client.get_collections()
except Exception as e:
    print(f"[ERROR] No se pudo conectar a Qdrant en localhost:6333: {e}")
    sys.exit(1)

embedding_model = SentenceTransformer("BAAI/bge-m3")


def ask_rag(question: str, top_k: int = 5, show_audit: bool = False):
    print(f"\n Buscando información relevante para: '{question}'...")

    # Generar vector para la pregunta del usuario
    query_vector = embedding_model.encode(question).tolist()

    # Buscar en Qdrant usando el método .query_points()
    response = client.query_points(
        collection_name=COLLECTION_NAME,
        query=query_vector,
        limit=top_k
    )

    search_result = response.points

    if not search_result:
        print("No se encontró información relacionada en la base de datos.")
        return

    # Construir el contexto uniendo los resultados
    context_blocks = []
    sources = []

    if show_audit:
        print("\n--- [AUDITORÍA DE CONTEXTO] FRAGMENTOS RECUPERADOS ---")

    for idx, res in enumerate(search_result, start=1):
        text = res.payload.get("text", "")
        page = res.payload.get("page", "?")
        source = res.payload.get("source", "Manual")

        context_blocks.append(f"[Página {page}]: {text}")
        sources.append(f"Página {page}")

        if show_audit:
            print(f"> Chunk #{idx} (Página {page}):\n{text}\n" + "-" * 40)

    context_str = "\n\n".join(context_blocks)

    # Prompt estructurado para el LLM
    prompt = f"""Eres un asistente técnico especializado en manuales de mantenimiento y lubricación.
Responde a la pregunta del usuario utilizando ÚNICAMENTE la información provista en los fragmentos del manual a continuación.
Si la respuesta no se encuentra en el texto provisto, indica claramente que la información no está disponible en el manual.

FRAGMENTOS DEL MANUAL:
{context_str}

PREGUNTA DEL USUARIO:
{question}

RESPUESTA DETALLADA (incluye referencias a los números de página si aplica):"""

    print(" Generando respuesta con Ollama...\n")

    try:
        res = ollama.generate(
            model=OLLAMA_MODEL,
            prompt=prompt
        )

        print("--- RESPUESTA ---")
        print(res["response"])
        print("\n--- FUENTES CONSULTADAS ---")
        print(", ".join(set(sources)))
    except Exception as e:
        print(f"[ERROR] Falló la comunicación con Ollama: {e}")


if __name__ == "__main__":
    print("=== ASISTENTE VIRTUAL DE MANUALES TÉCNICOS ===")
    while True:
        user_query = input("\n Escribe tu pregunta (o 'salir' para terminar): ")
        if user_query.strip().lower() in ["salir", "exit", "quit"]:
            break
        if user_query.strip():
            ask_rag(user_query, top_k=5, show_audit=False)