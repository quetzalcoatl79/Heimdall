.PHONY: help dev docker-up docker-down backend frontend migrate test lint

# Default target
help:
	@echo "NXO Engine V2 - Available commands:"
	@echo ""
	@echo "  make dev          - Start all services in development mode"
	@echo "  make docker-up    - Start Docker services (Postgres, Redis)"
	@echo "  make docker-down  - Stop Docker services"
	@echo "  make docker-debug - Start with debug tools (Adminer, Redis Commander)"
	@echo "  make backend      - Run backend server"
	@echo "  make frontend     - Run frontend dev server"
	@echo "  make migrate      - Run database migrations"
	@echo "  make migrate-new  - Create new migration (NAME=migration_name)"
	@echo "  make seed         - Seed database with initial data"
	@echo "  make test         - Run all tests"
	@echo "  make lint         - Run linters"
	@echo "  make build        - Build for production"

.PHONY: check-all check-go check-go-deps check-node check-fe-deps check-docker \
        docker-up db-init ensure-workers build-backend build-frontend install run start

# Convenience: run all checks
check-all: check-go check-go-deps check-node check-fe-deps check-docker

# Check Go installation; attempt install on Unix-like systems, instruct on Windows
check-go:
	@echo "-- Checking Go..."
	@if command -v go >/dev/null 2>&1; then \
		printf "Go found: %s\n" "$(shell go version)"; \
		exit 0; \
	fi; \
	OS=$$(uname -s 2>/dev/null || echo Windows); \
	if [ "$$OS" = "Windows" ]; then \
		echo "Go is not installed. Please install Go from https://go.dev/dl for Windows and re-run."; exit 1; \
	else \
		echo "Go not found — attempting to install Go (requires sudo)."; \
		GOVER=1.22.8; ARCH=amd64; URL=https://go.dev/dl/go$${GOVER}.linux-$${ARCH}.tar.gz; \
		curl -fsSL $$URL -o /tmp/go.tgz && sudo tar -C /usr/local -xzf /tmp/go.tgz && rm /tmp/go.tgz && echo "Go installed to /usr/local/go. Add /usr/local/go/bin to your PATH." || (echo "Automatic install failed. Install Go manually." && exit 1); \
	fi

# Ensure backend dependencies are present
check-go-deps:
	@echo "-- Checking Go dependencies (backend)..."
	cd backend && if [ -f go.mod ]; then go mod download || (echo "go mod download failed" && exit 1); else echo "No go.mod found in backend"; fi

# Check Node.js presence; instruct if missing
check-node:
	@echo "-- Checking Node.js/npm..."
	@if command -v node >/dev/null 2>&1 && command -v npm >/dev/null 2>&1; then \
		printf "Node found: %s\n" "$(shell node --version)"; \
	else \
		OS=$$(uname -s 2>/dev/null || echo Windows); \
		if [ "$$OS" = "Windows" ]; then \
			echo "Node.js not found. Please install Node.js (LTS) from https://nodejs.org/ for Windows."; exit 1; \
		else \
			echo "Node not found. Installing using nvm is recommended. See https://github.com/nvm-sh/nvm"; exit 1; \
		fi; \
	fi

# Install frontend dependencies (npm)
check-fe-deps:
	@echo "-- Checking frontend dependencies..."
	cd frontend && if [ -f package.json ]; then \
		if [ -f package-lock.json ]; then npm ci || npm install --legacy-peer-deps; else npm install --legacy-peer-deps; fi; \
	else echo "No package.json in frontend"; fi

# Check Docker and Docker Compose availability
check-docker:
	@echo "-- Checking Docker..."
	@if command -v docker >/dev/null 2>&1; then \
		if docker info >/dev/null 2>&1; then echo "Docker is running"; else echo "Docker found but not running. Start Docker Desktop or daemon."; exit 1; fi; \
	else echo "Docker not found. Install Docker Desktop: https://docs.docker.com/get-docker/"; exit 1; fi

# Development
# 'dev' now runs the full development stack in Docker (build + start)
dev:
	@echo "Starting development stack in Docker (postgres, redis, backend, worker, frontend)..."
	@docker compose up --build -d postgres redis backend worker frontend
	@echo "Showing combined logs (ctrl-c to detach)"
	@docker compose logs -f --tail=200

docker-up:
	docker compose up -d postgres redis
	@echo "Waiting for services to be healthy..."
	@sleep 3

docker-down:
	docker compose down

docker-debug:
	docker compose --profile debug up -d

# Backend
backend:
	cd backend && go run ./cmd/api

backend-watch:
	cd backend && air

# Frontend
frontend:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

# Database
migrate:
	cd backend && go run ./cmd/migrate up

migrate-down:
	cd backend && go run ./cmd/migrate down

migrate-new:
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-new NAME=migration_name"; exit 1; fi
	cd backend && go run ./cmd/migrate create $(NAME)

seed:
	cd backend && go run ./cmd/seed

# Apply DB migrations via CLI if available, otherwise use psql to apply SQL files
db-init:
	@echo "-- Initializing database (migrations)..."
	@if [ -f backend/cmd/migrate/main.go ]; then \
		cd backend && go run ./cmd/migrate up || (echo "Migration command failed" && exit 1); \
	else \
		# fallback: apply SQL files
		for f in backend/migrations/*.up.sql; do \
			echo "Applying $$f"; cat "$$f" | docker exec -i engine_postgres psql -U postgres -d engine_dev || (echo "Failed to apply $$f" && exit 1); \
		done; \
	fi

# Ensure workers: build Go worker binary and provide Celery hint
ensure-workers:
	@echo "-- Ensuring workers are available"
	cd backend && if [ -f cmd/worker/main.go ]; then \
		go build -o bin/worker ./cmd/worker || (echo "Failed to build Go worker" && exit 1); \
		echo "Go worker built: backend/bin/worker"; \
	else \
		echo "No Go worker found at backend/cmd/worker; skipping build."; \
	fi

# Build backend binaries
build-backend:
	@echo "-- Building backend (api + worker)"
	cd backend && CGO_ENABLED=0 go build -o bin/api ./cmd/api || (echo "Failed building api" && exit 1)
	cd backend && if [ -f cmd/worker/main.go ]; then CGO_ENABLED=0 go build -o bin/worker ./cmd/worker || (echo "Failed building worker" && exit 1); fi

# Build frontend
build-frontend:
	@echo "-- Building frontend (Next.js)"
	cd frontend && if [ -f package.json ]; then npm run build || (echo "Frontend build failed" && exit 1); else echo "No frontend to build"; fi

# Full install: checks + deps + builds
install: check-all check-go-deps check-fe-deps build-backend build-frontend

# Start everything: docker, db-init, backend, frontend, worker
run: docker-up db-init ensure-workers
	@echo "-- Starting backend, frontend and worker in background"
	# Start backend
	cd backend && nohup ./bin/api > ../logs/backend.log 2>&1 &
	# Start worker if present
	cd backend && if [ -f bin/worker ]; then nohup ./bin/worker > ../logs/worker.log 2>&1 & fi
	# Start frontend
	cd frontend && nohup node ./node_modules/next/dist/bin/next start -p 3001 > ../logs/frontend.log 2>&1 &
	@sleep 2
	@echo "All services started (logs in ./logs)"

# Testing
test:
	cd backend && go test -v ./...
	cd frontend && npm test

test-coverage:
	cd backend && go test -coverprofile=coverage.out ./...
	cd backend && go tool cover -html=coverage.out

# Linting
lint:
	cd backend && golangci-lint run
	cd frontend && npm run lint

# Build
build: frontend-build
	cd backend && CGO_ENABLED=0 go build -o bin/api ./cmd/api
	cd backend && CGO_ENABLED=0 go build -o bin/worker ./cmd/worker

# Clean
clean:
	rm -rf backend/bin
	rm -rf frontend/.next
	rm -rf frontend/node_modules/.cache
