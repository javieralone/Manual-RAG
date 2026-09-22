# 09 - MCP multi-colección y herramientas por dominio

## Prioridad

P2 - media-alta.

## Objetivo

Permitir que el servidor MCP consulte varias colecciones Qdrant de forma explícita, aislada y comprensible para el modelo de lenguaje, reutilizando el `RAGService` existente.

## Problema actual

El servidor MCP usa una única instancia de `QdrantAdapter` fijada a la colección por defecto `generic_manuals` y expone una sola herramienta genérica `search_manual`. No puede seleccionar colecciones ni expresar dominios de conocimiento distintos.

## Diseño recomendado

El MCP expondrá Tools especializadas por dominio o colección. Cada Tool tendrá una descripción clara para que el LLM pueda elegir el dominio correcto.

Ejemplos:

```text
search_technical_manuals
search_generic_manuals
```

Las Tools no implementarán embeddings ni acceso directo a Qdrant. Delegarán en el `RAGService`, que reutilizará el proveedor de embeddings y seleccionará el adapter Qdrant de la colección solicitada.

```text
MCP Tool
    -> RAGService
        -> EmbeddingPort (BAAI/bge-m3)
        -> VectorStorePort / QdrantAdapter
            -> Qdrant collection
```

## Compatibilidad de infraestructura

- Modelo de embeddings: `BAAI/bge-m3`.
- Dimensión vectorial: `1024`.
- Distancia: `COSINE`.
- Qdrant: host configurable mediante `QDRANT_HOST`.
- Qdrant: puerto configurable mediante `QDRANT_PORT`, por defecto `6333`.
- MCP: host `0.0.0.0` y puerto configurable mediante `MCP_PORT`, por defecto `8001`.
- Colección por defecto: `default_collection`, con valor `generic_manuals`.

Los vectores actuales fueron generados con `BAAI/bge-m3`, NO MODIFICAR.

## Selección de colección

El servidor debe mantener una caché de servicios o adapters por colección:

```text
collection name -> QdrantAdapter -> RAGService
```

La colección debe validarse antes de crear el adapter o ejecutar la consulta. No se permitirán nombres vacíos, separadores de ruta ni valores arbitrarios que permitan acceder a recursos no previstos.

Cuando no se indique una colección, se usará:

```text
default_collection = generic_manuals
```

Las colecciones se consultarán mediante `QdrantAdapter.search_similar`, que debe continuar usando la API actual de Qdrant:

```python
client.query_points(
    collection_name=collection_name,
    query=query_vector,
    limit=top_k,
    query_filter=query_filter,
)
```

## Herramientas MCP

La herramienta compatible con la colección por defecto debe conservarse para no romper clientes existentes:

```text
search_manual(query, top_k=3, document_id=None)
```

Las nuevas Tools específicas por dominio deben fijar explícitamente la colección que consultan:

```text
search_technical_manuals(query, top_k=3, document_id=None)
search_generic_manuals(query, top_k=3, document_id=None)
```

Cada Tool debe:

- Validar una consulta no vacía.
- Limitar `top_k` al rango soportado.
- Aplicar filtros `document_id`, `chapter` y `section` cuando corresponda.
- Delegar la búsqueda en `RAGService.execute_search`.
- Devolver una respuesta textual limpia y consistente.
- Incluir texto, score, página, fuente, `document_id`, colección y parte cuando estén disponibles.
- Indicar claramente cuando no haya resultados.

## Contrato conceptual

```text
search_technical_manuals(
    query="¿Cuál es la luz de válvulas del motor?",
    document_id="<nombre-manual>",
    top_k=3
)
```

La Tool resolverá internamente la colección configurada para el dominio y no expondrá al modelo una ruta de almacenamiento ni detalles internos de Qdrant.

## Reutilización del RAG

FastAPI y MCP deben reutilizar la misma abstracción `RAGService`. No se duplicará la generación de embeddings, la construcción de filtros ni el mapeo de resultados.

El proveedor `BGEEmbeddingAdapter` debe inicializarse una sola vez en el composition root y compartirse entre servicios.

## Filtros y metadatos

Cada resultado debe conservar los metadatos definidos por las features de indexación:

```json
{
  "collection": "manuales_tecnicos",
  "document_id": "<nombre-manual>",
  "source": "<nombre-manual>__parte-001.pdf",
  "part": 1,
  "page": 3
}
```

Los filtros de documento, capítulo y sección se aplican dentro de la colección elegida. La colección no debe mezclarse con un filtro de payload equivalente: es el límite de aislamiento principal.

## Tareas

- Añadir configuración `default_collection`.
- Crear validación de nombres de colección.
- Implementar resolución o caché de `RAGService` por colección.
- Adaptar el MCP actual para conservar `search_manual`.
- Añadir Tools especializadas por dominio.
- Propagar filtros existentes a cada Tool.
- Actualizar el formato de resultados con metadatos de colección y documento.
- Mantener una única instancia de `BAAI/bge-m3`.
- Usar `query_points` en todos los adapters Qdrant.
- Documentar puertos, variables y ejemplos de consumo.

## Pruebas requeridas

- `search_manual` usa `generic_manuals` por defecto.
- Cada Tool especializada consulta únicamente su colección configurada.
- Dos Tools con colecciones distintas no mezclan resultados.
- Una colección inválida produce un error controlado.
- Los filtros `document_id`, `chapter` y `section` llegan al adapter correcto.
- Se conserva la forma y el contenido de los metadatos de respuesta.
- El modelo de embeddings se inicializa una sola vez.
- MCP responde en el puerto `8001` y Qdrant en el puerto `6333`.
- Se prueba ausencia de resultados y consultas vacías.

## Criterio de finalización

- MCP puede consultar más de una colección sin mezclar resultados.
- El LLM dispone de Tools con descripciones específicas por dominio.
- `search_manual` sigue funcionando con `generic_manuals` como valor por defecto y permite consultar `manuales_tecnicos` mediante una Tool o selección explícita.
- La búsqueda reutiliza `RAGService`, `BAAI/bge-m3` y `query_points`.
- Los resultados incluyen metadatos suficientes para identificar colección, manual, parte y página.
- Existen pruebas de aislamiento, validación y compatibilidad.

## Dependencias

- Feature 05: soporte multi-colección y multi-manual.
- Feature 03: filtros y metadatos estables, ya implementada.

## Siguiente feature dependiente

[06 - Cola de indexación y procesamiento asíncrono](06-cola-indexacion.md)
