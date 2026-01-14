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
cd devdungeon

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

- [ ] SSH multiplayer via [Wish](https://github.com/charmbracelet/wish) *(in progress)*
- [ ] Leaderboards and async drops
- [ ] Co-op dungeon runs
- [ ] Unlockables shop (spend exit codes)

## Multiplayer (Coming Soon)

Connect via SSH and play from anywhere:

```bash
ssh player@devdungeon.io
```

Your SSH key is your identity. First connection prompts for username registration. Progress syncs across devices.

## Built With

- [Go](https://golang.org/) - Because it's fast and fun
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Wish](https://github.com/charmbracelet/wish) - SSH server for multiplayer

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
