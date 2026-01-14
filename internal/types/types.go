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
type Stats struct {
	PID  int // Health/HP
	CPU  int // Attack power
	MEM  int // Ability capacity / mana
	NICE int // Speed/priority (lower = faster)
	UID  int // Permission level
}

// MaxStats represents the maximum values for stats.
type MaxStats struct {
	MaxPID int
	MaxMEM int
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
