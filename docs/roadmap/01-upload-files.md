# FEATURE: Automatización de ingesta de manuales con cola de trabajos y worker aislado

## Resumen ejecutivo

La feature 01 tiene como objetivo reemplazar la ingesta manual de PDFs basada en carpetas locales por un flujo automático, trazable y seguro, gestionado por cola de trabajo, aislamiento por job y metadata obligatoria en Qdrant.

La intención no es reescribir el motor de OCR ni la lógica de embeddings, sino encapsular la ejecución actual del proyecto en un worker que procese cada documento en un entorno temporal aislado y deje un registro auditable de la operación.

Esta feature debe integrarse con la arquitectura existente del repositorio:
- Gateway Go en services/gateway-go
- Motor RAG en services/retrieval-python
- Qdrant como almacenamiento vectorial
- Redis como broker de tareas
- MinIO como almacenamiento de objetos, como evolución de la ingesta actual basada en directorios locales

La implementación debe ser compatible con el flujo actual y debe permitir una transición gradual sin romper la funcionalidad que ya existe.

---

## Estado actual del repositorio

El sistema actual ya tiene un flujo funcional de ingesta manual basado en carpetas:
- PDF cargado en data/documents/new
- proceso ejecutado por services/retrieval-python/scripts/process_manual_opt.py
- OCR y extracción en scripts del mismo servicio
- indexación y subida a Qdrant
- movimiento del archivo a reading y completed

Ese flujo se documenta en README.md y se apoya en scripts del servicio de retrieval. La limitación principal es que no soporta concurrencia real ni aislamiento por trabajo, ya que reutiliza archivos globales y usa un lock genérico sobre el proceso de ingesta.

Los problemas concretos que esta feature aborda son:
- concurrencia y escritura sobre archivos compartidos
- pérdida de trazabilidad por trabajo
- falta de reintentos y DLQ
- falta de deduplicación por documento
- falta de metadata minimalista y reutilizable en Qdrant
- ausencia de API o cola para encolar documentos

---

## Objetivo de negocio

Eliminar la ejecución manual de ingesta y convertir la inserción de documentos en un proceso automatizado, controlado y auditable, con estas propiedades:
- cada documento se procesa de forma aislada
- cada trabajo queda identificado por un ID único
- cada resultado en Qdrant conserva su trazabilidad
- el sistema es tolerante a fallos y puede reintentar trabajos fallidos
- los documentos duplicados no se vuelven a indexar
- la limpieza temporal de archivos y workspaces queda garantizada

---

## Objetivo técnico

Implementar una capa de ingestion basada en Redis Queue y workers de Python, que orqueste el pipeline actual del repositorio sin duplicar la lógica de negocio.

La primera versión debe:
1. aceptar documentos desde MinIO, o desde una entrada local compatible durante la transición
2. crear un trabajo en cola
3. ejecutar el pipeline real con archivos temporales aislados por job_id
4. subir vectores a Qdrant con metadata de trazabilidad
5. actualizar estado del job, reintentos y DLQ
6. limpiar recursos temporales aunque falle el acuerdo

---

## Alcance

### Incluido
- almacenamiento de documentos en MinIO
- cola de trabajos Redis RQ
- worker Python para procesar documentos
- directorio temporal aislado por job_id
- reutilización del pipeline existente con CLI parametrizada
- metadata obligatoria en cada chunk insertado en Qdrant
- manejo de estados del job
- reintentos y DLQ
- API REST para encolar y consultar trabajos
- logs estructurados y trazabilidad
- limpieza de archivos temporales

### No incluido en V1
- reindexación manual por job
- sincronización con Kafka, Airflow o Kubernetes
- separación del pipeline en microservicios por etapa
- integración compleja con un sistema de almacenamiento externo distinto a MinIO
- gestión de documentos eliminados en Qdrant como funcionalidad de negocio completa

---

## Arquitectura objetivo

MinIO / upload entrypoint
  ↓
RQ Queue (Redis)
  ↓
Ingestion Worker
  ↓
Pipeline parametrizado
  ├─ process_manual_opt.py
  ├─ index_manual.py
  └─ upload_to_qdrant.py
  ↓
Qdrant con metadata obligatoria por chunk
  ↓
Job state + logs + DLQ

La intención es mantener la lógica de OCR, chunking y embedding intacta, y mover solo la orquestación y el manejo de archivos temporales al worker.

---

## Requisitos funcionales

### RF-01: Entrada de documentos
La feature debe soportar al menos dos modos de entrada:
1. desde MinIO, con object_key y bucket
2. desde un path local, durante la fase de transición o pruebas

El sistema debe resolver la colección destino a partir de:
- el prefijo de la ruta en MinIO, o
- el nombre de colección explícito, o
- el valor por defecto generic_manuals

Ejemplo:
- manuals/manuales_tecnicos/manual_1.pdf => collection = manuales_tecnicos

### RF-02: Creación de jobs
Cuando se recibe un documento, el sistema debe generar un job con un UUID.

Payload mínimo del job:
{
  "job_id": "uuid-v4",
  "bucket": "manuals",
  "object_key": "manuales_tecnicos/manual_1.pdf",
  "collection": "manuales_tecnicos",
  "file_name": "manual_1.pdf",
  "file_sha256": "abc123...",
  "created_at": "2026-09-22T12:00:00Z",
  "status": "PENDING"
}

### RF-03: Aislamiento por job
Cada job debe crear un directorio temporal dedicado:
- /tmp/jobs/{job_id}/

Ese directorio debe contener:
- PDF original descargado
- JSON de páginas generado por OCR
- JSON de chunks generado por indexación
- logs temporales del trabajo

No deben compartirse rutas globales ni archivos comunes entre jobs.

### RF-04: Reutilización del pipeline existente
El worker debe reutilizar la lógica actual del repositorio, en concreto:
- services/retrieval-python/scripts/process_manual_opt.py
- services/retrieval-python/scripts/index_manual.py
- services/retrieval-python/scripts/upload_to_qdrant.py

La invocación debe hacerse con argumentos CLI dinámicos, no con rutas fijas.

Ejemplo:
- process_manual_opt.py --pdf /tmp/jobs/{job_id}/manual.pdf --collection manuales_tecnicos
- index_manual.py --pages-input /tmp/jobs/{job_id}/manual_pages.json --chunks-output /tmp/jobs/{job_id}/manual_chunks.json
- upload_to_qdrant.py --chunks-input /tmp/jobs/{job_id}/manual_chunks.json --collection manuales_tecnicos --job-id {job_id}

### RF-05: Estados del job
El sistema debe mantener un estado consistente por trabajo.

Estados permitidos:
- PENDING
- PROCESSING
- COMPLETED
- SKIPPED
- TIMEOUT
- FAILED
- FAILED_PERMANENTLY

Cada estado debe almacenar:
- created_at
- started_at
- finished_at
- error
- traceback
- retries
- worker_id

### RF-06: Timeouts y reintentos
Cada job debe tener timeout configurable, recomendado 900 segundos o 15 minutos.

Reglas:
- si el trabajo supera el timeout, se marca como TIMEOUT
- si falla por error o timeout, se reintenta automáticamente con backoff
- el valor inicial recomendado es max_retries = 3
- backoff sugerido: 30s, 60s, 120s
- si se supera el límite, el trabajo pasa a FAILED_PERMANENTLY y entra en DLQ

### RF-07: Idempotencia por documento
El sistema debe evitar reindexar el mismo PDF dos veces en la misma colección.

Regla principal:
- calcular SHA256 del archivo original
- comparar con registros ya procesados
- si existe un registro completo con el mismo file_sha256 y la misma collection, marcar el trabajo como SKIPPED o COMPLETED sin volver a insertar

### RF-08: Trazabilidad en Qdrant
Cada chunk insertado en Qdrant debe incluir obligatoriamente los siguientes campos:
- job_id
- minio_object_key
- file_sha256
- ingested_at
- collection
- document_id
- part
- page
- source

La metadata debe ser suficiente para:
- buscar todos los chunks de un documento
- borrar un documento por source o object_key
- reindexar un documento concreto
- hacer auditoría de contenidos insertados

### RF-09: Escalabilidad horizontal
Debe poder ejecutarse con varios workers al mismo tiempo.

Cada worker debe poder limitar su paralelismo interno con variables de entorno, por ejemplo:
- WORKERS_PER_PROCESS=2
- QUEUE_NAME=ingestion

Esto evita que un worker utilice demasiados hilos o memoria en entornos restringidos.

### RF-10: Dead Letter Queue
Cuando un job supera el número máximo de reintentos, debe pasar a DLQ.

Se debe conservar al menos:
- job_id
- error
- traceback
- timestamps
- collection
- object_key
- sha256

### RF-11: Limpieza garantizada
El workspace asociado a cada job debe eliminarse siempre con finally o cleanup, cualquiera que sea el resultado.

Estados que requieren limpieza:
- COMPLETED
- SKIPPED
- FAILED
- TIMEOUT
- FAILED_PERMANENTLY

### RF-12: API de ingestión
La feature debe exponer al menos estos endpoints:
- POST /ingestion/enqueue
- GET /ingestion/jobs
- GET /ingestion/jobs/{id}
- GET /ingestion/failed

La API debe devolver:
- job_id
- collection
- status
- created_at
- started_at
- finished_at
- error
- result summary

### RF-13: Observabilidad y logging
Los eventos deben registrarse en formato estructurado, por ejemplo:
{
  "job_id": "123e4567-e89b-12d3-a456-426614174000",
  "file": "manual_1.pdf",
  "collection": "manuales_tecnicos",
  "status": "PROCESSING",
  "worker_id": "worker-1",
  "duration_ms": 12345
}

Debe incluir:
- inicio del trabajo
- fin del trabajo
- errores con traceback
- timeout
- reintentos
- duración total

---

## Requisitos técnicos

### 1. Redis y RQ
Se debe usar Redis como broker principal y RQ como librería de encolado.

Soporte requerido:
- encolado de jobs
- job_timeout
- reintentos nativos
- almacenamiento de resultados de trabajo
- FailedJobRegistry para DLQ

### 2. MinIO
MinIO será el almacenamiento principal de objetos para la ingesta.

Bucket recomendado:
- manuals

Estructura sugerida:
- manuals/generic_manuals/file.pdf
- manuals/manuales_tecnicos/file.pdf

Se puede soportar también el modo local de transición con una carpeta para pruebas, pero el objetivo final es MinIO.

### 3. Docker Compose
Se deben añadir los servicios mínimos necesarios en docker-compose.yml:
- redis
- minio
- ingestion-worker
- ingestion-api

El worker debe poder escalarse con:
- docker compose up --scale ingestion-worker=2

### 4. Reutilización del código actual
No se debe reescribir la lógica de OCR o chunking. La integración debe respetar el patrón actual del repositorio, usando los scripts ya existentes como adaptadores del pipeline.

### 5. Compatibilidad con la arquitectura actual
La feature debe respetar la separación actual de servicios:
- API Gateway Go sigue siendo el punto público
- Engine RAG sigue siendo el servicio de procesamiento y consulta
- Qdrant sigue siendo la fuente de verdad de documentos indexados
- Redis ya usado por sesiones del gateway no debe mezclarse sin criterio con la cola de ingesta

Si conviene, se recomienda separar la cola de ingestión en una base distinta de Redis o en una keyspace distinta para evitar conflictos con sesiones del refresh token.

---

## Diseño de integración con este repositorio

### Servicio Python
El servicio real que debe ejecutar la ingesta es services/retrieval-python.

Archivos que deben verse afectados:
- services/retrieval-python/scripts/process_manual_opt.py
- services/retrieval-python/scripts/index_manual.py
- services/retrieval-python/scripts/upload_to_qdrant.py
- services/retrieval-python/app/main.py
- services/retrieval-python/app/mcp_server.py
- services/retrieval-python/requirements.txt

### Estructura de trabajo esperada
El worker no debe depender de rutas hardcodeadas de la máquina host. Debe crear todos los archivos dentro de un workspace temporal y usar ese path para toda la operación.

Ejemplo de diseño:
/tmp/jobs/{job_id}/
  ├─ input.pdf
  ├─ manual_pages.json
  ├─ manual_chunks.json
  └─ logs/

### Integración con Qdrant
Debe añadirse en la carga de chunks:
- job_id
- collection
- minio_object_key
- file_sha256
- ingested_at

Esto permite consultas futuras por documento y soporte para borrado o reindexación.

### Integración con API Gateway
La API Gateway no debe orquestar la lógica del documento, pero sí puede ofrecer una interfaz pública para:
- encolar un job
- consultar el estado del trabajo
- consultar los fallidos

La capa HTTP debe delegar la operación al worker o a la cola, manteniendo la separación de responsabilidades.

---

## Plan de implementación recomendado para Copilot

### Fase 1: Definir contratos del job y de la cola
Tareas:
- crear modelo de Job con campos mínimos
- definir enum de estados
- definir payload de entrada para enqueue
- definir resultados de la API y del worker

Archivos sugeridos:
- services/retrieval-python/app/core/models.py
- services/retrieval-python/app/core/job_service.py
- services/retrieval-python/app/core/job_states.py

### Fase 2: Crear worker de ingesta
Tareas:
- crear worker Python con Redis Queue
- definir una tarea enqueue_pdf
- crear trabajo temporal por job_id
- descargar PDF desde MinIO o input local
- llamar al pipeline con rutas temporales
- escribir JSONs temporales exclusivos de job
- actualizar estado del trabajo
- limpiar workspace con finally

Archivos sugeridos:
- services/retrieval-python/app/services/ingestion_worker.py
- services/retrieval-python/app/services/job_runner.py

### Fase 3: Parametrizar el pipeline actual
Tareas:
- añadir argumentos CLI para rutas de entrada/salida
- evitar rutas hardcodeadas
- supportar collection actual y nombre explícito
- adaptar upload_to_qdrant.py para aceptar job_id y metadata

Archivos sugeridos:
- services/retrieval-python/scripts/process_manual_opt.py
- services/retrieval-python/scripts/index_manual.py
- services/retrieval-python/scripts/upload_to_qdrant.py

### Fase 4: Integrar MinIO
Tareas:
- añadir configuración de MinIO
- añadir cliente para descarga de objetos
- generar object_key y SHA256
- validar bucket y colección

### Fase 5: Exponer la API
Tareas:
- crear endpoints de ingestion
- cuidar validaciones y manejo de errores
- devolver el estado del job y los fallidos

### Fase 6: Validar en Docker y con pruebas reales
Tareas:
- arrancar Redis, MinIO y worker
- subir un PDF real
- comprobar job COMPLETED
- comprobar metadata en Qdrant
- comprobar limpieza del workspace
- comprobar DLQ tras un fallo forzado

---

## Criterios de aceptación del producto

### Caso 1: flujo exitoso
1. Se envía un PDF a la cola de ingestión.
2. Se crea un job con un UUID.
3. El worker procesa el PDF en un directorio temporal aislado.
4. El OCR, indexación y carga en Qdrant se ejecutan usando rutas por job.
5. Qdrant recibe metadata válida: job_id, object_key, file_sha256 y ingested_at.
6. El job termina en COMPLETED.
7. El workspace temporal queda eliminado.

### Caso 2: fallo y reintento
1. Un job falla por error o timeout.
2. Se reintenta automáticamente hasta max_retries.
3. Si se supera el límite, el job pasa a FAILED_PERMANENTLY.
4. La información del error queda persistida en la DLQ.
5. El workspace temporal se elimina.

### Caso 3: concurrencia e idempotencia
1. Se lanzan varios workers simultáneamente.
2. Se procesan varios PDFs diferentes sin colisión.
3. Cada job usa su carpeta y su JSON temporal único.
4. Si el mismo PDF llega dos veces, el sistema detecta duplicado por SHA256 + collection.
5. El segundo trabajo se marca como SKIPPED o COMPLETED inmediato.

### Caso 4: trazabilidad
1. Se puede localizar todos los chunks de un documento en Qdrant por object_key o file_sha256.
2. Se puede reindexar un documento sin tocar el resto de la colección.
3. El historial del trabajo permite auditar qué archivo, cuándo y quién generó los vectores.

---

## Definition of Done

La feature 01 se considera completa cuando:
- existe cola de trabajo funcional con Redis RQ
- existe worker responsable de orquestar el pipeline
- el pipeline actual queda parametrizado para job temporales
- MinIO o una entrada compatible están integrados
- los chunks subidos a Qdrant llevan metadata mínima requerida
- la deduplicación por SHA256 + collection está implementada
- existe DLQ y reintentos configurados
- existe API de ingestión y consulta
- los trabajos quedan registrados con estado y timestamps
- el cleanup temporal está garantizado
- la feature se valida con un PDF real y un caso de fallo

---

## Recomendación final para la implementación

La mejor estrategia para este repositorio es una entrega incremental, no un rediseño completo de golpe:
1. primero encapsular el flujo actual en un worker con job_id
2. luego añadir Redis y reintentos
3. luego introducir MinIO
4. luego cubrir API y observabilidad

Esto permite mantener la funcionalidad actual en producción mientras se hace la transición a la nueva ingesta controlada.

---

## Cierre

Esta feature no debe interpretarse como una reescritura global del sistema. Debe ser una capa operativa de ingestión que reutiliza la lógica ya existente, separa la orquestación del negocio, añade trazabilidad y deja al proyecto preparado para crecer sin romper la base actual.

Si se sigue este enfoque, la feature 01 puede pasar de ser un diseño de roadmap a una solución funcional y mantenible dentro del repositorio actual.