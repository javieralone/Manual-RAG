Manual-RAG/
├── api-go/                     # Microservicio API REST (Go)
│   ├── cmd/api/main.go[cite: 1]
│   ├── internal/[cite: 1]
│   │   ├── handlers/[cite: 1]
│   │   ├── models/[cite: 1]
│   │   └── services/[cite: 1]
│   ├── Dockerfile[cite: 1]
│   └── go.mod[cite: 1]
│
├── rag-engine/                 # Microservicio RAG + MCP (Python)
│   ├── scripts/                # Tus scripts actuales[cite: 1]
│   │   ├── index_manual.py[cite: 1]
│   │   ├── ocr_manual.py[cite: 1]
│   │   ├── query_rag.py[cite: 1]
│   │   └── upload_to_qdrant.py[cite: 1]
│   ├── main.py                 # API interna con FastAPI (para llamar desde Go)
│   ├── mcp_server.py           # Servidor MCP para agentes
│   ├── requirements.txt        # Lista de dependencias del entorno
│   └── Dockerfile              # Dockerfile para el engine de Python
│
├── documents/                  # Archivos PDF originales[cite: 1]
├── output/                     # Salidas JSON temporales/procesadas[cite: 1]
├── qdrant_storage/             # Datos persistentes de Qdrant[cite: 1]
├── .gitignore
├── README.md
└── docker-compose.yml          # Orquestador global[cite: 1]

.\venv\Scripts\activate