.PHONY: dev dev-server dev-web dev-all build test docker-up docker-down play

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

# Build everything
build:
	go build -o ./bin/devdungeon ./cmd/devdungeon
	cd web && npm run build

# Run tests
test:
	go test ./...

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
	@echo "  make dev        - Start both server + frontend (requires overmind)"
	@echo "  make dev-server - Start Go server with Air live reload"
	@echo "  make dev-web    - Start Vite frontend dev server"
	@echo "  make build      - Build Go binary and web frontend"
	@echo "  make test       - Run Go tests"
	@echo "  make play       - Play the game locally"
	@echo "  make server     - Run server without live reload"
	@echo "  make docker-up  - Start Docker containers"
	@echo "  make docker-down - Stop Docker containers"
	@echo ""
	@echo "Development workflow:"
	@echo "  make dev         (both services in one terminal - requires overmind)"
	@echo "  - OR -"
	@echo "  Tab 1: make dev-server (Go server with live reload)"
	@echo "  Tab 2: make dev-web    (Vite with HMR)"
	@echo ""
	@echo "Install overmind: brew install overmind"
