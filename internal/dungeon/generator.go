// Package dungeon provides procedural dungeon generation for /dev/dungeon.
package dungeon

import (
	"math/rand"

	"github.com/iheanyi/devdungeon/internal/game"
	"github.com/iheanyi/devdungeon/internal/types"
)

// Config holds dungeon generation parameters.
type Config struct {
	// Floor dimensions
	Width  int
	Height int

	// BSP parameters
	MinCellSize  int // Minimum cell size before stopping splits
	MaxCellSize  int // Maximum cell size (forces split if larger)
	MinRoomSize  int // Minimum room dimensions
	MaxRoomSize  int // Maximum room dimensions
	RoomPadding  int // Minimum space between room and cell edge
	SplitChance  int // Percentage chance to split (0-100)

	// Population parameters
	MinRooms int
	MaxRooms int
}

// DefaultConfig returns default generation parameters.
func DefaultConfig() Config {
	return Config{
		Width:        80,
		Height:       40,
		MinCellSize:  10,
		MaxCellSize:  25,
		MinRoomSize:  4,
		MaxRoomSize:  10,
		RoomPadding:  1,
		SplitChance:  75,
		MinRooms:     5,
		MaxRooms:     12,
	}
}

// Generator implements the game.DungeonGenerator interface.
type Generator struct {
	config Config
}

// NewGenerator creates a new dungeon generator with the given config.
func NewGenerator(config Config) *Generator {
	return &Generator{config: config}
}

// Generate creates a new floor with the given parameters.
// The seed ensures reproducible generation.
func (g *Generator) Generate(floorType types.FloorType, depth int, seed int64) *game.Floor {
	rng := rand.New(rand.NewSource(seed))

	// Get floor-specific config adjustments
	cfg := g.getFloorConfig(floorType)

	// Create empty floor filled with walls
	floor := types.NewFloor(floorType, depth, cfg.Width, cfg.Height, seed)

	// Generate dungeon using BSP
	if floorType == types.FloorDevNull {
		// Boss arena: single large room
		g.generateBossArena(rng, floor)
	} else {
		// Normal dungeon: BSP-based rooms
		g.generateBSPDungeon(rng, floor, cfg)
	}

	// Place stairs
	g.placeStairs(floor)

	// Convert types.Floor to game.Floor
	return g.convertToGameFloor(floor)
}

// getFloorConfig adjusts config based on floor type.
func (g *Generator) getFloorConfig(floorType types.FloorType) Config {
	cfg := g.config

	switch floorType {
	case types.FloorHome:
		// Tutorial: smaller, simpler dungeon
		cfg.MinRooms = 3
		cfg.MaxRooms = 6
	case types.FloorTmp:
		// Chaotic: more varied room sizes
		cfg.MinRoomSize = 3
		cfg.MaxRoomSize = 12
		cfg.SplitChance = 85
	case types.FloorDevNull:
		// Boss arena: handled separately
	default:
		// Use default config
	}

	return cfg
}

// generateBSPDungeon creates a dungeon using Binary Space Partitioning.
func (g *Generator) generateBSPDungeon(rng *rand.Rand, floor *types.Floor, cfg Config) {
	// Create root BSP node covering the entire floor (with 1-tile border)
	root := &BSPNode{
		X:      1,
		Y:      1,
		Width:  floor.Width - 2,
		Height: floor.Height - 2,
	}

	// Recursively partition the space
	root.Split(rng, cfg)

	// Collect all leaf nodes (cells that will contain rooms)
	leaves := root.GetLeaves()

	// Create rooms in each leaf
	for _, leaf := range leaves {
		room := createRoomInCell(rng, leaf, cfg)
		if room != nil {
			floor.Rooms = append(floor.Rooms, *room)
			carveRoom(floor, room)
		}
	}

	// Connect rooms with corridors
	connectRooms(rng, floor, root)
}

// generateBossArena creates a single large room for the final boss.
func (g *Generator) generateBossArena(rng *rand.Rand, floor *types.Floor) {
	// Create a large central room
	roomWidth := floor.Width - 10
	roomHeight := floor.Height - 10
	roomX := (floor.Width - roomWidth) / 2
	roomY := (floor.Height - roomHeight) / 2

	room := types.Room{
		X:         roomX,
		Y:         roomY,
		Width:     roomWidth,
		Height:    roomHeight,
		Connected: true,
	}

	floor.Rooms = append(floor.Rooms, room)
	carveRoom(floor, &room)
}

// placeStairs places up and down stairs in the dungeon.
func (g *Generator) placeStairs(floor *types.Floor) {
	if len(floor.Rooms) == 0 {
		return
	}

	// Stairs up in first room (player start)
	firstRoom := floor.Rooms[0]
	floor.StairsUp = firstRoom.Center()
	floor.PlayerStart = floor.StairsUp
	floor.SetTile(floor.StairsUp, types.NewTile(types.TileStairsUp))

	// No stairs down on boss floor (/dev/null) - it's the final floor
	if floor.Type == types.FloorDevNull {
		return
	}

	// Stairs down in last room (exit)
	lastRoom := floor.Rooms[len(floor.Rooms)-1]
	floor.StairsDown = lastRoom.Center()

	// If same room, offset the stairs
	if floor.StairsUp == floor.StairsDown {
		floor.StairsDown.X++
		floor.StairsDown.Y++
	}
	floor.SetTile(floor.StairsDown, types.NewTile(types.TileStairsDown))
}

// carveRoom carves a room into the floor by setting tiles to floor type.
func carveRoom(floor *types.Floor, room *types.Room) {
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			floor.SetTile(types.Position{X: x, Y: y}, types.NewTile(types.TileFloor))
		}
	}
}

// convertToGameFloor converts a types.Floor to a game.Floor.
func (g *Generator) convertToGameFloor(floor *types.Floor) *game.Floor {
	// Convert tiles
	tiles := make([][]game.Tile, floor.Height)
	for y := range tiles {
		tiles[y] = make([]game.Tile, floor.Width)
		for x := range tiles[y] {
			tiles[y][x] = game.Tile{
				Type:    game.TileType(floor.Tiles[y][x].Type),
				Visible: floor.Tiles[y][x].Visible,
				Seen:    floor.Tiles[y][x].Explored,
			}
		}
	}

	// Convert rooms
	rooms := make([]*game.Room, len(floor.Rooms))
	for i, r := range floor.Rooms {
		rooms[i] = &game.Room{
			X:         r.X,
			Y:         r.Y,
			Width:     r.Width,
			Height:    r.Height,
			Connected: r.Connected,
		}
	}

	return &game.Floor{
		Type:     game.FloorType(floor.Type),
		Level:    floor.Depth,
		Tiles:    tiles,
		Width:    floor.Width,
		Height:   floor.Height,
		Rooms:    rooms,
		Entities: nil, // Populated separately
		Entrance: game.Position(floor.StairsUp),
		Exit:     game.Position(floor.StairsDown),
	}
}

// Ensure Generator implements game.DungeonGenerator.
var _ game.DungeonGenerator = (*Generator)(nil)
