// Package types provides shared types for /dev/dungeon.
// This package should have no internal dependencies to avoid import cycles.
package types

// Position represents a 2D coordinate in the game world.
type Position struct {
	X, Y int
}

// Direction represents movement directions.
type Direction int

const (
	DirNone Direction = iota
	DirUp
	DirDown
	DirLeft
	DirRight
)

// Stats represents the core stats for entities.
// Modeled after real Linux process attributes.
type Stats struct {
	RAM  int // Health - memory allocation (OOM = death)
	CPU  int // Attack power - processing cycles
	FD   int // Ability resource - file descriptors (mana)
	NICE int // Speed/priority (lower = faster, accurate to Linux)
	UID  int // Access level - permissions (0 = root = max power)
}

// MaxStats represents the maximum values for stats.
type MaxStats struct {
	MaxRAM int // Maximum memory allocation
	MaxFD  int // Maximum file descriptors (like ulimit -n)
}

// GameState represents the current state of the game.
type GameState int

const (
	StateMainMenu GameState = iota
	StateExploring
	StateCombat
	StateInventory
	StatePaused
	StateGameOver
	StateVictory
)

// ActionType represents the type of action in combat.
type ActionType int

const (
	ActionAttack ActionType = iota
	ActionHack
	ActionUseItem
	ActionFlee
)

// FloorType represents different dungeon zones.
type FloorType int

const (
	FloorHome FloorType = iota
	FloorTmp
	FloorVar
	FloorEtc
	FloorUsr
	FloorSys
	FloorDev
	FloorDevNull // Final boss
)

// FloorName returns the display name for a floor type.
func (f FloorType) FloorName() string {
	names := map[FloorType]string{
		FloorHome:    "/home",
		FloorTmp:     "/tmp",
		FloorVar:     "/var",
		FloorEtc:     "/etc",
		FloorUsr:     "/usr",
		FloorSys:     "/sys",
		FloorDev:     "/dev",
		FloorDevNull: "/dev/null",
	}
	return names[f]
}

// TileType represents the type of tile.
type TileType int

const (
	TileWall TileType = iota
	TileFloor
	TileDoor
	TileStairsDown
	TileStairsUp
	TileWater
	TileVoid
)

// TileGlyph returns the ASCII representation of a tile.
func (t TileType) TileGlyph() rune {
	glyphs := map[TileType]rune{
		TileWall:       '#',
		TileFloor:      '.',
		TileDoor:       '+',
		TileStairsDown: '>',
		TileStairsUp:   '<',
		TileWater:      '~',
		TileVoid:       ' ',
	}
	return glyphs[t]
}

// Tile represents a single cell in the dungeon.
type Tile struct {
	Type        TileType
	Visible     bool // Currently in player's FOV
	Explored    bool // Has been seen before
	Blocked     bool // Cannot walk through
	BlocksSight bool // Blocks line of sight
}

// NewTile creates a tile with appropriate properties for its type.
func NewTile(tileType TileType) Tile {
	blocked := tileType == TileWall || tileType == TileVoid
	blocksSight := tileType == TileWall
	return Tile{
		Type:        tileType,
		Visible:     false,
		Explored:    false,
		Blocked:     blocked,
		BlocksSight: blocksSight,
	}
}

// Room represents a rectangular room in the dungeon.
type Room struct {
	X, Y          int // Top-left corner
	Width, Height int
	Connected     bool // Whether this room is connected to the dungeon
}

// Center returns the center position of the room.
func (r Room) Center() Position {
	return Position{
		X: r.X + r.Width/2,
		Y: r.Y + r.Height/2,
	}
}

// Contains checks if a position is inside the room.
func (r Room) Contains(p Position) bool {
	return p.X >= r.X && p.X < r.X+r.Width &&
		p.Y >= r.Y && p.Y < r.Y+r.Height
}

// Intersects checks if this room overlaps with another.
func (r Room) Intersects(other Room) bool {
	return r.X < other.X+other.Width && r.X+r.Width > other.X &&
		r.Y < other.Y+other.Height && r.Y+r.Height > other.Y
}

// Floor represents a single dungeon level.
type Floor struct {
	Type       FloorType
	Depth      int // 1 = first floor, increases as you descend
	Width      int
	Height     int
	Tiles      [][]Tile
	Rooms      []Room
	Seed       int64    // For reproducibility
	PlayerStart Position
	StairsUp   Position
	StairsDown Position
}

// NewFloor creates an empty floor filled with walls.
func NewFloor(floorType FloorType, depth, width, height int, seed int64) *Floor {
	tiles := make([][]Tile, height)
	for y := range tiles {
		tiles[y] = make([]Tile, width)
		for x := range tiles[y] {
			tiles[y][x] = NewTile(TileWall)
		}
	}
	return &Floor{
		Type:   floorType,
		Depth:  depth,
		Width:  width,
		Height: height,
		Tiles:  tiles,
		Rooms:  make([]Room, 0),
		Seed:   seed,
	}
}

// InBounds checks if a position is within the floor boundaries.
func (f *Floor) InBounds(p Position) bool {
	return p.X >= 0 && p.X < f.Width && p.Y >= 0 && p.Y < f.Height
}

// GetTile returns the tile at a position, or nil if out of bounds.
func (f *Floor) GetTile(p Position) *Tile {
	if !f.InBounds(p) {
		return nil
	}
	return &f.Tiles[p.Y][p.X]
}

// SetTile sets the tile at a position.
func (f *Floor) SetTile(p Position, tile Tile) {
	if f.InBounds(p) {
		f.Tiles[p.Y][p.X] = tile
	}
}

// IsWalkable checks if a position can be walked on.
func (f *Floor) IsWalkable(p Position) bool {
	tile := f.GetTile(p)
	return tile != nil && !tile.Blocked
}

// IsTransparent checks if a position allows line of sight.
func (f *Floor) IsTransparent(p Position) bool {
	tile := f.GetTile(p)
	return tile != nil && !tile.BlocksSight
}
