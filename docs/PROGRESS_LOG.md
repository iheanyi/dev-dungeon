# /dev/dungeon Progress Log

A casual log of development progress, wins, and vibes.

---

## Session 1 - Project Genesis

**Date:** 2024

### What We Did

1. **Named the game** - `/dev/dungeon` (Unix-themed dungeon crawler, let's go)

2. **Gathered requirements** via structured questions:
   - Hybrid gameplay (real-time exploration, turn-based combat)
   - TUI widgets + ASCII visuals
   - Unix/hacker theme (daemons, zombies, grep scrolls, kill -9 weapons)
   - Meta-progression like Hades
   - Future SSH multiplayer via Wish

3. **Scaffolded the project**:
   - Go + Bubble Tea + Lip Gloss
   - Clean architecture with interfaces for testability
   - `internal/types` package to break import cycles (big brain move)

4. **Created core entity system**:
   - Player with 5 classes (init, cron, bash, vim, sudo)
   - 6 enemy types (zombie, daemon, fork_bomb, segfault, rootkit, kernel_panic)
   - 13 items with Unix-themed effects

5. **Built the docs**:
   - `docs/GAME_DESIGN.md` - Full GDD
   - `README.md` with sick ASCII art
   - `CLAUDE.md` for future AI assistants

### Commits
- `feat(types): add Floor, Tile, Room types for dungeon generation`
- `feat(dungeon): implement BSP procedural dungeon generator`
- `feat(game): implement core game engine with world management`

---

## Session 2 - Parallel Power

### What We Did

1. **Added shared types** for dungeon generation:
   - `Tile` with visibility/explored/blocked properties
   - `Room` with intersection detection
   - `Floor` with full tile grid

2. **Parallelized implementation** (two agents cooking simultaneously):

   **Agent 1 - Dungeon Generator:**
   - BSP (Binary Space Partitioning) algorithm
   - Seeded RNG for 100% reproducible dungeons
   - L-shaped corridor generation
   - Entity population scaling with depth
   - Floor-specific variations (/home = tutorial, /dev/null = boss arena)

   **Agent 2 - Game Engine:**
   - World management with entity tracking
   - Player movement + collision detection
   - Bump combat (walk into enemy = fight)
   - Floor transitions with state caching
   - Circular FOV with line-of-sight
   - Deterministic seed derivation (master -> floor -> entity)

3. **Recorded architectural decisions**:
   - Seeded RNG with FNV hash derivation
   - BSP over cellular automata for dungeon gen

### Key Insight
> "Always use `*rand.Rand` instances from seeds, never global rand, for reproducible procedural generation."

### Stats
- Dungeon generator: ~36 microseconds per floor
- Floor population: ~18 microseconds
- 17 engine tests passing
- 8 dungeon tests passing

---

## Next Up

- [ ] Wire up UI to game engine
- [ ] Implement turn-based combat system
- [ ] Add enemy AI behaviors
- [ ] Meta-progression and save system
- [ ] SSH multiplayer (the dream)

---

*"Navigate the filesystem. Survive."*
