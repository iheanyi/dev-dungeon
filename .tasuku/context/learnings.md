# Learnings

## 83355d - 2026-01-14T01:01:47Z
Always use *rand.Rand instances from seeds, never global rand, for reproducible procedural generation. Create RNG at generation entry point and pass to all helpers.

## 6555b8 - 2026-01-14T01:18:30Z
For Linux-themed games: RAM as health (OOM = death), FD as mana (limited resource), NICE as speed (accurate!), UID as power level (0 = root = god). These mappings are both thematic AND technically accurate.

## 1940fc - 2026-01-14T02:44:34Z
When setting a flag on a struct that was appended to a slice, the flag must be set BEFORE the append, not after. Structs are copied by value on append, so mutations after append don't affect the slice copy.

## ae99d3 - 2026-01-14T03:01:32Z
Equipment slots with the same EquipSlot enum value need auto-fill logic to find the next available slot. Without this, equipping two items of the same slot type will overwrite the first one instead of filling the second slot.

## 9f3ff4 - 2026-01-14T21:52:19Z
Always implement regression tests when fixing bugs - bugs without tests will regress

## bb9090 - 2026-01-14T23:11:01Z
Always use time.Now().UTC() when storing timestamps in PostgreSQL. Local time (e.g., CST) gets stored without timezone info, then retrieved as UTC, causing 6+ hour discrepancies. This breaks expiry checks where tokens appear expired immediately.

## 84c08b - 2026-01-14T23:35:20Z
Go 1.23 doesn't have the covdata tool required for -covermode=atomic. Use default coverage mode or upgrade to Go 1.24+ for atomic coverage in CI.

## 4e3b27 - 2026-01-16T02:40:06Z
When returning a new model from Bubble Tea's Update(), you must explicitly call Init() and return its result. Bubble Tea does NOT automatically call Init() on replacement models - only on the initial model. Failing to do this causes goroutine/resource leaks if Init() sets up background tasks.

## 4ebf4d - 2026-01-16T03:04:57Z
Always validate loaded player positions against current dungeon layout. Dungeon regeneration from seeds can produce different layouts when dimensions change, so saved positions may land in walls. Fallback to a known-safe position (like stairs) when loaded position is invalid.

## 604540 - 2026-01-17T00:48:33Z
Always use TemplateID (not ID()) when saving/loading items or comparing items for stacking. ID() returns unique instance identifiers that differ between item instances, while TemplateID identifies the item type for recreation and stacking.

## 8cf7aa - 2026-01-17T02:11:21Z
When modifying Stats in tests (e.g., Stats.RAM = 150), must also update MaxStats to allow the value. LoadGame validation caps stats to MaxStats bounds, so setting RAM higher than MaxRAM will silently clamp the value on load.

## ec9747 - 2026-01-17T02:22:33Z
When using pendingSave pattern for multiplayer saves, always update pendingSave after successful save callbacks. Otherwise Continue will load stale session-start data instead of the latest saved state. Get save data BEFORE calling save callback (engine may be destroyed after), then store it on success.

