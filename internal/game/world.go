package game

import (
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/types"
)

// GameWorld manages the game world state including floors, enemies, and items.
// It implements the World interface.
type GameWorld struct {
	CurrentFloor *types.Floor        // The currently active floor
	Enemies      []*entity.Enemy     // Enemies on the current floor
	Items        []*entity.Item      // Items on the current floor
	FloorCache   map[int]*FloorState // Cache of visited floors with their state
	CurrentDepth int                 // Current depth (1 = first floor)

	// Track removed entities for save/load floor deltas
	// Key is floor depth, value is list of entity IDs removed from that floor
	RemovedEnemies map[int][]string // Dead enemy IDs by floor depth
	RemovedItems   map[int][]string // Looted item IDs by floor depth
}

// FloorState stores the complete state of a floor for caching.
type FloorState struct {
	Floor   *types.Floor
	Enemies []*entity.Enemy
	Items   []*entity.Item
}

// NewGameWorld creates a new empty game world.
func NewGameWorld() *GameWorld {
	return &GameWorld{
		Enemies:        make([]*entity.Enemy, 0),
		Items:          make([]*entity.Item, 0),
		FloorCache:     make(map[int]*FloorState),
		CurrentDepth:   0,
		RemovedEnemies: make(map[int][]string),
		RemovedItems:   make(map[int][]string),
	}
}

// SetFloor sets the current floor and initializes the world state.
func (w *GameWorld) SetFloor(floor *types.Floor, enemies []*entity.Enemy, items []*entity.Item) {
	w.CurrentFloor = floor
	w.Enemies = enemies
	w.Items = items
	w.CurrentDepth = floor.Depth
}

// GetCurrentFloor returns the current floor.
func (w *GameWorld) GetCurrentFloor() *types.Floor {
	return w.CurrentFloor
}

// CacheCurrentFloor saves the current floor state to the cache.
func (w *GameWorld) CacheCurrentFloor() {
	if w.CurrentFloor == nil {
		return
	}
	w.FloorCache[w.CurrentDepth] = &FloorState{
		Floor:   w.CurrentFloor,
		Enemies: w.Enemies,
		Items:   w.Items,
	}
}

// LoadCachedFloor loads a floor from the cache if it exists.
func (w *GameWorld) LoadCachedFloor(depth int) bool {
	state, ok := w.FloorCache[depth]
	if !ok {
		return false
	}
	w.CurrentFloor = state.Floor
	w.Enemies = state.Enemies
	w.Items = state.Items
	w.CurrentDepth = depth
	return true
}

// AddEnemy adds an enemy to the current floor.
func (w *GameWorld) AddEnemy(enemy *entity.Enemy) {
	w.Enemies = append(w.Enemies, enemy)
}

// RemoveEnemy removes an enemy from the current floor by ID.
// Also tracks the removal for save/load floor deltas.
func (w *GameWorld) RemoveEnemy(id string) bool {
	for i, e := range w.Enemies {
		if e.ID() == id {
			w.Enemies = append(w.Enemies[:i], w.Enemies[i+1:]...)
			// Track removal for floor deltas
			if w.RemovedEnemies == nil {
				w.RemovedEnemies = make(map[int][]string)
			}
			w.RemovedEnemies[w.CurrentDepth] = append(w.RemovedEnemies[w.CurrentDepth], id)
			return true
		}
	}
	return false
}

// GetEnemyAt returns the enemy at the given position, or nil if none.
func (w *GameWorld) GetEnemyAt(pos types.Position) *entity.Enemy {
	for _, e := range w.Enemies {
		if e.Position() == pos && e.IsAlive() {
			return e
		}
	}
	return nil
}

// GetEnemyByID returns the enemy with the given ID, or nil if not found.
func (w *GameWorld) GetEnemyByID(id string) *entity.Enemy {
	for _, e := range w.Enemies {
		if e.ID() == id {
			return e
		}
	}
	return nil
}

// AddItem adds an item to the current floor.
func (w *GameWorld) AddItem(item *entity.Item) {
	w.Items = append(w.Items, item)
}

// RemoveItem removes an item from the current floor by ID.
// Also tracks the removal for save/load floor deltas.
func (w *GameWorld) RemoveItem(id string) bool {
	for i, item := range w.Items {
		if item.ID() == id {
			w.Items = append(w.Items[:i], w.Items[i+1:]...)
			// Track removal for floor deltas
			if w.RemovedItems == nil {
				w.RemovedItems = make(map[int][]string)
			}
			w.RemovedItems[w.CurrentDepth] = append(w.RemovedItems[w.CurrentDepth], id)
			return true
		}
	}
	return false
}

// GetItemAt returns the item at the given position, or nil if none.
func (w *GameWorld) GetItemAt(pos types.Position) *entity.Item {
	for _, item := range w.Items {
		if item.Position() == pos {
			return item
		}
	}
	return nil
}

// GetItemByID returns the item with the given ID, or nil if not found.
func (w *GameWorld) GetItemByID(id string) *entity.Item {
	for _, item := range w.Items {
		if item.ID() == id {
			return item
		}
	}
	return nil
}

// GetRemovedEnemies returns the list of enemy IDs removed from a floor.
func (w *GameWorld) GetRemovedEnemies(depth int) []string {
	if w.RemovedEnemies == nil {
		return nil
	}
	return w.RemovedEnemies[depth]
}

// GetRemovedItems returns the list of item IDs removed from a floor.
func (w *GameWorld) GetRemovedItems(depth int) []string {
	if w.RemovedItems == nil {
		return nil
	}
	return w.RemovedItems[depth]
}

// SetRemovedEntities sets the removed entities from saved floor state.
// Used when loading a save to restore tracking state.
func (w *GameWorld) SetRemovedEntities(depth int, enemies []string, items []string) {
	if w.RemovedEnemies == nil {
		w.RemovedEnemies = make(map[int][]string)
	}
	if w.RemovedItems == nil {
		w.RemovedItems = make(map[int][]string)
	}
	w.RemovedEnemies[depth] = enemies
	w.RemovedItems[depth] = items
}

// ClearRemovedEntities clears all entity removal tracking.
// Used when starting a new game.
func (w *GameWorld) ClearRemovedEntities() {
	w.RemovedEnemies = make(map[int][]string)
	w.RemovedItems = make(map[int][]string)
}

// IsPositionBlocked checks if a position is blocked by walls or entities.
func (w *GameWorld) IsPositionBlocked(pos types.Position) bool {
	// Check floor walkability
	if w.CurrentFloor == nil || !w.CurrentFloor.IsWalkable(pos) {
		return true
	}
	// Check for blocking enemies
	if enemy := w.GetEnemyAt(pos); enemy != nil && enemy.IsBlocking() {
		return true
	}
	return false
}

// GetFloorType returns the type of the current floor.
func (w *GameWorld) GetFloorType() types.FloorType {
	if w.CurrentFloor == nil {
		return types.FloorHome
	}
	return w.CurrentFloor.Type
}

// UpdateVisibility updates the visibility of tiles around a position.
// For now, this is a simple circular FOV. Can be enhanced later.
func (w *GameWorld) UpdateVisibility(center types.Position, radius int) {
	if w.CurrentFloor == nil {
		return
	}

	// First, mark all tiles as not visible
	for y := 0; y < w.CurrentFloor.Height; y++ {
		for x := 0; x < w.CurrentFloor.Width; x++ {
			w.CurrentFloor.Tiles[y][x].Visible = false
		}
	}

	// Simple circular FOV
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			// Check if within radius (circular)
			if dx*dx+dy*dy > radius*radius {
				continue
			}

			pos := types.Position{X: center.X + dx, Y: center.Y + dy}
			if !w.CurrentFloor.InBounds(pos) {
				continue
			}

			// Simple line of sight check
			if w.hasLineOfSight(center, pos) {
				tile := w.CurrentFloor.GetTile(pos)
				if tile != nil {
					tile.Visible = true
					tile.Explored = true
				}
			}
		}
	}
}

// hasLineOfSight checks if there's a clear line of sight between two positions.
// Uses a simple Bresenham-like algorithm.
func (w *GameWorld) hasLineOfSight(from, to types.Position) bool {
	dx := abs(to.X - from.X)
	dy := abs(to.Y - from.Y)
	sx := sign(to.X - from.X)
	sy := sign(to.Y - from.Y)

	x, y := from.X, from.Y
	err := dx - dy

	for {
		// Reached destination
		if x == to.X && y == to.Y {
			return true
		}

		// Check if current tile blocks sight (skip start position)
		if !(x == from.X && y == from.Y) {
			if !w.CurrentFloor.IsTransparent(types.Position{X: x, Y: y}) {
				return false
			}
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

// GetVisibleTiles returns the visible tiles for rendering.
func (w *GameWorld) GetVisibleTiles() [][]types.Tile {
	if w.CurrentFloor == nil {
		return nil
	}
	return w.CurrentFloor.Tiles
}

// abs returns the absolute value of an integer.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// sign returns -1, 0, or 1 depending on the sign of n.
func sign(n int) int {
	if n < 0 {
		return -1
	}
	if n > 0 {
		return 1
	}
	return 0
}
