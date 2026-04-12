# Documentation Conventions

This document defines the standards for writing and maintaining documentation in this repository.

---

## 1. README Convention

Every README must follow this section order (omit sections that do not apply):

```md
# Name

## Purpose
## Responsibilities
## Stack
## Project Structure
## Getting Started
## Available Scripts
## Environment Variables
## Integrations
## Related Documentation
```

**Root README** — project entry point. Contains overview, monorepo structure, stack summary, setup commands, and a documentation index. Does not contain detailed business rules or architecture explanations.

**Local README** (`apps/*/README.md`) — explains the specific app or package. Contains all sections above as applicable. Links to relevant `/docs` pages.

---

## 2. Business Rule Convention

Each file in `docs/business-rules/` must follow:

```md
# Rule Name

## Purpose
## Description
## Inputs
## Outputs
## Constraints
## Edge Cases
## Examples
## Related Code
## Related Decisions
```

---

## 3. ADR Convention

Each file in `docs/decisions/` must follow:

```md
# ADR XXX — Title

## Status
## Context
## Decision
## Consequences
## Alternatives Considered
```

Status values: `Proposed` | `Accepted` | `Deprecated` | `Superseded by ADR XXX`.

ADRs are append-only. Do not edit a past decision's **Decision** or **Context** sections. If a decision changes, write a new ADR and mark the old one as `Superseded by ADR XXX`.

---

## 4. Naming Convention

| Rule | Detail |
|---|---|
| File names | kebab-case: `terminal-commands.md`, `002-backend-architecture.md` |
| Concept names | Use domain language, not implementation detail: `terminal-commands` not `executeCommand-logic` |
| ADR numbering | Three-digit zero-padded sequential: `001`, `002`, `003` |
| Headings | Title case for top-level `#`; sentence case for `##` and below |

---

## 5. Code Comment Convention

Comments in source files explain **why**, not **what**. If an explanation requires more than a short rationale, it belongs in `/docs/business-rules` or `/docs/architecture`. Long block comments should be replaced with a link to the relevant doc.

---

## 6. When to Update Documentation

### Update root README when:

- Repository structure changes (new app, new top-level directory)
- Setup process changes (new prerequisite, changed port)
- Root scripts change
- Documentation index changes

### Update a local README when:

- App responsibility changes
- Local commands change
- Environment variables are added, removed, or renamed
- Main folders or entry points change
- Integrations change

### Add or update `docs/business-rules/` when:

- Business behaviour changes
- Command behaviour changes
- Validation rules change
- Edge cases are discovered or change

### Add an ADR when:

- Choosing a major architectural direction
- Changing a technical strategy
- Introducing an important constraint
- Replacing a previous decision

### Update `docs/architecture/` when:

- A component's responsibility or structure changes significantly
- A new cross-cutting concern is introduced

---

## 7. Documentation Review in Pull Requests

When delivering a feature or fix, ask:

- [ ] Does the root README need updating?
- [ ] Does any local README need updating?
- [ ] Is a new business rule document needed or an existing one updated?
- [ ] Is a new ADR needed?
- [ ] Does the API reference need updating?
- [ ] Does the glossary need a new term?

If none of the above apply, no documentation change is required.

---

## 8. Ownership

Documentation lives alongside the code it describes and is owned by whoever changes the code. There is no separate documentation role. A PR that changes observable behaviour without updating the relevant docs is incomplete.
