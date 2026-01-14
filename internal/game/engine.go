package game

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/save"
	"github.com/iheanyi/devdungeon/internal/types"
)

// MoveResult represents the outcome of a player movement action.
type MoveResult struct {
	Moved      bool           // Whether the player actually moved
	Combat     *entity.Enemy  // Non-nil if bumped into an enemy
	PickedUp   *entity.Item   // Non-nil if picked up an item
	UsedStairs bool           // Whether stairs were used
	Message    string         // Status message to display
}

// Engine is the core game engine that manages game state and logic.
// RunStats tracks statistics for the current run.
type RunStats struct {
	EnemiesKilled   map[string]int // Kills by enemy type
	TotalKills      int
	DamageDealt     int
	DamageTaken     int
	ItemsCollected  int
	ItemsUsed       int
	FloorsExplored  int
	MaxDepthReached int
	StepsWalked     int
}

// NewRunStats creates a new run stats tracker.
func NewRunStats() *RunStats {
	return &RunStats{
		EnemiesKilled: make(map[string]int),
	}
}

type Engine struct {
	config      *config.Config
	player      *entity.Player
	world       *GameWorld
	generator   DungeonGenerator
	state       types.GameState
	rng         *rand.Rand
	masterSeed  int64
	messages    []string // Recent game messages
	saveManager *save.Manager
	stats       *RunStats // Current run statistics
}

// NewEngine creates a new game engine with the given configuration and seed.
func NewEngine(cfg *config.Config, seed int64) *Engine {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	// Initialize save manager
	saveMgr, err := save.NewManager(save.DefaultConfig())
	if err != nil {
		// Log error but continue without saves
		saveMgr = nil
	}

	engine := &Engine{
		config:      cfg,
		world:       NewGameWorld(),
		state:       types.StateMainMenu,
		rng:         rand.New(rand.NewSource(seed)),
		masterSeed:  seed,
		messages:    make([]string, 0),
		saveManager: saveMgr,
	}

	// Start background save goroutine
	if saveMgr != nil {
		saveMgr.Start()
	}

	return engine
}

// SetGenerator sets the dungeon generator for the engine.
func (e *Engine) SetGenerator(gen DungeonGenerator) {
	e.generator = gen
}

// StartNewGame initializes a new game with the given player class.
func (e *Engine) StartNewGame(playerClass entity.PlayerClass) error {
	return e.StartNewGameWithBonuses(playerClass, entity.PermanentBonuses{})
}

// StartNewGameWithBonuses initializes a new game with the given player class and permanent bonuses.
func (e *Engine) StartNewGameWithBonuses(playerClass entity.PlayerClass, bonuses entity.PermanentBonuses) error {
	// Create the player with bonuses
	e.player = entity.NewPlayerWithBonuses(playerClass, bonuses)

	// Initialize run statistics
	e.stats = NewRunStats()
	e.stats.FloorsExplored = 1
	e.stats.MaxDepthReached = 1

	// Generate the first floor
	if err := e.generateFloor(1); err != nil {
		return fmt.Errorf("failed to generate first floor: %w", err)
	}

	// Place player at the floor's starting position
	e.player.SetPosition(e.world.CurrentFloor.PlayerStart)

	// Update visibility around player
	e.world.UpdateVisibility(e.player.Position(), e.getViewRadius())

	e.state = types.StateExploring
	e.addMessage("Welcome to /dev/dungeon. You are in %s.", e.world.CurrentFloor.Type.FloorName())

	return nil
}

// generateFloor generates a new floor at the given depth.
func (e *Engine) generateFloor(depth int) error {
	// Check if floor is already cached
	if e.world.LoadCachedFloor(depth) {
		return nil
	}

	// Derive deterministic seed for this floor
	floorSeed := DeriveFloorSeed(e.masterSeed, depth)

	// Determine floor type based on depth
	floorType := e.getFloorTypeForDepth(depth)

	// Generate the floor structure
	if e.generator == nil {
		// Create a simple default floor if no generator is set
		floor := e.createDefaultFloor(floorType, depth, floorSeed)
		e.world.SetFloor(floor, make([]*entity.Enemy, 0), make([]*entity.Item, 0))
	} else {
		floor := e.generator.Generate(floorType, depth, floorSeed)
		if floor == nil {
			return errors.New("generator returned nil floor")
		}

		// Convert game.Floor to types.Floor
		typesFloor := e.convertFloor(floor, floorType, depth, floorSeed)
		e.world.SetFloor(typesFloor, make([]*entity.Enemy, 0), make([]*entity.Item, 0))
	}

	// Populate the floor with enemies and items
	e.populateFloor(depth, floorSeed)

	return nil
}

// convertFloor converts a game.Floor to types.Floor.
func (e *Engine) convertFloor(gf *Floor, floorType types.FloorType, depth int, seed int64) *types.Floor {
	tf := types.NewFloor(floorType, depth, gf.Width, gf.Height, seed)

	// Copy tiles
	for y := 0; y < gf.Height; y++ {
		for x := 0; x < gf.Width; x++ {
			tf.Tiles[y][x] = types.Tile{
				Type:        gf.Tiles[y][x].Type,
				Visible:     gf.Tiles[y][x].Visible,
				Explored:    gf.Tiles[y][x].Seen,
				Blocked:     gf.Tiles[y][x].Type == types.TileWall || gf.Tiles[y][x].Type == types.TileVoid,
				BlocksSight: gf.Tiles[y][x].Type == types.TileWall,
			}
		}
	}

	// Copy rooms
	for _, r := range gf.Rooms {
		if r != nil {
			tf.Rooms = append(tf.Rooms, types.Room{
				X:         r.X,
				Y:         r.Y,
				Width:     r.Width,
				Height:    r.Height,
				Connected: r.Connected,
			})
		}
	}

	tf.PlayerStart = gf.Entrance
	tf.StairsUp = gf.Entrance
	tf.StairsDown = gf.Exit

	return tf
}

// createDefaultFloor creates a simple floor when no generator is available.
func (e *Engine) createDefaultFloor(floorType types.FloorType, depth int, seed int64) *types.Floor {
	width := e.config.Display.MapWidth
	height := e.config.Display.MapHeight

	floor := types.NewFloor(floorType, depth, width, height, seed)

	// Create a simple room in the center
	roomWidth := width - 10
	roomHeight := height - 6
	roomX := 5
	roomY := 3

	// Carve out the room
	for y := roomY; y < roomY+roomHeight; y++ {
		for x := roomX; x < roomX+roomWidth; x++ {
			floor.SetTile(types.Position{X: x, Y: y}, types.NewTile(types.TileFloor))
		}
	}

	// Add stairs
	floor.PlayerStart = types.Position{X: roomX + 2, Y: roomY + 2}
	floor.StairsUp = types.Position{X: roomX + 2, Y: roomY + 2}
	floor.StairsDown = types.Position{X: roomX + roomWidth - 3, Y: roomY + roomHeight - 3}

	// Place stairs tiles
	if depth > 1 {
		floor.SetTile(floor.StairsUp, types.NewTile(types.TileStairsUp))
	}
	floor.SetTile(floor.StairsDown, types.NewTile(types.TileStairsDown))

	// Add a room to the list
	floor.Rooms = append(floor.Rooms, types.Room{
		X:         roomX,
		Y:         roomY,
		Width:     roomWidth,
		Height:    roomHeight,
		Connected: true,
	})

	return floor
}

// populateFloor adds enemies and items to the current floor.
func (e *Engine) populateFloor(depth int, floorSeed int64) {
	floorRng := rand.New(rand.NewSource(floorSeed))

	// Final floor (depth 8 = /dev/null) is the boss floor
	isBossFloor := depth >= 8

	if isBossFloor {
		// Spawn the KERNEL PANIC boss
		pos := e.findEmptyPosition(floorRng)
		if pos != nil {
			boss := entity.NewEnemy(entity.EnemyKernelPanic, "boss_kernel_panic", *pos, depth)
			boss.IsBoss = true
			e.world.AddEnemy(boss)
			e.addMessage("WARNING: KERNEL PANIC detected in sector. Approach with extreme caution.")
		}

		// Spawn a few minions
		for i := 0; i < 3; i++ {
			pos := e.findEmptyPosition(floorRng)
			if pos == nil {
				continue
			}
			enemyType := e.selectEnemyType(depth, floorRng)
			enemy := entity.NewEnemy(enemyType, fmt.Sprintf("minion_%d", i), *pos, depth)
			e.world.AddEnemy(enemy)
		}
		return
	}

	// Normal floor population
	// Number of enemies scales with depth
	numEnemies := 2 + depth
	if numEnemies > 10 {
		numEnemies = 10
	}

	// Spawn enemies
	for i := 0; i < numEnemies; i++ {
		pos := e.findEmptyPosition(floorRng)
		if pos == nil {
			continue
		}

		enemyType := e.selectEnemyType(depth, floorRng)
		enemyID := fmt.Sprintf("enemy_%d_%d", depth, i)
		enemy := entity.NewEnemy(enemyType, enemyID, *pos, depth)
		e.world.AddEnemy(enemy)
	}

	// Number of items
	numItems := 1 + floorRng.Intn(3)

	// Spawn items
	for i := 0; i < numItems; i++ {
		pos := e.findEmptyPosition(floorRng)
		if pos == nil {
			continue
		}

		itemTemplate := e.selectItemTemplate(depth, floorRng)
		itemID := fmt.Sprintf("item_%d_%d", depth, i)
		item := entity.NewItem(itemTemplate, itemID, *pos)
		if item != nil {
			e.world.AddItem(item)
		}
	}
}

// findEmptyPosition finds a random walkable position not occupied by entities.
func (e *Engine) findEmptyPosition(rng *rand.Rand) *types.Position {
	floor := e.world.CurrentFloor
	if floor == nil || len(floor.Rooms) == 0 {
		return nil
	}

	// Try to find an empty position
	for attempts := 0; attempts < 100; attempts++ {
		// Pick a random room
		room := floor.Rooms[rng.Intn(len(floor.Rooms))]

		// Pick a random position in the room
		pos := types.Position{
			X: room.X + 1 + rng.Intn(room.Width-2),
			Y: room.Y + 1 + rng.Intn(room.Height-2),
		}

		// Check if position is valid and empty
		if !floor.IsWalkable(pos) {
			continue
		}
		if e.player != nil && e.player.Position() == pos {
			continue
		}
		if e.world.GetEnemyAt(pos) != nil {
			continue
		}
		if e.world.GetItemAt(pos) != nil {
			continue
		}
		// Don't spawn on stairs
		if pos == floor.StairsDown || pos == floor.StairsUp {
			continue
		}

		return &pos
	}

	return nil
}

// selectEnemyType selects an appropriate enemy type for the given depth.
func (e *Engine) selectEnemyType(depth int, rng *rand.Rand) entity.EnemyType {
	// Available enemies by depth tier
	tier1 := []entity.EnemyType{entity.EnemyZombie}
	tier2 := []entity.EnemyType{entity.EnemyZombie, entity.EnemyDaemon, entity.EnemyForkBomb}
	tier3 := []entity.EnemyType{entity.EnemyDaemon, entity.EnemyForkBomb, entity.EnemySegfault}
	tier4 := []entity.EnemyType{entity.EnemySegfault, entity.EnemyRootkit}

	var pool []entity.EnemyType
	switch {
	case depth <= 2:
		pool = tier1
	case depth <= 4:
		pool = tier2
	case depth <= 6:
		pool = tier3
	default:
		pool = tier4
	}

	return pool[rng.Intn(len(pool))]
}

// selectItemTemplate selects an appropriate item template for the given depth.
func (e *Engine) selectItemTemplate(depth int, rng *rand.Rand) string {
	// Common items available on all floors
	common := []string{"malloc", "fd_restore", "memory_fragment", "env_vars", "alias_file", "grep_glaive"}
	uncommon := []string{"grep_scroll", "service_token", "basic_script", "basic_shell", "pipe_wrench", "sed_saber", "awk_axe", "firewall", "sandbox", "ssh_key", "cron_tab", "config_file", "realloc", "nice_boost", "cpu_boost"}
	rare := []string{"sudo_potion", "kill_9", "chmod_x", "core_dump", "vim_blade", "selinux_shield", "container", "gpg_ring", "tmux_session", "mmap", "segfault_bomb", "cron_claw"}
	legendary := []string{"fork_bomb", "rm_rf", "sudo_armor"}

	roll := rng.Float64()
	var pool []string

	if depth <= 2 {
		if roll < 0.8 {
			pool = common
		} else {
			pool = uncommon
		}
	} else if depth <= 5 {
		if roll < 0.5 {
			pool = common
		} else if roll < 0.85 {
			pool = uncommon
		} else {
			pool = rare
		}
	} else {
		// Deep floors: better loot including legendary
		if roll < 0.2 {
			pool = common
		} else if roll < 0.5 {
			pool = uncommon
		} else if roll < 0.9 {
			pool = rare
		} else {
			pool = legendary
		}
	}

	return pool[rng.Intn(len(pool))]
}

// getFloorTypeForDepth returns the floor type for a given depth.
func (e *Engine) getFloorTypeForDepth(depth int) types.FloorType {
	switch {
	case depth == 1:
		return types.FloorHome
	case depth == 2:
		return types.FloorTmp
	case depth == 3:
		return types.FloorVar
	case depth == 4:
		return types.FloorEtc
	case depth == 5:
		return types.FloorUsr
	case depth == 6:
		return types.FloorSys
	case depth == 7:
		return types.FloorDev
	default:
		return types.FloorDevNull
	}
}

// getViewRadius returns the player's view radius.
func (e *Engine) getViewRadius() int {
	return 8 // Default view radius
}

// MovePlayer attempts to move the player in the given direction.
func (e *Engine) MovePlayer(dir types.Direction) MoveResult {
	if e.player == nil || e.state != types.StateExploring {
		return MoveResult{Moved: false, Message: "Cannot move right now."}
	}

	// Calculate new position
	newPos := e.getNewPosition(e.player.Position(), dir)

	// Check for out of bounds
	if !e.world.CurrentFloor.InBounds(newPos) {
		return MoveResult{Moved: false, Message: "You cannot go that way."}
	}

	// Check for enemy at position (bump attack)
	if enemy := e.world.GetEnemyAt(newPos); enemy != nil {
		return MoveResult{
			Moved:   false,
			Combat:  enemy,
			Message: fmt.Sprintf("You bump into a %s!", enemy.Name()),
		}
	}

	// Check if position is walkable
	if !e.world.CurrentFloor.IsWalkable(newPos) {
		return MoveResult{Moved: false, Message: "You cannot walk there."}
	}

	// Move the player
	e.player.SetPosition(newPos)

	// Track step
	if e.stats != nil {
		e.stats.StepsWalked++
	}

	// Update visibility
	e.world.UpdateVisibility(newPos, e.getViewRadius())

	result := MoveResult{Moved: true}

	// Check for item at new position
	if item := e.world.GetItemAt(newPos); item != nil {
		if e.player.Inventory.AddItem(item) {
			e.world.RemoveItem(item.ID())
			result.PickedUp = item
			result.Message = fmt.Sprintf("You picked up %s.", item.Name())
			e.addMessage("Picked up %s.", item.Name())
			// Track item pickup
			if e.stats != nil {
				e.stats.ItemsCollected++
			}
		} else {
			result.Message = "Your inventory is full."
		}
	}

	// Check for stairs
	tile := e.world.CurrentFloor.GetTile(newPos)
	if tile != nil {
		if tile.Type == types.TileStairsDown {
			result.Message = "You see stairs leading down. Press '>' to descend."
		} else if tile.Type == types.TileStairsUp && e.world.CurrentDepth > 1 {
			result.Message = "You see stairs leading up. Press '<' to ascend."
		}
	}

	return result
}

// getNewPosition calculates a new position given a direction.
func (e *Engine) getNewPosition(pos types.Position, dir types.Direction) types.Position {
	switch dir {
	case types.DirUp:
		return types.Position{X: pos.X, Y: pos.Y - 1}
	case types.DirDown:
		return types.Position{X: pos.X, Y: pos.Y + 1}
	case types.DirLeft:
		return types.Position{X: pos.X - 1, Y: pos.Y}
	case types.DirRight:
		return types.Position{X: pos.X + 1, Y: pos.Y}
	default:
		return pos
	}
}

// DescendStairs attempts to go down stairs.
func (e *Engine) DescendStairs() error {
	if e.player == nil || e.world.CurrentFloor == nil {
		return errors.New("no active game")
	}

	// Can't descend past the final floor (/dev/null at depth 8)
	if e.world.CurrentDepth >= 8 {
		return errors.New("you've reached the deepest level - defeat the Kernel Panic to escape")
	}

	playerPos := e.player.Position()
	tile := e.world.CurrentFloor.GetTile(playerPos)

	if tile == nil || tile.Type != types.TileStairsDown {
		return errors.New("no stairs down here")
	}

	// Save before floor transition
	e.Save(save.TriggerFloorTransition)

	// Cache current floor state
	e.world.CacheCurrentFloor()

	// Generate or load next floor
	nextDepth := e.world.CurrentDepth + 1
	if err := e.generateFloor(nextDepth); err != nil {
		return fmt.Errorf("failed to descend: %w", err)
	}

	// Place player at stairs up of new floor
	e.player.SetPosition(e.world.CurrentFloor.StairsUp)

	// Update visibility
	e.world.UpdateVisibility(e.player.Position(), e.getViewRadius())

	// Track floor exploration
	if e.stats != nil {
		e.stats.FloorsExplored++
		if nextDepth > e.stats.MaxDepthReached {
			e.stats.MaxDepthReached = nextDepth
		}
	}

	e.addMessage("You descend to %s (depth %d).", e.world.CurrentFloor.Type.FloorName(), nextDepth)

	return nil
}

// ForceDescend forces descent to the next floor regardless of player position (admin/debug).
func (e *Engine) ForceDescend() error {
	if e.player == nil || e.world.CurrentFloor == nil {
		return errors.New("no active game")
	}

	// Save before floor transition
	e.Save(save.TriggerFloorTransition)

	// Cache current floor state
	e.world.CacheCurrentFloor()

	// Generate or load next floor
	nextDepth := e.world.CurrentDepth + 1
	if err := e.generateFloor(nextDepth); err != nil {
		return fmt.Errorf("failed to descend: %w", err)
	}

	// Place player at stairs up of new floor
	e.player.SetPosition(e.world.CurrentFloor.StairsUp)

	// Update visibility
	e.world.UpdateVisibility(e.player.Position(), e.getViewRadius())

	e.addMessage("[ADMIN] Forced descent to %s (depth %d).", e.world.CurrentFloor.Type.FloorName(), nextDepth)

	return nil
}

// AscendStairs attempts to go up stairs.
func (e *Engine) AscendStairs() error {
	if e.player == nil || e.world.CurrentFloor == nil {
		return errors.New("no active game")
	}

	if e.world.CurrentDepth <= 1 {
		return errors.New("you cannot go up from here - this is the first floor")
	}

	playerPos := e.player.Position()
	tile := e.world.CurrentFloor.GetTile(playerPos)

	if tile == nil || tile.Type != types.TileStairsUp {
		return errors.New("no stairs up here")
	}

	// Save before floor transition
	e.Save(save.TriggerFloorTransition)

	// Cache current floor state
	e.world.CacheCurrentFloor()

	// Load previous floor (should be cached)
	prevDepth := e.world.CurrentDepth - 1
	if !e.world.LoadCachedFloor(prevDepth) {
		return errors.New("cannot return to previous floor - cache missing")
	}

	// Place player at stairs down of previous floor
	e.player.SetPosition(e.world.CurrentFloor.StairsDown)

	// Update visibility
	e.world.UpdateVisibility(e.player.Position(), e.getViewRadius())

	e.addMessage("You ascend to %s (depth %d).", e.world.CurrentFloor.Type.FloorName(), prevDepth)

	return nil
}

// GetVisibleTiles returns the current floor's tiles for rendering.
func (e *Engine) GetVisibleTiles() [][]types.Tile {
	return e.world.GetVisibleTiles()
}

// Update runs a single game loop tick.
func (e *Engine) Update() error {
	// Update enemy AI, cooldowns, effects, etc.
	// For now, this is a placeholder for future game loop logic
	return nil
}

// CurrentState returns the current game state.
func (e *Engine) CurrentState() types.GameState {
	return e.state
}

// SetState sets the game state.
func (e *Engine) SetState(state types.GameState) {
	e.state = state
}

// Player returns the player entity.
func (e *Engine) Player() *entity.Player {
	return e.player
}

// GetWorld returns the game world.
func (e *Engine) GetWorld() *GameWorld {
	return e.world
}

// Config returns the game configuration.
func (e *Engine) Config() *config.Config {
	return e.config
}

// MasterSeed returns the master seed used for this game.
func (e *Engine) MasterSeed() int64 {
	return e.masterSeed
}

// addMessage adds a message to the game log.
func (e *Engine) addMessage(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	e.messages = append(e.messages, msg)
	// Keep only the last 100 messages
	if len(e.messages) > 100 {
		e.messages = e.messages[len(e.messages)-100:]
	}
}

// Messages returns recent game messages.
func (e *Engine) Messages() []string {
	return e.messages
}

// ClearMessages clears all messages.
func (e *Engine) ClearMessages() {
	e.messages = make([]string, 0)
}

// RemoveEnemy removes an enemy from the world (e.g., after defeat).
func (e *Engine) RemoveEnemy(id string) {
	e.world.RemoveEnemy(id)
}

// GetEnemyAt returns the enemy at a position.
func (e *Engine) GetEnemyAt(pos types.Position) *entity.Enemy {
	return e.world.GetEnemyAt(pos)
}

// GetEnemies returns all enemies on the current floor.
func (e *Engine) GetEnemies() []*entity.Enemy {
	return e.world.Enemies
}

// GetItems returns all items on the current floor.
func (e *Engine) GetItems() []*entity.Item {
	return e.world.Items
}

// CurrentDepth returns the current dungeon depth.
func (e *Engine) CurrentDepth() int {
	return e.world.CurrentDepth
}

// CurrentFloorType returns the type of the current floor.
func (e *Engine) CurrentFloorType() types.FloorType {
	return e.world.GetFloorType()
}

// GetRunStats returns the current run statistics.
func (e *Engine) GetRunStats() *RunStats {
	return e.stats
}

// --- Save/Load Methods ---

// Save triggers a save with the given trigger type.
func (e *Engine) Save(trigger save.SaveTrigger) {
	if e.saveManager == nil || e.player == nil {
		return
	}

	data := e.toSaveData()
	e.saveManager.Save(data, trigger)
}

// SaveSync saves and waits for completion.
func (e *Engine) SaveSync(trigger save.SaveTrigger) error {
	if e.saveManager == nil {
		return errors.New("save manager not available")
	}
	if e.player == nil {
		return errors.New("no active game to save")
	}

	data := e.toSaveData()
	return e.saveManager.SaveSync(data, trigger)
}

// GetSaveData returns the current game state as save data (for external storage).
func (e *Engine) GetSaveData() *save.SaveData {
	if e.player == nil {
		return nil
	}
	return e.toSaveData()
}

// LoadGame loads a saved game state.
func (e *Engine) LoadGame(data *save.SaveData) error {
	if data == nil {
		return errors.New("no save data provided")
	}

	// Set the master seed from save
	e.masterSeed = data.MasterSeed
	e.rng = rand.New(rand.NewSource(e.masterSeed))

	// Recreate player from save data
	e.player = entity.NewPlayer(data.Player.Class)
	e.player.Stats = data.Player.Stats
	e.player.MaxStats = data.Player.MaxStats
	e.player.Level = data.Player.Level
	e.player.XP = data.Player.XP
	e.player.XPToLevel = data.Player.XPToLevel
	e.player.ExitCodes = data.Player.ExitCodes

	// Clear starting gear before loading saved inventory/equipment
	e.player.Inventory.Clear()
	e.player.Equipment.Weapon = nil
	e.player.Equipment.Armor = nil
	e.player.Equipment.Utility1 = nil
	e.player.Equipment.Utility2 = nil

	// Load inventory
	for _, itemData := range data.Player.Inventory {
		item := entity.NewItem(itemData.TemplateID, itemData.TemplateID, types.Position{})
		if item != nil {
			item.Quantity = itemData.Quantity
			e.player.Inventory.AddItem(item)
		}
	}

	// Load equipment
	if data.Player.Equipment.Weapon != "" {
		weapon := entity.NewItem(data.Player.Equipment.Weapon, "weapon", types.Position{})
		if weapon != nil {
			e.player.Equipment.Equip(weapon)
		}
	}
	if data.Player.Equipment.Armor != "" {
		armor := entity.NewItem(data.Player.Equipment.Armor, "armor", types.Position{})
		if armor != nil {
			e.player.Equipment.Equip(armor)
		}
	}
	if data.Player.Equipment.Utility1 != "" {
		util := entity.NewItem(data.Player.Equipment.Utility1, "utility1", types.Position{})
		if util != nil {
			e.player.Equipment.Equip(util)
		}
	}
	if data.Player.Equipment.Utility2 != "" {
		util := entity.NewItem(data.Player.Equipment.Utility2, "utility2", types.Position{})
		if util != nil {
			e.player.Equipment.Equip(util)
		}
	}

	// Cap depth at max floor (8 = boss floor)
	loadDepth := data.CurrentDepth
	if loadDepth > 8 {
		loadDepth = 8
		e.addMessage("Save was beyond final floor - resetting to /dev/null.")
	}

	// Regenerate the current floor
	if err := e.generateFloor(loadDepth); err != nil {
		return fmt.Errorf("failed to regenerate floor: %w", err)
	}

	// Apply floor state deltas (dead enemies, looted items, explored tiles)
	for _, floorState := range data.FloorStates {
		if floorState.Depth == loadDepth {
			e.applyFloorState(&floorState)
			break
		}
	}

	// Set player position (reset to stairs if we capped depth)
	if loadDepth != data.CurrentDepth {
		e.player.SetPosition(e.world.CurrentFloor.StairsUp)
	} else {
		e.player.SetPosition(data.Player.Position)
	}

	// Update visibility
	e.world.UpdateVisibility(e.player.Position(), e.getViewRadius())

	e.state = types.StateExploring
	e.addMessage("Game loaded. You are in %s (depth %d).", e.world.CurrentFloor.Type.FloorName(), loadDepth)

	return nil
}

// LoadLatestSave loads the most recent save file.
func (e *Engine) LoadLatestSave() error {
	if e.saveManager == nil {
		return errors.New("save manager not available")
	}

	data, err := e.saveManager.LoadLatest()
	if err != nil {
		return fmt.Errorf("failed to load save: %w", err)
	}
	if data == nil {
		return errors.New("no save file found")
	}

	return e.LoadGame(data)
}

// HasSaveFile returns true if a save file exists.
func (e *Engine) HasSaveFile() bool {
	if e.saveManager == nil {
		return false
	}
	saves, err := e.saveManager.ListSaves()
	if err != nil {
		return false
	}
	return len(saves) > 0
}

// Shutdown gracefully shuts down the engine.
func (e *Engine) Shutdown() {
	// Save on quit
	if e.player != nil && e.saveManager != nil {
		e.SaveSync(save.TriggerQuit)
	}

	// Stop the save manager
	if e.saveManager != nil {
		e.saveManager.Stop()
	}
}

// toSaveData converts the current game state to SaveData.
func (e *Engine) toSaveData() *save.SaveData {
	// Convert inventory
	var invData []save.ItemData
	for _, item := range e.player.Inventory.Items {
		invData = append(invData, save.ItemData{
			TemplateID: item.ID(),
			Quantity:   item.Quantity,
		})
	}

	// Convert equipment
	eqData := save.EquipmentData{}
	if e.player.Equipment.Weapon != nil {
		eqData.Weapon = e.player.Equipment.Weapon.ID()
	}
	if e.player.Equipment.Armor != nil {
		eqData.Armor = e.player.Equipment.Armor.ID()
	}
	if e.player.Equipment.Utility1 != nil {
		eqData.Utility1 = e.player.Equipment.Utility1.ID()
	}
	if e.player.Equipment.Utility2 != nil {
		eqData.Utility2 = e.player.Equipment.Utility2.ID()
	}

	// Build floor states
	floorStates := e.buildFloorStates()

	return &save.SaveData{
		Version:      save.Version,
		MasterSeed:   e.masterSeed,
		CurrentDepth: e.world.CurrentDepth,
		Player: save.PlayerData{
			Class:     e.player.Class,
			Stats:     e.player.Stats,
			MaxStats:  e.player.MaxStats,
			Level:     e.player.Level,
			XP:        e.player.XP,
			XPToLevel: e.player.XPToLevel,
			Position:  e.player.Position(),
			Inventory: invData,
			Equipment: eqData,
			ExitCodes: e.player.ExitCodes,
		},
		FloorStates: floorStates,
	}
}

// buildFloorStates builds floor state deltas for all visited floors.
func (e *Engine) buildFloorStates() []save.FloorState {
	var states []save.FloorState

	// For now, just save current floor state
	// In the future, track all visited floors
	if e.world.CurrentFloor != nil {
		state := save.FloorState{
			Depth:         e.world.CurrentDepth,
			ExploredTiles: e.getExploredTiles(),
			DeadEnemies:   []string{}, // Track killed enemies
			LootedItems:   []string{}, // Track picked up items
		}
		states = append(states, state)
	}

	return states
}

// getExploredTiles returns all explored tile positions.
func (e *Engine) getExploredTiles() []types.Position {
	var explored []types.Position
	if e.world.CurrentFloor == nil {
		return explored
	}

	for y, row := range e.world.CurrentFloor.Tiles {
		for x, tile := range row {
			if tile.Explored {
				explored = append(explored, types.Position{X: x, Y: y})
			}
		}
	}
	return explored
}

// applyFloorState applies saved floor state deltas.
func (e *Engine) applyFloorState(state *save.FloorState) {
	// Mark explored tiles
	for _, pos := range state.ExploredTiles {
		if tile := e.world.CurrentFloor.GetTile(pos); tile != nil {
			tile.Explored = true
		}
	}

	// Remove dead enemies
	for _, enemyID := range state.DeadEnemies {
		e.world.RemoveEnemy(enemyID)
	}

	// Remove looted items
	for _, itemID := range state.LootedItems {
		e.world.RemoveItem(itemID)
	}
}
