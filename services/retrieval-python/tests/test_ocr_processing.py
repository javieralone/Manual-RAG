import unittest

from manual_rag.application.ocr_processing import choose_ocr_text, normalize_ocr_text, ocr_quality_score


class OCRProcessingTests(unittest.TestCase):
    def test_normalize_ocr_text_joins_hyphenated_lines(self):
        text = "mantenimien-\nto del motor\n\n\n"

        self.assertEqual(normalize_ocr_text(text), "mantenimiento del motor")

    def test_quality_score_rejects_empty_text(self):
        self.assertEqual(ocr_quality_score(""), 0.0)

    def test_choose_ocr_text_uses_raw_fallback_when_better(self):
        text, mode, score = choose_ocr_text("###", "CHRYSLER motor")

        self.assertEqual(text, "CHRYSLER motor")
        self.assertEqual(mode, "raw")
        self.assertGreater(score, 0.0)


if __name__ == "__main__":
    unittest.main()