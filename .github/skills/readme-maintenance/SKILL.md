---
name: readme-maintenance
description: 'Maintain and review Manual-RAG README.md files. Use when documenting setup, Docker Compose, service ports, environment variables, API contracts, operational procedures, feature status, or repository structure; verify claims against implementation before editing.'
argument-hint: 'Describe the README or repository behavior that needs to be updated.'
user-invocable: true
---

# README Maintenance

Keep README.md files aligned with the behavior that is actually present in the repository.

## When to Use

- A service, script, deployment flow, or user workflow changed.
- A README contains stale setup, port, environment, API, or operational instructions.
- The user asks for a README review, documentation sync, or documentation gap analysis.
- A feature status or roadmap claim needs to be checked against the implementation.

## Procedure

1. **Locate the owning README**
   - Check the repository root, the nearest service or app directory, and any README linked by the affected component.
   - Update only the README files that own the behavior. Keep cross-links relative and valid.

2. **Inspect the source of truth**
   - For runtime behavior, inspect `docker-compose*.yml`, Dockerfiles, entrypoints, environment examples, and health checks.
   - For API behavior, inspect handlers, schemas, route registration, and relevant tests.
   - For scripts, inspect CLI help, argument parsing, defaults, and exit behavior.
   - For status claims, inspect the implementation and tests rather than relying on roadmap text alone.

3. **Classify findings**
   - Verified: directly supported by code or a successful local command.
   - Environment-dependent: requires Docker, models, credentials, or external services that are unavailable locally.
   - Stale or incorrect: contradicted by the current implementation.
   - Missing: useful information absent from the owning README.

4. **Edit narrowly**
   - Preserve the README's existing language and organization.
   - Prefer concrete commands and tables for setup, ports, variables, and service responsibilities.
   - Use placeholders for secrets and never copy values from local `.env` files.
   - Keep generated artifacts, caches, vector stores, and local runtime data out of source documentation unless explaining their role.
   - Do not claim a feature is complete when important paths or validation remain unverified.

5. **Validate**
   - Check referenced files and relative links exist when practical.
   - Run a focused command such as `docker compose --env-file .env.example config --quiet`, a script's `--help`, or the relevant build/test check when it is inexpensive.
   - Separate successful documentation validation from checks blocked by missing services, models, credentials, or Docker resources.

## Required Report

Return:

- README files changed and the behavior documented.
- Evidence or commands used to verify the updates.
- Environment-dependent checks that were not run or could not pass.
- Remaining documentation gaps, contradictions, or implementation issues.
