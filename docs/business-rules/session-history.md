# Business Rule — Session History

## Purpose

Define how the terminal's line history is stored, updated, and cleared within a browser session.

## Description

The terminal history is a reactive in-memory list of `TerminalLine` objects managed by the `terminalHistory` Svelte store. It is initialised as an empty array when the page loads and accumulates lines as the user interacts with the terminal. History is never persisted to `localStorage` or sent to the backend; it exists only for the lifetime of the browser tab.

## Inputs

- `addLine(line)` — appends a single line.
- `addLines(lines[])` — appends multiple lines atomically (used by `executeCommand()`).
- `clear()` — replaces the list with an empty array.

Each added line receives a randomly generated `id` (7-character base-36 string) and a `timestamp` (`Date`) assigned at insertion time.

## Outputs

The store exposes a `subscribe` interface. `TerminalHistory.svelte` subscribes and re-renders the list on every update.

## Constraints

- History is bounded only by available browser memory; there is currently no maximum line count.
- Unique `id` values are generated with `Math.random().toString(36).slice(2, 9)`. Collisions are theoretically possible but inconsequential (IDs are used only as Svelte `key` values for list rendering).
- The `clear` command empties the history via `terminalHistory.clear()`; this action is irreversible within the session.
- The boot sequence lines (animated startup text) are the first entries written to history; they appear before any user input is accepted.

## Edge Cases

- **Tab reload** — history resets; the boot sequence replays.
- **`clear` during boot** — the `clear` command is only accessible after `isReady` is `true`, so it cannot interrupt the boot animation.

## Related Code

- `src/lib/stores/terminal.ts` — `terminalHistory` store, `isBooting`, `isReady`
- `src/lib/utils/boot.ts` — writes boot sequence lines to history
- `src/lib/components/TerminalHistory.svelte` — renders the history list
- `src/lib/commands/index.ts` — `clear` command emits the `__CLEAR__` sentinel

## Related Decisions

- [Architecture — Frontend](../architecture/frontend.md)
