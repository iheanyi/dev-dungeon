# Save/Load State Bug Audit Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix state persistence bugs identified in the save/load system audit

**Architecture:** The save system uses delta-based floor persistence (only changes are saved, floors regenerate from seed). Key data flows: GameWorld tracks removed entities -> buildFloorStates() collects deltas -> save file stores deltas -> LoadGame() regenerates floor + applies deltas via applyFloorState().

**Tech Stack:** Go, JSON file-based saves, seeded RNG for reproducible floor generation

---

## Bug Summary

### Bug 1: Floor delta tracking not restored on load (CRITICAL)
**Location:** `internal/game/engine.go:LoadGame()` (lines 993-1004)
**Impact:** When loading a save, if you descend/ascend and then save again, floor deltas from previous floors are lost because `world.RemovedEnemies`/`world.RemovedItems` maps are never populated from the save data.

**Scenario:**
1. Play floor 1, kill enemies, they're tracked in `RemovedEnemies[1]`
2. Save on floor 1
3. Load the save - `applyFloorState()` removes dead enemies from world, but tracking maps are empty
4. Descend to floor 2, then save
5. `buildFloorStates()` checks `world.GetRemovedEnemies(1)` - returns nil
6. Floor 1 deltas are lost from the save file
7. Next load: dead enemies respawn on floor 1

**Root cause:** `SetRemovedEntities()` exists but is never called during `LoadGame()`.

### Bug 2: Test incorrectly sets RAM without MaxRAM (MINOR - Test Bug)
**Location:** `internal/game/engine_test.go:TestSaveLoadE2E_FileSystem` (line 1094)
**Impact:** Test sets `RAM = 150` but MaxRAM is still 100, so load clamping makes test fail.

### Non-Bug: StatBonuses naming (COSMETIC)
**Location:** `internal/save/types.go:StatBonuses` uses `PID/MEM` while game uses `RAM/FD`
**Impact:** Confusing but works correctly via mapping in UI. Low priority cleanup.

---

## Task 1: Fix floor delta tracking restoration

**Files:**
- Modify: `internal/game/engine.go:LoadGame()` (around line 1004)
- Test: `internal/game/engine_test.go`

**Step 1: Write the failing test**

Create a test that reproduces the bug: load a save, descend, re-save, and verify floor 1 deltas persist.

Add this test to `internal/game/engine_test.go`:

```go
// TestFloorDeltasPreservedAfterLoadAndFloorChange tests that floor deltas from the save
// are preserved in tracking maps after load, so they persist through floor changes.
// Regression test for: floor deltas lost when saving after descending post-load.
func TestFloorDeltasPreservedAfterLoadAndFloorChange(t *testing.T) {
	tempDir := t.TempDir()
	saveCfg := save.Config{
		SaveDir:          tempDir,
		AutoSaveInterval: 1 * time.Second,
		MinSaveInterval:  0,
	}
	saveMgr, err := save.NewManager(saveCfg)
	if err != nil {
		t.Fatalf("failed to create save manager: %v", err)
	}
	saveMgr.Start()
	defer saveMgr.Stop()

	cfg := config.DefaultConfig()

	// Create initial engine to get enemy ID from seeded floor
	seedEngine := NewEngine(cfg, 77777)
	if err := seedEngine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}
	if len(seedEngine.world.Enemies) == 0 {
		t.Skip("no enemies on floor 1")
	}
	deadEnemyID := seedEngine.world.Enemies[0].ID()

	// Create save with floor 1 dead enemy delta
	saveData := &save.SaveData{
		Version:      save.Version,
		MasterSeed:   77777,
		CurrentDepth: 1,
		Player: save.PlayerData{
			Class:     entity.ClassInit,
			Level:     1,
			XP:        0,
			XPToLevel: 100,
			Stats:     seedEngine.Player().Stats,
			MaxStats:  seedEngine.Player().MaxStats,
			Position:  seedEngine.Player().Position(),
			Inventory: []save.ItemData{},
			Equipment: save.EquipmentData{},
		},
		FloorStates: []save.FloorState{
			{
				Depth:         1,
				ExploredTiles: []types.Position{},
				DeadEnemies:   []string{deadEnemyID},
				LootedItems:   []string{},
			},
		},
	}
	if err := saveMgr.SaveSync(saveData, save.TriggerManual); err != nil {
		t.Fatalf("SaveSync failed: %v", err)
	}

	// Load the save
	engine := NewEngine(cfg, 77777)
	engine.saveManager = saveMgr
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}
	if err := engine.LoadLatestSave(); err != nil {
		t.Fatalf("LoadLatestSave failed: %v", err)
	}

	// Verify floor 1 deltas are in tracking maps (this is the bug - they won't be)
	removedEnemies := engine.world.GetRemovedEnemies(1)
	if len(removedEnemies) == 0 {
		t.Error("BUG: floor 1 dead enemies not restored to tracking maps after load")
	}

	// Now simulate: descend to floor 2
	// Move player to stairs down first
	engine.player.SetPosition(engine.world.CurrentFloor.StairsDown)
	if err := engine.DescendStairs(); err != nil {
		t.Fatalf("DescendStairs failed: %v", err)
	}

	// Save on floor 2
	if err := engine.SaveSync(save.TriggerManual); err != nil {
		t.Fatalf("SaveSync on floor 2 failed: %v", err)
	}

	// Load the floor 2 save to check if floor 1 deltas were preserved
	data, err := saveMgr.LoadLatest()
	if err != nil {
		t.Fatalf("LoadLatest failed: %v", err)
	}

	// Check floor states include floor 1 deltas
	foundFloor1Deltas := false
	for _, fs := range data.FloorStates {
		if fs.Depth == 1 && len(fs.DeadEnemies) > 0 {
			foundFloor1Deltas = true
			break
		}
	}
	if !foundFloor1Deltas {
		t.Error("BUG: floor 1 deltas lost after descending and re-saving")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestFloorDeltasPreservedAfterLoadAndFloorChange ./internal/game/`
Expected: FAIL with "BUG: floor 1 dead enemies not restored to tracking maps after load"

**Step 3: Implement the fix**

In `internal/game/engine.go`, after applying floor state deltas (around line 1004), add code to restore the tracking maps for ALL floors in the save:

```go
// Apply floor state deltas (dead enemies, looted items, explored tiles)
for _, floorState := range data.FloorStates {
	if floorState.Depth == loadDepth {
		e.applyFloorState(&floorState)
	}
	// Restore removal tracking for ALL floors (not just current)
	// This ensures floor deltas persist if we descend/ascend and re-save
	e.world.SetRemovedEntities(floorState.Depth, floorState.DeadEnemies, floorState.LootedItems)
}
```

This changes the existing code from:
```go
for _, floorState := range data.FloorStates {
	if floorState.Depth == loadDepth {
		e.applyFloorState(&floorState)
		break
	}
}
```

To:
```go
for _, floorState := range data.FloorStates {
	if floorState.Depth == loadDepth {
		e.applyFloorState(&floorState)
	}
	// Restore removal tracking for ALL floors (not just current)
	// This ensures floor deltas persist if we descend/ascend and re-save
	e.world.SetRemovedEntities(floorState.Depth, floorState.DeadEnemies, floorState.LootedItems)
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestFloorDeltasPreservedAfterLoadAndFloorChange ./internal/game/`
Expected: PASS

**Step 5: Run all save/load tests to verify no regressions**

Run: `go test -v -run "Save|Load|Floor" ./internal/game/`
Expected: All pass except TestSaveLoadE2E_FileSystem (known test bug)

**Step 6: Commit**

```bash
git add internal/game/engine.go internal/game/engine_test.go
git commit -m "$(cat <<'EOF'
fix: restore floor delta tracking on save load

When loading a save, the removal tracking maps (RemovedEnemies/RemovedItems)
were not being populated from the save data. This caused floor deltas to be
lost if the player descended/ascended and then saved again.

The fix calls SetRemovedEntities() for ALL floor states in the save, not just
the current floor. This ensures that if you load a save on floor 1, descend
to floor 2, and save, the floor 1 deltas are still tracked and will be saved.

Includes regression test: TestFloorDeltasPreservedAfterLoadAndFloorChange

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Fix TestSaveLoadE2E_FileSystem test bug

**Files:**
- Modify: `internal/game/engine_test.go:TestSaveLoadE2E_FileSystem` (line 1094)

**Step 1: Write the failing test (already failing)**

The test is already failing. Skip this step.

**Step 2: Understand the bug**

The test sets `Stats.RAM = 150` but doesn't set `MaxStats.MaxRAM = 150`. On load, the validation code bounds RAM to MaxRAM (default 100 for bash class).

**Step 3: Fix the test**

Change line 1094 from just setting RAM to also setting MaxRAM:

```go
// Set player to known state
engine.Player().Level = 7
engine.Player().XP = 350
engine.Player().Stats.RAM = 150
engine.Player().MaxStats.MaxRAM = 150  // Must set MaxRAM too, otherwise load will clamp
engine.Player().Stats.CPU = 25
engine.Player().ExitCodes = 42
engine.Player().SetPosition(types.Position{X: 10, Y: 15})
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestSaveLoadE2E_FileSystem ./internal/game/`
Expected: PASS

**Step 5: Run all tests to verify no regressions**

Run: `go test ./internal/game/...`
Expected: All pass

**Step 6: Commit**

```bash
git add internal/game/engine_test.go
git commit -m "$(cat <<'EOF'
fix: set MaxRAM in TestSaveLoadE2E_FileSystem

The test was setting RAM=150 but not MaxRAM, so the load validation
correctly clamped RAM to the default MaxRAM=100. Fixed by also setting
MaxRAM=150 to match the desired test state.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Run full test suite and verify all passes

**Files:**
- None (verification only)

**Step 1: Run full game package tests**

Run: `go test -v ./internal/game/...`
Expected: All tests pass

**Step 2: Run full save package tests**

Run: `go test -v ./internal/save/...`
Expected: All tests pass

**Step 3: Run full project tests**

Run: `go test ./...`
Expected: All tests pass

**Step 4: Build the game**

Run: `go build ./cmd/devdungeon/`
Expected: Build succeeds

---

## Summary of Changes

1. **Bug Fix (Critical):** Floor delta tracking now restored on load via `SetRemovedEntities()` for all saved floor states
2. **Test Fix:** `TestSaveLoadE2E_FileSystem` now sets MaxRAM when testing RAM persistence

## Testing Checklist

- [ ] New regression test passes: `TestFloorDeltasPreservedAfterLoadAndFloorChange`
- [ ] Fixed test passes: `TestSaveLoadE2E_FileSystem`
- [ ] All existing save/load tests pass
- [ ] Full test suite passes
- [ ] Game builds successfully
- [ ] Manual verification: kill enemy on floor 1, save, load, descend to floor 2, save, load - enemy on floor 1 should still be dead
