---
status: ready
priority: 3
created_at: 2026-01-14T00:36:11.326932Z
updated_at: 2026-01-14T03:14:14.175856Z
---

# Implement unlockables system - class unlock checks, unlock shop UI, permanent...

Implement unlockables system - class unlock checks, unlock shop UI, permanent stat bonuses, unlockable items in loot pool. Data structures exist in save/types.go but aren't wired up.

## Notes

### 2026-01-14T03:14:14Z [c123c4]
Current state: MetaProgress struct exists with UnlockedClasses, UnlockedItems, PermanentBonuses fields. LoadMetaProgress/SaveMetaProgress methods exist. But nothing calls them - class select doesn't check unlocks, no unlock shop UI, no way to spend exit codes on upgrades.

