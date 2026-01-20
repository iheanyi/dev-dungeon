---
status: ready
priority: 3
tags: [refactor, tech-debt]
created_at: 2026-01-20T17:29:21.365896Z
updated_at: 2026-01-20T17:29:31.901983Z
---

# Refactor app.go (4,225 lines) into modular components by extracting view...

Refactor app.go (4,225 lines) into modular components by extracting view renderers, update handlers, and domain logic into separate files

## Notes

### 2026-01-20T17:29:31Z [ce729b]
Suggested split based on function analysis (125 functions total):

1. **views.go** (~1500 lines) - All `view*()` rendering functions
   - viewMainMenu, viewGame, viewCombat, viewInventory, viewPause, etc.
   - renderStats, renderMap, renderLog, renderTile

2. **updates.go** (~800 lines) - All `update*()` state handlers
   - updateMainMenu, updateGame, updateCombat, updateInventory, etc.

3. **combat.go** (~300 lines) - Combat-specific logic
   - startCombat, endCombat, executeCombatAction, executeSkill
   - getValidTargetIndex, cycleTarget

4. **inventory.go** (~300 lines) - Inventory/equipment management
   - useOrEquipItem, useConsumable, equipItem, dropItem, unequipItem

5. **shop.go** (~400 lines) - Shop and unlock shop logic
   - openShop, updateShop, buyItem, viewShop
   - openUnlockShop, purchaseUnlock, viewUnlockShop

6. **leaderboard.go** (~300 lines) - Leaderboard functionality
   - openLeaderboard, updateLeaderboard, viewLeaderboard
   - showDailyLeaderboard, refreshDailyLeaderboard

7. **styles.go** (~100 lines) - Already partially separated, move Styles struct

8. **model.go** (~500 lines) - Core Model struct, Init, Update router, setters

