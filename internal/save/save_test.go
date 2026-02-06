package save

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/types"
)

func TestNewMetaProgress(t *testing.T) {
	meta := NewMetaProgress()

	if len(meta.UnlockedClasses) != 1 {
		t.Errorf("expected 1 unlocked class, got %d", len(meta.UnlockedClasses))
	}

	if meta.UnlockedClasses[0] != "init" {
		t.Errorf("expected 'init' class, got '%s'", meta.UnlockedClasses[0])
	}

	if meta.TotalExitCodes != 0 {
		t.Errorf("expected 0 exit codes, got %d", meta.TotalExitCodes)
	}

	if len(meta.UnlockedItems) != 0 {
		t.Errorf("expected 0 unlocked items, got %d", len(meta.UnlockedItems))
	}
}

func TestSaveTriggerString(t *testing.T) {
	testCases := []struct {
		trigger  SaveTrigger
		expected string
	}{
		{TriggerFloorTransition, "floor_transition"},
		{TriggerCombatVictory, "combat_victory"},
		{TriggerRareItemPickup, "rare_item_pickup"},
		{TriggerAutoSave, "auto_save"},
		{TriggerManual, "manual"},
		{TriggerQuit, "quit"},
		{SaveTrigger(99), "unknown"}, // Invalid trigger
	}

	for _, tc := range testCases {
		result := tc.trigger.String()
		if result != tc.expected {
			t.Errorf("expected '%s', got '%s'", tc.expected, result)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.SaveDir == "" {
		t.Error("SaveDir should not be empty")
	}

	if cfg.AutoSaveInterval != 60*time.Second {
		t.Errorf("expected 60s auto save interval, got %v", cfg.AutoSaveInterval)
	}

	if cfg.MinSaveInterval != 5*time.Second {
		t.Errorf("expected 5s min interval, got %v", cfg.MinSaveInterval)
	}
}

func TestNewManager(t *testing.T) {
	// Use temp directory
	tempDir := t.TempDir()

	cfg := Config{
		SaveDir:          tempDir,
		AutoSaveInterval: 1 * time.Second,
		MinSaveInterval:  100 * time.Millisecond,
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if manager == nil {
		t.Fatal("manager should not be nil")
	}
}

func TestNewManagerCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	saveDir := filepath.Join(tempDir, "new_saves_dir")

	cfg := Config{
		SaveDir:          saveDir,
		AutoSaveInterval: 1 * time.Second,
		MinSaveInterval:  100 * time.Millisecond,
	}

	_, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Directory should exist
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		t.Error("save directory should have been created")
	}
}

func TestManagerSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		SaveDir:          tempDir,
		AutoSaveInterval: 1 * time.Second,
		MinSaveInterval:  10 * time.Millisecond,
	}

	manager, _ := NewManager(cfg)
	manager.Start()
	defer manager.Stop()

	// Create save data
	saveData := &SaveData{
		Version:      Version,
		MasterSeed:   12345,
		CurrentDepth: 3,
		Player: PlayerData{
			Class:     entity.ClassBash,
			Level:     5,
			XP:        100,
			XPToLevel: 200,
			ExitCodes: 50,
			Stats: types.Stats{
				RAM:  80,
				CPU:  15,
				FD:   20,
				NICE: 5,
				UID:  1000,
			},
		},
	}

	// Save synchronously
	err := manager.SaveSync(saveData, TriggerManual)
	if err != nil {
		t.Fatalf("SaveSync failed: %v", err)
	}

	// Load the save
	loaded, err := manager.Load(12345)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded == nil {
		t.Fatal("loaded save should not be nil")
	}

	if loaded.MasterSeed != 12345 {
		t.Errorf("expected seed 12345, got %d", loaded.MasterSeed)
	}

	if loaded.CurrentDepth != 3 {
		t.Errorf("expected depth 3, got %d", loaded.CurrentDepth)
	}

	if loaded.Player.Class != entity.ClassBash {
		t.Errorf("expected class 'bash', got '%s'", loaded.Player.Class)
	}

	if loaded.Player.Level != 5 {
		t.Errorf("expected level 5, got %d", loaded.Player.Level)
	}
}

func TestManagerLoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir}

	manager, _ := NewManager(cfg)

	// Load non-existent save
	loaded, err := manager.Load(99999)
	if err != nil {
		t.Fatalf("Load should not error for non-existent save: %v", err)
	}

	if loaded != nil {
		t.Error("loaded should be nil for non-existent save")
	}
}

func TestManagerLoadLatest(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		SaveDir:         tempDir,
		MinSaveInterval: 0, // No debounce for test
	}

	manager, _ := NewManager(cfg)
	manager.Start()
	defer manager.Stop()

	// Create two saves
	save1 := &SaveData{
		Version:      Version,
		MasterSeed:   111,
		CurrentDepth: 2,
	}
	_ = manager.SaveSync(save1, TriggerManual)

	time.Sleep(50 * time.Millisecond) // Ensure different timestamps

	save2 := &SaveData{
		Version:      Version,
		MasterSeed:   222,
		CurrentDepth: 5,
	}
	_ = manager.SaveSync(save2, TriggerManual)

	// Load latest
	latest, err := manager.LoadLatest()
	if err != nil {
		t.Fatalf("LoadLatest failed: %v", err)
	}

	if latest == nil {
		t.Fatal("latest should not be nil")
	}

	if latest.MasterSeed != 222 {
		t.Errorf("expected latest seed 222, got %d", latest.MasterSeed)
	}
}

func TestManagerDeleteSave(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir}

	manager, _ := NewManager(cfg)
	manager.Start()
	defer manager.Stop()

	// Create save
	saveData := &SaveData{
		Version:    Version,
		MasterSeed: 12345,
	}
	_ = manager.SaveSync(saveData, TriggerManual)

	// Verify it exists
	loaded, _ := manager.Load(12345)
	if loaded == nil {
		t.Fatal("save should exist")
	}

	// Delete it
	err := manager.DeleteSave(12345)
	if err != nil {
		t.Fatalf("DeleteSave failed: %v", err)
	}

	// Verify it's gone
	loaded, _ = manager.Load(12345)
	if loaded != nil {
		t.Error("save should be deleted")
	}
}

func TestManagerListSaves(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir, MinSaveInterval: 0}

	manager, _ := NewManager(cfg)
	manager.Start()
	defer manager.Stop()

	// Create multiple saves
	for _, seed := range []int64{111, 222, 333} {
		saveData := &SaveData{
			Version:      Version,
			MasterSeed:   seed,
			CurrentDepth: int(seed % 10),
			Player: PlayerData{
				Class: entity.ClassInit,
				Level: 1,
			},
		}
		_ = manager.SaveSync(saveData, TriggerManual)
	}

	// List saves
	saves, err := manager.ListSaves()
	if err != nil {
		t.Fatalf("ListSaves failed: %v", err)
	}

	if len(saves) != 3 {
		t.Errorf("expected 3 saves, got %d", len(saves))
	}
}

func TestManagerMetaProgress(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir}

	manager, _ := NewManager(cfg)

	// Load initial meta (should create default)
	meta, err := manager.LoadMetaProgress()
	if err != nil {
		t.Fatalf("LoadMetaProgress failed: %v", err)
	}

	if meta == nil {
		t.Fatal("meta should not be nil")
	}

	// Modify and save
	meta.TotalExitCodes = 100
	meta.UnlockedClasses = append(meta.UnlockedClasses, "bash")

	err = manager.SaveMetaProgress(meta)
	if err != nil {
		t.Fatalf("SaveMetaProgress failed: %v", err)
	}

	// Load again
	loaded, _ := manager.LoadMetaProgress()
	if loaded.TotalExitCodes != 100 {
		t.Errorf("expected 100 exit codes, got %d", loaded.TotalExitCodes)
	}

	if len(loaded.UnlockedClasses) != 2 {
		t.Errorf("expected 2 unlocked classes, got %d", len(loaded.UnlockedClasses))
	}
}

func TestManagerAsyncSave(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		SaveDir:         tempDir,
		MinSaveInterval: 0,
	}

	manager, _ := NewManager(cfg)
	manager.Start()
	defer manager.Stop()

	// Queue async save
	saveData := &SaveData{
		Version:    Version,
		MasterSeed: 12345,
	}
	manager.Save(saveData, TriggerAutoSave)

	// Wait a bit for background processing
	time.Sleep(100 * time.Millisecond)

	// Verify save was written
	loaded, _ := manager.Load(12345)
	if loaded == nil {
		t.Error("async save should have been written")
	}
}

func TestManagerDebounce(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		SaveDir:         tempDir,
		MinSaveInterval: 500 * time.Millisecond, // Long debounce
	}

	manager, _ := NewManager(cfg)
	manager.Start()
	defer manager.Stop()

	// First save should go through
	save1 := &SaveData{Version: Version, MasterSeed: 111}
	_ = manager.SaveSync(save1, TriggerAutoSave)

	// Immediate second auto-save should be debounced (skipped)
	save2 := &SaveData{Version: Version, MasterSeed: 222}
	manager.Save(save2, TriggerAutoSave) // Non-sync, will be skipped

	time.Sleep(50 * time.Millisecond)

	// First should exist
	loaded, _ := manager.Load(111)
	if loaded == nil {
		t.Error("first save should exist")
	}

	// Second should NOT exist (debounced)
	loaded, _ = manager.Load(222)
	if loaded != nil {
		t.Error("second save should have been debounced")
	}
}

// === Data Structure Tests ===

func TestSaveDataStruct(t *testing.T) {
	data := SaveData{
		Version:      1,
		MasterSeed:   12345,
		Timestamp:    time.Now(),
		CurrentDepth: 5,
		Player: PlayerData{
			Class:     entity.ClassVim,
			Level:     10,
			XP:        500,
			ExitCodes: 75,
		},
		FloorStates: []FloorState{
			{Depth: 1, DeadEnemies: []string{"e1", "e2"}},
		},
	}

	if data.Version != 1 {
		t.Error("Version not set correctly")
	}
	if data.MasterSeed != 12345 {
		t.Error("MasterSeed not set correctly")
	}
	if len(data.FloorStates) != 1 {
		t.Error("FloorStates not set correctly")
	}
}

func TestPlayerDataStruct(t *testing.T) {
	player := PlayerData{
		Class:     entity.ClassSudo,
		Level:     99,
		XP:        9999,
		XPToLevel: 10000,
		ExitCodes: 1000,
		Position:  types.Position{X: 10, Y: 20},
		Stats: types.Stats{
			RAM:  200,
			CPU:  50,
			FD:   100,
			NICE: -5,
			UID:  0, // Root!
		},
		Inventory: []ItemData{
			{TemplateID: "malloc", Quantity: 5},
		},
		Equipment: EquipmentData{
			Weapon: "vim_blade",
			Armor:  "firewall",
		},
		Skills: []string{"sudo", "kill"},
	}

	if player.Class != entity.ClassSudo {
		t.Error("Class not set correctly")
	}
	if player.Stats.UID != 0 {
		t.Error("Should be root!")
	}
	if len(player.Inventory) != 1 {
		t.Error("Inventory not set correctly")
	}
}

func TestItemDataStruct(t *testing.T) {
	item := ItemData{
		TemplateID: "malloc",
		Quantity:   10,
	}

	if item.TemplateID != "malloc" {
		t.Error("TemplateID not set correctly")
	}
	if item.Quantity != 10 {
		t.Error("Quantity not set correctly")
	}
}

func TestEquipmentDataStruct(t *testing.T) {
	equip := EquipmentData{
		Weapon:   "vim_blade",
		Armor:    "firewall",
		Utility1: "malloc",
		Utility2: "realloc",
	}

	if equip.Weapon != "vim_blade" {
		t.Error("Weapon not set correctly")
	}
	if equip.Utility2 != "realloc" {
		t.Error("Utility2 not set correctly")
	}
}

func TestFloorStateStruct(t *testing.T) {
	state := FloorState{
		Depth:         3,
		ExploredTiles: []types.Position{{X: 1, Y: 2}, {X: 3, Y: 4}},
		DeadEnemies:   []string{"zombie_1", "daemon_2"},
		LootedItems:   []string{"item_a"},
	}

	if state.Depth != 3 {
		t.Error("Depth not set correctly")
	}
	if len(state.ExploredTiles) != 2 {
		t.Error("ExploredTiles not set correctly")
	}
	if len(state.DeadEnemies) != 2 {
		t.Error("DeadEnemies not set correctly")
	}
}

func TestStatBonusesStruct(t *testing.T) {
	bonuses := StatBonuses{
		RAM:  10,
		CPU:  5,
		FD:   8,
		NICE: -2,
		UID:  0,
	}

	if bonuses.RAM != 10 {
		t.Error("RAM not set correctly")
	}
	if bonuses.NICE != -2 {
		t.Error("NICE not set correctly")
	}
}

func TestSaveInfoStruct(t *testing.T) {
	info := SaveInfo{
		Seed:        12345,
		Timestamp:   time.Now(),
		Depth:       5,
		PlayerClass: "bash",
		Level:       10,
	}

	if info.Seed != 12345 {
		t.Error("Seed not set correctly")
	}
	if info.PlayerClass != "bash" {
		t.Error("PlayerClass not set correctly")
	}
}

// === Additional Coverage Tests ===

func TestLoadFromPathNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir}

	manager, _ := NewManager(cfg)

	// Load from non-existent path
	data, err := manager.LoadFromPath("/nonexistent/path/save.json")
	if err != nil {
		t.Errorf("LoadFromPath should not error for non-existent file: %v", err)
	}
	if data != nil {
		t.Error("should return nil for non-existent file")
	}
}

func TestLoadFromPathInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir}

	manager, _ := NewManager(cfg)

	// Write invalid JSON
	invalidPath := filepath.Join(tempDir, "invalid.json")
	_ = os.WriteFile(invalidPath, []byte("not valid json"), 0644)

	data, err := manager.LoadFromPath(invalidPath)
	if err == nil {
		t.Error("LoadFromPath should error for invalid JSON")
	}
	if data != nil {
		t.Error("should return nil for invalid JSON")
	}
}

func TestLoadFromPathNewerVersion(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir}

	manager, _ := NewManager(cfg)

	// Write save with future version
	futurePath := filepath.Join(tempDir, "future.json")
	futureData := fmt.Sprintf(`{"version": %d}`, Version+1)
	_ = os.WriteFile(futurePath, []byte(futureData), 0644)

	data, err := manager.LoadFromPath(futurePath)
	if err == nil {
		t.Error("LoadFromPath should error for newer version")
	}
	if data != nil {
		t.Error("should return nil for newer version")
	}
}

func TestLoadLatestEmpty(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir}

	manager, _ := NewManager(cfg)

	// Load latest from empty directory
	data, err := manager.LoadLatest()
	if err != nil {
		t.Errorf("LoadLatest should not error for empty dir: %v", err)
	}
	if data != nil {
		t.Error("should return nil for empty directory")
	}
}

func TestLoadLatestNonExistentDir(t *testing.T) {
	tempDir := t.TempDir()
	saveDir := filepath.Join(tempDir, "saves")
	cfg := Config{SaveDir: saveDir}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Remove the directory after manager creation
	os.RemoveAll(saveDir)

	// Load latest from removed directory should return nil
	data, err := manager.LoadLatest()
	if err != nil {
		t.Errorf("LoadLatest should not error for removed dir: %v", err)
	}
	if data != nil {
		t.Error("should return nil for non-existent directory")
	}
}

func TestLoadLatestSkipsDirectories(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir, MinSaveInterval: 0}

	manager, _ := NewManager(cfg)
	manager.Start()
	defer manager.Stop()

	// Create a subdirectory
	_ = os.MkdirAll(filepath.Join(tempDir, "subdir"), 0755)

	// Create a save
	saveData := &SaveData{
		Version:    Version,
		MasterSeed: 111,
	}
	_ = manager.SaveSync(saveData, TriggerManual)

	// Load latest should ignore the directory
	latest, err := manager.LoadLatest()
	if err != nil {
		t.Fatalf("LoadLatest failed: %v", err)
	}
	if latest == nil {
		t.Fatal("should find the save file")
	}
	if latest.MasterSeed != 111 {
		t.Errorf("expected seed 111, got %d", latest.MasterSeed)
	}
}

func TestLoadLatestSkipsNonJSONFiles(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir, MinSaveInterval: 0}

	manager, _ := NewManager(cfg)
	manager.Start()
	defer manager.Stop()

	// Create a non-JSON file
	_ = os.WriteFile(filepath.Join(tempDir, "readme.txt"), []byte("hello"), 0644)

	// Create a save
	saveData := &SaveData{
		Version:    Version,
		MasterSeed: 222,
	}
	_ = manager.SaveSync(saveData, TriggerManual)

	// Load latest should ignore the txt file
	latest, err := manager.LoadLatest()
	if err != nil {
		t.Fatalf("LoadLatest failed: %v", err)
	}
	if latest == nil {
		t.Fatal("should find the save file")
	}
	if latest.MasterSeed != 222 {
		t.Errorf("expected seed 222, got %d", latest.MasterSeed)
	}
}

func TestListSavesEmpty(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir}

	manager, _ := NewManager(cfg)

	saves, err := manager.ListSaves()
	if err != nil {
		t.Errorf("ListSaves should not error for empty dir: %v", err)
	}
	if len(saves) != 0 {
		t.Errorf("expected 0 saves, got %d", len(saves))
	}
}

func TestListSavesNonExistentDir(t *testing.T) {
	tempDir := t.TempDir()
	saveDir := filepath.Join(tempDir, "saves")
	cfg := Config{SaveDir: saveDir}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Remove the directory after manager creation
	os.RemoveAll(saveDir)

	saves, err := manager.ListSaves()
	if err != nil {
		t.Errorf("ListSaves should not error for removed dir: %v", err)
	}
	if len(saves) != 0 {
		t.Error("expected nil or empty for removed directory")
	}
}

func TestListSavesSkipsCorruptedFiles(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir, MinSaveInterval: 0}

	manager, _ := NewManager(cfg)
	manager.Start()
	defer manager.Stop()

	// Create a corrupted JSON file
	_ = os.WriteFile(filepath.Join(tempDir, "corrupted.json"), []byte("invalid json"), 0644)

	// Create a valid save
	saveData := &SaveData{
		Version:    Version,
		MasterSeed: 333,
		Player: PlayerData{
			Class: "init",
			Level: 1,
		},
	}
	_ = manager.SaveSync(saveData, TriggerManual)

	// List should skip corrupted and return valid save
	saves, err := manager.ListSaves()
	if err != nil {
		t.Fatalf("ListSaves failed: %v", err)
	}
	if len(saves) != 1 {
		t.Errorf("expected 1 valid save, got %d", len(saves))
	}
}

func TestSaveMetaProgressWrites(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	meta := NewMetaProgress()
	meta.TotalExitCodes = 100

	err = manager.SaveMetaProgress(&meta)
	if err != nil {
		t.Fatalf("SaveMetaProgress failed: %v", err)
	}

	// Verify file was created
	metaPath := filepath.Join(tempDir, "meta_progress.json")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error("meta_progress.json should exist")
	}
}

func TestLoadMetaProgressInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Write invalid JSON to meta file
	metaPath := filepath.Join(tempDir, "meta_progress.json")
	_ = os.WriteFile(metaPath, []byte("invalid json"), 0644)

	_, err = manager.LoadMetaProgress()
	if err == nil {
		t.Error("LoadMetaProgress should error for invalid JSON")
	}
}

func TestDeleteSaveNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{SaveDir: tempDir}

	manager, _ := NewManager(cfg)

	// Delete non-existent save should not error
	err := manager.DeleteSave(99999)
	if err != nil {
		t.Errorf("DeleteSave should not error for non-existent save: %v", err)
	}
}

func TestSaveFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	saveDir := filepath.Join(tempDir, "restricted_saves")

	cfg := Config{
		SaveDir:         saveDir,
		MinSaveInterval: 0,
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	manager.Start()
	defer manager.Stop()

	// Verify directory was created with 0700
	info, err := os.Stat(saveDir)
	if err != nil {
		t.Fatalf("failed to stat save directory: %v", err)
	}
	dirPerm := info.Mode().Perm()
	if dirPerm != 0700 {
		t.Errorf("save directory permissions = %o, want 0700", dirPerm)
	}

	// Save a file and check its permissions
	saveData := &SaveData{
		Version:    Version,
		MasterSeed: 12345,
		Player: PlayerData{
			Class: "init",
			Level: 1,
		},
	}
	if err := manager.SaveSync(saveData, TriggerManual); err != nil {
		t.Fatalf("SaveSync failed: %v", err)
	}

	// Check save file permissions
	savePath := filepath.Join(saveDir, fmt.Sprintf("run_%d.json", 12345))
	fileInfo, err := os.Stat(savePath)
	if err != nil {
		t.Fatalf("failed to stat save file: %v", err)
	}
	filePerm := fileInfo.Mode().Perm()
	if filePerm != 0600 {
		t.Errorf("save file permissions = %o, want 0600", filePerm)
	}

	// Check meta progress file permissions
	meta := NewMetaProgress()
	meta.TotalExitCodes = 50
	if err := manager.SaveMetaProgress(&meta); err != nil {
		t.Fatalf("SaveMetaProgress failed: %v", err)
	}

	metaPath := filepath.Join(saveDir, "meta_progress.json")
	metaInfo, err := os.Stat(metaPath)
	if err != nil {
		t.Fatalf("failed to stat meta file: %v", err)
	}
	metaPerm := metaInfo.Mode().Perm()
	if metaPerm != 0600 {
		t.Errorf("meta progress file permissions = %o, want 0600", metaPerm)
	}
}

func TestSaveSyncTimeout(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		SaveDir:         tempDir,
		MinSaveInterval: 0,
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	// Do NOT call manager.Start() - the background saver is not running,
	// so the save channel will not be drained and SaveSync should time out.

	saveData := &SaveData{
		Version:    Version,
		MasterSeed: 99999,
	}

	// Fill the channel buffer (capacity 10) to ensure the send also blocks
	for i := 0; i < 10; i++ {
		select {
		case manager.saveChan <- saveRequest{data: saveData, trigger: TriggerManual}:
		default:
		}
	}

	// SaveSync should time out since no background saver is draining the channel
	start := time.Now()
	err = manager.SaveSync(saveData, TriggerManual)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("SaveSync should return an error when it times out")
	}

	// Should not block for more than ~6 seconds (5s timeout + some slack)
	if elapsed > 7*time.Second {
		t.Errorf("SaveSync blocked for %v, expected ~5s timeout", elapsed)
	}
}
