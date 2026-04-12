# gabriel-q7 — Portfolio Monorepo

A monorepo containing the frontend and backend for a hacker/terminal-style personal portfolio.

## What It Is

An interactive portfolio presented as a fake terminal. Visitors type commands (`help`, `about`, `projects`, `contact`) and see styled output. The backend exposes a REST API that serves project data and health endpoints.

## Monorepo Structure

```
first-portfolio/
├── apps/
│   ├── frontend/          # SvelteKit 5 hacker-terminal UI
│   └── backend/           # Go REST API
├── docs/                  # Architecture, business rules, decisions, API docs
├── docker-compose.yml     # Full-stack local environment
├── Makefile               # Root convenience targets
└── README.md
```

## Stack

| Layer    | Technology                                      |
|----------|-------------------------------------------------|
| Frontend | SvelteKit 5 · Tailwind CSS v4 · TypeScript      |
| Backend  | Go 1.23 · net/http · PostgreSQL · Redis         |
| Infra    | Docker · Docker Compose                         |

## Getting Started

### Prerequisites

- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/)
- [Make](https://www.gnu.org/software/make/)

### Run with Docker Compose (recommended)

```bash
# Build all images
make build

# Start all services (frontend, backend, postgres, redis)
make up

# Stream logs
make logs

# Stop all services
make down
```

| Service  | URL                        |
|----------|----------------------------|
| Frontend | http://localhost:3000       |
| Backend  | http://localhost:8080       |
| Postgres | localhost:5432              |
| Redis    | localhost:6379              |

### Run locally without Docker

```bash
# Frontend
cd apps/frontend
cp .env.example .env
npm install
npm run dev          # http://localhost:5173

# Backend
cd apps/backend
cp .env.example .env
make run             # http://localhost:8080
```

## Available Scripts (root)

| Command              | Description                              |
|----------------------|------------------------------------------|
| `make build`         | Build all Docker images                  |
| `make up`            | Start all services in detached mode      |
| `make down`          | Stop all services                        |
| `make logs`          | Tail logs for all services               |
| `make backend-build` | Compile the Go binary                    |
| `make backend-run`   | Run the backend directly (no Docker)     |
| `make backend-test`  | Run Go tests                             |
| `make backend-lint`  | Run golangci-lint on backend             |
| `make frontend-dev`  | Start frontend dev server                |

## Apps

- **[apps/frontend](apps/frontend/README.md)** — SvelteKit terminal UI
- **[apps/backend](apps/backend/README.md)** — Go REST API

## Documentation

| Document | Description |
|----------|-------------|
| [docs/architecture/overview.md](docs/architecture/overview.md) | System-level architecture |
| [docs/architecture/frontend.md](docs/architecture/frontend.md) | Frontend architecture |
| [docs/architecture/backend.md](docs/architecture/backend.md) | Backend architecture |
| [docs/business-rules/terminal-commands.md](docs/business-rules/terminal-commands.md) | Terminal command rules |
| [docs/business-rules/session-history.md](docs/business-rules/session-history.md) | Session history behaviour |
| [docs/business-rules/error-handling.md](docs/business-rules/error-handling.md) | Error handling rules |
| [docs/api/endpoints.md](docs/api/endpoints.md) | REST API reference |
| [docs/decisions/](docs/decisions/) | Architectural Decision Records |
| [docs/glossary.md](docs/glossary.md) | Domain glossary |
| [docs/conventions.md](docs/conventions.md) | Documentation conventions |