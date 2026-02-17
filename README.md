# Mini RAG Chat

AI-powered chat with your documents using RAG (Retrieval-Augmented Generation).

Upload documents (PDF, Markdown, TXT), and ask questions — the system retrieves relevant chunks via vector similarity search and generates answers using OpenAI.

## Architecture

```
Browser → Next.js (SSR) → Go API (Fiber) → PostgreSQL + pgvector → OpenAI
```

### Tech Stack

| Layer    | Technology                                        |
| -------- | ------------------------------------------------- |
| Frontend | Next.js 16, TypeScript, Tailwind CSS v4, shadcn/ui |
| Backend  | Go 1.25, Fiber v2, golang-migrate                 |
| Database | PostgreSQL 16 + pgvector                          |
| AI       | OpenAI (chat completions + embeddings)             |
| i18n     | next-intl (en, ru, ky)                            |
| DevOps   | Docker, docker-compose                            |

### Architecture Decisions

- **SSE streaming** — Chat responses stream via Server-Sent Events for real-time UX
- **pgvector** — Vector similarity search runs in PostgreSQL, no separate vector DB needed
- **JWT auth** — Access + refresh token pair; tokens stored in memory (not localStorage)
- **Standalone Next.js output** — Minimal Docker image via `output: "standalone"`
- **Auto-migrations** — Database schema applied automatically on backend startup

## Quick Start

```bash
cp .env.example .env
# Fill in OPENAI_API_KEY and JWT_SECRET
docker compose up --build
# Open http://localhost:3000
```

## API Documentation

Swagger UI available at `http://localhost:8080/swagger/` when the backend is running.

### Endpoints

| Method | Path                          | Auth | Description                |
| ------ | ----------------------------- | ---- | -------------------------- |
| POST   | /api/auth/register            | No   | Register a new user        |
| POST   | /api/auth/login               | No   | Login, returns JWT tokens  |
| POST   | /api/auth/refresh             | No   | Refresh access token       |
| POST   | /api/documents                | Yes  | Upload a document          |
| GET    | /api/documents                | Yes  | List user's documents      |
| DELETE | /api/documents/:id            | Yes  | Delete a document          |
| POST   | /api/chat                     | Yes  | Ask a question (SSE)       |
| GET    | /api/chat/history             | Yes  | List chat sessions         |
| GET    | /api/chat/history/:sessionId  | Yes  | Get messages in a session  |
| GET    | /health                       | No   | Health check               |

## Project Structure

```
rag-chat/
├── backend/
│   ├── cmd/server/          # Entrypoint
│   ├── internal/
│   │   ├── config/          # Environment config
│   │   ├── handler/         # HTTP handlers
│   │   ├── middleware/      # Auth, CORS, rate limiting
│   │   ├── model/           # Domain models
│   │   ├── repository/      # Database layer
│   │   ├── service/         # Business logic
│   │   └── pkg/             # Shared packages
│   │       ├── chunker/     # Text chunking
│   │       ├── llm/         # OpenAI client
│   │       └── parser/      # Document parsers (PDF, MD, TXT)
│   ├── migrations/          # SQL migration files
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── app/[locale]/    # Pages (chat, documents, auth)
│   │   ├── components/      # UI components
│   │   ├── hooks/           # use-auth, use-chat, use-documents
│   │   ├── lib/             # API client, utilities
│   │   ├── i18n/            # Internationalization config
│   │   └── messages/        # Translation files (en, ru, ky)
│   └── Dockerfile
├── docker-compose.yml
├── .env.example
└── .gitignore
```

## Running Tests

```bash
# Backend
cd backend && go test ./...

# Frontend
cd frontend && pnpm test
```

## Environment Variables

| Variable         | Description                       | Default                          |
| ---------------- | --------------------------------- | -------------------------------- |
| `DB_USER`        | PostgreSQL username               | `ragchat`                        |
| `DB_PASSWORD`    | PostgreSQL password               | —                                |
| `OPENAI_API_KEY` | OpenAI API key                    | —                                |
| `JWT_SECRET`     | JWT signing secret (min 32 chars) | —                                |
| `PORT`           | Backend server port               | `8080`                           |
