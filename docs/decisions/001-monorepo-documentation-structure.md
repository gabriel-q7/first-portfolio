# ADR 001 — Monorepo Documentation Structure

## Status

Accepted

## Context

The repository grew from a single frontend app into a full-stack monorepo with a Go backend, Docker Compose infrastructure, and multiple cross-cutting concerns. Documentation was scattered across a single root `README.md` and a placeholder frontend `README.md`. Business rules were implicit in code comments; architectural decisions were not recorded anywhere. Onboarding and future maintenance required reading source files rather than dedicated documentation.

## Decision

Introduce a structured `/docs` directory alongside updated READMEs for the root, frontend, and backend. The structure follows the layout described in the problem statement:

```
docs/
  architecture/     ← system, frontend, backend architecture
  business-rules/   ← domain rules extracted from code
  decisions/        ← ADRs numbered sequentially
  api/              ← REST API reference
  glossary.md       ← shared vocabulary
  conventions.md    ← documentation standards
```

Each README is scoped to its app or package and links to the relevant `/docs` pages. The root README serves as the project entry point with a documentation index.

## Consequences

- New contributors have a single entry point (root README) and clear paths to deeper documentation.
- Business rules are no longer buried in code comments; they are explicitly maintained.
- Architectural decisions are traceable; future decisions add a new numbered ADR rather than editing existing docs.
- Documentation must be updated alongside code; conventions and update rules are defined in `docs/conventions.md`.

## Alternatives Considered

- **Single root README with all content**: Rejected because a single file becomes unwieldy and mixes audience concerns (quick start vs. deep architecture).
- **Wiki (GitHub Wiki)**: Rejected because it is disconnected from code and cannot be reviewed alongside PRs.
- **No `/docs` directory**: Rejected because business rules and decisions embedded in code comments are invisible to non-engineers and hard to navigate.
