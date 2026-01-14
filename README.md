# /dev/dungeon

```
  /██████╗ ███████╗██╗   ██╗/██████╗ ██╗   ██╗███╗   ██╗ ██████╗ ███████╗ ██████╗ ███╗   ██╗
  /██╔══██╗██╔════╝██║   ██║/██╔══██╗██║   ██║████╗  ██║██╔════╝ ██╔════╝██╔═══██╗████╗  ██║
  /██║  ██║█████╗  ██║   ██║/██║  ██║██║   ██║██╔██╗ ██║██║  ███╗█████╗  ██║   ██║██╔██╗ ██║
  /██║  ██║██╔══╝  ╚██╗ ██╔╝/██║  ██║██║   ██║██║╚██╗██║██║   ██║██╔══╝  ██║   ██║██║╚██╗██║
  /██████╔╝███████╗ ╚████╔╝ /██████╔╝╚██████╔╝██║ ╚████║╚██████╔╝███████╗╚██████╔╝██║ ╚████║
  /╚═════╝ ╚══════╝  ╚═══╝  /╚═════╝  ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═══╝

                         Navigate the filesystem. Survive.
```

> *You are a rogue process. The system is corrupted. Descend through the filesystem, battle daemons and zombie processes, and reach `/dev/null` to defeat the Kernel Panic.*

## The Concept

**/dev/dungeon** is a terminal-based roguelike that reimagines classic dungeon crawling through the lens of Unix systems. Every element draws from computing concepts:

```
┌─────────────────────────────────────────────────────────────┐
│  TRADITIONAL RPG          →      /dev/dungeon               │
├─────────────────────────────────────────────────────────────┤
│  Health Points (HP)       →      PID (Process ID)           │
│  Mana/Magic               →      MEM (Memory)               │
│  Attack Power             →      CPU (Processing Power)     │
│  Speed/Agility            →      NICE (Lower = Faster)      │
│  Access Level             →      UID (User Permissions)     │
├─────────────────────────────────────────────────────────────┤
│  Skeletons & Goblins      →      Zombies & Daemons          │
│  Healing Potions          →      PID Restore                │
│  Magic Scrolls            →      grep Scrolls               │
│  Legendary Sword          →      kill -9                    │
│  Dungeon Floors           →      /home → /tmp → /dev/null   │
└─────────────────────────────────────────────────────────────┘
```

## Gameplay

**Hybrid Exploration & Combat**
- Explore dungeons in real-time
- Combat is turn-based and menu-driven
- Die, unlock permanent upgrades, try again (roguelike meta-progression)

**The Descent**
```
/home      ──→  Tutorial, minor bugs
    │
/tmp       ──→  Orphan processes, chaos
    │
/var       ──→  Log daemons, cron jobs
    │
/etc       ──→  Config corruptions
    │
/usr       ──→  Privileged processes
    │
/sys       ──→  Kernel threads
    │
/dev       ──→  Device drivers, I/O errors
    │
/dev/null  ──→  FINAL BOSS: Kernel Panic
```

## Getting Started

### Prerequisites

- Go 1.21+ installed
- A terminal with Unicode support

### Installation

```bash
# Clone the repository
git clone https://github.com/iheanyi/dev-dungeon.git
cd dev-dungeon

# Build the game
go build -o devdungeon ./cmd/devdungeon/

# Run it
./devdungeon
```

Or run directly:

```bash
go run ./cmd/devdungeon/
```

### Running as SSH Server

```bash
# Start PostgreSQL (Docker)
docker run -d --name devdungeon-db \
  -e POSTGRES_PASSWORD=dev \
  -e POSTGRES_DB=devdungeon \
  -p 5432:5432 postgres:18

# Run as SSH server
export DATABASE_URL="postgres://postgres:dev@localhost:5432/devdungeon?sslmode=disable"
go run ./cmd/devdungeon/ --server --port 2222

# Connect from another terminal
ssh -p 2222 yourname@localhost
```

### Controls

```
┌──────────────────────────────────────┐
│  EXPLORATION                         │
├──────────────────────────────────────┤
│  W / ↑ / K     Move up               │
│  S / ↓ / J     Move down             │
│  A / ← / H     Move left             │
│  D / → / L     Move right            │
│  I             Open inventory        │
│  P / Esc       Pause menu            │
├──────────────────────────────────────┤
│  COMBAT                              │
├──────────────────────────────────────┤
│  1 / A         Attack                │
│  2 / H         Hack (use MEM)        │
│  3 / I         Use item              │
│  4 / F         Flee                  │
├──────────────────────────────────────┤
│  GENERAL                             │
├──────────────────────────────────────┤
│  Enter         Confirm               │
│  Esc           Back / Cancel         │
│  Q             Quit                  │
└──────────────────────────────────────┘
```

## Character Classes

Choose your process type:

| Class | Specialty | Starting Bonus |
|-------|-----------|----------------|
| `init` | Balanced | Jack of all trades |
| `cron` | Speed | Lower NICE (faster turns) |
| `bash` | Offense | Higher CPU (more damage) |
| `vim` | Abilities | Higher MEM (more skills) |
| `sudo` | Access | Higher UID (bypass barriers) |

## Enemies

```
 z - zombie process     Slow but swarms you
 d - daemon             Persistent, may respawn
 f - fork bomb          Multiplies each turn
 s - segfault           Erratic, unpredictable
 r - rootkit            Stealthy ambusher
 K - KERNEL PANIC       The final boss
```

## Items

```
 ! - sudo potion        Temporary invincibility
 ? - grep scroll        Reveal all items on floor
 + - PID restore        Heal your process
 * - MEM restore        Restore ability points
 ) - weapons            Increase CPU damage
 [ - armor              Increase max PID
```

## Meta-Progression

Death isn't the end—it's an upgrade opportunity.

Earn **exit codes** based on how far you descend. Spend them on:
- Unlocking new character classes
- Permanent stat bonuses
- New items in the loot pool
- Starting equipment

## Current Features

- [x] Core architecture & UI framework
- [x] Entity system (player, enemies, items)
- [x] Configuration system
- [x] Procedural dungeon generation (BSP algorithm)
- [x] Turn-based combat system with abilities
- [x] Enemy AI with varied behaviors
- [x] Field of view and fog of war
- [x] Equipment system (weapons, armor, accessories)
- [x] Skill system (process-themed abilities)
- [x] Meta-progression with exit codes
- [x] Save/load system
- [x] 8 floor types from `/home` to `/dev/null`
- [x] Boss fights (Kernel Panic final boss)

## Roadmap

- [x] SSH multiplayer via [Wish](https://github.com/charmbracelet/wish)
- [x] Web portal for registration and leaderboards
- [x] Unlockables shop (spend exit codes)
- [ ] Async drops (messages/items between players)
- [ ] Co-op dungeon runs
- [ ] Daily seeded runs

## Multiplayer

Connect via SSH and play from anywhere:

```bash
ssh player@dev-dungeon.com
```

Your SSH key is your identity. First connection prompts for username registration. Progress syncs across devices.

## Development

### Prerequisites

- Go 1.25+
- Node.js 22+
- PostgreSQL 18+ (or Docker)

### Quick Start

```bash
# Install development tools
brew install overmind air  # macOS
# or: go install github.com/DarthSim/overmind/v2@latest github.com/air-verse/air@latest

# Clone and setup
git clone https://github.com/iheanyi/dev-dungeon.git
cd dev-dungeon
go mod download
cd web && npm install && cd ..

# Start database with Docker
docker compose up -d postgres

# Create .env file
cat > .env << 'EOF'
DATABASE_URL=postgres://postgres:dev@localhost:5432/devdungeon?sslmode=disable
SSH_PORT=2222
HTTP_PORT=8080
EOF

# Run migrations
go run ./cmd/migrate up

# Start development (both services with live reload)
make dev
```

### Available Commands

```bash
make dev          # Start both server + frontend (requires overmind)
make dev-server   # Start Go server with Air live reload
make dev-web      # Start Vite frontend dev server
make build        # Build Go binary and web frontend
make test         # Run Go tests
make play         # Play the game locally (single player)
```

## Deployment

### Docker Compose (Recommended for Self-Hosting)

```bash
# Build and start all services
docker compose up -d

# View logs
docker compose logs -f devdungeon

# SSH to play
ssh -p 2222 yourname@localhost

# Web portal
open http://localhost:8080
```

### Production Deployment (Fly.io)

```bash
# Install Fly CLI
brew install flyctl

# Login and create app
fly auth login
fly apps create dev-dungeon

# Create PostgreSQL database
fly postgres create --name devdungeon-db
fly postgres attach devdungeon-db

# Set secrets
fly secrets set \
  DATABASE_URL="postgres://..."

# Create persistent volume for SSH host key
fly volumes create devdungeon_data --region ord --size 1

# Deploy
fly deploy

# SSH to your server
ssh player@dev-dungeon.com
```

### Custom Domain Setup (dev-dungeon.com)

```bash
# 1. Add SSL certificate for web
fly certs add dev-dungeon.com
fly certs add www.dev-dungeon.com

# 2. Get your Fly.io IP addresses
fly ips list
```

Then add these DNS records at your domain registrar:

| Type | Name | Value |
|------|------|-------|
| A | @ | `<fly-ipv4>` |
| AAAA | @ | `<fly-ipv6>` |
| CNAME | www | dev-dungeon.fly.dev |

**For SSH on the custom domain** (so users can `ssh player@dev-dungeon.com`):

The A/AAAA records above already point to your Fly app. Fly.io routes port 22 traffic to your SSH service automatically based on the `fly.toml` configuration.

Verify with:
```bash
# Test web
curl https://dev-dungeon.com/api/health

# Test SSH
ssh -v player@dev-dungeon.com
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `SSH_PORT` | SSH server port | `2222` |
| `HTTP_PORT` | HTTP server port | `8080` |
| `SSH_HOST_KEY_PATH` | Path to SSH host key | `.ssh/host_key` |
| `SENTRY_DSN` | Sentry DSN for error tracking | (optional) |
| `SENTRY_ENVIRONMENT` | Sentry environment name | `development` |

### Architecture

```
┌─────────────────────────────────────────────┐
│              Load Balancer                  │
│         (SSH port 22, HTTP port 443)        │
└─────────────────────────────────────────────┘
                    │
      ┌─────────────┴─────────────┐
      │                           │
┌─────────────┐           ┌─────────────┐
│   SSH :2222 │           │  HTTP :8080 │
│   (Wish)    │           │  (Web API)  │
└─────────────┘           └─────────────┘
      │                           │
      └─────────────┬─────────────┘
                    │
          ┌─────────────────┐
          │   PostgreSQL    │
          │  (accounts,     │
          │   saves,        │
          │   leaderboards, │
          │   sessions)     │
          └─────────────────┘
```

## Built With

- [Go](https://golang.org/) - Because it's fast and fun
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Wish](https://github.com/charmbracelet/wish) - SSH server for multiplayer
- [SvelteKit](https://kit.svelte.dev/) - Web portal

## License

MIT

---

```
    $ ./devdungeon

    ┌─────────────────────────────────────────┐
    │         You have been spawned.          │
    │                                         │
    │              PID: 1337                  │
    │           Status: RUNNING               │
    │                                         │
    │         Press any key to begin          │
    │              your descent.              │
    └─────────────────────────────────────────┘
```

*Good luck, process. You'll need it.*
