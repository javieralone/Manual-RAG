import argparse
import json
import os
import sys
import time
from pathlib import Path
from typing import Dict, List

SERVICE_DIR = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(SERVICE_DIR / "src"))

from manual_rag.application.evaluation_service import EvaluationService
from manual_rag.domain.evaluation import EvalCase, EvalThresholds
from manual_rag.domain.schemas import ChunkResult
from manual_rag.ports.embedding_port import EmbeddingPort
from manual_rag.ports.vector_store_port import VectorStorePort
from manual_rag.application.rag_service import RAGService

BASE_DIR = SERVICE_DIR.parent.parent
EVAL_DIR = SERVICE_DIR / "eval"
OUTPUT_DIR = BASE_DIR / "data" / "artifacts" / "evaluations"


def load_cases(path: Path):
    with open(path, "r", encoding="utf-8") as handle:
        raw_cases = json.load(handle)
    return [EvalCase(**item) for item in raw_cases]


def load_thresholds(path: Path) -> EvalThresholds:
    if not path.exists():
        return EvalThresholds()
    with open(path, "r", encoding="utf-8") as handle:
        return EvalThresholds(**json.load(handle))


class FixtureEmbeddingAdapter(EmbeddingPort):
    """Maps each golden query to a deterministic vector for hermetic CI runs."""

    def __init__(self, chunks_by_query: Dict[str, List[ChunkResult]]):
        self._vectors = {
            query: [float(index)]
            for index, query in enumerate(chunks_by_query, start=1)
        }

    def generate_embedding(self, text: str) -> List[float]:
        return self._vectors.get(text, [0.0])


class FixtureVectorStore(VectorStorePort):
    def __init__(self, chunks_by_query: Dict[str, List[ChunkResult]]):
        self._chunks_by_vector = {
            (float(index),): chunks
            for index, chunks in enumerate(chunks_by_query.values(), start=1)
        }

    def search_similar(
        self,
        query_vector: List[float],
        top_k: int,
        filters=None,
        collection_name=None,
    ) -> List[ChunkResult]:
        return self._chunks_by_vector.get(tuple(query_vector), [])[:top_k]


def load_fixture(path: Path) -> Dict[str, List[ChunkResult]]:
    with open(path, "r", encoding="utf-8") as handle:
        raw_fixture = json.load(handle)
    return {
        query: [ChunkResult(**chunk) for chunk in chunks]
        for query, chunks in raw_fixture.items()
    }


def main():
    parser = argparse.ArgumentParser(description="Ejecuta el runner de evaluación de calidad del RAG.")
    parser.add_argument("--dataset", default=str(EVAL_DIR / "golden_dataset.json"))
    parser.add_argument("--thresholds", default=str(EVAL_DIR / "thresholds.json"))
    parser.add_argument("--qdrant-host", default=os.getenv("QDRANT_HOST", "localhost"))
    parser.add_argument("--qdrant-port", type=int, default=int(os.getenv("QDRANT_PORT", 6333)))
    parser.add_argument("--collection", default="manuales_tecnicos")
    parser.add_argument("--ollama-model", default=os.getenv("OLLAMA_MODEL", "qwen2.5:1.5b"))
    parser.add_argument("--ollama-host", default=os.getenv("OLLAMA_HOST", ""))
    parser.add_argument(
        "--skip-generation",
        action="store_true",
        help="Evalúa solo recuperación, sin generar respuestas con Ollama.",
    )
    parser.add_argument("--output", default="")
    parser.add_argument(
        "--fixture",
        default="",
        help="Usa un corpus determinista versionado, pensado para CI sin Qdrant ni modelos.",
    )
    args = parser.parse_args()

    cases = load_cases(Path(args.dataset))
    thresholds = load_thresholds(Path(args.thresholds))

    if args.fixture:
        chunks_by_query = load_fixture(Path(args.fixture))
        embedding_adapter = FixtureEmbeddingAdapter(chunks_by_query)
        vector_store = FixtureVectorStore(chunks_by_query)
    else:
        from manual_rag.adapters.bge_embedding_adapter import BGEEmbeddingAdapter
        from manual_rag.adapters.ollama_answer_adapter import OllamaAnswerAdapter
        from manual_rag.adapters.qdrant_adapter import QdrantAdapter

        embedding_adapter = BGEEmbeddingAdapter(model_name="BAAI/bge-m3")
        vector_store = QdrantAdapter(host=args.qdrant_host, port=args.qdrant_port, collection_name=args.collection)

    rag_service = RAGService(embedding_provider=embedding_adapter, vector_store=vector_store)

    # Descarta el costo de carga en frío del modelo para que no infle la latencia del primer caso.
    if not args.fixture:
        embedding_adapter.generate_embedding("warm up")

    answer_generator = None
    if not args.skip_generation:
        if args.fixture:
            raise ValueError("--fixture requiere --skip-generation")
        answer_generator = OllamaAnswerAdapter(model=args.ollama_model, host=args.ollama_host)

    evaluation_service = EvaluationService(
        rag_service=rag_service,
        thresholds=thresholds,
        answer_generator=answer_generator,
    )

    report = evaluation_service.evaluate(cases)

    output_path = Path(args.output) if args.output else OUTPUT_DIR / f"eval_report_{int(time.time())}.json"
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(report.model_dump_json(indent=2), encoding="utf-8")

    print(f"Casos evaluados: {report.total_cases}")
    print(f"Casos aprobados: {report.passed_cases} ({report.pass_rate:.0%})")
    print(f"Precisión de contexto promedio: {report.avg_context_precision:.2f}")
    print(f"Cobertura de contexto promedio: {report.avg_context_recall:.2f}")
    print(f"Groundedness promedio: {report.avg_groundedness:.2f}")
    print(f"Latencia promedio (s): {report.avg_latency_seconds:.2f}")
    print(f"Reporte guardado en: {output_path}")

    if report.pass_rate < 1.0:
        failing = [result.case_id for result in report.results if not result.passed]
        print(f"Casos que no superaron los thresholds: {', '.join(failing)}")
        sys.exit(1)


if __name__ == "__main__":
    main()
