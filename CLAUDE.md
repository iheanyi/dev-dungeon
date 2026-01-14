# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**/dev/dungeon** is a Unix/hacker-themed terminal roguelike built in Go with Bubble Tea. Players navigate procedurally generated filesystem dungeons (`/home` → `/dev/null`), battling rogue processes (zombie, daemon, fork_bomb, etc.) with Unix-themed stats (PID=health, CPU=attack, MEM=mana, NICE=speed, UID=permissions).

## Build & Run Commands

```bash
# Build
go build ./...

# Run the game
go run ./cmd/devdungeon/

# Run tests
go test ./...

# Run single test
go test -run TestName ./internal/package/
```

## Architecture

### Import Hierarchy (to avoid cycles)
```
types (shared types, no dependencies)
  ↓
entity (player, enemy, item - imports types)
  ↓
game (interfaces, engine - imports types only, defines Entity/Combatant interfaces)
  ↓
ui (Bubble Tea model - imports types, entity, config)
```

The `internal/types` package exists specifically to break import cycles. All shared types (Position, Stats, GameState, FloorType, etc.) live there.

### Key Packages

- **internal/types**: Shared types with zero dependencies
- **internal/game**: Core interfaces (Engine, World, CombatSystem, DungeonGenerator) - does NOT import entity to avoid cycles
- **internal/entity**: Player, Enemy, Item with template-based creation (`EnemyTemplates`, `ItemTemplates` maps)
- **internal/ui**: Bubble Tea model with ViewType-based rendering (MainMenu, Game, Combat, Inventory, Pause, GameOver, Victory)
- **internal/config**: JSON config loading with defaults

### Entity System

Entities use a base struct pattern:
```go
type BaseEntity struct { id, name, position, glyph, blocking }
type Player struct { *BaseEntity, Stats, Inventory, Equipment, Skills... }
type Enemy struct { *BaseEntity, Stats, Behavior, LootTable... }
```

Templates define enemy/item types in maps (`EnemyTemplates`, `ItemTemplates`), and factory functions create instances with scaling.

### Game State Machine

The UI manages state via `ViewType` (which screen is shown) and `types.GameState` (game logic state). Key transitions:
- MainMenu → Game (New Game)
- Game ↔ Combat (enemy encounter)
- Game ↔ Inventory (toggle)
- Game/Combat → GameOver (death)

## Key Interfaces

All major systems have interfaces for testability:
- `game.Engine`: State management, game loop, combat initiation
- `game.World`: Floor/entity management, visibility
- `game.CombatSystem`: Turn-based combat lifecycle
- `game.DungeonGenerator`: Procedural floor generation
- `game.Renderer`: Render game state to string (for testing)

## Task Tracking

This project uses Tasuku for task management. View tasks with `tk task list`.
