# Roadmap de features pendientes

Este directorio organiza las mejoras pendientes del proyecto por prioridad de implementación y por dependencias técnicas.

## Orden recomendado

1. [01 - Base de observabilidad y métricas de consulta](01-base-observabilidad.md)
2. [02 - Evaluación automática de calidad del RAG](02-evaluacion-calidad-rag.md)
3. [03 - Filtros por documento, capítulo y sección](03-filtros-documentos.md)
4. [04 - Mejora de OCR para manuales escaneados](04-mejora-ocr.md)
5. [05 - Soporte multi-colección y multi-manual](05-soporte-multi-coleccion.md)
6. [06 - Cola de indexación y procesamiento asíncrono](06-cola-indexacion.md)
7. [07 - UI de consulta y administración](07-ui-consulta-admin.md)
8. [08 - Tests de integración y regresión end-to-end](08-tests-integracion.md)

## Principio de priorización

Las tareas se han ordenado con dos criterios:

- prioridad funcional: qué aporta más valor operativo al producto,
- dependencia lógica: qué tareas necesitan que termine la anterior para poder evolucionar sin rework.

### Dependencias clave

- La feature de observabilidad es base para medir impacto, latencia y calidad del RAG.
- La evaluación de calidad es necesaria antes de mejorar filtros o reindexación avanzada.
- Los filtros y multi-colección dependen de un vocabulario estable de metadatos y de una estrategia de indexación consistente.
- La cola de procesamiento es más útil cuando ya hay una estructura clara de documentos y metadatos.
- La UI y los tests de integración suelen hacerse después de estabilizar la capa de negocio y la infraestructura.

## Estado sugerido por fase

### Fase 1: estabilización y medición
- 01 - Base de observabilidad y métricas de consulta
- 02 - Evaluación automática de calidad del RAG

### Fase 2: mejora del contenido y la búsqueda
- 03 - Filtros por documento, capítulo y sección
- 04 - Mejora de OCR para manuales escaneados
- 05 - Soporte multi-colección y multi-manual

### Fase 3: escalabilidad y producto
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
