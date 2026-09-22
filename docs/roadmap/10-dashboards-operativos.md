# 10 - Dashboards operativos de consulta

## Prioridad

P2 - media-alta.

## Dependencias

- Feature 06: métricas y estados de la cola de indexación.
- Feature 07: métricas de uso del panel administrativo.
- Feature 08: contratos de integración y nombres de métricas estabilizados.

## Objetivo

Ampliar la observabilidad existente con dashboards orientados a operación diaria, diagnóstico de consultas y seguimiento de la latencia de generación.

## Alcance

- Panel de volumen, errores y latencia del gateway.
- Panel de retrieval con duración, resultados vacíos, colección y filtros usados sin datos de alta cardinalidad.
- Panel de tiempo hasta el primer token, duración total de generación y saturación del worker pool.
- Panel de cola de indexación con pendientes, fallos, reintentos y tiempo de permanencia.
- Enlaces de Grafana entre métricas, logs y trazas mediante `trace_id`.
- Alertas para degradación de retrieval, generación, cola y dependencias.

## Restricciones

- No incluir preguntas completas, prompts, documentos, JWTs, usuarios ni direcciones IP como labels.
- Reutilizar métricas existentes antes de crear nuevas.
- Mantener dashboards provisionados y reproducibles desde el repositorio.

## Criterio de finalización

- Los dashboards se cargan automáticamente con Docker Compose.
- Cada panel usa métricas existentes o métricas nuevas documentadas y testeadas.
- Las alertas tienen umbrales, severidad y procedimiento de diagnóstico.
- Se puede seguir una consulta desde gateway hasta RAG, Ollama, logs y trazas.