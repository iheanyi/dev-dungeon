package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig should not return nil")
	}

	// Check game defaults
	if cfg.Game.StartingClass != "init" {
		t.Errorf("expected starting class 'init', got '%s'", cfg.Game.StartingClass)
	}
	if cfg.Game.DifficultyMultiplier != 1.0 {
		t.Errorf("expected difficulty 1.0, got %f", cfg.Game.DifficultyMultiplier)
	}
	if !cfg.Game.EnablePermadeath {
		t.Error("permadeath should be enabled by default")
	}
	if cfg.Game.RandomSeed != 0 {
		t.Errorf("expected seed 0, got %d", cfg.Game.RandomSeed)
	}

	// Check display defaults
	if cfg.Display.MapWidth != 80 {
		t.Errorf("expected map width 80, got %d", cfg.Display.MapWidth)
	}
	if cfg.Display.MapHeight != 24 {
		t.Errorf("expected map height 24, got %d", cfg.Display.MapHeight)
	}
	if !cfg.Display.ShowMinimap {
		t.Error("minimap should be shown by default")
	}
	if !cfg.Display.ShowStats {
		t.Error("stats should be shown by default")
	}
	if cfg.Display.AnimationSpeed != 50 {
		t.Errorf("expected animation speed 50, got %d", cfg.Display.AnimationSpeed)
	}
	if cfg.Display.ColorScheme != "default" {
		t.Errorf("expected color scheme 'default', got '%s'", cfg.Display.ColorScheme)
	}

	// Check controls defaults
	if cfg.Controls.MoveUp != "w" {
		t.Errorf("expected move up 'w', got '%s'", cfg.Controls.MoveUp)
	}
	if cfg.Controls.MoveDown != "s" {
		t.Errorf("expected move down 's', got '%s'", cfg.Controls.MoveDown)
	}
	if cfg.Controls.MoveLeft != "a" {
		t.Errorf("expected move left 'a', got '%s'", cfg.Controls.MoveLeft)
	}
	if cfg.Controls.MoveRight != "d" {
		t.Errorf("expected move right 'd', got '%s'", cfg.Controls.MoveRight)
	}
	if cfg.Controls.Inventory != "i" {
		t.Errorf("expected inventory 'i', got '%s'", cfg.Controls.Inventory)
	}
	if cfg.Controls.Pause != "p" {
		t.Errorf("expected pause 'p', got '%s'", cfg.Controls.Pause)
	}

	// Check debug defaults
	if cfg.Debug.Enabled {
		t.Error("debug should be disabled by default")
	}
	if cfg.Debug.GodMode {
		t.Error("god mode should be disabled by default")
	}
	if cfg.Debug.LogLevel != "info" {
		t.Errorf("expected log level 'info', got '%s'", cfg.Debug.LogLevel)
	}
}

func TestLoadNonExistent(t *testing.T) {
	// Load from non-existent path should return defaults
	cfg, err := Load("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("Load should not error for non-existent file: %v", err)
	}

	if cfg == nil {
		t.Fatal("config should not be nil")
	}

	// Should have default values
	if cfg.Game.StartingClass != "init" {
		t.Error("should have default starting class")
	}
}

func TestLoadAndSave(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Create custom config
	cfg := DefaultConfig()
	cfg.Game.StartingClass = "bash"
	cfg.Game.DifficultyMultiplier = 1.5
	cfg.Display.MapWidth = 100
	cfg.Display.MapHeight = 30
	cfg.Controls.MoveUp = "k"
	cfg.Debug.Enabled = true

	// Save it
	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file should exist after save")
	}

	// Load it back
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify values
	if loaded.Game.StartingClass != "bash" {
		t.Errorf("expected 'bash', got '%s'", loaded.Game.StartingClass)
	}
	if loaded.Game.DifficultyMultiplier != 1.5 {
		t.Errorf("expected 1.5, got %f", loaded.Game.DifficultyMultiplier)
	}
	if loaded.Display.MapWidth != 100 {
		t.Errorf("expected 100, got %d", loaded.Display.MapWidth)
	}
	if loaded.Display.MapHeight != 30 {
		t.Errorf("expected 30, got %d", loaded.Display.MapHeight)
	}
	if loaded.Controls.MoveUp != "k" {
		t.Errorf("expected 'k', got '%s'", loaded.Controls.MoveUp)
	}
	if !loaded.Debug.Enabled {
		t.Error("debug should be enabled")
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nested", "dir", "config.json")

	cfg := DefaultConfig()
	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Save should create nested directories: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file should exist")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.json")

	// Write invalid JSON
	_ = os.WriteFile(configPath, []byte("not valid json {{{"), 0644)

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load should error on invalid JSON")
	}
}

func TestLoadPartialConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "partial.json")

	// Write partial config (only game section)
	partialJSON := `{
		"game": {
			"starting_class": "vim",
			"difficulty_multiplier": 2.0
		}
	}`
	_ = os.WriteFile(configPath, []byte(partialJSON), 0644)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Specified values should be loaded
	if cfg.Game.StartingClass != "vim" {
		t.Errorf("expected 'vim', got '%s'", cfg.Game.StartingClass)
	}
	if cfg.Game.DifficultyMultiplier != 2.0 {
		t.Errorf("expected 2.0, got %f", cfg.Game.DifficultyMultiplier)
	}

	// Unspecified values should have defaults
	if cfg.Display.MapWidth != 80 {
		t.Errorf("expected default map width 80, got %d", cfg.Display.MapWidth)
	}
	if cfg.Controls.MoveUp != "w" {
		t.Errorf("expected default move up 'w', got '%s'", cfg.Controls.MoveUp)
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()

	if path == "" {
		t.Error("ConfigPath should not return empty string")
	}

	// ConfigPath uses home dir which should be absolute
	// But on some systems it might not be, so just check it's not empty

	// Should end with config.json
	if filepath.Base(path) != "config.json" {
		t.Errorf("ConfigPath should end with 'config.json', got '%s'", filepath.Base(path))
	}

	// Should contain devdungeon directory
	if filepath.Base(filepath.Dir(path)) != "devdungeon" {
		t.Errorf("ConfigPath should be in 'devdungeon' directory")
	}
}

// === Struct Tests ===

func TestGameConfigStruct(t *testing.T) {
	gc := GameConfig{
		StartingClass:        "sudo",
		DifficultyMultiplier: 2.5,
		EnablePermadeath:     false,
		RandomSeed:           12345,
	}

	if gc.StartingClass != "sudo" {
		t.Error("StartingClass not set correctly")
	}
	if gc.DifficultyMultiplier != 2.5 {
		t.Error("DifficultyMultiplier not set correctly")
	}
	if gc.EnablePermadeath {
		t.Error("EnablePermadeath not set correctly")
	}
	if gc.RandomSeed != 12345 {
		t.Error("RandomSeed not set correctly")
	}
}

func TestDisplayConfigStruct(t *testing.T) {
	dc := DisplayConfig{
		MapWidth:       120,
		MapHeight:      40,
		ShowMinimap:    false,
		ShowStats:      false,
		AnimationSpeed: 100,
		ColorScheme:    "retro",
	}

	if dc.MapWidth != 120 {
		t.Error("MapWidth not set correctly")
	}
	if dc.ShowMinimap {
		t.Error("ShowMinimap not set correctly")
	}
	if dc.ColorScheme != "retro" {
		t.Error("ColorScheme not set correctly")
	}
}

func TestControlsConfigStruct(t *testing.T) {
	cc := ControlsConfig{
		MoveUp:    "up",
		MoveDown:  "down",
		MoveLeft:  "left",
		MoveRight: "right",
		Inventory: "tab",
		Attack:    "space",
		Hack:      "h",
		UseItem:   "u",
		Flee:      "f",
		Confirm:   "return",
		Cancel:    "escape",
		Pause:     "p",
	}

	if cc.MoveUp != "up" {
		t.Error("MoveUp not set correctly")
	}
	if cc.Attack != "space" {
		t.Error("Attack not set correctly")
	}
}

func TestDebugConfigStruct(t *testing.T) {
	dc := DebugConfig{
		Enabled:   true,
		ShowFPS:   true,
		GodMode:   true,
		RevealMap: true,
		LogLevel:  "debug",
	}

	if !dc.Enabled {
		t.Error("Enabled not set correctly")
	}
	if !dc.GodMode {
		t.Error("GodMode not set correctly")
	}
	if dc.LogLevel != "debug" {
		t.Error("LogLevel not set correctly")
	}
}
