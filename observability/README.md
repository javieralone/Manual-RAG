# Observability

The stack provides Prometheus metrics, Grafana dashboards, Loki logs, and Tempo traces.

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin by default; change `GRAFANA_ADMIN_PASSWORD`)
- Loki: http://localhost:3100
- Tempo: http://localhost:3200

Go and Python services expose `/metrics`, `/health`, and `/ready`. Prometheus scrapes the two application services. Alerts are defined in `prometheus/alerts.yml`.

Application logs are JSON on stdout. Docker logging can be shipped to Loki with a production log driver or an Alloy/Promtail deployment; this Compose stack keeps Loki available without changing the Docker daemon globally.
