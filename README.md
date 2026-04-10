# gabriel-q7 — Portfolio Monorepo

A monorepo containing the frontend, backend, infrastructure and automation scripts for the portfolio project.

## Structure

```
first-portfolio/
├── apps/
│   ├── frontend/      # SvelteKit hacker-terminal UI (TypeScript + Tailwind CSS v4)
│   └── backend/       # Backend service (to be implemented)
│
├── infra/             # Docker, Terraform, etc.
├── scripts/           # Automation scripts
├── docs/              # Project documentation
│
├── docker-compose.yml
├── Makefile
└── README.md
```

## Apps

### `apps/frontend`

A portfolio interface built with a **hacker/terminal aesthetic**. Features:

- **Stack**: SvelteKit 5 · Tailwind CSS v4 · TypeScript · `@sveltejs/adapter-node`
- Boot sequence animation on load
- Interactive terminal with commands: `help`, `about`, `projects`, `contact`, `clear`
- Blinking cursor, CRT scanlines overlay, subtle green glow
- JetBrains Mono font, macOS-style window chrome
- API service layer ready to connect to the backend via `VITE_API_URL`

#### Running locally

```bash
cd apps/frontend
cp .env.example .env     # set VITE_API_URL if needed
npm install
npm run dev              # starts on http://localhost:5173
```

#### Environment variables

| Variable       | Default                   | Description              |
|----------------|---------------------------|--------------------------|
| `VITE_API_URL` | `http://localhost:8080`   | Backend API base URL     |

### `apps/backend`

Backend service — to be implemented. Will expose:
- `GET /health` — healthcheck
- `GET /projects` — projects list

---

## Getting Started (Docker)

### Prerequisites

- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/)
- [Make](https://www.gnu.org/software/make/)

### Running the project

```bash
# Build all services
make build

# Start all services in detached mode
make up

# View logs
make logs

# Stop all services
make down
```

The frontend will be available at **http://localhost:3000**.

---

## Architecture decisions

- **Monorepo layout**: `apps/` for deployable services; `packages/` (future) for shared types/configs.
- **Frontend isolation**: the frontend has its own `package.json`, `Dockerfile`, and `.env` — fully independent build and deploy.
- **No hardcoded endpoints**: all API calls go through `src/lib/services/api.ts`, reading from `VITE_API_URL`.
- **Svelte 5 runes**: uses the modern `$state`, `$props`, and `$derived` primitives for clean reactivity.
- **Tailwind v4**: CSS-first configuration via `@import "tailwindcss"` in `app.css`; custom design tokens as CSS variables.

## Future evolution

- Add `packages/ui` for shared headless components across apps
- Add `packages/types` for shared TypeScript types between frontend and backend
- Implement the backend and wire up `fetchProjects()` and `fetchHealthcheck()`
- Add CI/CD pipeline (e.g., GitHub Actions) for independent builds per app
- Add E2E tests with Playwright