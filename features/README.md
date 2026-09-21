# Roadmap de features pendientes

Este directorio organiza las mejoras pendientes del proyecto por prioridad de implementación y por dependencias técnicas. Las features 01, 02 y 03 ya están implementadas y no forman parte de este roadmap.

## Orden recomendado

1. [04 - Mejora de OCR para manuales escaneados](04-mejora-ocr.md)
2. [05 - Soporte multi-colección y multi-manual](05-soporte-multi-coleccion.md)
3. [09 - MCP multi-colección y herramientas por dominio](09-mcp-multi-coleccion.md)
4. [06 - Cola de indexación y procesamiento asíncrono](06-cola-indexacion.md)
5. [07 - UI de consulta y administración](07-ui-consulta-admin.md)
6. [08 - Tests de integración y regresión end-to-end](08-tests-integracion.md)

## Principio de priorización

Las tareas se han ordenado con dos criterios:

- prioridad funcional: qué aporta más valor operativo al producto,
- dependencia lógica: qué tareas necesitan que termine la anterior para poder evolucionar sin rework.

### Dependencias clave

- La multi-colección depende del vocabulario estable de metadatos y de la estrategia de indexación consistente incorporados en la feature 03.
- La cola de procesamiento es más útil cuando ya hay una estructura clara de documentos y metadatos.
- La UI y los tests de integración suelen hacerse después de estabilizar la capa de negocio y la infraestructura.

## Estado sugerido por fase

### Fase 1: estabilización y medición
- 01 - Base de observabilidad y métricas de consulta
- 02 - Evaluación automática de calidad del RAG

### Fase 1: mejora del contenido y la búsqueda
- 03 - Filtros por documento, capítulo y sección (implementada)
- 04 - Mejora de OCR para manuales escaneados
- 05 - Soporte multi-colección y multi-manual
- 09 - MCP multi-colección y herramientas por dominio

### Fase 2: escalabilidad y producto
- 06 - Cola de indexación y procesamiento asíncrono
- 07 - UI de consulta y administración
- 08 - Tests de integración y regresión end-to-end

---

## Criterio de cierre de una feature

Cada feature debe cerrar con:

- implementación funcional,
- pruebas específicas,
- documentación mínima del uso,
- medición del impacto sobre latencia, error o precisión,
- evidencia de compatibilidad con la arquitectura actual.
