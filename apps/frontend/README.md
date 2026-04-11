# Portfolio Frontend

## Purpose

Renders a hacker/terminal-style portfolio interface in the browser. Visitors interact with a simulated terminal by typing commands to explore profile information, projects, and contact details.

## Responsibilities

- Present a boot sequence animation on first load
- Accept user input in a terminal prompt and route it to the command registry
- Render command output as styled terminal lines
- Call the backend API to fetch live project data (once wired up)
- Apply a CRT / green-phosphor visual theme via Tailwind CSS v4

## Stack

| Technology | Role |
|---|---|
| SvelteKit 5 | App framework, routing, SSR/adapter |
| Tailwind CSS v4 | Styling (CSS-first, CSS variables) |
| TypeScript | Type safety across the app |
| `@sveltejs/adapter-node` | Production Node.js server |
| Vite 8 | Dev server and bundler |

## Project Structure

```
src/
├── app.css                    # Global styles, Tailwind import, CSS design tokens
├── app.html                   # HTML shell
├── app.d.ts                   # SvelteKit ambient types
├── lib/
│   ├── assets/                # Static assets (favicon, etc.)
│   ├── commands/
│   │   └── index.ts           # Command registry and executeCommand()
│   ├── components/
│   │   ├── Cursor.svelte      # Blinking cursor
│   │   ├── TerminalHistory.svelte  # Scrollable output area
│   │   ├── TerminalInput.svelte    # Input prompt
│   │   ├── TerminalLine.svelte     # Single output line renderer
│   │   └── TerminalWindow.svelte  # Window chrome (title bar, body)
│   ├── services/
│   │   └── api.ts             # HTTP client + fetchProjects / fetchHealthcheck
│   ├── stores/
│   │   └── terminal.ts        # terminalHistory store, isBooting, isReady flags
│   ├── types/
│   │   └── terminal.ts        # TerminalLine and Command type definitions
│   ├── utils/
│   │   └── boot.ts            # Boot sequence logic
│   └── index.ts               # Library re-exports
└── routes/
    ├── +layout.svelte         # Root layout
    └── +page.svelte           # Terminal page (entry point)
```

## Getting Started

```bash
cp .env.example .env     # configure VITE_API_URL if needed
npm install
npm run dev              # http://localhost:5173
```

## Available Scripts

| Script | Description |
|---|---|
| `npm run dev` | Start development server |
| `npm run build` | Build production bundle |
| `npm run preview` | Preview the production build locally |
| `npm run check` | Type-check with svelte-check |
| `npm run check:watch` | Type-check in watch mode |

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `VITE_API_URL` | `http://localhost:8080` | Backend API base URL |

## Integrations

- **Backend API** — `src/lib/services/api.ts` wraps all HTTP calls. `fetchProjects()` calls `GET /api/v1/projects`; `fetchHealthcheck()` calls `GET /health`. The base URL is read from `VITE_API_URL`.

## Related Documentation

- [Architecture — Frontend](../../docs/architecture/frontend.md)
- [Business Rules — Terminal Commands](../../docs/business-rules/terminal-commands.md)
- [Business Rules — Session History](../../docs/business-rules/session-history.md)
- [API Endpoints](../../docs/api/endpoints.md)
