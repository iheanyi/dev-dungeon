# Learnings

## 83355d - 2026-01-14T01:01:47Z
Always use *rand.Rand instances from seeds, never global rand, for reproducible procedural generation. Create RNG at generation entry point and pass to all helpers.

## 6555b8 - 2026-01-14T01:18:30Z
For Linux-themed games: RAM as health (OOM = death), FD as mana (limited resource), NICE as speed (accurate!), UID as power level (0 = root = god). These mappings are both thematic AND technically accurate.

