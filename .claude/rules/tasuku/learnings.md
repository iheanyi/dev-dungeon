# Tasuku Learnings

_Auto-synced from .tasuku/context/learnings.md_

## Rules

- Always use *rand.Rand instances from seeds, never global rand, for reproducible procedural generation. Create RNG at generation entry point and pass to all helpers.

