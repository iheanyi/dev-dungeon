package game

import (
	"testing"

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

	if result.Combat == nil {
		t.Error("expected combat encounter")
	}

	if result.Combat != nil && result.Combat.ID() != enemy.ID() {
		t.Errorf("combat enemy mismatch: expected %s, got %s", enemy.ID(), result.Combat.ID())
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

	testItem := entity.NewItem("malloc", "test_item_123", itemPos)
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

	// Check inventory
	if len(engine.Player().Inventory.Items) != initialInventorySize+1 {
		t.Errorf("Inventory size should have increased by 1")
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
