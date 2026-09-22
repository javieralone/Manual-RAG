# Roadmap de features pendientes

Esta carpeta contiene únicamente trabajo pendiente. Las capacidades de observabilidad base, evaluación RAG, filtros, OCR optimizado, multi-colección, MCP multi-colección y el MVP de la UI ya están implementadas y se documentan en los README correspondientes.

## Bloqueadores P0/P1

Antes de las features funcionales, debe cerrarse [12 - Seguridad de producción y contratos](12-seguridad-produccion-contratos.md). El reverse proxy, TLS, redes internas, sesiones Redis y límite de body están implementados; quedan su validación operativa y los controles descritos en esa feature.

## Orden de implementación

El orden sigue las dependencias técnicas entre funcionalidades:

1. [12 - Seguridad de producción y contratos](12-seguridad-produccion-contratos.md)
2. [06 - Cola de indexación y procesamiento asíncrono](06-cola-indexacion.md)
3. [07 - Panel administrativo y gestión avanzada](07-ui-consulta-admin.md) — completa el MVP existente de frontend
4. [10 - Dashboards operativos de consulta](10-dashboards-operativos.md)
5. [08 - Tests de integración y regresión end-to-end](08-tests-integracion.md)
6. [11 - Caché y reindexación incremental](11-cache-reindexacion.md)

## Principio de priorización

Las tareas se han ordenado con dos criterios:

- prioridad funcional: qué aporta más valor operativo al producto,
- dependencia lógica: qué tareas necesitan que termine la anterior para poder evolucionar sin rework.

### Dependencias clave

- La cola debe estabilizar el procesamiento de documentos antes de exponer su estado en una UI administrativa.
- Los dashboards deben construirse sobre métricas y contratos estables del gateway, la ingesta y el retrieval.
- Los tests end-to-end deben cubrir primero la cola, la administración y los principales flujos observables.
- La caché y la reindexación incremental necesitan métricas y pruebas de regresión para medir mejoras sin ocultar errores ni servir resultados obsoletos.

## Estado sugerido por fase

### Fase 1: procesamiento y producto
- 12 - Seguridad de producción y contratos
- 06 - Cola de indexación y procesamiento asíncrono
- 07 - Panel administrativo y gestión avanzada

### Fase 2: operación y calidad
- 10 - Dashboards operativos de consulta
- 08 - Tests de integración y regresión end-to-end

### Fase 3: rendimiento
- 11 - Caché y reindexación incremental

---

## Criterio de cierre de una feature

Cada feature debe cerrar con:

- implementación funcional,
- pruebas específicas,
- documentación mínima del uso,
- medición del impacto sobre latencia, error o precisión,
- evidencia de compatibilidad con la arquitectura actual.
