import argparse
from pathlib import Path


def resolve_feature(roadmap_dir: Path, requested: str) -> Path:
    requested = requested.strip().lower()
    candidates = sorted(
        path
        for path in roadmap_dir.glob("*.md")
        if path.name.lower() != "readme.md"
        and (path.stem.lower() == requested or path.stem.lower().startswith(f"{requested}-"))
    )
    if not candidates:
        raise SystemExit(f"No se encontró la feature '{requested}' en {roadmap_dir}.")
    if len(candidates) > 1:
        names = ", ".join(path.stem for path in candidates)
        raise SystemExit(f"La feature '{requested}' es ambigua: {names}")
    return candidates[0]


def build_plan(feature_path: Path) -> str:
    lines = feature_path.read_text(encoding="utf-8").splitlines()
    title = next((line[2:].strip() for line in lines if line.startswith("# ")), feature_path.stem)
    sections = []
    current = None
    for line in lines:
        if line.startswith("## "):
            current = {"title": line[3:].strip(), "items": []}
            sections.append(current)
        elif current is not None and line.strip():
            current["items"].append(line.strip())

    output = [
        f"# Implementation Plan: {title}",
        "",
        f"Source: `{feature_path.as_posix()}`",
        "",
        "## Scope extracted from the feature",
    ]
    for section in sections:
        output.extend(["", f"### {section['title']}"])
        output.extend(f"- {item.lstrip('-* ').strip()}" for item in section["items"])
    output.extend(
        [
            "",
            "## Implementation checklist",
            "",
            "- [ ] Confirm affected service and architecture boundary.",
            "- [ ] Implement the smallest end-to-end slice.",
            "- [ ] Add or update focused tests.",
            "- [ ] Update documentation and configuration.",
            "- [ ] Run project validation and RAG evaluation when relevant.",
            "- [ ] Open a pull request against `dev` for review.",
        ]
    )
    return "\n".join(output) + "\n"


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("feature")
    parser.add_argument("--roadmap", default="docs/roadmap")
    parser.add_argument("--output", default="implementation-plan.md")
    args = parser.parse_args()

    feature_path = resolve_feature(Path(args.roadmap), args.feature)
    Path(args.output).write_text(build_plan(feature_path), encoding="utf-8")
    print(f"Feature: {feature_path}")
    print(f"Plan: {args.output}")


if __name__ == "__main__":
    main()