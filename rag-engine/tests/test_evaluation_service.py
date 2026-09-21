import unittest
from typing import List

from app.core.domain.evaluation import EvalCase, EvalThresholds
from app.core.domain.schemas import ChunkResult
from app.core.ports.answer_generator_port import AnswerGeneratorPort
from app.core.ports.embedding_port import EmbeddingPort
from app.core.ports.vector_store_port import VectorStorePort
from app.services.evaluation_service import EvaluationService
from app.services.rag_service import RAGService


class FakeEmbeddingAdapter(EmbeddingPort):
    def generate_embedding(self, text: str) -> List[float]:
        return [0.0]


class FakeVectorStore(VectorStorePort):
    def __init__(self, chunks: List[ChunkResult]):
        self._chunks = chunks

    def search_similar(self, query_vector: List[float], top_k: int) -> List[ChunkResult]:
        return self._chunks[:top_k]


class FakeAnswerGenerator(AnswerGeneratorPort):
    def __init__(self, answer: str):
        self._answer = answer

    def generate_answer(self, query: str, context: str) -> str:
        return self._answer


class EvaluationServiceTests(unittest.TestCase):
    def _build_rag_service(self, chunks):
        return RAGService(embedding_provider=FakeEmbeddingAdapter(), vector_store=FakeVectorStore(chunks))

    def test_case_passes_when_thresholds_are_met(self):
        chunks = [ChunkResult(page=2, source="manual.pdf", text="el número de motor está en el bloque", score=0.9)]
        rag_service = self._build_rag_service(chunks)
        service = EvaluationService(
            rag_service=rag_service,
            thresholds=EvalThresholds(
                min_context_recall=1.0,
                min_context_precision=1.0,
                min_groundedness=1.0,
                max_latency_seconds=5.0,
            ),
            answer_generator=FakeAnswerGenerator("el número de motor está en el bloque de cilindros"),
        )
        case = EvalCase(
            id="case-1",
            query="¿Dónde está el número de motor?",
            expected_sources=["manual.pdf"],
            expected_keywords=["número de motor"],
        )

        result = service.evaluate_case(case)

        self.assertTrue(result.passed)
        self.assertEqual(result.context_recall, 1.0)
        self.assertEqual(result.groundedness, 1.0)

    def test_case_fails_when_expected_source_is_missing(self):
        chunks = [ChunkResult(page=5, source="otro.pdf", text="texto no relacionado", score=0.5)]
        rag_service = self._build_rag_service(chunks)
        service = EvaluationService(rag_service=rag_service, thresholds=EvalThresholds())
        case = EvalCase(
            id="case-2",
            query="¿Cuál es la cilindrada?",
            expected_sources=["manual.pdf"],
            expected_keywords=[],
        )

        result = service.evaluate_case(case)

        self.assertFalse(result.passed)
        self.assertEqual(result.context_recall, 0.0)

    def test_report_aggregates_pass_rate(self):
        chunks = [ChunkResult(page=1, source="manual.pdf", text="contenido", score=0.8)]
        rag_service = self._build_rag_service(chunks)
        service = EvaluationService(
            rag_service=rag_service,
            thresholds=EvalThresholds(min_context_recall=1.0, min_context_precision=1.0, min_groundedness=0.0),
        )
        cases = [
            EvalCase(id="case-1", query="q1", expected_sources=["manual.pdf"]),
            EvalCase(id="case-2", query="q2", expected_sources=["otro.pdf"]),
        ]

        report = service.evaluate(cases)

        self.assertEqual(report.total_cases, 2)
        self.assertEqual(report.passed_cases, 1)
        self.assertAlmostEqual(report.pass_rate, 0.5)


if __name__ == "__main__":
    unittest.main()
