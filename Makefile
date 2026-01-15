.PHONY: dev dev-server dev-web dev-all build test test-integration docker-up docker-down play

# Development with live reload (both services)
dev: dev-all

# Start both server and frontend together
dev-all:
	@if command -v overmind >/dev/null 2>&1; then \
		echo "Starting both services with Overmind..."; \
		overmind start -f Procfile.dev; \
	else \
		echo "Overmind not found. Install with: brew install overmind"; \
		echo "Or run in separate terminals:"; \
		echo "  Tab 1: make dev-server"; \
		echo "  Tab 2: make dev-web"; \
		exit 1; \
	fi

dev-server:
	@echo "Starting Go server with Air (live reload)..."
	@set -a && source .env && set +a && air

dev-web:
	@echo "Starting Vite dev server with HMR..."
	cd web && npm run dev

# Build everything (development - uses disk static files)
build:
	cd web && npm run build
	go build -o ./bin/devdungeon ./cmd/devdungeon

# Build for production (embedded static files in single binary)
build-prod:
	@echo "Building frontend..."
	cd web && npm run build
	@echo "Copying static files for embedding..."
	rm -rf cmd/devdungeon/static && mkdir -p cmd/devdungeon/static
	cp -r web/build/* cmd/devdungeon/static/
	@echo "Building Go binary with embedded frontend..."
	go build -o ./bin/devdungeon ./cmd/devdungeon
	@echo "Restoring placeholder..."
	rm -rf cmd/devdungeon/static && mkdir -p cmd/devdungeon/static && touch cmd/devdungeon/static/.gitkeep
	@echo "Done! Binary at ./bin/devdungeon"

# Run tests
test:
	go test ./...

# Run integration tests (requires Docker)
test-integration:
	@echo "Starting PostgreSQL container..."
	@docker run -d --name devdungeon-test-db \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=devdungeon_test \
		-p 5434:5432 \
		postgres:18 > /dev/null
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@echo "Running integration tests..."
	@DATABASE_URL="postgres://postgres:postgres@localhost:5434/devdungeon_test?sslmode=disable" \
		go test -v -tags=integration ./internal/db/... || (docker stop devdungeon-test-db > /dev/null && docker rm devdungeon-test-db > /dev/null && exit 1)
	@echo "Cleaning up..."
	@docker stop devdungeon-test-db > /dev/null && docker rm devdungeon-test-db > /dev/null
	@echo "Done!"

# Docker commands
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f devdungeon

docker-build:
	docker compose build

# Play the game locally (no server)
play:
	go run ./cmd/devdungeon

# Run server locally (no live reload)
server:
	source .env && go run ./cmd/devdungeon --server

# Help
help:
	@echo "Available commands:"
	@echo "  make dev          - Start both server + frontend (requires overmind)"
	@echo "  make dev-server   - Start Go server with Air live reload"
	@echo "  make dev-web      - Start Vite frontend dev server"
	@echo "  make build        - Build Go binary and web frontend (dev mode)"
	@echo "  make build-prod   - Build single binary with embedded frontend"
	@echo "  make test         - Run Go unit tests"
	@echo "  make test-integration - Run integration tests (requires Docker)"
	@echo "  make play         - Play the game locally"
	@echo "  make server       - Run server without live reload"
	@echo "  make docker-up    - Start Docker containers"
	@echo "  make docker-down  - Stop Docker containers"
	@echo ""
	@echo "Development workflow:"
	@echo "  make dev         (both services in one terminal - requires overmind)"
	@echo "  - OR -"
	@echo "  Tab 1: make dev-server (Go server with live reload)"
	@echo "  Tab 2: make dev-web    (Vite with HMR)"
	@echo ""
	@echo "Install overmind: brew install overmind"
