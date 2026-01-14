# Tasuku Learnings

_Auto-synced from .tasuku/context/learnings.md_

## Rules

- Always use *rand.Rand instances from seeds, never global rand, for reproducible procedural generation. Create RNG at generation entry point and pass to all helpers.
- When setting a flag on a struct that was appended to a slice, the flag must be set BEFORE the append, not after. Structs are copied by value on append, so mutations after append don't affect the slice copy.
- Always implement regression tests when fixing bugs - bugs without tests will regress
- Always use time.Now().UTC() when storing timestamps in PostgreSQL. Local time (e.g., CST) gets stored without timezone info, then retrieved as UTC, causing 6+ hour discrepancies. This breaks expiry checks where tokens appear expired immediately.

## Insights

- For Linux-themed games: RAM as health (OOM = death), FD as mana (limited resource), NICE as speed (accurate!), UID as power level (0 = root = god). These mappings are both thematic AND technically accurate.
- Equipment slots with the same EquipSlot enum value need auto-fill logic to find the next available slot. Without this, equipping two items of the same slot type will overwrite the first one instead of filling the second slot.

