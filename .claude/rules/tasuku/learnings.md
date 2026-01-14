# Tasuku Learnings

_Auto-synced from .tasuku/context/learnings.md_

## Rules

- Always use *rand.Rand instances from seeds, never global rand, for reproducible procedural generation. Create RNG at generation entry point and pass to all helpers.
- When setting a flag on a struct that was appended to a slice, the flag must be set BEFORE the append, not after. Structs are copied by value on append, so mutations after append don't affect the slice copy.

## Insights

- For Linux-themed games: RAM as health (OOM = death), FD as mana (limited resource), NICE as speed (accurate!), UID as power level (0 = root = god). These mappings are both thematic AND technically accurate.

