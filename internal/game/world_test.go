package game

import (
	"testing"

	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/types"
)

// === Spatial Indexing Tests ===
// These tests verify the O(1) spatial indexing for enemy and item lookups.

func TestSpatialIndex_GetEnemyAt(t *testing.T) {
	world := NewGameWorld()

	// Create a simple floor
	floor := &types.Floor{
		Width:  20,
		Height: 20,
		Tiles:  make([][]types.Tile, 20),
	}
	for y := 0; y < 20; y++ {
		floor.Tiles[y] = make([]types.Tile, 20)
		for x := 0; x < 20; x++ {
			floor.Tiles[y][x] = types.Tile{Type: types.TileFloor}
		}
	}
	floor.Depth = 1

	// Create enemies at known positions
	enemy1 := entity.NewEnemy("zombie", "enemy1", types.Position{X: 5, Y: 5}, 1)
	enemy2 := entity.NewEnemy("daemon", "enemy2", types.Position{X: 10, Y: 10}, 1)

	// Set floor with enemies - this should build spatial index
	world.SetFloor(floor, []*entity.Enemy{enemy1, enemy2}, nil)

	// Test GetEnemyAt returns correct enemies (O(1) lookup)
	if e := world.GetEnemyAt(types.Position{X: 5, Y: 5}); e == nil {
		t.Error("expected enemy1 at (5,5)")
	} else if e.ID() != "enemy1" {
		t.Errorf("expected enemy1, got %s", e.ID())
	}

	if e := world.GetEnemyAt(types.Position{X: 10, Y: 10}); e == nil {
		t.Error("expected enemy2 at (10,10)")
	} else if e.ID() != "enemy2" {
		t.Errorf("expected enemy2, got %s", e.ID())
	}

	// Test empty position returns nil
	if e := world.GetEnemyAt(types.Position{X: 15, Y: 15}); e != nil {
		t.Error("expected nil at empty position")
	}
}

func TestSpatialIndex_GetItemAt(t *testing.T) {
	world := NewGameWorld()

	// Create a simple floor
	floor := &types.Floor{
		Width:  20,
		Height: 20,
		Tiles:  make([][]types.Tile, 20),
	}
	for y := 0; y < 20; y++ {
		floor.Tiles[y] = make([]types.Tile, 20)
		for x := 0; x < 20; x++ {
			floor.Tiles[y][x] = types.Tile{Type: types.TileFloor}
		}
	}
	floor.Depth = 1

	// Create items at known positions
	item1 := entity.NewItem("malloc", "item1", types.Position{X: 3, Y: 3})
	item2 := entity.NewItem("realloc", "item2", types.Position{X: 7, Y: 7})

	// Set floor with items - this should build spatial index
	world.SetFloor(floor, nil, []*entity.Item{item1, item2})

	// Test GetItemAt returns correct items (O(1) lookup)
	if i := world.GetItemAt(types.Position{X: 3, Y: 3}); i == nil {
		t.Error("expected item1 at (3,3)")
	} else if i.ID() != "item1" {
		t.Errorf("expected item1, got %s", i.ID())
	}

	if i := world.GetItemAt(types.Position{X: 7, Y: 7}); i == nil {
		t.Error("expected item2 at (7,7)")
	} else if i.ID() != "item2" {
		t.Errorf("expected item2, got %s", i.ID())
	}

	// Test empty position returns nil
	if i := world.GetItemAt(types.Position{X: 15, Y: 15}); i != nil {
		t.Error("expected nil at empty position")
	}
}

func TestSpatialIndex_AddEnemy(t *testing.T) {
	world := NewGameWorld()

	// Create a simple floor
	floor := &types.Floor{
		Width:  20,
		Height: 20,
		Tiles:  make([][]types.Tile, 20),
		Depth:  1,
	}
	for y := 0; y < 20; y++ {
		floor.Tiles[y] = make([]types.Tile, 20)
	}

	world.SetFloor(floor, nil, nil)

	// Add enemy - should update spatial index
	enemy := entity.NewEnemy("zombie", "new_enemy", types.Position{X: 8, Y: 8}, 1)
	world.AddEnemy(enemy)

	// Verify spatial index was updated
	if e := world.GetEnemyAt(types.Position{X: 8, Y: 8}); e == nil {
		t.Error("expected enemy at (8,8) after AddEnemy")
	} else if e.ID() != "new_enemy" {
		t.Errorf("expected new_enemy, got %s", e.ID())
	}
}

func TestSpatialIndex_RemoveEnemy(t *testing.T) {
	world := NewGameWorld()

	// Create floor with an enemy
	floor := &types.Floor{
		Width:  20,
		Height: 20,
		Tiles:  make([][]types.Tile, 20),
		Depth:  1,
	}
	for y := 0; y < 20; y++ {
		floor.Tiles[y] = make([]types.Tile, 20)
	}

	enemy := entity.NewEnemy("zombie", "to_remove", types.Position{X: 5, Y: 5}, 1)
	world.SetFloor(floor, []*entity.Enemy{enemy}, nil)

	// Verify enemy is in index
	if world.GetEnemyAt(types.Position{X: 5, Y: 5}) == nil {
		t.Fatal("enemy should be in index before removal")
	}

	// Remove enemy - should update spatial index
	removed := world.RemoveEnemy("to_remove")
	if !removed {
		t.Fatal("RemoveEnemy should return true")
	}

	// Verify spatial index was updated
	if world.GetEnemyAt(types.Position{X: 5, Y: 5}) != nil {
		t.Error("enemy should be removed from spatial index")
	}

	// Verify removal is tracked
	if removedIDs := world.GetRemovedEnemies(1); len(removedIDs) == 0 || removedIDs[0] != "to_remove" {
		t.Error("removed enemy should be tracked")
	}
}

func TestSpatialIndex_AddItem(t *testing.T) {
	world := NewGameWorld()

	// Create a simple floor
	floor := &types.Floor{
		Width:  20,
		Height: 20,
		Tiles:  make([][]types.Tile, 20),
		Depth:  1,
	}
	for y := 0; y < 20; y++ {
		floor.Tiles[y] = make([]types.Tile, 20)
	}

	world.SetFloor(floor, nil, nil)

	// Add item - should update spatial index
	item := entity.NewItem("malloc", "new_item", types.Position{X: 6, Y: 6})
	world.AddItem(item)

	// Verify spatial index was updated
	if i := world.GetItemAt(types.Position{X: 6, Y: 6}); i == nil {
		t.Error("expected item at (6,6) after AddItem")
	} else if i.ID() != "new_item" {
		t.Errorf("expected new_item, got %s", i.ID())
	}
}

func TestSpatialIndex_RemoveItem(t *testing.T) {
	world := NewGameWorld()

	// Create floor with an item
	floor := &types.Floor{
		Width:  20,
		Height: 20,
		Tiles:  make([][]types.Tile, 20),
		Depth:  1,
	}
	for y := 0; y < 20; y++ {
		floor.Tiles[y] = make([]types.Tile, 20)
	}

	item := entity.NewItem("malloc", "to_loot", types.Position{X: 4, Y: 4})
	world.SetFloor(floor, nil, []*entity.Item{item})

	// Verify item is in index
	if world.GetItemAt(types.Position{X: 4, Y: 4}) == nil {
		t.Fatal("item should be in index before removal")
	}

	// Remove item - should update spatial index
	removed := world.RemoveItem("to_loot")
	if !removed {
		t.Fatal("RemoveItem should return true")
	}

	// Verify spatial index was updated
	if world.GetItemAt(types.Position{X: 4, Y: 4}) != nil {
		t.Error("item should be removed from spatial index")
	}

	// Verify removal is tracked
	if removedIDs := world.GetRemovedItems(1); len(removedIDs) == 0 || removedIDs[0] != "to_loot" {
		t.Error("removed item should be tracked")
	}
}

func TestSpatialIndex_DeadEnemyNotIndexed(t *testing.T) {
	world := NewGameWorld()

	// Create a simple floor
	floor := &types.Floor{
		Width:  20,
		Height: 20,
		Tiles:  make([][]types.Tile, 20),
		Depth:  1,
	}
	for y := 0; y < 20; y++ {
		floor.Tiles[y] = make([]types.Tile, 20)
	}

	// Create a dead enemy (RAM = 0)
	deadEnemy := entity.NewEnemy("zombie", "dead", types.Position{X: 5, Y: 5}, 1)
	deadEnemy.Stats.RAM = 0 // Kill it

	aliveEnemy := entity.NewEnemy("daemon", "alive", types.Position{X: 10, Y: 10}, 1)

	world.SetFloor(floor, []*entity.Enemy{deadEnemy, aliveEnemy}, nil)

	// Dead enemy should NOT be in spatial index
	if e := world.GetEnemyAt(types.Position{X: 5, Y: 5}); e != nil {
		t.Error("dead enemy should not be in spatial index")
	}

	// Alive enemy should be in spatial index
	if e := world.GetEnemyAt(types.Position{X: 10, Y: 10}); e == nil {
		t.Error("alive enemy should be in spatial index")
	}
}

func TestSpatialIndex_LoadCachedFloorRebuildsIndex(t *testing.T) {
	world := NewGameWorld()

	// Create floor 1 with enemy
	floor1 := &types.Floor{
		Width:  20,
		Height: 20,
		Tiles:  make([][]types.Tile, 20),
		Depth:  1,
	}
	for y := 0; y < 20; y++ {
		floor1.Tiles[y] = make([]types.Tile, 20)
	}
	enemy1 := entity.NewEnemy("zombie", "floor1_enemy", types.Position{X: 5, Y: 5}, 1)
	world.SetFloor(floor1, []*entity.Enemy{enemy1}, nil)

	// Cache floor 1
	world.CacheCurrentFloor()

	// Create floor 2 with different enemy
	floor2 := &types.Floor{
		Width:  20,
		Height: 20,
		Tiles:  make([][]types.Tile, 20),
		Depth:  2,
	}
	for y := 0; y < 20; y++ {
		floor2.Tiles[y] = make([]types.Tile, 20)
	}
	enemy2 := entity.NewEnemy("daemon", "floor2_enemy", types.Position{X: 10, Y: 10}, 1)
	world.SetFloor(floor2, []*entity.Enemy{enemy2}, nil)

	// Verify floor 2 enemy is in index, floor 1 enemy is not
	if world.GetEnemyAt(types.Position{X: 5, Y: 5}) != nil {
		t.Error("floor 1 enemy should not be in index on floor 2")
	}
	if world.GetEnemyAt(types.Position{X: 10, Y: 10}) == nil {
		t.Error("floor 2 enemy should be in index")
	}

	// Load cached floor 1
	if !world.LoadCachedFloor(1) {
		t.Fatal("should load cached floor 1")
	}

	// Verify floor 1 enemy is now in index, floor 2 enemy is not
	if world.GetEnemyAt(types.Position{X: 5, Y: 5}) == nil {
		t.Error("floor 1 enemy should be in index after loading cache")
	}
	if world.GetEnemyAt(types.Position{X: 10, Y: 10}) != nil {
		t.Error("floor 2 enemy should not be in index after loading floor 1")
	}
}

// === Concurrency Tests ===
// These tests verify the engine mutex properly protects state during concurrent access.

func TestEngineMutex_ConcurrentMoveAndSave(t *testing.T) {
	// This test verifies that concurrent moves and saves don't cause data races.
	// Run with: go test -race ./internal/game/...

	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Run concurrent operations
	done := make(chan bool, 2)

	// Goroutine 1: Move player repeatedly
	go func() {
		for i := 0; i < 100; i++ {
			engine.MovePlayer(types.DirRight)
			engine.MovePlayer(types.DirLeft)
		}
		done <- true
	}()

	// Goroutine 2: Get save data repeatedly
	go func() {
		for i := 0; i < 100; i++ {
			_ = engine.toSaveData()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// If we get here without race detector panic, the mutex is working
}
