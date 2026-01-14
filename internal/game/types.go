// Package game provides the core game logic for /dev/dungeon.
package game

import "github.com/iheanyi/devdungeon/internal/types"

// Re-export commonly used types for convenience
type (
	Position   = types.Position
	Direction  = types.Direction
	Stats      = types.Stats
	MaxStats   = types.MaxStats
	GameState  = types.GameState
	FloorType  = types.FloorType
	TileType   = types.TileType
	ActionType = types.ActionType
)

// Re-export constants
const (
	DirNone  = types.DirNone
	DirUp    = types.DirUp
	DirDown  = types.DirDown
	DirLeft  = types.DirLeft
	DirRight = types.DirRight

	StateMainMenu  = types.StateMainMenu
	StateExploring = types.StateExploring
	StateCombat    = types.StateCombat
	StateInventory = types.StateInventory
	StatePaused    = types.StatePaused
	StateGameOver  = types.StateGameOver
	StateVictory   = types.StateVictory

	FloorHome    = types.FloorHome
	FloorTmp     = types.FloorTmp
	FloorVar     = types.FloorVar
	FloorEtc     = types.FloorEtc
	FloorUsr     = types.FloorUsr
	FloorSys     = types.FloorSys
	FloorDev     = types.FloorDev
	FloorDevNull = types.FloorDevNull

	TileWall       = types.TileWall
	TileFloor      = types.TileFloor
	TileDoor       = types.TileDoor
	TileStairsDown = types.TileStairsDown
	TileStairsUp   = types.TileStairsUp
	TileWater      = types.TileWater
	TileVoid       = types.TileVoid

	ActionAttack  = types.ActionAttack
	ActionHack    = types.ActionHack
	ActionUseItem = types.ActionUseItem
	ActionFlee    = types.ActionFlee
)

// Action represents a combat action.
type Action struct {
	Type     ActionType
	TargetID string
	ItemID   string
}

// Result represents the outcome of an action.
type Result struct {
	Success    bool
	Damage     int
	Message    string
	TargetDied bool
	PlayerFled bool
	ItemUsed   string
	ExtraTurns int
}

// Rewards represents rewards from combat.
type Rewards struct {
	ExitCodes int      // Meta currency
	Items     []string // Item IDs
	XP        int
}
