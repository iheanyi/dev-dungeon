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
FROM golang:1.24-alpine AS backend

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /devdungeon ./cmd/devdungeon

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from backend builder
COPY --from=backend /devdungeon /app/devdungeon

# Copy web frontend from frontend builder
COPY --from=frontend /app/web/build /app/web/build

# Create .ssh directory for host key
RUN mkdir -p /app/.ssh

# SSH and HTTP ports
EXPOSE 2222 8080

# Default to running both servers
ENTRYPOINT ["/app/devdungeon"]
CMD ["--server"]
