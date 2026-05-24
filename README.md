# Hackathon QIA - Backend Serfinanza

Backend del Agente 360 para Banco Serfinanza.

## Stack

- Go 1.25+ (chi router)
- Anthropic Claude API (LLM)
- RAG con knowledge.json

## Estructura

```
hackathon-qia/
├── main.go
├── go.mod
├── .env.example
├── internal/
│   ├── server/router.go
│   ├── handlers/handlers.go
│   ├── llm/anthropic.go
│   ├── rag/retrieve.go
│   ├── session/store.go
│   └── domain/types.go
└── data/
    ├── knowledge.json
    └── profiles.json
```

## Inicio rápido

1. Copiar `.env.example` a `.env.dev` y configurar variables
2. `go mod tidy`
3. `go run main.go`

## Docker

```bash
docker compose up --build -d
```

API expuesta en `http://localhost:8090`

## Endpoints

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/ask` | Pregunta al Agente 360 |
| POST | `/simulate-cdt` | Simulador CDT |
| POST | `/recommend` | Recomendación producto |
| POST | `/whatsapp/webhook` | Webhook WhatsApp |

## API Envelope

- Éxito: `{ "data": <payload> }`
- Error: `{ "error": { "code", "message" } }`