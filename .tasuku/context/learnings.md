# Learnings

## 83355d - 2026-01-14T01:01:47Z
Always use *rand.Rand instances from seeds, never global rand, for reproducible procedural generation. Create RNG at generation entry point and pass to all helpers.

