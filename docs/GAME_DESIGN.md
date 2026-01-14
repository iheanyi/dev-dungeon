# /dev/dungeon - Game Design Document

## Overview

**/dev/dungeon** is a Unix/hacker-themed terminal roguelike dungeon crawler built in Go with Bubble Tea. Players navigate procedurally generated filesystem dungeons, battling rogue processes and daemons while collecting permissions and upgrading their system access.

## Core Gameplay Loop

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│   ┌──────────┐    ┌──────────┐    ┌──────────┐            │
│   │  EXPLORE │───▶│ ENCOUNTER│───▶│  COMBAT  │            │
│   │ (real-   │    │  enemy?  │    │  (turn-  │            │
│   │  time)   │    │          │    │  based)  │            │
│   └────┬─────┘    └────┬─────┘    └────┬─────┘            │
│        │               │               │                   │
│        │               │ no            │                   │
│        │               ▼               │                   │
│        │         ┌──────────┐          │                   │
│        │         │  LOOT /  │◀─────────┘                   │
│        │         │  ITEMS   │   victory                    │
│        │         └────┬─────┘                              │
│        │              │                                    │
│        │              ▼                                    │
│        │         ┌──────────┐                              │
│        │         │ DESCEND  │                              │
│        │         │  FLOOR?  │                              │
│        │         └────┬─────┘                              │
│        │              │                                    │
│        └──────────────┘                                    │
│                                                             │
│   ON DEATH:                                                │
│   ┌──────────┐    ┌──────────┐    ┌──────────┐            │
│   │   DIE    │───▶│  UNLOCK  │───▶│ NEW RUN  │            │
│   │          │    │  META    │    │          │            │
│   └──────────┘    └──────────┘    └──────────┘            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Theme & Flavor

### Setting
The player is a rogue process navigating the depths of a corrupted Unix filesystem. Starting at `/home`, they must descend through increasingly dangerous directories to reach `/dev/null` and defeat the Kernel Panic.

### Dungeon Floors (Zones)
| Floor | Theme | Difficulty | Enemies |
|-------|-------|------------|---------|
| `/home` | Tutorial/Safe | Easy | Minor bugs, lint errors |
| `/tmp` | Temporary chaos | Easy-Medium | Orphan processes, temp files |
| `/var` | Variable dangers | Medium | Log daemons, cron jobs |
| `/etc` | Configuration maze | Medium-Hard | Config corruptions, permission errors |
| `/usr` | User space | Hard | Privileged processes, sudo attempts |
| `/sys` | System level | Very Hard | Kernel threads, interrupts |
| `/dev` | Device layer | Brutal | Device drivers, I/O errors |
| `/dev/null` | Final boss | Boss | The Kernel Panic |

### Enemy Types
| Enemy | HP | Behavior | Drops |
|-------|-----|----------|-------|
| `zombie` | Low | Slow, swarm | Memory fragments |
| `daemon` | Medium | Persistent, respawns | Service tokens |
| `fork_bomb` | Low | Multiplies each turn | CPU cycles |
| `segfault` | Medium | Erratic movement | Core dumps |
| `rootkit` | High | Stealth, ambush | Root shards |
| `kernel_panic` | Boss | Multi-phase | Victory |

### Items & Equipment
| Item | Type | Effect |
|------|------|--------|
| `sudo` potion | Consumable | Temporary invincibility |
| `grep` scroll | Consumable | Reveal all items on floor |
| `kill -9` | Weapon | Instant kill on low HP enemies |
| `chmod +x` | Armor | Increase execution permissions |
| `man page` | Skill unlock | Learn new ability |
| `pipe \|` | Utility | Chain attacks |
| `&&` operator | Utility | Combo attacks |

### Stats
- **PID** (Process ID): Health/HP
- **CPU**: Attack power
- **MEM**: Ability capacity / mana
- **NICE**: Speed/priority (lower = faster)
- **UID**: Permission level (unlocks areas)

## Systems

### Combat (Turn-Based Menu)
```
┌─────────────────────────────────────┐
│ COMBAT: zombie (PID: 45/100)        │
├─────────────────────────────────────┤
│                                     │
│  [A] Attack      - 15 CPU damage    │
│  [H] Hack        - 20 MEM cost      │
│  [I] Inventory   - Use item         │
│  [F] Flee        - NICE check       │
│                                     │
│  Your PID: 80/100  MEM: 50/50       │
└─────────────────────────────────────┘
```

### Character Classes (Unlockable)
| Class | Starting Stats | Playstyle |
|-------|---------------|-----------|
| `init` | Balanced | Default starter |
| `cron` | High NICE | Scheduled abilities, time-based |
| `bash` | High CPU | Aggressive, script combos |
| `vim` | High MEM | Complex abilities, modal |
| `sudo` | High UID | Access shortcuts, permission focus |

### Meta-Progression
Earned currency: **`exit codes`** (gained on death based on progress)

**Unlockables:**
- New character classes
- Permanent stat bonuses (+5 base PID, etc.)
- New items added to loot pool
- Starting equipment options
- Lore/man pages

### Inventory
- Limited slots (expandable via meta)
- Equipment slots: Weapon, Armor, Utility x2
- Consumables stack

## Technical Architecture

### Design Principles
1. **Testable**: Interfaces for all major systems, dependency injection
2. **Extensible**: Entity-component style for game objects, data-driven content
3. **Configurable**: External config files, easy balancing tweaks
4. **Modular**: Clear separation between game logic and presentation

### Package Structure
```
/dev/dungeon/
├── cmd/
│   └── devdungeon/      # Main entry point
├── internal/
│   ├── game/            # Core game logic
│   │   ├── engine.go    # Game loop, state management
│   │   ├── world.go     # World/level management
│   │   └── combat.go    # Combat system
│   ├── entity/          # Game entities
│   │   ├── player.go
│   │   ├── enemy.go
│   │   └── item.go
│   ├── dungeon/         # Procedural generation
│   │   ├── generator.go
│   │   ├── room.go
│   │   └── floor.go
│   ├── ui/              # Bubble Tea UI components
│   │   ├── app.go       # Main Bubble Tea model
│   │   ├── views/       # Different screens
│   │   └── components/  # Reusable UI pieces
│   ├── config/          # Configuration loading
│   └── save/            # Save/load, meta progression
├── assets/
│   └── data/            # JSON/YAML game data
├── docs/
│   └── GAME_DESIGN.md   # This file
├── go.mod
└── go.sum
```

### Key Interfaces
```go
// Testable combat system
type CombatSystem interface {
    StartCombat(player *Player, enemies []*Enemy) *Combat
    ExecuteAction(combat *Combat, action Action) Result
    EndCombat(combat *Combat) Rewards
}

// Pluggable dungeon generation
type DungeonGenerator interface {
    Generate(floor FloorType, seed int64) *Dungeon
}

// Swappable renderer for testing
type Renderer interface {
    Render(state *GameState) string
}
```

## Future Considerations

### SSH Multiplayer (via Wish)
- Leaderboards
- Ghost runs (see other players' paths)
- Shared message system
- Potential co-op

### Content Expansion
- More floor types
- Boss variants
- Challenge modes
- Daily runs with fixed seeds

---

## Open Questions / Revisit Points

- [ ] Should exploration be grid-based or room-based navigation?
- [ ] Exact balance numbers for stats/damage
- [ ] How many floors per zone?
- [ ] Inventory slot count
- [ ] Meta-progression curve

---

*Document Version: 0.1*
*Last Updated: 2024-01-13*
