package game

// Entity is the minimal interface for game entities.
// This avoids importing the entity package to prevent cycles.
type Entity interface {
	ID() string
	Name() string
	Position() Position
	SetPosition(pos Position)
	Glyph() rune
	IsBlocking() bool
}

// Combatant is the interface for entities that can fight.
type Combatant interface {
	Entity
	GetStats() Stats
	TakeDamage(amount int) bool
	IsAlive() bool
}

// Engine is the core game engine interface.
// This interface allows for testing and swapping implementations.
type Engine interface {
	// State management
	CurrentState() GameState
	SetState(state GameState)

	// Game world
	World() World
	Player() Combatant

	// Game loop
	Update() error
	HandleInput(input string) error

	// Combat
	StartCombat(enemies []Combatant)
	InCombat() bool
}

// World manages the game world and dungeon floors.
type World interface {
	// Floor management
	CurrentFloor() *Floor
	CurrentFloorType() FloorType
	DescendFloor() error

	// Entity management
	GetEntitiesAt(pos Position) []Entity
	MoveEntity(e Entity, dir Direction) bool
	RemoveEntity(id string)

	// Visibility
	IsVisible(pos Position) bool
	UpdateVisibility(center Position, radius int)
}

// Floor represents a single dungeon floor.
type Floor struct {
	Type     FloorType
	Level    int
	Tiles    [][]Tile
	Width    int
	Height   int
	Rooms    []*Room
	Entities []Entity
	Entrance Position
	Exit     Position
}

// Tile represents a single tile in the dungeon.
type Tile struct {
	Type    TileType
	Visible bool
	Seen    bool // Previously seen (fog of war)
}

// Room represents a room in the dungeon.
type Room struct {
	X, Y          int
	Width, Height int
	Connected     bool
}

// CombatSystem handles turn-based combat.
type CombatSystem interface {
	// Combat lifecycle
	StartCombat(player Combatant, enemies []Combatant) *Combat
	ExecuteAction(combat *Combat, action Action) Result
	EndCombat(combat *Combat) Rewards

	// Combat state
	IsPlayerTurn(combat *Combat) bool
	GetValidTargets(combat *Combat) []Combatant
	GetAvailableActions(combat *Combat) []ActionType
}

// Combat represents an active combat encounter.
type Combat struct {
	Player      Combatant
	Enemies     []Combatant
	TurnOrder   []string // Entity IDs in turn order
	CurrentTurn int
	Round       int
	IsActive    bool
	CombatLog   []string
}

// DungeonGenerator generates procedural dungeons.
type DungeonGenerator interface {
	Generate(floorType FloorType, level int, seed int64) *Floor
}

// Renderer renders the game state to a string.
// This interface allows for testing without a real terminal.
type Renderer interface {
	RenderWorld(world World, player Combatant) string
	RenderCombat(combat *Combat) string
	RenderInventory(player Combatant) string
	RenderStats(player Combatant) string
}

// SaveSystem handles saving and loading game state.
type SaveSystem interface {
	SaveRun(engine Engine) error
	LoadRun() (Engine, error)
	SaveMeta(meta *MetaProgress) error
	LoadMeta() (*MetaProgress, error)
}

// MetaProgress tracks persistent progress across runs.
type MetaProgress struct {
	TotalExitCodes   int
	UnlockedClasses  []string
	UnlockedItems    []string
	PermanentBonuses map[string]int
	TotalRuns        int
	BestFloorReached FloorType
	TotalKills       int
}
