package save

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Manager handles save/load operations with background auto-save.
type Manager struct {
	saveDir          string
	autoSaveInterval time.Duration

	// Background save state
	saveChan chan saveRequest
	stopChan chan struct{}
	wg       sync.WaitGroup

	// Debouncing
	lastSave    time.Time
	minInterval time.Duration
	mu          sync.Mutex
}

type saveRequest struct {
	data    *SaveData
	trigger SaveTrigger
	done    chan error
}

// Config for the save manager.
type Config struct {
	SaveDir          string
	AutoSaveInterval time.Duration
	MinSaveInterval  time.Duration // Debounce threshold
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()
	return Config{
		SaveDir:          filepath.Join(homeDir, ".devdungeon", "saves"),
		AutoSaveInterval: 60 * time.Second,
		MinSaveInterval:  5 * time.Second,
	}
}

// NewManager creates a new save manager.
func NewManager(cfg Config) (*Manager, error) {
	// Create save directory if it doesn't exist
	if err := os.MkdirAll(cfg.SaveDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create save directory: %w", err)
	}

	m := &Manager{
		saveDir:          cfg.SaveDir,
		autoSaveInterval: cfg.AutoSaveInterval,
		minInterval:      cfg.MinSaveInterval,
		saveChan:         make(chan saveRequest, 10),
		stopChan:         make(chan struct{}),
	}

	return m, nil
}

// Start begins the background save goroutine.
func (m *Manager) Start() {
	m.wg.Add(1)
	go m.backgroundSaver()
}

// Stop gracefully shuts down the background saver.
func (m *Manager) Stop() {
	close(m.stopChan)
	m.wg.Wait()
}

// backgroundSaver handles save requests in the background.
func (m *Manager) backgroundSaver() {
	defer m.wg.Done()

	for {
		select {
		case req := <-m.saveChan:
			err := m.doSave(req.data, req.trigger)
			if req.done != nil {
				req.done <- err
			}
		case <-m.stopChan:
			// Drain remaining save requests
			for {
				select {
				case req := <-m.saveChan:
					err := m.doSave(req.data, req.trigger)
					if req.done != nil {
						req.done <- err
					}
				default:
					return
				}
			}
		}
	}
}

// Save queues a save request (non-blocking).
func (m *Manager) Save(data *SaveData, trigger SaveTrigger) {
	// Check debounce (except for critical saves)
	if trigger == TriggerAutoSave {
		m.mu.Lock()
		if time.Since(m.lastSave) < m.minInterval {
			m.mu.Unlock()
			return // Skip debounced auto-save
		}
		m.mu.Unlock()
	}

	select {
	case m.saveChan <- saveRequest{data: data, trigger: trigger}:
	default:
		// Channel full, skip this save (shouldn't happen often)
	}
}

// SaveSync saves and waits for completion.
func (m *Manager) SaveSync(data *SaveData, trigger SaveTrigger) error {
	done := make(chan error, 1)
	m.saveChan <- saveRequest{data: data, trigger: trigger, done: done}
	return <-done
}

// doSave performs the actual save operation.
func (m *Manager) doSave(data *SaveData, trigger SaveTrigger) error {
	m.mu.Lock()
	m.lastSave = time.Now()
	m.mu.Unlock()

	data.Timestamp = time.Now()

	// Determine filename
	filename := m.getRunSavePath(data.MasterSeed)

	// Marshal to JSON (pretty-printed for debugging)
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal save data: %w", err)
	}

	// Write atomically (write to temp, then rename)
	tempFile := filename + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}

	if err := os.Rename(tempFile, filename); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to finalize save file: %w", err)
	}

	return nil
}

// Load loads a save file for a given seed.
func (m *Manager) Load(seed int64) (*SaveData, error) {
	filename := m.getRunSavePath(seed)
	return m.LoadFromPath(filename)
}

// LoadFromPath loads a save file from a specific path.
func (m *Manager) LoadFromPath(path string) (*SaveData, error) {
	jsonData, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No save file exists
		}
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}

	var data SaveData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to parse save file: %w", err)
	}

	// Version check for future migrations
	if data.Version > Version {
		return nil, fmt.Errorf("save file version %d is newer than supported version %d", data.Version, Version)
	}

	return &data, nil
}

// LoadLatest loads the most recent save file.
func (m *Manager) LoadLatest() (*SaveData, error) {
	entries, err := os.ReadDir(m.saveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read save directory: %w", err)
	}

	var latestPath string
	var latestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestPath = filepath.Join(m.saveDir, entry.Name())
		}
	}

	if latestPath == "" {
		return nil, nil
	}

	return m.LoadFromPath(latestPath)
}

// DeleteSave removes a save file for a given seed.
func (m *Manager) DeleteSave(seed int64) error {
	filename := m.getRunSavePath(seed)
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete save file: %w", err)
	}
	return nil
}

// ListSaves returns all available save files.
func (m *Manager) ListSaves() ([]SaveInfo, error) {
	entries, err := os.ReadDir(m.saveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read save directory: %w", err)
	}

	var saves []SaveInfo
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(m.saveDir, entry.Name())
		data, err := m.LoadFromPath(path)
		if err != nil {
			continue // Skip corrupted saves
		}

		saves = append(saves, SaveInfo{
			Seed:        data.MasterSeed,
			Timestamp:   data.Timestamp,
			Depth:       data.CurrentDepth,
			PlayerClass: string(data.Player.Class),
			Level:       data.Player.Level,
		})
	}

	return saves, nil
}

// SaveInfo provides summary info about a save file.
type SaveInfo struct {
	Seed        int64
	Timestamp   time.Time
	Depth       int
	PlayerClass string
	Level       int
}

// getRunSavePath returns the path for a run's save file.
func (m *Manager) getRunSavePath(seed int64) string {
	return filepath.Join(m.saveDir, fmt.Sprintf("run_%d.json", seed))
}

// LoadMetaProgress loads the persistent meta-progress file.
func (m *Manager) LoadMetaProgress() (*MetaProgress, error) {
	path := filepath.Join(m.saveDir, "meta_progress.json")

	jsonData, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			meta := NewMetaProgress()
			return &meta, nil
		}
		return nil, fmt.Errorf("failed to read meta progress: %w", err)
	}

	var meta MetaProgress
	if err := json.Unmarshal(jsonData, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse meta progress: %w", err)
	}

	return &meta, nil
}

// SaveMetaProgress saves the persistent meta-progress file.
func (m *Manager) SaveMetaProgress(meta *MetaProgress) error {
	path := filepath.Join(m.saveDir, "meta_progress.json")

	jsonData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal meta progress: %w", err)
	}

	tempFile := path + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write meta progress: %w", err)
	}

	if err := os.Rename(tempFile, path); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to finalize meta progress: %w", err)
	}

	return nil
}
