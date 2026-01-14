# Multiplayer Feature Spec: SSH-Based /dev/dungeon

## Overview

Transform /dev/dungeon into an SSH-accessible multiplayer roguelike where players connect via `ssh username@devdungeon.io` and their identity is tied to their SSH public key.

---

## Identity & Authentication

### SSH Key-Based Identity

**How it works:**
1. Player generates SSH keypair locally: `ssh-keygen -t ed25519 -f ~/.ssh/id_devdungeon`
2. Player registers on web portal: uploads public key + chooses username
3. Server stores: `username -> public_key_fingerprint` mapping
4. On SSH connect: server validates key matches registered username
5. Player identity = their public key fingerprint (immutable, secure)

**Key fingerprint as identity:**
```
SHA256:abc123... -> "iheanyi" (username)
                 -> player_id: uuid
                 -> meta_progress, saves, stats
```

### Security Model

| Layer | Protection |
|-------|------------|
| **Transport** | SSH encryption (no MITM possible) |
| **Auth** | Ed25519 public key (no passwords) |
| **Identity** | Key fingerprint -> DB user record |
| **Session** | One Bubble Tea program per connection |
| **Data** | PostgreSQL encryption at rest |

**Key security measures:**
- Rate limit: max 5 failed auth attempts per IP per minute
- Session timeout: 30 min idle disconnect
- Logging: IP, key fingerprint, timestamp, duration for all connections
- Admin commands: ban/kick by fingerprint
- No shell access: Wish only runs the game, nothing else

---

## Architecture

### Tech Stack

```
Player Terminal
    |
    | ssh player@devdungeon.io
    v
┌─────────────────────────────────────────────┐
│         Wish SSH Server (Go)                │
│  ├─ charmbracelet/wish                      │
│  ├─ Public key auth middleware              │
│  └─ Session handler                         │
└─────────────────────────────────────────────┘
    |
    | Per-session
    v
┌─────────────────────────────────────────────┐
│         Game Session                        │
│  ├─ Bubble Tea program (existing UI)        │
│  ├─ Game Engine instance                    │
│  └─ PTY connected to SSH session            │
└─────────────────────────────────────────────┘
    |
    v
┌──────────────────────────────────┐
│           PostgreSQL             │
├──────────────────────────────────┤
│ accounts      │ leaderboard      │
│ game_saves    │ active_sessions  │
│ meta_progress │ async_drops      │
│ statistics    │ session_locks    │
└──────────────────────────────────┘
```

### New Packages

```
internal/
├── server/           # NEW: SSH server setup
│   ├── server.go     # Wish configuration
│   ├── auth.go       # Public key -> user lookup
│   └── session.go    # Per-connection handler
├── db/               # NEW: Database layer
│   ├── postgres.go   # Player accounts, saves, sessions, leaderboards
│   └── models.go     # DB models
├── accounts/         # NEW: User management
│   ├── register.go   # Web registration flow
│   └── profile.go    # Player profiles
```

---

## Multiplayer Features

### Phase 1: Single-Player over SSH
- Players SSH in with their key
- Game state persists in PostgreSQL (not local files)
- Play from any machine with your key
- Global leaderboards (PostgreSQL)

### Phase 2: Async Multiplayer
- **Message drops**: Leave notes in dungeons for others to find
- **Item drops**: Leave items at locations (persist 24h)
- **Seed sharing**: Share seeds for competitive runs
- **Ghost data**: See where other players died (gravestones)

### Phase 3: Sync Multiplayer (Future)
- Shared dungeon instances
- Turn order by NICE stat (lower = faster)
- Co-op boss fights
- PvP arenas

---

## Database Schema (PostgreSQL)

**IDs**: Use go-nanoid (alphanumeric, 21 chars) instead of UUIDs for readability.
**Usernames**: CITEXT for case-insensitive matching, changes allowed.

```sql
-- Enable citext extension
CREATE EXTENSION IF NOT EXISTS citext;

-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    nanoid VARCHAR(21) UNIQUE NOT NULL,  -- go-nanoid for external refs
    username CITEXT UNIQUE NOT NULL,      -- case-insensitive, changeable
    public_key_fingerprint VARCHAR(64) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    last_login TIMESTAMP,
    is_banned BOOLEAN DEFAULT FALSE
);

-- Game saves (replaces local JSON)
CREATE TABLE game_saves (
    id SERIAL PRIMARY KEY,
    nanoid VARCHAR(21) UNIQUE NOT NULL,
    user_id INT REFERENCES users(id),
    save_data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Meta progression (replaces local meta.json)
CREATE TABLE meta_progress (
    user_id INT PRIMARY KEY REFERENCES users(id),
    total_exit_codes INT DEFAULT 0,
    unlocked_classes TEXT[] DEFAULT ARRAY['init'],
    permanent_bonuses JSONB DEFAULT '{}',
    unlocked_items TEXT[] DEFAULT ARRAY[]::TEXT[],
    runs_completed INT DEFAULT 0,
    deepest_floor INT DEFAULT 0,
    total_deaths INT DEFAULT 0
);

-- Leaderboards
CREATE TABLE leaderboard_entries (
    id SERIAL PRIMARY KEY,
    nanoid VARCHAR(21) UNIQUE NOT NULL,
    user_id INT REFERENCES users(id),
    run_type VARCHAR(32),  -- 'standard', 'daily', 'seeded'
    seed BIGINT,
    score INT,
    floors_cleared INT,
    time_seconds INT,
    class VARCHAR(32),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Async drops (messages/items left for others)
CREATE TABLE world_drops (
    id SERIAL PRIMARY KEY,
    nanoid VARCHAR(21) UNIQUE NOT NULL,
    user_id INT REFERENCES users(id),
    floor_type VARCHAR(32),
    position_x INT,
    position_y INT,
    drop_type VARCHAR(16),  -- 'message', 'item'
    content TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Daily seeds table
CREATE TABLE daily_seeds (
    date DATE PRIMARY KEY,
    seed BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

## Connection Flow

```
1. SSH Handshake
   └─ Client: ssh iheanyi@devdungeon.io
   └─ Server: Present host key
   └─ Client: Accept/verify host key

2. Public Key Auth
   └─ Client: Present public key
   └─ Server: Query DB for fingerprint
   └─ If found: Auth success, get user_id
   └─ If not: Reject with "Register at devdungeon.io/register"

3. Session Setup
   └─ Allocate PTY for terminal
   └─ Load user's game state from PostgreSQL
   └─ Create Engine with DB-backed save manager
   └─ Launch Bubble Tea program

4. Gameplay
   └─ Normal game loop (existing code)
   └─ Saves go to PostgreSQL instead of filesystem
   └─ Leaderboard updates via PostgreSQL

5. Disconnect
   └─ Auto-save to PostgreSQL
   └─ Update last_login timestamp
   └─ Clean up session resources
```

---

## Registration Flow (Web)

Simple web portal at `devdungeon.io`:

1. User visits `/register`
2. Enters desired username
3. Pastes their public key (`~/.ssh/id_ed25519.pub`)
4. Server validates:
   - Username available
   - Key is valid Ed25519/RSA format
   - Fingerprint not already registered
5. Creates user record
6. Shows: "You can now connect with `ssh username@devdungeon.io`"

---

## Migration Path

### Existing Code Changes

| Component | Change |
|-----------|--------|
| `cmd/devdungeon/main.go` | Add `--server` flag to run as SSH server |
| `internal/save/manager.go` | Add PostgreSQL backend alongside filesystem |
| `internal/ui/app.go` | Accept `io.Reader`/`io.Writer` for SSH PTY |
| `internal/game/engine.go` | No changes (already decoupled) |

### New Dependencies

```go
// go.mod additions
require (
    github.com/charmbracelet/wish v1.x.x
    github.com/charmbracelet/ssh v0.x.x
    github.com/jackc/pgx/v5 v5.x.x          // PostgreSQL
    github.com/matoous/go-nanoid/v2 v2.x.x  // Alphanumeric IDs
)
```

**NanoID usage:**
```go
import gonanoid "github.com/matoous/go-nanoid/v2"

// Generate alphanumeric ID (default 21 chars)
id, _ := gonanoid.Generate("0123456789abcdefghijklmnopqrstuvwxyz", 21)
// Result: "v0dxsk3skk4oht1pqv5j2"
```

---

## Security Checklist

- [ ] Ed25519 keys only (reject weak RSA)
- [ ] Rate limit auth attempts (5/min/IP)
- [ ] Session timeout (30 min idle)
- [ ] No shell access (Wish app-only)
- [ ] Log all connections (IP, fingerprint, duration)
- [ ] PostgreSQL connection over TLS
- [ ] Host key stored securely (not in repo)
- [ ] Ban system by fingerprint
- [ ] Input validation on all player actions

---

## Decisions Made

1. **Registration flow**: In-terminal first (connect without key, guided setup), fallback to web portal if too complex
2. **Key mapping**: One key = one account (fingerprint IS identity)
3. **Scope**: Start with SSH + Leaderboards, add async drops, then co-op

## Co-op Room System

```
Player A                          Player B
    |                                 |
    | /host                           |
    v                                 |
  Creates room: "FORK-BOMB"           |
  Seed: 12345                         |
  Waiting for player...               |
    |                                 |
    |                           /join FORK-BOMB
    |                                 |
    v                                 v
  ┌─────────────────────────────────────┐
  │         SHARED DUNGEON              │
  │  - Same floor (seed 12345)          │
  │  - Both players visible on map      │
  │  - Shared enemy HP pools            │
  │  - Turn order by NICE stat          │
  │  - Combat: both attack same enemies │
  └─────────────────────────────────────┘
```

**Turn order in co-op combat:**
- All players + all enemies sorted by NICE (lower = faster)
- Example: Player A (NICE 5), Enemy (NICE 10), Player B (NICE 15)
- A attacks -> Enemy attacks -> B attacks -> repeat

**If one player dies:**
- Option 1: They spectate until floor complete
- Option 2: Run ends for both (hardcore mode)
- Option 3: Respawn at stairs with penalty

## Async Drops System

When entering a floor, 10% chance to see a random drop from another player:
- **Messages**: "iheanyi was here" or "Watch out for the fork bomb!"
- **Items**: Low-tier consumables left behind
- **Gravestones**: "Here lies bash_warrior, slain by daemon"

Drops persist 24-48 hours, selected randomly from pool.

## Decisions Made (Continued)

4. **Username changes**: Allowed (just a row update). Using CITEXT for case-insensitive.
5. **Key loss**: Too bad, start over. Admin panel later for manual intervention.
6. **Daily runs**: Yes! Fixed seed announced daily, separate leaderboard.
7. **Co-op loot**: Random assignment - for M item drops with N players, each drop randomly assigned to one player.

---

## Hosting

**Recommended Stack:**
- **SSH Server**: Fly.io (supports TCP on custom ports, Wish-friendly)
- **Database**: Fly Postgres or Neon ($5/mo starter)

**Fly.io SSH config** (`fly.toml`):
```toml
[services]
  internal_port = 2222
  protocol = "tcp"

[[services.ports]]
  port = 22
  handlers = []  # Raw TCP, no TLS termination
```

**PlanetScale Postgres connection:**
```bash
# Connection string from PlanetScale dashboard
export DATABASE_URL="postgres://user:pass@region.connect.psdb.cloud:5432/devdungeon?sslmode=require"
```

**Alternatives:**
| Platform | SSH Support | Notes |
|----------|-------------|-------|
| Fly.io | YES | Best PaaS option for SSH |
| DigitalOcean | YES | Full control, droplet + managed DB |
| Hetzner | YES | Cheapest VPS option |

---

## Local Testing

```bash
# 1. Start local PostgreSQL 18 (Docker)
docker run -d --name devdungeon-db \
  -e POSTGRES_PASSWORD=dev \
  -e POSTGRES_DB=devdungeon \
  -p 5432:5432 postgres:18

# 2. Run migrations
go run ./cmd/migrate up

# 3. Start SSH server on localhost:2222
go run ./cmd/devdungeon --server --port 2222

# 4. Generate test key (if needed)
ssh-keygen -t ed25519 -f ~/.ssh/id_devdungeon_test -N ""

# 5. Connect from another terminal
ssh -i ~/.ssh/id_devdungeon_test -p 2222 testuser@localhost

# First connection triggers in-terminal registration flow
```

**Environment variables for local dev:**
```bash
export DATABASE_URL="postgres://postgres:dev@localhost:5432/devdungeon?sslmode=disable"
export SSH_HOST_KEY_PATH="./dev_host_key"  # Auto-generated if missing
export SSH_PORT=2222
```

---

## Implementation Order

1. **SSH Server Skeleton** - Wish setup, basic auth
2. **PostgreSQL Integration** - Migrate save system
3. **Web Registration** - Simple Go HTTP server
4. **Leaderboards** - PostgreSQL queries with indexes
5. **Async Drops** - Messages/items between players
6. **Polish** - Timeouts, rate limits, logging

**Future consideration:** PostgreSQL LISTEN/NOTIFY can be used for real-time co-op features when Phase 3 is implemented.

---

## Verification

**Local Testing Checklist:**
- [ ] Docker PostgreSQL running
- [ ] SSH server starts on port 2222
- [ ] First connection triggers registration flow
- [ ] Can create account with username + public key
- [ ] Rejected on second registration attempt with same key
- [ ] Game state persists after disconnect/reconnect
- [ ] Leaderboard entry created on death/victory
- [ ] Username change works (case-insensitive)
- [ ] Idle timeout disconnects after 30 min
- [ ] NanoIDs generated correctly (alphanumeric, 21 chars)

**Integration Tests:**
```bash
go test ./internal/server/...  # SSH connection tests
go test ./internal/db/...      # Database operations
go test ./internal/accounts/...# Registration flow
```
