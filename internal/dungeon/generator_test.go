package dungeon

import (
	"math/rand"
	"testing"

	"github.com/iheanyi/devdungeon/internal/game"
	"github.com/iheanyi/devdungeon/internal/types"
)

func TestDeterministicGeneration(t *testing.T) {
	gen := NewGenerator(DefaultConfig())

	// Generate two floors with the same seed
	floor1 := gen.Generate(types.FloorHome, 1, 12345)
	floor2 := gen.Generate(types.FloorHome, 1, 12345)

	// Verify floors are identical
	if floor1.Width != floor2.Width || floor1.Height != floor2.Height {
		t.Errorf("Dimensions differ: %dx%d vs %dx%d",
			floor1.Width, floor1.Height, floor2.Width, floor2.Height)
	}

	if len(floor1.Rooms) != len(floor2.Rooms) {
		t.Errorf("Room count differs: %d vs %d", len(floor1.Rooms), len(floor2.Rooms))
	}

	// Check all tiles match
	for y := 0; y < floor1.Height; y++ {
		for x := 0; x < floor1.Width; x++ {
			if floor1.Tiles[y][x].Type != floor2.Tiles[y][x].Type {
				t.Errorf("Tile mismatch at (%d, %d): %v vs %v",
					x, y, floor1.Tiles[y][x].Type, floor2.Tiles[y][x].Type)
			}
		}
	}

	// Check entrance and exit match
	if floor1.Entrance != floor2.Entrance {
		t.Errorf("Entrance differs: %v vs %v", floor1.Entrance, floor2.Entrance)
	}

	if floor1.Exit != floor2.Exit {
		t.Errorf("Exit differs: %v vs %v", floor1.Exit, floor2.Exit)
	}
}

func TestDifferentSeedsProduceDifferentFloors(t *testing.T) {
	gen := NewGenerator(DefaultConfig())

	floor1 := gen.Generate(types.FloorTmp, 3, 12345)
	floor2 := gen.Generate(types.FloorTmp, 3, 67890)

	// Floors should be different (could theoretically be same, but extremely unlikely)
	same := true
	for y := 0; y < floor1.Height && same; y++ {
		for x := 0; x < floor1.Width && same; x++ {
			if floor1.Tiles[y][x].Type != floor2.Tiles[y][x].Type {
				same = false
			}
		}
	}

	if same {
		t.Error("Different seeds produced identical floors")
	}
}

func TestRoomsAreConnected(t *testing.T) {
	gen := NewGenerator(DefaultConfig())
	floor := gen.Generate(types.FloorVar, 2, 42)

	if len(floor.Rooms) == 0 {
		t.Fatal("No rooms generated")
	}

	// All rooms should be marked as connected
	for i, room := range floor.Rooms {
		if !room.Connected {
			t.Errorf("Room %d is not marked as connected", i)
		}
	}

	// Verify reachability using flood fill from entrance
	visited := floodFill(floor, floor.Entrance)

	// Check that exit is reachable
	if !visited[floor.Exit.Y][floor.Exit.X] {
		t.Error("Exit is not reachable from entrance")
	}

	// Check that all room centers are reachable
	for i, room := range floor.Rooms {
		center := types.Position{X: room.X + room.Width/2, Y: room.Y + room.Height/2}
		if !visited[center.Y][center.X] {
			t.Errorf("Room %d center at %v is not reachable from entrance", i, center)
		}
	}
}

func TestStairsPlacement(t *testing.T) {
	gen := NewGenerator(DefaultConfig())
	floor := gen.Generate(types.FloorEtc, 1, 99999)

	// Entrance should have stairs up tile
	entranceTile := floor.Tiles[floor.Entrance.Y][floor.Entrance.X]
	if entranceTile.Type != types.TileStairsUp {
		t.Errorf("Entrance tile is %v, expected TileStairsUp", entranceTile.Type)
	}

	// Exit should have stairs down tile
	exitTile := floor.Tiles[floor.Exit.Y][floor.Exit.X]
	if exitTile.Type != types.TileStairsDown {
		t.Errorf("Exit tile is %v, expected TileStairsDown", exitTile.Type)
	}
}

func TestBossArena(t *testing.T) {
	gen := NewGenerator(DefaultConfig())
	floor := gen.Generate(types.FloorDevNull, 10, 11111)

	// Boss arena should have exactly one room
	if len(floor.Rooms) != 1 {
		t.Errorf("Boss arena has %d rooms, expected 1", len(floor.Rooms))
	}

	// The room should be large
	room := floor.Rooms[0]
	if room.Width < 50 || room.Height < 20 {
		t.Errorf("Boss room is too small: %dx%d", room.Width, room.Height)
	}
}

func TestRoomSizeConstraints(t *testing.T) {
	gen := NewGenerator(DefaultConfig())
	floor := gen.Generate(types.FloorUsr, 5, 54321)

	cfg := DefaultConfig()
	for i, room := range floor.Rooms {
		if room.Width < cfg.MinRoomSize {
			t.Errorf("Room %d width %d is less than minimum %d", i, room.Width, cfg.MinRoomSize)
		}
		if room.Height < cfg.MinRoomSize {
			t.Errorf("Room %d height %d is less than minimum %d", i, room.Height, cfg.MinRoomSize)
		}
		if room.Width > cfg.MaxRoomSize {
			t.Errorf("Room %d width %d exceeds maximum %d", i, room.Width, cfg.MaxRoomSize)
		}
		if room.Height > cfg.MaxRoomSize {
			t.Errorf("Room %d height %d exceeds maximum %d", i, room.Height, cfg.MaxRoomSize)
		}
	}
}

func TestPopulateFloor(t *testing.T) {
	// Create a floor using types.Floor for population
	rng := rand.New(rand.NewSource(12345))

	gen := NewGenerator(DefaultConfig())
	gameFloor := gen.Generate(types.FloorTmp, 3, 12345)

	// Convert back to types.Floor for populate test
	typesFloor := types.NewFloor(types.FloorTmp, 3, gameFloor.Width, gameFloor.Height, 12345)
	for y := 0; y < gameFloor.Height; y++ {
		for x := 0; x < gameFloor.Width; x++ {
			typesFloor.Tiles[y][x] = types.Tile{
				Type:        types.TileType(gameFloor.Tiles[y][x].Type),
				Visible:     gameFloor.Tiles[y][x].Visible,
				Explored:    gameFloor.Tiles[y][x].Seen,
				Blocked:     types.TileType(gameFloor.Tiles[y][x].Type) == types.TileWall,
				BlocksSight: types.TileType(gameFloor.Tiles[y][x].Type) == types.TileWall,
			}
		}
	}
	for _, r := range gameFloor.Rooms {
		typesFloor.Rooms = append(typesFloor.Rooms, types.Room{
			X:         r.X,
			Y:         r.Y,
			Width:     r.Width,
			Height:    r.Height,
			Connected: r.Connected,
		})
	}
	typesFloor.StairsUp = types.Position(gameFloor.Entrance)
	typesFloor.StairsDown = types.Position(gameFloor.Exit)
	typesFloor.PlayerStart = typesFloor.StairsUp

	data := PopulateFloor(rng, typesFloor, 3)

	// Should have enemies and items
	if len(data.Enemies) == 0 {
		t.Error("No enemies spawned")
	}

	if len(data.Items) == 0 {
		t.Error("No items spawned")
	}

	// All spawns should be on floor tiles, not on stairs
	for _, enemy := range data.Enemies {
		tile := typesFloor.GetTile(enemy.Position)
		if tile == nil || tile.Type != types.TileFloor {
			t.Errorf("Enemy spawned on non-floor tile at %v", enemy.Position)
		}
		if enemy.Position == typesFloor.StairsUp || enemy.Position == typesFloor.StairsDown {
			t.Errorf("Enemy spawned on stairs at %v", enemy.Position)
		}
	}

	for _, item := range data.Items {
		tile := typesFloor.GetTile(item.Position)
		if tile == nil || tile.Type != types.TileFloor {
			t.Errorf("Item spawned on non-floor tile at %v", item.Position)
		}
	}
}

func TestPopulateFloorDeterministic(t *testing.T) {
	gen := NewGenerator(DefaultConfig())

	// Generate the same floor twice
	gameFloor1 := gen.Generate(types.FloorHome, 1, 11111)
	gameFloor2 := gen.Generate(types.FloorHome, 1, 11111)

	// Create types.Floor for population
	typesFloor1 := convertGameFloorToTypesFloor(gameFloor1)
	typesFloor2 := convertGameFloorToTypesFloor(gameFloor2)

	// Populate with same seed
	rng1 := rand.New(rand.NewSource(22222))
	rng2 := rand.New(rand.NewSource(22222))

	data1 := PopulateFloor(rng1, typesFloor1, 1)
	data2 := PopulateFloor(rng2, typesFloor2, 1)

	// Should produce identical spawns
	if len(data1.Enemies) != len(data2.Enemies) {
		t.Errorf("Enemy count differs: %d vs %d", len(data1.Enemies), len(data2.Enemies))
	}

	if len(data1.Items) != len(data2.Items) {
		t.Errorf("Item count differs: %d vs %d", len(data1.Items), len(data2.Items))
	}

	for i := range data1.Enemies {
		if data1.Enemies[i].Position != data2.Enemies[i].Position {
			t.Errorf("Enemy %d position differs: %v vs %v",
				i, data1.Enemies[i].Position, data2.Enemies[i].Position)
		}
		if data1.Enemies[i].TemplateID != data2.Enemies[i].TemplateID {
			t.Errorf("Enemy %d template differs: %v vs %v",
				i, data1.Enemies[i].TemplateID, data2.Enemies[i].TemplateID)
		}
	}
}

func TestFloorTypeVariations(t *testing.T) {
	gen := NewGenerator(DefaultConfig())

	// Generate floors of different types
	homeFloor := gen.Generate(types.FloorHome, 1, 12345)
	tmpFloor := gen.Generate(types.FloorTmp, 1, 12345)
	devNullFloor := gen.Generate(types.FloorDevNull, 1, 12345)

	// Home should have at least some rooms (BSP produces variable count)
	if len(homeFloor.Rooms) < 3 {
		t.Errorf("Home floor has too few rooms: %d", len(homeFloor.Rooms))
	}

	// DevNull should have exactly 1 room (boss arena)
	if len(devNullFloor.Rooms) != 1 {
		t.Errorf("DevNull floor has %d rooms, expected 1", len(devNullFloor.Rooms))
	}

	// Tmp might have more varied rooms (harder to test precisely)
	if len(tmpFloor.Rooms) == 0 {
		t.Error("Tmp floor has no rooms")
	}
}

// floodFill performs flood fill from a starting position and returns a 2D visited map.
func floodFill(floor *game.Floor, start types.Position) [][]bool {
	visited := make([][]bool, floor.Height)
	for y := range visited {
		visited[y] = make([]bool, floor.Width)
	}

	stack := []types.Position{start}

	for len(stack) > 0 {
		pos := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if pos.X < 0 || pos.X >= floor.Width || pos.Y < 0 || pos.Y >= floor.Height {
			continue
		}

		if visited[pos.Y][pos.X] {
			continue
		}

		tile := floor.Tiles[pos.Y][pos.X]
		if tile.Type == types.TileWall {
			continue
		}

		visited[pos.Y][pos.X] = true

		// Add neighbors
		stack = append(stack,
			types.Position{X: pos.X + 1, Y: pos.Y},
			types.Position{X: pos.X - 1, Y: pos.Y},
			types.Position{X: pos.X, Y: pos.Y + 1},
			types.Position{X: pos.X, Y: pos.Y - 1},
		)
	}

	return visited
}

// convertGameFloorToTypesFloor converts a game.Floor to types.Floor for testing.
func convertGameFloorToTypesFloor(gameFloor *game.Floor) *types.Floor {
	typesFloor := types.NewFloor(types.FloorType(gameFloor.Type), gameFloor.Level, gameFloor.Width, gameFloor.Height, 0)
	for y := 0; y < gameFloor.Height; y++ {
		for x := 0; x < gameFloor.Width; x++ {
			typesFloor.Tiles[y][x] = types.Tile{
				Type:        types.TileType(gameFloor.Tiles[y][x].Type),
				Visible:     gameFloor.Tiles[y][x].Visible,
				Explored:    gameFloor.Tiles[y][x].Seen,
				Blocked:     types.TileType(gameFloor.Tiles[y][x].Type) == types.TileWall,
				BlocksSight: types.TileType(gameFloor.Tiles[y][x].Type) == types.TileWall,
			}
		}
	}
	for _, r := range gameFloor.Rooms {
		typesFloor.Rooms = append(typesFloor.Rooms, types.Room{
			X:         r.X,
			Y:         r.Y,
			Width:     r.Width,
			Height:    r.Height,
			Connected: r.Connected,
		})
	}
	typesFloor.StairsUp = types.Position(gameFloor.Entrance)
	typesFloor.StairsDown = types.Position(gameFloor.Exit)
	typesFloor.PlayerStart = typesFloor.StairsUp
	return typesFloor
}

// Benchmark tests
func BenchmarkGenerate(b *testing.B) {
	gen := NewGenerator(DefaultConfig())
	for i := 0; i < b.N; i++ {
		gen.Generate(types.FloorVar, 5, int64(i))
	}
}

func BenchmarkPopulate(b *testing.B) {
	gen := NewGenerator(DefaultConfig())
	gameFloor := gen.Generate(types.FloorVar, 5, 12345)
	typesFloor := convertGameFloorToTypesFloor(gameFloor)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewSource(int64(i)))
		PopulateFloor(rng, typesFloor, 5)
	}
}
