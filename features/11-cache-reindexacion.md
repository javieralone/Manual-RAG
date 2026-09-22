# 11 - Caché y reindexación incremental

## Prioridad

P3 - media.

## Dependencias

- Feature 06: procesamiento de documentos desacoplado y trazable.
- Feature 08: pruebas de regresión para garantizar coherencia de resultados.
- Feature 10: métricas para medir hit rate, latencia y trabajo evitado.

## Objetivo

Reducir la latencia y el coste de consultas e ingestas evitando recalcular embeddings o reprocesar documentos que no han cambiado.

## Alcance

- Definir una clave de caché que incluya consulta normalizada, colección, filtros, `top_k` y versión del índice.
- Añadir TTL y límites de tamaño configurables.
- Invalidar resultados cuando cambien documentos, colecciones, modelo de embeddings o configuración relevante.
- Detectar cambios por hash o metadatos de archivo para reindexar solo documentos modificados.
- Mantener IDs deterministas y evitar duplicados durante reintentos.
- Exponer métricas de aciertos, fallos, invalidaciones y tiempo ahorrado.

## Restricciones

- No servir resultados de una colección o versión de índice distinta.
- No ocultar errores de Qdrant, embeddings u Ollama detrás de un valor cacheado.
- La caché debe poder desactivarse para diagnóstico y pruebas.

## Criterio de finalización

- Las consultas repetidas respetan la clave, el TTL y la invalidación configurados.
- Modificar un documento solo reprocesa los documentos afectados.
- Las pruebas verifican aislamiento por colección, filtros, versión y concurrencia.
- Las métricas permiten comparar latencia y carga antes y después de activar la caché.