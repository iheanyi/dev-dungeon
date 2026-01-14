// Package config provides configuration management for /dev/dungeon.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds all game configuration.
type Config struct {
	Game     GameConfig     `json:"game"`
	Display  DisplayConfig  `json:"display"`
	Controls ControlsConfig `json:"controls"`
	Debug    DebugConfig    `json:"debug"`
}

// GameConfig holds gameplay-related settings.
type GameConfig struct {
	StartingClass        string  `json:"starting_class"`
	DifficultyMultiplier float64 `json:"difficulty_multiplier"`
	EnablePermadeath     bool    `json:"enable_permadeath"`
	RandomSeed           int64   `json:"random_seed"` // 0 = random
}

// DisplayConfig holds display settings.
type DisplayConfig struct {
	MapWidth       int    `json:"map_width"`
	MapHeight      int    `json:"map_height"`
	ShowMinimap    bool   `json:"show_minimap"`
	ShowStats      bool   `json:"show_stats"`
	AnimationSpeed int    `json:"animation_speed"` // ms between frames
	ColorScheme    string `json:"color_scheme"`
}

// ControlsConfig holds key bindings.
type ControlsConfig struct {
	MoveUp    string `json:"move_up"`
	MoveDown  string `json:"move_down"`
	MoveLeft  string `json:"move_left"`
	MoveRight string `json:"move_right"`
	Inventory string `json:"inventory"`
	Attack    string `json:"attack"`
	Hack      string `json:"hack"`
	UseItem   string `json:"use_item"`
	Flee      string `json:"flee"`
	Confirm   string `json:"confirm"`
	Cancel    string `json:"cancel"`
	Pause     string `json:"pause"`
}

// DebugConfig holds debug settings.
type DebugConfig struct {
	Enabled   bool   `json:"enabled"`
	ShowFPS   bool   `json:"show_fps"`
	GodMode   bool   `json:"god_mode"`
	RevealMap bool   `json:"reveal_map"`
	LogLevel  string `json:"log_level"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Game: GameConfig{
			StartingClass:        "init",
			DifficultyMultiplier: 1.0,
			EnablePermadeath:     true,
			RandomSeed:           0,
		},
		Display: DisplayConfig{
			MapWidth:       80,
			MapHeight:      24,
			ShowMinimap:    true,
			ShowStats:      true,
			AnimationSpeed: 50,
			ColorScheme:    "default",
		},
		Controls: ControlsConfig{
			MoveUp:    "w",
			MoveDown:  "s",
			MoveLeft:  "a",
			MoveRight: "d",
			Inventory: "i",
			Attack:    "1",
			Hack:      "2",
			UseItem:   "3",
			Flee:      "4",
			Confirm:   "enter",
			Cancel:    "esc",
			Pause:     "p",
		},
		Debug: DebugConfig{
			Enabled:   false,
			ShowFPS:   false,
			GodMode:   false,
			RevealMap: false,
			LogLevel:  "info",
		},
	}
}

// Load loads configuration from a file, falling back to defaults.
func Load(path string) (*Config, error) {
	config := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil // Use defaults
		}
		return nil, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// Save saves configuration to a file.
func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ConfigPath returns the default config file path.
func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "devdungeon", "config.json")
}
