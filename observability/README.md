# Observability

The stack provides Prometheus metrics, Grafana dashboards, Loki logs, and Tempo traces.

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin by default; change `GRAFANA_ADMIN_PASSWORD`)
- Loki: http://localhost:3100
- Tempo: http://localhost:3200

Promtail reads Docker container logs and sends them to Loki. In Grafana Explore, select Loki and query `{service="api-go"}` or `{service="rag-engine"}`. JSON fields such as `level` and `trace_id` are extracted when present; secrets, authorization headers, prompts, and document text are not added as labels. The provisioned dashboard includes logs and traces panels.

Go and Python services expose `/metrics`, `/health`, and `/ready`. Prometheus scrapes the two application services. Alerts are defined in `prometheus/alerts.yml`.

Application logs are JSON on stdout. Promtail is included in Compose and discovers Docker containers through the Docker socket. On Docker Desktop, if the socket or container log directory is not exposed to the Linux VM, use `docker compose logs` or configure a host-level Alloy/Promtail collector instead.
