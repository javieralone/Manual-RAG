api-go/
├── cmd/
│   └── api/
│       └── main.go                   # Bootstrap e Inyección de Dependencias
└── internal/
    ├── core/                         # NÚCLEO (Sin dependencias externas)
    │   ├── domain/
    │   │   └── query.go              # Entidades y Structs puros
    │   ├── ports/
    │   │   ├── rag_port.go           # Interfaz para el motor Python
    │   │   ├── llm_port.go           # Interfaz para Ollama
    │   │   └── query_service.go      # Interfaz para el Caso de Uso
    │   └── services/
    │       └── query_orchestrator.go # Implementación del Caso de Uso (Lógica)
    └── adapters/                     # INFRAESTRUCTURA Y DETALLES
        ├── http/
        │   ├── handlers/
        │   │   └── query_handler.go  # Controller/Handler HTTP
        │   └── router.go             # Configuración de Rutas
        └── clients/
            ├── python_rag_client.go  # Implementa rag_port.go
            └── ollama_client.go      # Implementa llm_port.go