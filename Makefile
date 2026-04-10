.PHONY: up down build logs backend-build backend-run backend-test backend-lint frontend-dev

# ── Docker Compose ────────────────────────────────────────────────────────────

up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose build

logs:
	docker compose logs -f

# ── Backend ───────────────────────────────────────────────────────────────────

backend-build:
	$(MAKE) -C apps/backend build

backend-run:
	$(MAKE) -C apps/backend run

backend-test:
	$(MAKE) -C apps/backend test

backend-lint:
	$(MAKE) -C apps/backend lint

# ── Frontend ──────────────────────────────────────────────────────────────────

frontend-dev:
	cd apps/frontend && npm run dev
