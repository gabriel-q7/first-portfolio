# Frontend Architecture

## Purpose

Provides the browser-side user interface: a simulated terminal that accepts typed commands and renders output in a hacker/CRT aesthetic.

## Entry Points

| File | Role |
|---|---|
| `src/routes/+page.svelte` | Single page that mounts `TerminalWindow` |
| `src/routes/+layout.svelte` | Root layout (global styles, fonts) |
| `src/app.html` | HTML shell |

## Component Hierarchy

```
+page.svelte
└── TerminalWindow.svelte       # Window chrome (title bar, body)
    ├── TerminalHistory.svelte  # Scrollable output area (reads terminalHistory store)
    │   └── TerminalLine.svelte # Renders a single line with type-based styling
    ├── TerminalInput.svelte    # Prompt + input field; calls executeCommand()
    └── Cursor.svelte           # Blinking block cursor
```

## State Management

State is held in Svelte stores (`src/lib/stores/terminal.ts`):

| Store | Type | Description |
|---|---|---|
| `terminalHistory` | `TerminalLine[]` | All visible terminal lines |
| `isBooting` | `boolean` | True during the startup animation |
| `isReady` | `boolean` | True once boot is complete and input is active |

The stores are writable and expose typed mutator methods (`addLine`, `addLines`, `clear`). Components subscribe directly; no prop drilling.

## Command Execution Flow

```
User types a command and presses Enter
        │
TerminalInput.svelte
        │ calls executeCommand(input)
src/lib/commands/index.ts
        │
        ├─ if "clear"  → return [{content: '__CLEAR__', type: 'system'}]
        ├─ if unknown  → return [inputLine, errorLine]
        └─ if known    → return [inputLine, ...command.handler()]
        │
terminalHistory.addLines(result)
        │
TerminalHistory.svelte re-renders
```

The special sentinel value `__CLEAR__` is detected by the store/component layer and triggers `terminalHistory.clear()`.

## API Integration

`src/lib/services/api.ts` wraps all backend calls. It reads `VITE_API_URL` from the environment at build time.

| Function | Method | Path |
|---|---|---|
| `fetchHealthcheck()` | GET | `/health` |
| `fetchProjects()` | GET | `/projects` |

`fetchProjects()` is not yet called from any component; it is wired up once the backend is live.

## Styling

- Tailwind CSS v4 loaded via `@import "tailwindcss"` in `app.css`.
- Design tokens (colour palette, font, spacing) are defined as CSS custom properties on `:root`.
- The green-phosphor / CRT effect is achieved entirely through CSS (colour variables + a scanlines pseudo-element).
- Font: JetBrains Mono (loaded from Google Fonts or bundled).

## Build and Deployment

- Dev server: `vite dev` (hot module replacement, port 5173).
- Production build: `vite build` → `@sveltejs/adapter-node` creates a Node.js server.
- Docker: `apps/frontend/Dockerfile` builds and serves the Node adapter output on port 3000.

## See Also

- [Business Rules — Terminal Commands](../business-rules/terminal-commands.md)
- [Business Rules — Session History](../business-rules/session-history.md)
- [API Endpoints](../api/endpoints.md)
