import unittest

from manual_rag.application.evaluation_metrics import context_precision, context_recall, keyword_coverage


class EvaluationMetricsTests(unittest.TestCase):
    def test_context_precision_counts_relevant_sources(self):
        retrieved = ["manual_a.pdf", "manual_b.pdf", "manual_a.pdf"]
        self.assertAlmostEqual(context_precision(retrieved, ["manual_a.pdf"]), 2 / 3)

    def test_context_precision_without_expected_sources_is_perfect(self):
        self.assertEqual(context_precision(["manual_a.pdf"], []), 1.0)

    def test_context_precision_without_retrieved_sources_is_zero(self):
        self.assertEqual(context_precision([], ["manual_a.pdf"]), 0.0)

    def test_context_recall_finds_all_expected_sources(self):
        retrieved = ["manual_a.pdf", "manual_c.pdf"]
        expected = ["manual_a.pdf", "manual_b.pdf"]
        self.assertEqual(context_recall(retrieved, expected), 0.5)

    def test_context_recall_without_expected_sources_is_perfect(self):
        self.assertEqual(context_recall([], []), 1.0)

    def test_keyword_coverage_matches_case_insensitively(self):
        text = "El número de motor está en el bloque de cilindros."
        self.assertEqual(keyword_coverage(text, ["NÚMERO DE MOTOR", "bloque"]), 1.0)

    def test_keyword_coverage_without_keywords_is_perfect(self):
        self.assertEqual(keyword_coverage("cualquier texto", []), 1.0)

    def test_keyword_coverage_empty_text_is_zero(self):
        self.assertEqual(keyword_coverage("", ["motor"]), 0.0)


if __name__ == "__main__":
    unittest.main()
