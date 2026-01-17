#!/bin/sh
set -e

# SSH Host Key Persistence for /dev/dungeon
# This script ensures a persistent SSH host key exists to prevent
# "HOST IDENTIFICATION HAS CHANGED" warnings for users.

# Default host key path (can be overridden by SSH_HOST_KEY_PATH env var)
HOST_KEY_PATH="${SSH_HOST_KEY_PATH:-/app/.ssh/devdungeon_host_key}"
HOST_KEY_DIR=$(dirname "$HOST_KEY_PATH")

# Ensure the directory exists
mkdir -p "$HOST_KEY_DIR"

# Generate host key if it doesn't exist
if [ ! -f "$HOST_KEY_PATH" ]; then
    echo "==> SSH host key not found at $HOST_KEY_PATH"
    echo "==> Generating new ed25519 host key..."
    ssh-keygen -t ed25519 -f "$HOST_KEY_PATH" -N "" -C "devdungeon-host-key"
    echo "==> Host key generated successfully"
    echo "==> Fingerprint: $(ssh-keygen -lf "$HOST_KEY_PATH")"
else
    echo "==> Using existing SSH host key at $HOST_KEY_PATH"
    echo "==> Fingerprint: $(ssh-keygen -lf "$HOST_KEY_PATH")"
fi

# Set appropriate permissions
chmod 600 "$HOST_KEY_PATH"
chmod 644 "$HOST_KEY_PATH.pub" 2>/dev/null || true

# Export the host key path for the application
export SSH_HOST_KEY_PATH="$HOST_KEY_PATH"

echo "==> Starting /dev/dungeon server..."

# Check for migration force (to fix dirty state)
MIGRATE_ARGS=""
if [ -n "$MIGRATE_FORCE_VERSION" ]; then
    echo "==> Forcing migration version to $MIGRATE_FORCE_VERSION"
    MIGRATE_ARGS="--migrate-force $MIGRATE_FORCE_VERSION"
fi

# Execute the main application with all arguments
exec /app/devdungeon $MIGRATE_ARGS "$@"
