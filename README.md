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
git clone https://github.com/iheanyi/devdungeon.git
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

## Roadmap

- [x] Core architecture & UI framework
- [x] Entity system (player, enemies, items)
- [x] Configuration system
- [ ] Procedural dungeon generation
- [ ] Turn-based combat system
- [ ] Enemy AI behaviors
- [ ] Meta-progression & save system
- [ ] SSH multiplayer via [Wish](https://github.com/charmbracelet/wish)

## Built With

- [Go](https://golang.org/) - Because it's fast and fun
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

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
