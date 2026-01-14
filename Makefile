.PHONY: dev dev-server dev-web build test docker-up docker-down play

# Development with live reload
dev: dev-server

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
	@echo "  make dev        - Start Go server with Air live reload"
	@echo "  make dev-web    - Start Vite frontend dev server"
	@echo "  make build      - Build Go binary and web frontend"
	@echo "  make test       - Run Go tests"
	@echo "  make play       - Play the game locally"
	@echo "  make server     - Run server without live reload"
	@echo "  make docker-up  - Start Docker containers"
	@echo "  make docker-down - Stop Docker containers"
	@echo ""
	@echo "Development workflow:"
	@echo "  Tab 1: make dev      (Go server with live reload)"
	@echo "  Tab 2: make dev-web  (Vite with HMR)"
