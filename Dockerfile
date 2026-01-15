# Frontend build stage
FROM node:22-alpine AS frontend

WORKDIR /app/web

# Copy package files
COPY web/package*.json ./
RUN npm ci

# Copy frontend source and build
COPY web/ ./
RUN npm run build

# Backend build stage
FROM golang:1.25-alpine AS backend

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy frontend build for embedding
COPY --from=frontend /app/web/build /app/cmd/devdungeon/static

# Build the binary with embedded frontend
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /devdungeon ./cmd/devdungeon

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates openssh-keygen

WORKDIR /app

# Copy binary from backend builder (includes embedded frontend)
COPY --from=backend /devdungeon /app/devdungeon

# Copy entrypoint script
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Create directories for host key (will be mounted as volumes)
# /data is used in production (Fly.io), /app/.ssh for local dev
RUN mkdir -p /app/.ssh /data

# SSH and HTTP ports
EXPOSE 2222 8080

# Use entrypoint script to handle host key setup
ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["--server"]
