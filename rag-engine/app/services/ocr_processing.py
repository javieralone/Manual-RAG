import re


def normalize_ocr_text(text: str) -> str:
    normalized = text.replace("\r\n", "\n").replace("\r", "\n")
    normalized = re.sub(r"(?<=\w)-\n(?=\w)", "", normalized)
    normalized = re.sub(r"[ \t]+", " ", normalized)
    normalized = re.sub(r"\n{3,}", "\n\n", normalized)
    return "\n".join(line.strip() for line in normalized.splitlines()).strip()


def ocr_quality_score(text: str) -> float:
    normalized = normalize_ocr_text(text)
    if not normalized:
        return 0.0

    alphanumeric = sum(character.isalnum() for character in normalized)
    readable_ratio = alphanumeric / len(normalized)
    length_score = min(len(normalized) / 200.0, 1.0)
    return round((readable_ratio * 0.6) + (length_score * 0.4), 4)


def choose_ocr_text(primary: str, fallback: str) -> tuple[str, str, float]:
    primary_normalized = normalize_ocr_text(primary)
    fallback_normalized = normalize_ocr_text(fallback)
    primary_score = ocr_quality_score(primary_normalized)
    fallback_score = ocr_quality_score(fallback_normalized)

    if fallback_score > primary_score:
        return fallback_normalized, "raw", fallback_score
    return primary_normalized, "preprocessed", primary_score