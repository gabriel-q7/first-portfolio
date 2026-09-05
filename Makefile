.PHONY: up down build backend-image nginx-image logs backend-build backend-run backend-test backend-lint frontend-dev

# ── Docker Compose ────────────────────────────────────────────────────────────

up:
	docker compose up -d

down:
	docker compose down

build:
	$(MAKE) backend-image nginx-image

backend-image:
	docker build --file apps/backend/Dockerfile --build-arg VERSION=dev --tag portfolio-backend:dev apps/backend

nginx-image:
	docker build --file apps/frontend/Dockerfile --build-arg VERSION=dev --tag portfolio-nginx:dev .

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
