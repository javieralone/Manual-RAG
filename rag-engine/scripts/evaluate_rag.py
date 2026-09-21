import argparse
import json
import os
import sys
import time
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from app.adapters.bge_embedding_adapter import BGEEmbeddingAdapter
from app.adapters.ollama_answer_adapter import OllamaAnswerAdapter
from app.adapters.qdrant_adapter import QdrantAdapter
from app.core.domain.evaluation import EvalCase, EvalThresholds
from app.services.evaluation_service import EvaluationService
from app.services.rag_service import RAGService

BASE_DIR = Path(__file__).resolve().parent.parent
EVAL_DIR = BASE_DIR / "eval"
OUTPUT_DIR = BASE_DIR / "output"


def load_cases(path: Path):
    with open(path, "r", encoding="utf-8") as handle:
        raw_cases = json.load(handle)
    return [EvalCase(**item) for item in raw_cases]


def load_thresholds(path: Path) -> EvalThresholds:
    if not path.exists():
        return EvalThresholds()
    with open(path, "r", encoding="utf-8") as handle:
        return EvalThresholds(**json.load(handle))


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
    args = parser.parse_args()

    cases = load_cases(Path(args.dataset))
    thresholds = load_thresholds(Path(args.thresholds))

    embedding_adapter = BGEEmbeddingAdapter(model_name="BAAI/bge-m3")
    vector_store = QdrantAdapter(host=args.qdrant_host, port=args.qdrant_port, collection_name=args.collection)
    rag_service = RAGService(embedding_provider=embedding_adapter, vector_store=vector_store)

    # Descarta el costo de carga en frío del modelo para que no infle la latencia del primer caso.
    embedding_adapter.generate_embedding("warm up")

    answer_generator = None
    if not args.skip_generation:
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
