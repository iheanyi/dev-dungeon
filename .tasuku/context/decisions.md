# Decisions

## seeded-rng-architecture - 2026-01-14T01:01:46Z
**Chose**: Deterministic seed derivation with FNV hash: masterSeed -> floorSeed -> entitySeed
**Over**: Global rand with SetSeed calls, Passing RNG instances everywhere, Time-based seeds only
**Because**: Enables reproducible dungeons for testing/debugging, allows players to share seeds for cool runs, and derived seeds mean changing floor 1 generation won't affect floor 5

## bsp-dungeon-generation - 2026-01-14T01:01:46Z
**Chose**: Binary Space Partitioning (BSP) for dungeon generation
**Over**: Cellular automata, Drunkard's walk, Simple random room placement
**Because**: BSP guarantees non-overlapping rooms, ensures connectivity through tree structure, and produces classic roguelike layouts. Can still vary room shapes within cells for different floor types.

## save-system-strategy - 2026-01-14T01:06:29Z
**Chose**: Hybrid save system: checkpoint on floor transition + background auto-save every 60s
**Over**: Save only on floor transition, Continuous goroutine save, Manual save only
**Because**: Floor transitions are natural checkpoints for roguelikes, but background auto-save prevents frustrating crash losses. Only save deltas (dead enemies, looted items) since floors regenerate from seed.

## stat-system-redesign - 2026-01-14T01:13:13Z
**Chose**: RAM (health), CPU (attack), FD (ability resource), NICE (speed), UID (access) - Linux-accurate stats
**Over**: Original PID/CPU/MEM/NICE/UID, Full terminal ps aux style, Traditional HP/ATK/MP/SPD
**Because**: RAM as health is Linux-accurate (OOM = death), FD as mana reflects real process limits, keeps NICE/UID which were already perfect. Balances accuracy with accessibility.

## multiplayer-architecture - 2026-01-14T01:46:05Z
**Chose**: Wish (SSH-based) for authentic terminal multiplayer
**Over**: WebSocket server + web client, P2P networking, REST API with polling
**Because**: Wish is Charm's SSH library - players SSH into the server and get a terminal session. Fits the Unix theme perfectly (ssh user@devdungeon.io), handles auth via SSH keys, and Bubble Tea works natively. Storage: PostgreSQL for accounts/leaderboards, Redis for real-time state. Co-op could be async (leave items/messages) or sync (shared dungeon, turn order by NICE stat). Seed sharing for competitive runs is easy - same seed = same dungeon.

## group-spawn-design - 2026-01-14T20:32:53Z
**Chose**: Template-based MinSpawn/MaxSpawn fields on EnemyTemplate
**Over**: Combat-time spawning (fork bomb style), Floor generation weighted spawning, Separate GroupedEnemy entity type
**Because**: Keeps spawning logic deterministic with seeded RNG, requires minimal code changes, each enemy type can have unique group behavior, and data-driven design makes balance tuning easy. Breaking change is acceptable since game isn't live yet.

## test-db-layer - 2026-01-14T21:52:56Z
**Chose**: Repository pattern with in-memory implementation for unit tests
**Over**: E2E tests with real database, Mocking with testify/mock, Test containers with Dockerized PostgreSQL
**Because**: In-memory repositories keep tests fast, isolated, and don't require external dependencies. The repository interface also improves code organization and makes the database layer more testable.

## cursor-based-leaderboard-pagination - 2026-01-15T22:14:14Z
**Chose**: Cursor-based pagination using (score, id) tuple for leaderboards
**Over**: Offset/limit pagination, Page number pagination, Keyset pagination on ID only
**Because**: Leaderboards can have concurrent score submissions while users browse. Offset/limit would cause duplicate entries or missed entries when new scores are inserted. The (score, id) cursor is stable because scores are immutable once submitted, and ID breaks ties for equal scores.

