import time
from typing import List, Optional

from app.core.domain.evaluation import EvalCase, EvalReport, EvalResult, EvalThresholds
from app.core.ports.answer_generator_port import AnswerGeneratorPort
from app.services.evaluation_metrics import context_precision, context_recall, keyword_coverage
from app.services.rag_service import RAGService


class EvaluationService:
    """Ejecuta un conjunto de casos de referencia contra el RAGService y mide su calidad."""

    def __init__(
        self,
        rag_service: RAGService,
        thresholds: Optional[EvalThresholds] = None,
        answer_generator: Optional[AnswerGeneratorPort] = None,
    ):
        self._rag_service = rag_service
        self._thresholds = thresholds or EvalThresholds()
        self._answer_generator = answer_generator

    def evaluate_case(self, case: EvalCase) -> EvalResult:
        started = time.perf_counter()
        response = self._rag_service.execute_search(case.query, case.top_k)
        latency_seconds = time.perf_counter() - started

        retrieved_sources = [chunk.source for chunk in response.results]
        context_text = "\n\n".join(chunk.text for chunk in response.results)
        precision = context_precision(retrieved_sources, case.expected_sources)
        recall = context_recall(retrieved_sources, case.expected_sources)

        answer = None
        if self._answer_generator is not None and response.results:
            answer = self._answer_generator.generate_answer(case.query, context_text)
            groundedness = keyword_coverage(answer, case.expected_keywords)
        else:
            # Sin generación de respuesta, se mide cobertura del tema sobre el contexto recuperado.
            groundedness = keyword_coverage(context_text, case.expected_keywords)

        passed = (
            recall >= self._thresholds.min_context_recall
            and precision >= self._thresholds.min_context_precision
            and groundedness >= self._thresholds.min_groundedness
            and latency_seconds <= self._thresholds.max_latency_seconds
        )

        return EvalResult(
            case_id=case.id,
            query=case.query,
            context_precision=precision,
            context_recall=recall,
            groundedness=groundedness,
            latency_seconds=latency_seconds,
            retrieved_sources=retrieved_sources,
            answer=answer,
            passed=passed,
        )

    def evaluate(self, cases: List[EvalCase]) -> EvalReport:
        results = [self.evaluate_case(case) for case in cases]
        total = len(results)
        passed = sum(1 for result in results if result.passed)
        return EvalReport(
            results=results,
            total_cases=total,
            passed_cases=passed,
            pass_rate=(passed / total) if total else 0.0,
            avg_context_precision=_average(r.context_precision for r in results),
            avg_context_recall=_average(r.context_recall for r in results),
            avg_groundedness=_average(r.groundedness for r in results),
            avg_latency_seconds=_average(r.latency_seconds for r in results),
        )


def _average(values) -> float:
    values = list(values)
    return sum(values) / len(values) if values else 0.0
