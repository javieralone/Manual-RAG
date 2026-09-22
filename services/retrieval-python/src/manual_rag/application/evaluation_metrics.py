from typing import List


def context_precision(retrieved_sources: List[str], expected_sources: List[str]) -> float:
    """Fracción de fragmentos recuperados que provienen de una fuente esperada."""
    if not retrieved_sources:
        return 0.0
    if not expected_sources:
        return 1.0
    expected = set(expected_sources)
    relevant = sum(1 for source in retrieved_sources if source in expected)
    return relevant / len(retrieved_sources)


def context_recall(retrieved_sources: List[str], expected_sources: List[str]) -> float:
    """Fracción de fuentes esperadas que aparecen entre los fragmentos recuperados."""
    if not expected_sources:
        return 1.0
    expected = set(expected_sources)
    found = expected & set(retrieved_sources)
    return len(found) / len(expected)


def keyword_coverage(text: str, expected_keywords: List[str]) -> float:
    """Fracción de palabras clave esperadas presentes en el texto (case-insensitive)."""
    if not expected_keywords:
        return 1.0
    if not text:
        return 0.0
    normalized = text.lower()
    found = sum(1 for keyword in expected_keywords if keyword.lower() in normalized)
    return found / len(expected_keywords)
