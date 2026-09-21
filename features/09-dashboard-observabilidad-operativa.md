# Dashboard de observabilidad operativa

## Objetivo

Reorganizar el dashboard `Manual-RAG Overview` para que el estado, rendimiento y diagnóstico del sistema se entiendan de un vistazo. La implementación inicial solo modifica el dashboard provisionado y reutiliza métricas existentes; no cambia contratos HTTP ni la instrumentación de los servicios.

El resultado esperado está en `observability/grafana/dashboards/manual-rag.json` y se carga mediante el aprovisionamiento actual de Grafana.

## Decisión de alcance

La métrica **Preguntas Totales Hoy** se implementará inicialmente como el número de solicitudes `POST` exitosas (`2xx`) a los endpoints de consulta:

- `/api/v1/query`
- `/query/stream`

Es una aproximación operativa suficiente con `api_go_http_requests_total{method,route,status}`. Incluye cada intento exitoso, no solo preguntas semánticamente válidas o respuestas completadas. No se debe modificar Go para esta primera entrega.

Como mejora posterior, si se necesita la semántica estricta de pregunta validada, crear `api_go_questions_total{route,result}` en el handler después de validar el campo `question`. Nunca usar la pregunta, el prompt ni la respuesta como labels de Prometheus.

## Métricas confirmadas

El dashboard debe usar solamente estas métricas ya expuestas:

| Métrica | Uso |
| --- | --- |
| `api_go_readiness` | Disponibilidad del Gateway Go. |
| `rag_engine_readiness` | Disponibilidad del Motor RAG Python. |
| `api_go_http_requests_total{method,route,status}` | Tráfico y preguntas procesadas. |
| `api_go_http_request_duration_seconds` | Latencia HTTP del Gateway. |
| `rag_engine_qdrant_duration_seconds` | Tiempo de búsqueda/operación de Qdrant. |
| `api_go_generation_duration_seconds` | Tiempo total de generación en Ollama. |
| `api_go_dependency_requests_total{dependency,result}` | Errores de dependencias. |

No usar `api_go_time_to_first_token_seconds` como TTFT global: en la ruta no streaming representa el tiempo de finalización y no es un TTFT real.

## Diseño final del dashboard

Conservar `uid: "manual-rag-overview"`, la fuente de datos por defecto y el refresco de 10 segundos. Añadir paneles de tipo `row` y ubicar los paneles por debajo de cada fila para que Grafana los agrupe visualmente.

### Fila 1: Salud del Sistema

Usar cuatro paneles `stat`, de ancho 6 y altura 5.

| Título | Consulta | Notas |
| --- | --- | --- |
| Estado API Gateway (Go) | `api_go_readiness` | Mostrar 1 como OK y 0 como caído. |
| Estado Motor RAG (Python) | `rag_engine_readiness` | Mostrar 1 como OK y 0 como caído. |
| Preguntas Totales Hoy | `sum(increase(api_go_http_requests_total{method="POST",route=~"/api/v1/query|/query/stream",status=~"2.."}[$__range]))` | Stat grande, sin unidades. El selector temporal del dashboard debe fijarse en "Today so far" para representar el día; con otro rango el título deja de ser literal. |
| Tasa de Errores de Consulta (%) | `100 * sum(rate(api_go_http_requests_total{method="POST",route=~"/api/v1/query|/query/stream",status=~"5.."}[$__rate_interval])) / clamp_min(sum(rate(api_go_http_requests_total{method="POST",route=~"/api/v1/query|/query/stream"}[$__rate_interval])), 1e-9)` | Stat con unidad `percent (0-100)`. Mide errores 5xx sobre solicitudes de consulta. |

Usar umbrales coherentes en los readiness: rojo para 0 y verde para 1. Para tasa de errores, verde en 0, amarillo desde 1 y rojo desde 5. No usar `No data` como OK.

### Fila 2: Tiempos y Rendimiento

Usar tres paneles `timeseries`, ancho 8 y altura 8. Formatear el eje y y tooltip en segundos o milisegundos de manera consistente; preferir segundos si se mantienen las métricas base.

| Título | Consulta PromQL | Leyenda |
| --- | --- | --- |
| Tiempo de Respuesta API (p95) | `histogram_quantile(0.95, sum by (le) (rate(api_go_http_request_duration_seconds_bucket[$__rate_interval])))` | `p95` |
| Tiempo Búsqueda Vectorial (Qdrant) | `histogram_quantile(0.95, sum by (le) (rate(rag_engine_qdrant_duration_seconds_bucket[$__rate_interval])))` | `p95` |
| Tiempo de Generación del LLM (Ollama p95) | `histogram_quantile(0.95, sum by (le) (rate(api_go_generation_duration_seconds_bucket[$__rate_interval])))` | `p95` |

El panel de Qdrant reemplaza el actual **RAG Retrieval Latency**, ya que este último mide retrieval de forma más amplia y no expresa específicamente la búsqueda vectorial solicitada.

### Fila 3: Diagnóstico y Auditoría

Usar dos paneles de ancho 12 y altura 10.

| Tipo | Título | Fuente y consulta |
| --- | --- | --- |
| `logs` | Logs de Consultas y Errores | Loki (`uid: "loki"`): `{service=~"api-go|rag-engine"} !~ "(GET|POST) /(health|ready|metrics)(\\?| |$)"` |
| `table` | Trazas de Peticiones Lentas | Tempo (`uid: "tempo"`), TraceQL: `{ resource.service.name = "api-go" }` |

Mantener el límite de trazas en 20. No afirmar que son lentas hasta que exista un filtro TraceQL compatible con la versión de Tempo desplegada. Si esa versión permite comparar duración, añadir el filtro y verificarlo contra trazas reales; si no, titular inicialmente el panel **Trazas de Peticiones API** para no inducir a error.

### Paneles adicionales conservados

Mantener un panel `timeseries` de dependencias con el título:

- **Errores con Ollama / Qdrant**

Consulta:

```promql
sum by (dependency) (
  rate(api_go_dependency_requests_total{dependency=~"ollama|qdrant",result="error"}[$__rate_interval])
)
```

Ubicarlo en la fila 1 o en la fila 2 según la legibilidad final. Si no existe la etiqueta `qdrant` porque el Gateway solo llama a RAG y Ollama, conservar solamente las dependencias presentes (`ollama` y/o `rag-engine`) y ajustar el título a los nombres reales. Verificar los valores de `dependency` con `api_go_dependency_requests_total` antes de finalizar.

El panel **Peticiones por Segundo (RPS)** es útil para contextualizar las latencias. Mantenerlo como `timeseries`, con consulta:

```promql
sum(rate(api_go_http_requests_total{route=~"/api/v1/query|/query/stream"}[$__rate_interval]))
```

Ubicarlo al final de la fila 2 o crear una fila compacta adicional solo si no cabe sin degradar la lectura. Excluir rutas de salud, readiness y métricas.

## Pasos de implementación

1. Abrir `observability/grafana/dashboards/manual-rag.json` y conservar su formato JSON compacto si no hay necesidad de reformatearlo en todo el archivo.
2. Incrementar `version` del dashboard en uno.
3. Añadir tres paneles `row` titulados exactamente `Salud del Sistema`, `Tiempos y Rendimiento` y `Diagnóstico y Auditoría`.
4. Renombrar los paneles actuales de readiness, RPS, latencia API, latencia RAG, errores de dependencias, logs y Tempo con los títulos indicados arriba.
5. Reemplazar la consulta de latencia de retrieval por la métrica específica de Qdrant.
6. Añadir los paneles de preguntas totales, tasa de errores y latencia de generación de Ollama.
7. Cambiar los intervalos fijos `[5m]` por `[$__rate_interval]` en series basadas en `rate`, para que el dashboard responda al rango temporal y al scrape interval.
8. Cambiar el selector de logs por el filtro Loki definido en esta feature. No introducir `route` como label de Loki: actualmente no existe y extraerlo podría elevar la cardinalidad.
9. Reordenar `gridPos` sin superposiciones. Confirmar en Grafana que las filas se pliegan y despliegan correctamente.

## Validación obligatoria

1. Validar el JSON localmente:

```powershell
Get-Content observability/grafana/dashboards/manual-rag.json -Raw | ConvertFrom-Json | Out-Null
```

2. Reiniciar o reprovisionar Grafana según `docker-compose.yml`, y comprobar que el dashboard se carga sin errores de JSON ni de datasource.
3. En Prometheus, ejecutar cada consulta PromQL. Para los histogramas, generar al menos una consulta real antes de concluir que un panel vacío es un fallo del dashboard.
4. Verificar que **Preguntas Totales Hoy** no aumenta al consultar `/health`, `/ready` o `/metrics`, y sí aumenta tras una consulta exitosa a uno de los endpoints de query.
5. Verificar que los logs de health, ready y metrics no aparecen en el panel de Loki y que sí aparecen logs de una consulta real o error.
6. Comprobar que no hay solapamientos visuales en Grafana, con ancho de escritorio y vista reducida.

## Criterios de aceptación

- Los paneles tienen nombres de negocio claros en español, salvo nombres de productos como Go, Python, Qdrant y Ollama.
- La primera fila permite saber si el Gateway y RAG están disponibles, cuántas consultas se procesaron en el rango del día y si hay errores de consulta.
- La segunda fila muestra p95 de API, Qdrant y Ollama con datos cuando existen llamadas reales.
- Las solicitudes de health, ready y metrics no contaminan RPS, preguntas ni logs de diagnóstico.
- El dashboard no depende de métricas ni labels inexistentes.
- La versión del dashboard aumenta y el archivo sigue siendo JSON válido.

## Fuera de alcance de esta entrega

- Crear un contador semántico de preguntas válidas en Go.
- Implementar TTFT real para solicitudes no streaming.
- Extraer rutas como labels de Loki.
- Cambiar instrumentación, Docker Compose, Prometheus, Tempo o Promtail.