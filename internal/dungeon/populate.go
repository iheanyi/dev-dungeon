package dungeon

import (
	"math/rand"

	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/types"
)

// SpawnData contains positions and template IDs for spawning entities.
type SpawnData struct {
	Enemies []EnemySpawn
	Items   []ItemSpawn
}

// EnemySpawn represents an enemy spawn point.
type EnemySpawn struct {
	Position   types.Position
	TemplateID entity.EnemyType
}

// ItemSpawn represents an item spawn point.
type ItemSpawn struct {
	Position   types.Position
	TemplateID string
}

// PopulationConfig holds parameters for entity placement.
type PopulationConfig struct {
	BaseEnemyCount    int
	EnemyCountPerDepth int
	BaseItemCount     int
	ItemCountPerRoom  float32
}

// DefaultPopulationConfig returns default population parameters.
func DefaultPopulationConfig() PopulationConfig {
	return PopulationConfig{
		BaseEnemyCount:    2,
		EnemyCountPerDepth: 2,
		BaseItemCount:     2,
		ItemCountPerRoom:  0.5,
	}
}

// PopulateFloor calculates enemy and item spawn positions for a floor.
// Returns spawn data without creating actual entities.
func PopulateFloor(rng *rand.Rand, floor *types.Floor, depth int) SpawnData {
	return PopulateFloorWithConfig(rng, floor, depth, DefaultPopulationConfig())
}

// PopulateFloorWithConfig calculates spawns using custom configuration.
func PopulateFloorWithConfig(rng *rand.Rand, floor *types.Floor, depth int, cfg PopulationConfig) SpawnData {
	data := SpawnData{
		Enemies: make([]EnemySpawn, 0),
		Items:   make([]ItemSpawn, 0),
	}

	if len(floor.Rooms) == 0 {
		return data
	}

	// Adjust based on floor type
	enemyMultiplier := getEnemyMultiplier(floor.Type)
	itemMultiplier := getItemMultiplier(floor.Type)

	// Calculate enemy count
	enemyCount := cfg.BaseEnemyCount + (cfg.EnemyCountPerDepth * depth)
	enemyCount = int(float32(enemyCount) * enemyMultiplier)

	// Calculate item count
	itemCount := cfg.BaseItemCount + int(float32(len(floor.Rooms))*cfg.ItemCountPerRoom)
	itemCount = int(float32(itemCount) * itemMultiplier)

	// Get valid spawn positions (floor tiles in rooms, excluding stairs)
	spawnPositions := getValidSpawnPositions(floor)
	if len(spawnPositions) == 0 {
		return data
	}

	// Shuffle spawn positions
	rng.Shuffle(len(spawnPositions), func(i, j int) {
		spawnPositions[i], spawnPositions[j] = spawnPositions[j], spawnPositions[i]
	})

	// Track used positions
	usedPositions := make(map[types.Position]bool)
	usedPositions[floor.StairsUp] = true
	usedPositions[floor.StairsDown] = true
	usedPositions[floor.PlayerStart] = true

	// Spawn enemies
	enemyTypes := getEnemyTypesForFloor(floor.Type, depth)
	for i := 0; i < enemyCount && i < len(spawnPositions); i++ {
		pos := spawnPositions[i]
		if usedPositions[pos] {
			continue
		}

		enemyType := enemyTypes[rng.Intn(len(enemyTypes))]
		data.Enemies = append(data.Enemies, EnemySpawn{
			Position:   pos,
			TemplateID: enemyType,
		})
		usedPositions[pos] = true
	}

	// Spawn items (use remaining positions)
	itemTemplates := getItemTemplatesForFloor(floor.Type, depth)
	itemIndex := len(data.Enemies)
	for i := 0; i < itemCount && itemIndex+i < len(spawnPositions); i++ {
		pos := spawnPositions[itemIndex+i]
		if usedPositions[pos] {
			continue
		}

		itemTemplate := itemTemplates[rng.Intn(len(itemTemplates))]
		data.Items = append(data.Items, ItemSpawn{
			Position:   pos,
			TemplateID: itemTemplate,
		})
		usedPositions[pos] = true
	}

	return data
}

// getValidSpawnPositions returns all floor tiles inside rooms that can be spawned on.
func getValidSpawnPositions(floor *types.Floor) []types.Position {
	positions := make([]types.Position, 0)

	for _, room := range floor.Rooms {
		// Exclude the outermost tiles of each room for better spacing
		for y := room.Y + 1; y < room.Y+room.Height-1; y++ {
			for x := room.X + 1; x < room.X+room.Width-1; x++ {
				pos := types.Position{X: x, Y: y}
				tile := floor.GetTile(pos)
				if tile != nil && tile.Type == types.TileFloor {
					positions = append(positions, pos)
				}
			}
		}
	}

	return positions
}

// getEnemyMultiplier returns enemy count multiplier based on floor type.
func getEnemyMultiplier(floorType types.FloorType) float32 {
	switch floorType {
	case types.FloorHome:
		return 0.5 // Tutorial: fewer enemies
	case types.FloorTmp:
		return 1.2 // Chaotic: more enemies
	case types.FloorDevNull:
		return 0.0 // Boss only, handled separately
	default:
		return 1.0
	}
}

// getItemMultiplier returns item count multiplier based on floor type.
func getItemMultiplier(floorType types.FloorType) float32 {
	switch floorType {
	case types.FloorHome:
		return 1.5 // Tutorial: more items
	case types.FloorTmp:
		return 0.8 // Less items
	default:
		return 1.0
	}
}

// getEnemyTypesForFloor returns enemy types appropriate for the floor.
func getEnemyTypesForFloor(floorType types.FloorType, depth int) []entity.EnemyType {
	// Early floors have weaker enemies
	if depth <= 2 {
		return []entity.EnemyType{
			entity.EnemyZombie,
			entity.EnemyForkBomb,
		}
	}

	// Mid-game adds more variety
	if depth <= 5 {
		return []entity.EnemyType{
			entity.EnemyZombie,
			entity.EnemyDaemon,
			entity.EnemyForkBomb,
			entity.EnemySegfault,
		}
	}

	// Late-game has all enemy types except boss
	return []entity.EnemyType{
		entity.EnemyZombie,
		entity.EnemyDaemon,
		entity.EnemyForkBomb,
		entity.EnemySegfault,
		entity.EnemyRootkit,
	}
}

// getItemTemplatesForFloor returns item templates appropriate for the floor.
func getItemTemplatesForFloor(floorType types.FloorType, depth int) []string {
	// Basic items always available
	items := []string{
		"pid_restore",
		"mem_restore",
		"memory_fragment",
	}

	// Add more items based on depth
	if depth >= 2 {
		items = append(items, "grep_scroll", "service_token", "cpu_cycle")
	}

	if depth >= 4 {
		items = append(items, "basic_script", "basic_shell")
	}

	if depth >= 6 {
		items = append(items, "sudo_potion", "core_dump", "chmod_x")
	}

	if depth >= 8 {
		items = append(items, "kill_9", "root_shard")
	}

	return items
}
