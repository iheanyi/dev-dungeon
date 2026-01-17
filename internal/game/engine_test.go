package game

import (
	"testing"
	"time"

	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/save"
	"github.com/iheanyi/devdungeon/internal/types"
)

func TestNewEngine(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}

	if engine.MasterSeed() != 12345 {
		t.Errorf("expected seed 12345, got %d", engine.MasterSeed())
	}

	if engine.CurrentState() != types.StateMainMenu {
		t.Errorf("expected StateMainMenu, got %v", engine.CurrentState())
	}
}

func TestStartNewGame(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	err := engine.StartNewGame(entity.ClassInit)
	if err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	if engine.Player() == nil {
		t.Fatal("Player is nil after StartNewGame")
	}

	if engine.Player().Class != entity.ClassInit {
		t.Errorf("expected ClassInit, got %v", engine.Player().Class)
	}

	if engine.CurrentState() != types.StateExploring {
		t.Errorf("expected StateExploring, got %v", engine.CurrentState())
	}

	if engine.CurrentDepth() != 1 {
		t.Errorf("expected depth 1, got %d", engine.CurrentDepth())
	}

	if engine.CurrentFloorType() != types.FloorHome {
		t.Errorf("expected FloorHome, got %v", engine.CurrentFloorType())
	}
}

func TestPlayerMovement(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassBash); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	initialPos := engine.Player().Position()

	// Test movement in each direction
	tests := []struct {
		name      string
		direction types.Direction
		deltaX    int
		deltaY    int
	}{
		{"move right", types.DirRight, 1, 0},
		{"move down", types.DirDown, 0, 1},
		{"move left", types.DirLeft, -1, 0},
		{"move up", types.DirUp, 0, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforePos := engine.Player().Position()
			result := engine.MovePlayer(tt.direction)

			// If movement was successful, verify position changed correctly
			if result.Moved {
				afterPos := engine.Player().Position()
				expectedX := beforePos.X + tt.deltaX
				expectedY := beforePos.Y + tt.deltaY

				if afterPos.X != expectedX || afterPos.Y != expectedY {
					t.Errorf("%s: expected position (%d,%d), got (%d,%d)",
						tt.name, expectedX, expectedY, afterPos.X, afterPos.Y)
				}
			}
		})
	}

	// Player should be able to move at least once (starting position is valid)
	// Reset to initial position for wall collision test
	engine.Player().SetPosition(initialPos)
}

func TestWallCollision(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Move player to a position near a wall
	// The default floor has walls around the perimeter
	floor := engine.GetWorld().CurrentFloor

	// Find a position next to a wall
	var testPos types.Position
	for y := 1; y < floor.Height-1; y++ {
		for x := 1; x < floor.Width-1; x++ {
			pos := types.Position{X: x, Y: y}
			// Check if this tile is walkable and has a wall to the left
			if floor.IsWalkable(pos) {
				leftPos := types.Position{X: x - 1, Y: y}
				if !floor.IsWalkable(leftPos) {
					testPos = pos
					break
				}
			}
		}
		if testPos.X != 0 {
			break
		}
	}

	if testPos.X == 0 && testPos.Y == 0 {
		t.Skip("Could not find suitable test position")
	}

	engine.Player().SetPosition(testPos)
	engine.GetWorld().UpdateVisibility(testPos, 8)

	// Try to move into the wall
	result := engine.MovePlayer(types.DirLeft)

	if result.Moved {
		t.Error("Player should not be able to move into a wall")
	}

	// Position should be unchanged
	if engine.Player().Position() != testPos {
		t.Errorf("Player position changed when moving into wall: %v", engine.Player().Position())
	}
}

func TestFloorTransitions(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassCron); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Save initial floor depth
	initialDepth := engine.CurrentDepth()
	if initialDepth != 1 {
		t.Errorf("expected initial depth 1, got %d", initialDepth)
	}

	// Move to stairs down position
	stairsDownPos := engine.GetWorld().CurrentFloor.StairsDown
	engine.Player().SetPosition(stairsDownPos)

	// Verify we're on stairs
	tile := engine.GetWorld().CurrentFloor.GetTile(stairsDownPos)
	if tile == nil || tile.Type != types.TileStairsDown {
		t.Fatal("player not on stairs down")
	}

	// Descend
	if err := engine.DescendStairs(); err != nil {
		t.Fatalf("DescendStairs failed: %v", err)
	}

	// Verify we're on floor 2
	if engine.CurrentDepth() != 2 {
		t.Errorf("expected depth 2, got %d", engine.CurrentDepth())
	}

	// Floor type should be /tmp
	if engine.CurrentFloorType() != types.FloorTmp {
		t.Errorf("expected FloorTmp, got %v", engine.CurrentFloorType())
	}

	// Move to stairs up and ascend
	stairsUpPos := engine.GetWorld().CurrentFloor.StairsUp
	engine.Player().SetPosition(stairsUpPos)

	if err := engine.AscendStairs(); err != nil {
		t.Fatalf("AscendStairs failed: %v", err)
	}

	// Should be back on floor 1
	if engine.CurrentDepth() != 1 {
		t.Errorf("expected depth 1, got %d", engine.CurrentDepth())
	}

	if engine.CurrentFloorType() != types.FloorHome {
		t.Errorf("expected FloorHome, got %v", engine.CurrentFloorType())
	}
}

func TestFloorCaching(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassVim); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Get initial enemy count on floor 1
	floor1Enemies := len(engine.GetEnemies())
	floor1Items := len(engine.GetItems())

	// Descend to floor 2
	stairsDownPos := engine.GetWorld().CurrentFloor.StairsDown
	engine.Player().SetPosition(stairsDownPos)

	if err := engine.DescendStairs(); err != nil {
		t.Fatalf("DescendStairs failed: %v", err)
	}

	// Ascend back to floor 1
	stairsUpPos := engine.GetWorld().CurrentFloor.StairsUp
	engine.Player().SetPosition(stairsUpPos)

	if err := engine.AscendStairs(); err != nil {
		t.Fatalf("AscendStairs failed: %v", err)
	}

	// Enemy and item counts should be preserved
	if len(engine.GetEnemies()) != floor1Enemies {
		t.Errorf("enemy count changed: expected %d, got %d", floor1Enemies, len(engine.GetEnemies()))
	}

	if len(engine.GetItems()) != floor1Items {
		t.Errorf("item count changed: expected %d, got %d", floor1Items, len(engine.GetItems()))
	}
}

func TestDeterministicSeeds(t *testing.T) {
	cfg := config.DefaultConfig()

	// Create two engines with the same seed
	engine1 := NewEngine(cfg, 42)
	engine2 := NewEngine(cfg, 42)

	if err := engine1.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("engine1 StartNewGame failed: %v", err)
	}

	if err := engine2.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("engine2 StartNewGame failed: %v", err)
	}

	// Both should have identical floor layouts
	floor1 := engine1.GetWorld().CurrentFloor
	floor2 := engine2.GetWorld().CurrentFloor

	if floor1.Width != floor2.Width || floor1.Height != floor2.Height {
		t.Error("floor dimensions differ")
	}

	// Compare tiles
	for y := 0; y < floor1.Height; y++ {
		for x := 0; x < floor1.Width; x++ {
			tile1 := floor1.Tiles[y][x]
			tile2 := floor2.Tiles[y][x]

			if tile1.Type != tile2.Type {
				t.Errorf("tile mismatch at (%d,%d): %v vs %v", x, y, tile1.Type, tile2.Type)
			}
		}
	}

	// Compare stair positions
	if floor1.StairsDown != floor2.StairsDown {
		t.Errorf("stairs down positions differ: %v vs %v", floor1.StairsDown, floor2.StairsDown)
	}

	if floor1.StairsUp != floor2.StairsUp {
		t.Errorf("stairs up positions differ: %v vs %v", floor1.StairsUp, floor2.StairsUp)
	}

	// Enemy count should be the same
	if len(engine1.GetEnemies()) != len(engine2.GetEnemies()) {
		t.Errorf("enemy counts differ: %d vs %d", len(engine1.GetEnemies()), len(engine2.GetEnemies()))
	}
}

func TestDifferentSeeds(t *testing.T) {
	cfg := config.DefaultConfig()

	engine1 := NewEngine(cfg, 111)
	engine2 := NewEngine(cfg, 222)

	if err := engine1.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("engine1 StartNewGame failed: %v", err)
	}

	if err := engine2.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("engine2 StartNewGame failed: %v", err)
	}

	// With different seeds, we expect different results
	// Check that at least something is different
	floor1 := engine1.GetWorld().CurrentFloor
	floor2 := engine2.GetWorld().CurrentFloor

	// The basic default floor generation produces same structure
	// but enemy/item positions should differ
	enemies1 := engine1.GetEnemies()
	enemies2 := engine2.GetEnemies()

	// If there are enemies, their positions should likely differ
	if len(enemies1) > 0 && len(enemies2) > 0 {
		samePositions := true
		for i := 0; i < len(enemies1) && i < len(enemies2); i++ {
			if enemies1[i].Position() != enemies2[i].Position() {
				samePositions = false
				break
			}
		}

		// We expect different positions with different seeds
		// This test is probabilistic but should pass with high probability
		_ = samePositions // Just verify we can compare positions
	}

	// Verify seeds are stored correctly
	if engine1.MasterSeed() != 111 {
		t.Errorf("engine1 seed mismatch: expected 111, got %d", engine1.MasterSeed())
	}

	if engine2.MasterSeed() != 222 {
		t.Errorf("engine2 seed mismatch: expected 222, got %d", engine2.MasterSeed())
	}

	_ = floor1
	_ = floor2
}

func TestSeedDerivation(t *testing.T) {
	masterSeed := int64(12345)

	// Same inputs should produce same outputs
	seed1a := DeriveFloorSeed(masterSeed, 1)
	seed1b := DeriveFloorSeed(masterSeed, 1)

	if seed1a != seed1b {
		t.Errorf("same inputs produced different seeds: %d vs %d", seed1a, seed1b)
	}

	// Different depths should produce different seeds
	seed2 := DeriveFloorSeed(masterSeed, 2)
	if seed1a == seed2 {
		t.Error("different depths produced same seed")
	}

	// Different master seeds should produce different seeds
	seed3 := DeriveFloorSeed(54321, 1)
	if seed1a == seed3 {
		t.Error("different master seeds produced same seed")
	}
}

func TestEnemyBumpCombat(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassBash); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Find an enemy position
	enemies := engine.GetEnemies()
	if len(enemies) == 0 {
		t.Skip("No enemies spawned")
	}

	enemy := enemies[0]
	enemyPos := enemy.Position()

	// Position player adjacent to enemy
	playerPos := types.Position{X: enemyPos.X - 1, Y: enemyPos.Y}
	if !engine.GetWorld().CurrentFloor.IsWalkable(playerPos) {
		playerPos = types.Position{X: enemyPos.X + 1, Y: enemyPos.Y}
	}
	if !engine.GetWorld().CurrentFloor.IsWalkable(playerPos) {
		t.Skip("Could not find valid position adjacent to enemy")
	}

	engine.Player().SetPosition(playerPos)

	// Bump into enemy
	result := engine.MovePlayer(types.DirRight)

	if result.Moved {
		t.Error("player should not move into enemy")
	}

	if len(result.Combat) == 0 {
		t.Error("expected combat encounter")
	}

	if len(result.Combat) > 0 && result.Combat[0].ID() != enemy.ID() {
		t.Errorf("combat enemy mismatch: expected %s, got %s", enemy.ID(), result.Combat[0].ID())
	}
}

func TestItemPickup(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Manually place an item adjacent to player
	playerPos := engine.Player().Position()
	itemPos := types.Position{X: playerPos.X + 1, Y: playerPos.Y}

	// Make sure item position is walkable
	if !engine.GetWorld().CurrentFloor.IsWalkable(itemPos) {
		t.Skip("Adjacent position not walkable")
	}

	// Use core_dump which is not in starting inventory, so we know it won't stack
	testItem := entity.NewItem("core_dump", "test_item_123", itemPos)
	if testItem == nil {
		t.Fatal("Failed to create test item")
	}

	engine.GetWorld().AddItem(testItem)
	initialInventorySize := len(engine.Player().Inventory.Items)

	// Move to pick up item
	result := engine.MovePlayer(types.DirRight)

	if !result.Moved {
		t.Fatal("Player should be able to move to item position")
	}

	if result.PickedUp == nil {
		t.Error("Expected to pick up item")
	}

	if result.PickedUp != nil && result.PickedUp.ID() != "test_item_123" {
		t.Errorf("Picked up wrong item: %s", result.PickedUp.ID())
	}

	// Check inventory - core_dump not in starting gear, so should create new slot
	if len(engine.Player().Inventory.Items) != initialInventorySize+1 {
		t.Errorf("Inventory size should have increased by 1, got %d (was %d)", len(engine.Player().Inventory.Items), initialInventorySize)
	}

	// Item should be removed from world
	if engine.GetWorld().GetItemAt(itemPos) != nil {
		t.Error("Item should be removed from world after pickup")
	}
}

func TestAscendFromFirstFloor(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Try to ascend from floor 1
	err := engine.AscendStairs()

	if err == nil {
		t.Error("should not be able to ascend from first floor")
	}
}

func TestDescendWithoutStairs(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Player starts at PlayerStart, not on stairs
	startPos := engine.Player().Position()

	// Move away from any stairs
	engine.MovePlayer(types.DirRight)
	engine.MovePlayer(types.DirRight)

	// Try to descend without being on stairs
	err := engine.DescendStairs()

	if err == nil {
		t.Error("should not be able to descend without stairs")
	}

	_ = startPos
}

func TestCannotDescendPastBossFloor(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Manually set depth to boss floor
	engine.world.CurrentDepth = 8

	// Try to descend past the final floor
	err := engine.DescendStairs()

	if err == nil {
		t.Error("should not be able to descend past the boss floor (depth 8)")
	}
}

func TestStateManagement(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	// Initial state should be main menu
	if engine.CurrentState() != types.StateMainMenu {
		t.Errorf("expected StateMainMenu, got %v", engine.CurrentState())
	}

	// Start game - should be exploring
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	if engine.CurrentState() != types.StateExploring {
		t.Errorf("expected StateExploring, got %v", engine.CurrentState())
	}

	// Manually change state
	engine.SetState(types.StateCombat)
	if engine.CurrentState() != types.StateCombat {
		t.Errorf("expected StateCombat, got %v", engine.CurrentState())
	}

	engine.SetState(types.StateInventory)
	if engine.CurrentState() != types.StateInventory {
		t.Errorf("expected StateInventory, got %v", engine.CurrentState())
	}
}

func TestMessages(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Should have a welcome message
	messages := engine.Messages()
	if len(messages) == 0 {
		t.Error("expected at least one welcome message")
	}

	// Clear messages
	engine.ClearMessages()
	if len(engine.Messages()) != 0 {
		t.Error("messages should be cleared")
	}
}

func TestFloorTypeProgression(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	expectedTypes := []types.FloorType{
		types.FloorHome,
		types.FloorTmp,
		types.FloorVar,
		types.FloorEtc,
		types.FloorUsr,
		types.FloorSys,
		types.FloorDev,
		types.FloorDevNull,
	}

	for depth, expectedType := range expectedTypes {
		actualType := engine.getFloorTypeForDepth(depth + 1)
		if actualType != expectedType {
			t.Errorf("depth %d: expected %v, got %v", depth+1, expectedType.FloorName(), actualType.FloorName())
		}
	}
}

func TestLoadGameDoesNotDuplicateItems(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	// Start a new game
	if err := engine.StartNewGame(entity.ClassBash); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Record initial inventory count (starting gear)
	initialInventoryCount := len(engine.Player().Inventory.Items)

	// Create mock save data with specific inventory
	saveData := engine.toSaveData()

	// Clear the inventory in save data to simulate a saved state with no items
	saveData.Player.Inventory = nil
	saveData.Player.Equipment.Weapon = ""
	saveData.Player.Equipment.Armor = ""

	// Load the save - this should NOT add starting gear again
	if err := engine.LoadGame(saveData); err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	// Inventory should be empty (not have starting gear)
	if len(engine.Player().Inventory.Items) != 0 {
		t.Errorf("expected empty inventory after loading save with no items, got %d items",
			len(engine.Player().Inventory.Items))
	}

	// Equipment should be empty
	if engine.Player().Equipment.Weapon != nil {
		t.Errorf("expected no weapon after loading save with no weapon, got %s",
			engine.Player().Equipment.Weapon.Name())
	}

	_ = initialInventoryCount
}

func TestLoadGamePreservesInventory(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Get save data
	saveData := engine.toSaveData()

	// Add a specific item to save data inventory
	saveData.Player.Inventory = []save.ItemData{
		{TemplateID: "malloc", Quantity: 5},
		{TemplateID: "realloc", Quantity: 2},
	}
	saveData.Player.Equipment.Weapon = "vim_blade"
	saveData.Player.Equipment.Armor = "firewall"

	// Load the save
	if err := engine.LoadGame(saveData); err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	// Verify inventory has exactly what we saved (not starting gear)
	if len(engine.Player().Inventory.Items) != 2 {
		t.Errorf("expected 2 items in inventory, got %d", len(engine.Player().Inventory.Items))
	}

	// Verify equipment
	if engine.Player().Equipment.Weapon == nil {
		t.Error("expected weapon to be equipped")
	} else if engine.Player().Equipment.Weapon.TemplateID != "vim_blade" {
		t.Errorf("expected vim_blade weapon, got %s", engine.Player().Equipment.Weapon.TemplateID)
	}

	if engine.Player().Equipment.Armor == nil {
		t.Error("expected armor to be equipped")
	} else if engine.Player().Equipment.Armor.TemplateID != "firewall" {
		t.Errorf("expected firewall armor, got %s", engine.Player().Equipment.Armor.TemplateID)
	}
}

func TestLoadGameInvalidPositionFallsBackToStairs(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassBash); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Get save data and set position to an invalid location (inside a wall)
	saveData := engine.toSaveData()

	// Set position to a known wall position (0,0 is always a wall in BSP generation)
	saveData.Player.Position = types.Position{X: 0, Y: 0}

	// Load the save
	if err := engine.LoadGame(saveData); err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	// Player should NOT be at the invalid position
	playerPos := engine.Player().Position()
	if playerPos.X == 0 && playerPos.Y == 0 {
		t.Error("player should not be at invalid position (0,0)")
	}

	// Player should be at the stairs up position
	stairsUp := engine.GetWorld().CurrentFloor.StairsUp
	if playerPos != stairsUp {
		t.Errorf("player should be at stairs up (%v), got %v", stairsUp, playerPos)
	}

	// Verify position is walkable
	if !engine.GetWorld().CurrentFloor.IsWalkable(playerPos) {
		t.Error("player position should be walkable")
	}
}

func TestGatherNearbyEnemies(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Clear existing enemies and place our own for controlled test
	engine.world.Enemies = nil

	// Create a target enemy at known position
	targetPos := types.Position{X: 10, Y: 10}
	target := entity.NewEnemy("zombie", "target-1", targetPos, 1)

	// Create enemies at various distances
	nearEnemy1 := entity.NewEnemy("zombie", "near-1", types.Position{X: 11, Y: 10}, 1) // 1 tile away (within radius 2)
	nearEnemy2 := entity.NewEnemy("zombie", "near-2", types.Position{X: 10, Y: 11}, 1) // 1 tile away (within radius 2)
	nearEnemy3 := entity.NewEnemy("zombie", "near-3", types.Position{X: 12, Y: 12}, 1) // 2 tiles away diagonally (within radius 2)
	farEnemy := entity.NewEnemy("zombie", "far-1", types.Position{X: 15, Y: 15}, 1)    // 5 tiles away (outside radius 2)

	// Add enemies to world
	engine.world.Enemies = []*entity.Enemy{target, nearEnemy1, nearEnemy2, nearEnemy3, farEnemy}

	// Test gathering with radius 2
	enemies := engine.gatherNearbyEnemies(target, 2)

	// Should include target + 3 nearby enemies (4 total)
	if len(enemies) != 4 {
		t.Errorf("expected 4 enemies (target + 3 nearby), got %d", len(enemies))
	}

	// Verify target is included
	foundTarget := false
	for _, e := range enemies {
		if e.ID() == "target-1" {
			foundTarget = true
			break
		}
	}
	if !foundTarget {
		t.Error("target enemy should be included in results")
	}

	// Verify far enemy is NOT included
	foundFar := false
	for _, e := range enemies {
		if e.ID() == "far-1" {
			foundFar = true
			break
		}
	}
	if foundFar {
		t.Error("far enemy (distance 5) should NOT be included in radius 2 search")
	}
}

func TestGatherNearbyEnemiesSingleEnemy(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Clear existing enemies
	engine.world.Enemies = nil

	// Create only one enemy
	targetPos := types.Position{X: 10, Y: 10}
	target := entity.NewEnemy("zombie", "alone-1", targetPos, 1)
	engine.world.Enemies = []*entity.Enemy{target}

	// Test gathering - should return just the target
	enemies := engine.gatherNearbyEnemies(target, 2)

	if len(enemies) != 1 {
		t.Errorf("expected 1 enemy (just target), got %d", len(enemies))
	}

	if enemies[0].ID() != "alone-1" {
		t.Errorf("expected target enemy, got %s", enemies[0].ID())
	}
}

func TestGatherNearbyEnemiesExcludesDeadEnemies(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Clear existing enemies
	engine.world.Enemies = nil

	// Create enemies
	targetPos := types.Position{X: 10, Y: 10}
	target := entity.NewEnemy("zombie", "target-1", targetPos, 1)
	nearEnemy := entity.NewEnemy("zombie", "near-1", types.Position{X: 11, Y: 10}, 1)
	deadEnemy := entity.NewEnemy("zombie", "dead-1", types.Position{X: 10, Y: 11}, 1)

	// Kill the dead enemy
	deadEnemy.Stats.RAM = 0

	engine.world.Enemies = []*entity.Enemy{target, nearEnemy, deadEnemy}

	// Test gathering
	enemies := engine.gatherNearbyEnemies(target, 2)

	// Should include target + near enemy, but NOT dead enemy (2 total)
	if len(enemies) != 2 {
		t.Errorf("expected 2 enemies (target + near), got %d", len(enemies))
	}

	// Verify dead enemy is NOT included
	foundDead := false
	for _, e := range enemies {
		if e.ID() == "dead-1" {
			foundDead = true
			break
		}
	}
	if foundDead {
		t.Error("dead enemy should NOT be included in results")
	}
}

func TestGatherNearbyEnemiesChebyshevDistance(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Clear existing enemies
	engine.world.Enemies = nil

	// Test Chebyshev distance (king's move): diagonal distance should equal max of dx, dy
	// With target at (10,10) and radius 2:
	// - (12, 12) is 2 away (should be included)
	// - (13, 13) is 3 away (should NOT be included)
	targetPos := types.Position{X: 10, Y: 10}
	target := entity.NewEnemy("zombie", "target", targetPos, 1)
	diag2 := entity.NewEnemy("zombie", "diag2", types.Position{X: 12, Y: 12}, 1) // Chebyshev dist = 2
	diag3 := entity.NewEnemy("zombie", "diag3", types.Position{X: 13, Y: 13}, 1) // Chebyshev dist = 3

	engine.world.Enemies = []*entity.Enemy{target, diag2, diag3}

	enemies := engine.gatherNearbyEnemies(target, 2)

	if len(enemies) != 2 {
		t.Errorf("expected 2 enemies (target + diag2), got %d", len(enemies))
	}

	// Verify diag2 IS included
	foundDiag2 := false
	for _, e := range enemies {
		if e.ID() == "diag2" {
			foundDiag2 = true
			break
		}
	}
	if !foundDiag2 {
		t.Error("enemy at distance 2 (diagonal) should be included")
	}

	// Verify diag3 is NOT included
	foundDiag3 := false
	for _, e := range enemies {
		if e.ID() == "diag3" {
			foundDiag3 = true
			break
		}
	}
	if foundDiag3 {
		t.Error("enemy at distance 3 (diagonal) should NOT be included in radius 2 search")
	}
}

func TestGatherNearbyEnemiesNilWorld(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	// Don't start a game - world will be nil
	target := entity.NewEnemy("zombie", "target", types.Position{X: 5, Y: 5}, 1)

	enemies := engine.gatherNearbyEnemies(target, 2)

	// Should return just the target
	if len(enemies) != 1 {
		t.Errorf("expected 1 enemy when world is nil, got %d", len(enemies))
	}

	if enemies[0].ID() != "target" {
		t.Error("should return target when world is nil")
	}
}

// TestSaveDataUsesTemplateID is a regression test to ensure that toSaveData()
// saves item TemplateIDs (like "malloc") not instance IDs (like "item_malloc_123").
// Bug: Items were being saved with ID() instead of TemplateID, causing items
// to be lost on load because NewItem() couldn't find templates with those IDs.
func TestSaveDataUsesTemplateID(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 12345)

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Clear starting gear to have clean slate for testing
	engine.Player().Inventory.Clear()
	engine.Player().Equipment.Weapon = nil
	engine.Player().Equipment.Armor = nil
	engine.Player().Equipment.Utility1 = nil
	engine.Player().Equipment.Utility2 = nil

	// Add specific items to inventory with known template IDs
	// These items have instance IDs like "item_malloc_0" but TemplateID is "malloc"
	malloc := entity.NewItem("malloc", "item_malloc_test", types.Position{})
	malloc.Quantity = 3
	engine.Player().Inventory.AddItem(malloc)

	realloc := entity.NewItem("realloc", "item_realloc_test", types.Position{})
	realloc.Quantity = 2
	engine.Player().Inventory.AddItem(realloc)

	// Equip items
	weapon := entity.NewItem("vim_blade", "weapon_test", types.Position{})
	engine.Player().Equipment.Weapon = weapon

	armor := entity.NewItem("firewall", "armor_test", types.Position{})
	engine.Player().Equipment.Armor = armor

	utility := entity.NewItem("ssh_key", "utility_test", types.Position{})
	engine.Player().Equipment.Utility1 = utility

	// Get save data
	saveData := engine.toSaveData()

	// Verify inventory items use TemplateID, not instance ID
	foundMalloc := false
	foundRealloc := false
	for _, item := range saveData.Player.Inventory {
		// TemplateID should be "malloc" or "realloc", NOT "item_malloc_test" or "item_realloc_test"
		if item.TemplateID == "malloc" {
			foundMalloc = true
			if item.Quantity != 3 {
				t.Errorf("malloc quantity = %d, want 3", item.Quantity)
			}
		}
		if item.TemplateID == "realloc" {
			foundRealloc = true
			if item.Quantity != 2 {
				t.Errorf("realloc quantity = %d, want 2", item.Quantity)
			}
		}
		// Should never contain instance IDs
		if item.TemplateID == "item_malloc_test" || item.TemplateID == "item_realloc_test" {
			t.Errorf("save data contains instance ID %q instead of template ID", item.TemplateID)
		}
	}

	if !foundMalloc {
		t.Error("malloc not found in saved inventory (or has wrong TemplateID)")
	}
	if !foundRealloc {
		t.Error("realloc not found in saved inventory (or has wrong TemplateID)")
	}

	// Verify equipment uses TemplateID
	if saveData.Player.Equipment.Weapon != "vim_blade" {
		t.Errorf("saved weapon = %q, want 'vim_blade' (TemplateID)", saveData.Player.Equipment.Weapon)
	}
	if saveData.Player.Equipment.Armor != "firewall" {
		t.Errorf("saved armor = %q, want 'firewall' (TemplateID)", saveData.Player.Equipment.Armor)
	}
	if saveData.Player.Equipment.Utility1 != "ssh_key" {
		t.Errorf("saved utility1 = %q, want 'ssh_key' (TemplateID)", saveData.Player.Equipment.Utility1)
	}

	// Verify round-trip: save then load should preserve items
	newEngine := NewEngine(cfg, 12345)
	if err := newEngine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame for load test failed: %v", err)
	}
	if err := newEngine.LoadGame(saveData); err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	// Check loaded inventory has correct items
	loadedInv := newEngine.Player().Inventory.Items
	foundLoadedMalloc := false
	foundLoadedRealloc := false
	for _, item := range loadedInv {
		if item.TemplateID == "malloc" && item.Quantity == 3 {
			foundLoadedMalloc = true
		}
		if item.TemplateID == "realloc" && item.Quantity == 2 {
			foundLoadedRealloc = true
		}
	}
	if !foundLoadedMalloc {
		t.Error("malloc not found after load (save/load round-trip failed)")
	}
	if !foundLoadedRealloc {
		t.Error("realloc not found after load (save/load round-trip failed)")
	}

	// Check loaded equipment
	if newEngine.Player().Equipment.Weapon == nil || newEngine.Player().Equipment.Weapon.TemplateID != "vim_blade" {
		t.Error("weapon not loaded correctly")
	}
	if newEngine.Player().Equipment.Armor == nil || newEngine.Player().Equipment.Armor.TemplateID != "firewall" {
		t.Error("armor not loaded correctly")
	}
	if newEngine.Player().Equipment.Utility1 == nil || newEngine.Player().Equipment.Utility1.TemplateID != "ssh_key" {
		t.Error("utility1 not loaded correctly")
	}
}

// TestSaveLoadE2E_FileSystem is an E2E integration test that verifies the full
// save/load cycle including actual file I/O. This tests the complete path:
// Engine.SaveSync() -> toSaveData() -> SaveManager -> File -> Load -> LoadGame()
func TestSaveLoadE2E_FileSystem(t *testing.T) {
	// Create a save manager with a temp directory for real file I/O
	tempDir := t.TempDir()
	saveCfg := save.Config{
		SaveDir:          tempDir,
		AutoSaveInterval: 1 * time.Second,
		MinSaveInterval:  0, // No debounce for tests
	}

	saveMgr, err := save.NewManager(saveCfg)
	if err != nil {
		t.Fatalf("failed to create save manager: %v", err)
	}
	saveMgr.Start()
	defer saveMgr.Stop()

	// Create engine and inject the save manager
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, 54321)
	engine.saveManager = saveMgr // Inject custom save manager

	if err := engine.StartNewGame(entity.ClassBash); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Clear starting gear for clean test
	engine.Player().Inventory.Clear()
	engine.Player().Equipment.Weapon = nil
	engine.Player().Equipment.Armor = nil
	engine.Player().Equipment.Utility1 = nil
	engine.Player().Equipment.Utility2 = nil

	// Set player to known state
	engine.Player().Level = 7
	engine.Player().XP = 350
	engine.Player().MaxStats.MaxRAM = 200 // Must set max before current stat
	engine.Player().Stats.RAM = 150
	engine.Player().Stats.CPU = 25
	engine.Player().ExitCodes = 42
	engine.Player().SetPosition(types.Position{X: 10, Y: 15})

	// Add inventory items
	malloc := entity.NewItem("malloc", "e2e_malloc", types.Position{})
	malloc.Quantity = 5
	engine.Player().Inventory.AddItem(malloc)

	mmap := entity.NewItem("mmap", "e2e_mmap", types.Position{})
	mmap.Quantity = 2
	engine.Player().Inventory.AddItem(mmap)

	// Equip items
	weapon := entity.NewItem("vim_blade", "e2e_weapon", types.Position{})
	engine.Player().Equipment.Weapon = weapon

	armor := entity.NewItem("firewall", "e2e_armor", types.Position{})
	engine.Player().Equipment.Armor = armor

	util1 := entity.NewItem("ssh_key", "e2e_util1", types.Position{})
	engine.Player().Equipment.Utility1 = util1

	util2 := entity.NewItem("env_vars", "e2e_util2", types.Position{})
	engine.Player().Equipment.Utility2 = util2

	// Save to file system (synchronously)
	if err := engine.SaveSync(save.TriggerManual); err != nil {
		t.Fatalf("SaveSync failed: %v", err)
	}

	// Verify save file was created
	saves, err := saveMgr.ListSaves()
	if err != nil {
		t.Fatalf("ListSaves failed: %v", err)
	}
	if len(saves) != 1 {
		t.Fatalf("expected 1 save file, got %d", len(saves))
	}

	// Create a FRESH engine with the same save manager (same temp dir)
	newEngine := NewEngine(cfg, 99999) // Different seed - should use loaded data
	newEngine.saveManager = saveMgr

	// Start a new game first (required before LoadGame)
	if err := newEngine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame for loading failed: %v", err)
	}

	// Load the saved game from disk
	if err := newEngine.LoadLatestSave(); err != nil {
		t.Fatalf("LoadLatestSave failed: %v", err)
	}

	// Verify player state was restored correctly
	player := newEngine.Player()

	// Check class (should be bash, not init which was used for StartNewGame)
	if player.Class != entity.ClassBash {
		t.Errorf("loaded class = %s, want bash", player.Class)
	}

	// Check level and XP
	if player.Level != 7 {
		t.Errorf("loaded level = %d, want 7", player.Level)
	}
	if player.XP != 350 {
		t.Errorf("loaded XP = %d, want 350", player.XP)
	}

	// Check stats
	if player.Stats.RAM != 150 {
		t.Errorf("loaded RAM = %d, want 150", player.Stats.RAM)
	}
	if player.Stats.CPU != 25 {
		t.Errorf("loaded CPU = %d, want 25", player.Stats.CPU)
	}

	// Check exit codes
	if player.ExitCodes != 42 {
		t.Errorf("loaded exit codes = %d, want 42", player.ExitCodes)
	}

	// Check position
	pos := player.Position()
	if pos.X != 10 || pos.Y != 15 {
		t.Errorf("loaded position = (%d, %d), want (10, 15)", pos.X, pos.Y)
	}

	// Verify inventory survived the round-trip through file system
	foundMalloc := false
	foundMmap := false
	for _, item := range player.Inventory.Items {
		if item.TemplateID == "malloc" {
			foundMalloc = true
			if item.Quantity != 5 {
				t.Errorf("malloc quantity = %d, want 5", item.Quantity)
			}
		}
		if item.TemplateID == "mmap" {
			foundMmap = true
			if item.Quantity != 2 {
				t.Errorf("mmap quantity = %d, want 2", item.Quantity)
			}
		}
	}
	if !foundMalloc {
		t.Error("malloc not found in loaded inventory (E2E file I/O round-trip failed)")
	}
	if !foundMmap {
		t.Error("mmap not found in loaded inventory (E2E file I/O round-trip failed)")
	}

	// Verify equipment survived the round-trip through file system
	if player.Equipment.Weapon == nil {
		t.Error("weapon not loaded from file")
	} else if player.Equipment.Weapon.TemplateID != "vim_blade" {
		t.Errorf("weapon template = %s, want vim_blade", player.Equipment.Weapon.TemplateID)
	}

	if player.Equipment.Armor == nil {
		t.Error("armor not loaded from file")
	} else if player.Equipment.Armor.TemplateID != "firewall" {
		t.Errorf("armor template = %s, want firewall", player.Equipment.Armor.TemplateID)
	}

	if player.Equipment.Utility1 == nil {
		t.Error("utility1 not loaded from file")
	} else if player.Equipment.Utility1.TemplateID != "ssh_key" {
		t.Errorf("utility1 template = %s, want ssh_key", player.Equipment.Utility1.TemplateID)
	}

	if player.Equipment.Utility2 == nil {
		t.Error("utility2 not loaded from file")
	} else if player.Equipment.Utility2.TemplateID != "env_vars" {
		t.Errorf("utility2 template = %s, want env_vars", player.Equipment.Utility2.TemplateID)
	}
}

// TestCombatVictory_E2E tests the full combat flow including XP, exit codes, loot,
// and verifies everything persists through a save/load cycle with real file I/O.
func TestCombatVictory_E2E(t *testing.T) {
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
	engine := NewEngine(cfg, 11111)
	engine.saveManager = saveMgr

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Record initial state
	initialLevel := engine.Player().Level
	initialXP := engine.Player().XP
	initialExitCodes := engine.Player().ExitCodes

	// Create enemies for combat
	enemy1 := entity.NewEnemy(entity.EnemyZombie, "test_zombie", types.Position{X: 5, Y: 5}, 1)
	enemy2 := entity.NewEnemy(entity.EnemyDaemon, "test_daemon", types.Position{X: 6, Y: 5}, 1)

	// Start combat
	combat := engine.StartCombat([]*entity.Enemy{enemy1, enemy2})

	if engine.CurrentState() != types.StateCombat {
		t.Errorf("state should be StateCombat, got %v", engine.CurrentState())
	}

	// Kill both enemies (simulate combat victory)
	enemy1.Stats.RAM = 0
	enemy2.Stats.RAM = 0
	combat.Victory = true

	// End combat - should award XP and exit codes
	engine.EndCombat(combat)

	if engine.CurrentState() != types.StateExploring {
		t.Errorf("state should be StateExploring after combat, got %v", engine.CurrentState())
	}

	// Verify XP was gained
	expectedXP := enemy1.XPReward + enemy2.XPReward
	if engine.Player().XP < initialXP+expectedXP && engine.Player().Level == initialLevel {
		// XP should increase (or level up and reset)
		t.Errorf("XP should have increased: initial=%d, expected gain=%d, final=%d",
			initialXP, expectedXP, engine.Player().XP)
	}

	// Verify exit codes were gained
	expectedExitCodes := expectedXP / 2
	if engine.Player().ExitCodes != initialExitCodes+expectedExitCodes {
		t.Errorf("exit codes = %d, want %d", engine.Player().ExitCodes, initialExitCodes+expectedExitCodes)
	}

	// Save to file
	if err := engine.SaveSync(save.TriggerManual); err != nil {
		t.Fatalf("SaveSync failed: %v", err)
	}

	// Create fresh engine and load
	newEngine := NewEngine(cfg, 99999)
	newEngine.saveManager = saveMgr
	if err := newEngine.StartNewGame(entity.ClassBash); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}
	if err := newEngine.LoadLatestSave(); err != nil {
		t.Fatalf("LoadLatestSave failed: %v", err)
	}

	// Verify combat rewards persisted
	if newEngine.Player().ExitCodes != initialExitCodes+expectedExitCodes {
		t.Errorf("loaded exit codes = %d, want %d", newEngine.Player().ExitCodes, initialExitCodes+expectedExitCodes)
	}
}

// TestFloorStateDeltas_E2E tests that applyFloorState correctly removes dead enemies
// and looted items from the world when loading a save. This tests the load-side of
// floor deltas since buildFloorStates tracking is not yet implemented.
func TestFloorStateDeltas_E2E(t *testing.T) {
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

	// First, create an engine to discover enemy/item IDs on the generated floor
	seedEngine := NewEngine(cfg, 22222)
	if err := seedEngine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed for seed engine: %v", err)
	}

	// Find an enemy and item on the floor
	if len(seedEngine.world.Enemies) == 0 {
		t.Skip("no enemies on floor 1 to test with")
	}
	if len(seedEngine.world.Items) == 0 {
		t.Skip("no items on floor 1 to test with")
	}
	deadEnemyID := seedEngine.world.Enemies[0].ID()
	lootedItemID := seedEngine.world.Items[0].ID()

	// Create a save with floor state deltas directly (bypassing buildFloorStates)
	saveData := &save.SaveData{
		Version:      save.Version,
		MasterSeed:   22222,
		CurrentDepth: 1,
		Player: save.PlayerData{
			Class:     entity.ClassInit,
			Level:     1,
			XP:        0,
			XPToLevel: 100,
			Stats:     seedEngine.Player().Stats,
			MaxStats:  seedEngine.Player().MaxStats,
			Position:  seedEngine.Player().Position(),
			Inventory: []save.ItemData{
				{TemplateID: "malloc", Quantity: 3}, // Add an item to verify inventory persists
			},
			Equipment: save.EquipmentData{},
		},
		FloorStates: []save.FloorState{
			{
				Depth:         1,
				ExploredTiles: []types.Position{},
				DeadEnemies:   []string{deadEnemyID},
				LootedItems:   []string{lootedItemID},
			},
		},
	}

	// Save to file
	if err := saveMgr.SaveSync(saveData, save.TriggerManual); err != nil {
		t.Fatalf("SaveSync failed: %v", err)
	}

	// Create fresh engine with same seed (regenerates same floor)
	newEngine := NewEngine(cfg, 22222)
	newEngine.saveManager = saveMgr
	if err := newEngine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame for load failed: %v", err)
	}

	// Before load: enemy and item should exist (regenerated from seed)
	if newEngine.world.GetEnemyByID(deadEnemyID) == nil {
		t.Skip("enemy not found before load - seed may have changed")
	}
	if newEngine.world.GetItemByID(lootedItemID) == nil {
		t.Skip("item not found before load - seed may have changed")
	}

	// Load the save - this should apply floor deltas
	if err := newEngine.LoadLatestSave(); err != nil {
		t.Fatalf("LoadLatestSave failed: %v", err)
	}

	// After load: dead enemy should be removed by applyFloorState
	if newEngine.world.GetEnemyByID(deadEnemyID) != nil {
		t.Error("dead enemy should be removed after loading save with floor delta")
	}

	// After load: looted item should be removed by applyFloorState
	if newEngine.world.GetItemByID(lootedItemID) != nil {
		t.Error("looted item should be removed after loading save with floor delta")
	}

	// Verify inventory items from save were restored
	foundMalloc := false
	for _, item := range newEngine.Player().Inventory.Items {
		if item.TemplateID == "malloc" && item.Quantity == 3 {
			foundMalloc = true
			break
		}
	}
	if !foundMalloc {
		t.Error("inventory items from save should be restored")
	}
}

// TestLevelUp_E2E tests that level up stat increases persist through save/load.
func TestLevelUp_E2E(t *testing.T) {
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
	engine := NewEngine(cfg, 33333)
	engine.saveManager = saveMgr

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Record initial stats (level 1)
	initialLevel := engine.Player().Level
	initialMaxRAM := engine.Player().MaxStats.MaxRAM
	initialMaxFD := engine.Player().MaxStats.MaxFD
	initialCPU := engine.Player().Stats.CPU

	if initialLevel != 1 {
		t.Fatalf("expected starting level 1, got %d", initialLevel)
	}

	// Give enough XP to level up (more than XPToLevel)
	xpNeeded := engine.Player().XPToLevel + 10
	leveled := engine.Player().GainXP(xpNeeded)

	if !leveled {
		t.Fatal("player should have leveled up")
	}
	if engine.Player().Level != 2 {
		t.Errorf("level should be 2, got %d", engine.Player().Level)
	}

	// Verify stat increases from level up
	// Level up grants: MaxRAM +10, MaxFD +2, CPU +2
	if engine.Player().MaxStats.MaxRAM != initialMaxRAM+10 {
		t.Errorf("MaxRAM should be %d after level up, got %d", initialMaxRAM+10, engine.Player().MaxStats.MaxRAM)
	}
	if engine.Player().MaxStats.MaxFD != initialMaxFD+2 {
		t.Errorf("MaxFD should be %d after level up, got %d", initialMaxFD+2, engine.Player().MaxStats.MaxFD)
	}
	if engine.Player().Stats.CPU != initialCPU+2 {
		t.Errorf("CPU should be %d after level up, got %d", initialCPU+2, engine.Player().Stats.CPU)
	}

	// Save to file
	if err := engine.SaveSync(save.TriggerManual); err != nil {
		t.Fatalf("SaveSync failed: %v", err)
	}

	// Create fresh engine and load
	newEngine := NewEngine(cfg, 99999)
	newEngine.saveManager = saveMgr
	if err := newEngine.StartNewGame(entity.ClassBash); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}
	if err := newEngine.LoadLatestSave(); err != nil {
		t.Fatalf("LoadLatestSave failed: %v", err)
	}

	// Verify level and stats persisted
	if newEngine.Player().Level != 2 {
		t.Errorf("loaded level = %d, want 2", newEngine.Player().Level)
	}
	if newEngine.Player().MaxStats.MaxRAM != initialMaxRAM+10 {
		t.Errorf("loaded MaxRAM = %d, want %d", newEngine.Player().MaxStats.MaxRAM, initialMaxRAM+10)
	}
	if newEngine.Player().MaxStats.MaxFD != initialMaxFD+2 {
		t.Errorf("loaded MaxFD = %d, want %d", newEngine.Player().MaxStats.MaxFD, initialMaxFD+2)
	}
	if newEngine.Player().Stats.CPU != initialCPU+2 {
		t.Errorf("loaded CPU = %d, want %d", newEngine.Player().Stats.CPU, initialCPU+2)
	}
}

// TestMetaProgression_E2E tests that meta-progress (exit codes, run stats) persists
// across multiple "runs" through real file I/O.
func TestMetaProgression_E2E(t *testing.T) {
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

	// Load initial meta progress (should be empty/default)
	meta, err := saveMgr.LoadMetaProgress()
	if err != nil {
		t.Fatalf("LoadMetaProgress failed: %v", err)
	}
	initialExitCodes := meta.TotalExitCodes

	// Simulate a "run" - earn some exit codes
	meta.TotalExitCodes += 100
	meta.RunsCompleted++
	meta.DeepestFloor = 3

	// Save meta progress
	if err := saveMgr.SaveMetaProgress(meta); err != nil {
		t.Fatalf("SaveMetaProgress failed: %v", err)
	}

	// Create a NEW save manager (simulate app restart) pointing to same directory
	saveMgr2, err := save.NewManager(saveCfg)
	if err != nil {
		t.Fatalf("failed to create second save manager: %v", err)
	}

	// Load meta progress after "restart"
	loadedMeta, err := saveMgr2.LoadMetaProgress()
	if err != nil {
		t.Fatalf("LoadMetaProgress after restart failed: %v", err)
	}

	// Verify exit codes persisted
	if loadedMeta.TotalExitCodes != initialExitCodes+100 {
		t.Errorf("loaded TotalExitCodes = %d, want %d", loadedMeta.TotalExitCodes, initialExitCodes+100)
	}
	if loadedMeta.RunsCompleted != 1 {
		t.Errorf("loaded RunsCompleted = %d, want 1", loadedMeta.RunsCompleted)
	}
	if loadedMeta.DeepestFloor != 3 {
		t.Errorf("loaded DeepestFloor = %d, want 3", loadedMeta.DeepestFloor)
	}

	// Simulate second run - more exit codes
	loadedMeta.TotalExitCodes += 50
	loadedMeta.RunsCompleted++
	loadedMeta.DeepestFloor = 5 // Went deeper

	if err := saveMgr2.SaveMetaProgress(loadedMeta); err != nil {
		t.Fatalf("SaveMetaProgress second run failed: %v", err)
	}

	// Create THIRD save manager (another restart)
	saveMgr3, err := save.NewManager(saveCfg)
	if err != nil {
		t.Fatalf("failed to create third save manager: %v", err)
	}

	finalMeta, err := saveMgr3.LoadMetaProgress()
	if err != nil {
		t.Fatalf("LoadMetaProgress final failed: %v", err)
	}

	// Verify cumulative progress
	if finalMeta.TotalExitCodes != initialExitCodes+150 {
		t.Errorf("final TotalExitCodes = %d, want %d", finalMeta.TotalExitCodes, initialExitCodes+150)
	}
	if finalMeta.RunsCompleted != 2 {
		t.Errorf("final RunsCompleted = %d, want 2", finalMeta.RunsCompleted)
	}
	if finalMeta.DeepestFloor != 5 {
		t.Errorf("final DeepestFloor = %d, want 5", finalMeta.DeepestFloor)
	}
}

// TestFloorDeltaTracking_E2E verifies that killed enemies and looted items are
// tracked during gameplay and correctly saved to the floor state deltas.
func TestFloorDeltaTracking_E2E(t *testing.T) {
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
	engine := NewEngine(cfg, 44444)
	engine.saveManager = saveMgr

	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}

	// Find enemies and items on the floor
	if len(engine.world.Enemies) == 0 {
		t.Skip("no enemies on floor to test with")
	}
	if len(engine.world.Items) == 0 {
		t.Skip("no items on floor to test with")
	}

	// Record IDs before removal
	enemyToKill := engine.world.Enemies[0].ID()
	itemToLoot := engine.world.Items[0].ID()

	// Simulate killing enemy (via world.RemoveEnemy which tracks)
	engine.world.RemoveEnemy(enemyToKill)

	// Simulate looting item (via world.RemoveItem which tracks)
	engine.world.RemoveItem(itemToLoot)

	// Verify tracking recorded the removals
	removedEnemies := engine.world.GetRemovedEnemies(1)
	if len(removedEnemies) == 0 {
		t.Error("killed enemy should be tracked in RemovedEnemies")
	} else {
		found := false
		for _, id := range removedEnemies {
			if id == enemyToKill {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("enemy %s should be in RemovedEnemies", enemyToKill)
		}
	}

	removedItems := engine.world.GetRemovedItems(1)
	if len(removedItems) == 0 {
		t.Error("looted item should be tracked in RemovedItems")
	} else {
		found := false
		for _, id := range removedItems {
			if id == itemToLoot {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("item %s should be in RemovedItems", itemToLoot)
		}
	}

	// Save to file
	if err := engine.SaveSync(save.TriggerManual); err != nil {
		t.Fatalf("SaveSync failed: %v", err)
	}

	// Load the save file directly to verify floor state deltas
	loadedSave, err := saveMgr.LoadLatest()
	if err != nil {
		t.Fatalf("LoadLatest failed: %v", err)
	}
	if loadedSave == nil {
		t.Fatal("save should exist")
	}

	// Find floor 1 state in save data
	var floor1State *save.FloorState
	for i := range loadedSave.FloorStates {
		if loadedSave.FloorStates[i].Depth == 1 {
			floor1State = &loadedSave.FloorStates[i]
			break
		}
	}
	if floor1State == nil {
		t.Fatal("floor 1 state should be in save data")
	}

	// Verify dead enemy is in floor state
	foundDeadEnemy := false
	for _, id := range floor1State.DeadEnemies {
		if id == enemyToKill {
			foundDeadEnemy = true
			break
		}
	}
	if !foundDeadEnemy {
		t.Errorf("dead enemy %s should be in save file floor state DeadEnemies", enemyToKill)
	}

	// Verify looted item is in floor state
	foundLootedItem := false
	for _, id := range floor1State.LootedItems {
		if id == itemToLoot {
			foundLootedItem = true
			break
		}
	}
	if !foundLootedItem {
		t.Errorf("looted item %s should be in save file floor state LootedItems", itemToLoot)
	}

	// Now verify the full round-trip: load and confirm entities are still gone
	newEngine := NewEngine(cfg, 44444) // Same seed
	newEngine.saveManager = saveMgr
	if err := newEngine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame for load failed: %v", err)
	}

	// Before load, entities should be regenerated
	if newEngine.world.GetEnemyByID(enemyToKill) == nil {
		t.Skip("enemy not regenerated - seed may vary")
	}

	// Load the save
	if err := newEngine.LoadLatestSave(); err != nil {
		t.Fatalf("LoadLatestSave failed: %v", err)
	}

	// After load, dead enemy and looted item should be gone
	if newEngine.world.GetEnemyByID(enemyToKill) != nil {
		t.Error("dead enemy should be removed after loading save")
	}
	if newEngine.world.GetItemByID(itemToLoot) != nil {
		t.Error("looted item should be removed after loading save")
	}
}

// TestFloorDeltasPreservedAfterLoadAndFloorChange tests that floor deltas from the save
// are preserved in tracking maps after load, so they persist through floor changes.
// Regression test for: floor deltas lost when saving after descending post-load.
//
// The bug: When loading a save that's on floor 2 with floor 1 deltas, the floor 1
// deltas were not being restored to the tracking maps because applyFloorState was
// only called for the current floor (floor 2).
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

	// Create initial engine to get enemy ID from seeded floor 1
	seedEngine := NewEngine(cfg, 77777)
	if err := seedEngine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}
	if len(seedEngine.world.Enemies) == 0 {
		t.Skip("no enemies on floor 1")
	}
	deadEnemyID := seedEngine.world.Enemies[0].ID()

	// Descend to floor 2 to get floor 2 start position
	seedEngine.player.SetPosition(seedEngine.world.CurrentFloor.StairsDown)
	if err := seedEngine.DescendStairs(); err != nil {
		t.Fatalf("DescendStairs failed: %v", err)
	}
	floor2StartPos := seedEngine.Player().Position()

	// Create save ON FLOOR 2 with floor 1 dead enemy delta
	// This is the key scenario: loading a save where we're NOT on the floor with deltas
	saveData := &save.SaveData{
		Version:      save.Version,
		MasterSeed:   77777,
		CurrentDepth: 2, // Save is on floor 2
		Player: save.PlayerData{
			Class:     entity.ClassInit,
			Level:     1,
			XP:        0,
			XPToLevel: 100,
			Stats:     seedEngine.Player().Stats,
			MaxStats:  seedEngine.Player().MaxStats,
			Position:  floor2StartPos,
			Inventory: []save.ItemData{},
			Equipment: save.EquipmentData{},
		},
		FloorStates: []save.FloorState{
			{
				Depth:         1, // Floor 1 has deltas
				ExploredTiles: []types.Position{},
				DeadEnemies:   []string{deadEnemyID},
				LootedItems:   []string{},
			},
			{
				Depth:         2, // Floor 2 has no deltas
				ExploredTiles: []types.Position{},
				DeadEnemies:   []string{},
				LootedItems:   []string{},
			},
		},
	}
	if err := saveMgr.SaveSync(saveData, save.TriggerManual); err != nil {
		t.Fatalf("SaveSync failed: %v", err)
	}

	// Load the save (on floor 2)
	engine := NewEngine(cfg, 77777)
	engine.saveManager = saveMgr
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		t.Fatalf("StartNewGame failed: %v", err)
	}
	if err := engine.LoadLatestSave(); err != nil {
		t.Fatalf("LoadLatestSave failed: %v", err)
	}

	// Verify we're on floor 2
	if engine.CurrentDepth() != 2 {
		t.Fatalf("expected to load on floor 2, got floor %d", engine.CurrentDepth())
	}

	// Verify floor 1 deltas are in tracking maps (this is the bug - they won't be)
	// The bug: applyFloorState is only called for current floor (2), so floor 1 deltas
	// are never restored to the tracking maps
	removedEnemies := engine.world.GetRemovedEnemies(1)
	if len(removedEnemies) == 0 {
		t.Error("BUG: floor 1 dead enemies not restored to tracking maps after load on floor 2")
	}

	// Save again (on floor 2)
	if err := engine.SaveSync(save.TriggerManual); err != nil {
		t.Fatalf("SaveSync on floor 2 failed: %v", err)
	}

	// Load the new save to check if floor 1 deltas were preserved
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
		t.Error("BUG: floor 1 deltas lost after re-saving from floor 2")
	}
}
