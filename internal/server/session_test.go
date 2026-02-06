package server

import (
	"encoding/json"
	"testing"

	"github.com/iheanyi/devdungeon/internal/db"
	"github.com/iheanyi/devdungeon/internal/save"
)

func TestDbMetaToSaveMeta(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result := dbMetaToSaveMeta(nil)
		if result != nil {
			t.Error("expected nil for nil input")
		}
	})

	t.Run("empty meta returns defaults", func(t *testing.T) {
		dbMeta := &db.MetaProgress{
			UserID: 1,
		}
		result := dbMetaToSaveMeta(dbMeta)
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		// Should have default unlocked classes
		if len(result.UnlockedClasses) != 1 || result.UnlockedClasses[0] != "init" {
			t.Errorf("UnlockedClasses = %v, want [init]", result.UnlockedClasses)
		}
		if result.UnlockedItems == nil {
			t.Error("UnlockedItems should not be nil")
		}
		// Zero bonuses
		if result.PermanentBonuses.RAM != 0 || result.PermanentBonuses.CPU != 0 {
			t.Error("bonuses should be zero for empty meta")
		}
	})

	t.Run("full meta with bonuses", func(t *testing.T) {
		bonuses := save.StatBonuses{
			RAM:  20,
			CPU:  5,
			FD:   10,
			NICE: -2,
			UID:  0,
		}
		bonusJSON, _ := json.Marshal(bonuses)

		dbMeta := &db.MetaProgress{
			UserID:           42,
			TotalExitCodes:   500,
			UnlockedClasses:  []string{"init", "bash", "vim"},
			PermanentBonuses: bonusJSON,
			UnlockedItems:    []string{"advanced_malloc", "thread_pool"},
			RunsCompleted:    3,
			DeepestFloor:     7,
			TotalDeaths:      5,
		}

		result := dbMetaToSaveMeta(dbMeta)
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.TotalExitCodes != 500 {
			t.Errorf("TotalExitCodes = %d, want 500", result.TotalExitCodes)
		}
		if result.PermanentBonuses.RAM != 20 {
			t.Errorf("RAM = %d, want 20", result.PermanentBonuses.RAM)
		}
		if result.PermanentBonuses.CPU != 5 {
			t.Errorf("CPU = %d, want 5", result.PermanentBonuses.CPU)
		}
		if result.PermanentBonuses.FD != 10 {
			t.Errorf("FD = %d, want 10", result.PermanentBonuses.FD)
		}
		if result.PermanentBonuses.NICE != -2 {
			t.Errorf("NICE = %d, want -2", result.PermanentBonuses.NICE)
		}
		if len(result.UnlockedClasses) != 3 {
			t.Errorf("UnlockedClasses = %v, want 3 items", result.UnlockedClasses)
		}
		if len(result.UnlockedItems) != 2 {
			t.Errorf("UnlockedItems = %v, want 2 items", result.UnlockedItems)
		}
		if result.RunsCompleted != 3 {
			t.Errorf("RunsCompleted = %d, want 3", result.RunsCompleted)
		}
		if result.DeepestFloor != 7 {
			t.Errorf("DeepestFloor = %d, want 7", result.DeepestFloor)
		}
		if result.TotalDeaths != 5 {
			t.Errorf("TotalDeaths = %d, want 5", result.TotalDeaths)
		}
	})

	t.Run("invalid JSON bonuses uses zero values", func(t *testing.T) {
		dbMeta := &db.MetaProgress{
			UserID:           1,
			TotalExitCodes:   100,
			UnlockedClasses:  []string{"init"},
			PermanentBonuses: []byte("invalid json"),
			UnlockedItems:    []string{},
		}

		result := dbMetaToSaveMeta(dbMeta)
		if result == nil {
			t.Fatal("expected non-nil result even with invalid JSON")
		}
		// Should fallback to zero bonuses
		if result.PermanentBonuses.RAM != 0 {
			t.Errorf("RAM = %d, want 0 (fallback)", result.PermanentBonuses.RAM)
		}
		// Other fields should still be set
		if result.TotalExitCodes != 100 {
			t.Errorf("TotalExitCodes = %d, want 100", result.TotalExitCodes)
		}
	})

	t.Run("backward compatible JSON tags", func(t *testing.T) {
		// The JSONB in the database uses old field names "pid" and "mem"
		// due to backward-compatible JSON tags on StatBonuses
		bonusJSON := []byte(`{"pid": 15, "cpu": 3, "mem": 8, "nice": -1, "uid": 0}`)

		dbMeta := &db.MetaProgress{
			UserID:           1,
			UnlockedClasses:  []string{"init"},
			PermanentBonuses: bonusJSON,
			UnlockedItems:    []string{},
		}

		result := dbMetaToSaveMeta(dbMeta)
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		// "pid" JSON key should map to RAM field
		if result.PermanentBonuses.RAM != 15 {
			t.Errorf("RAM (from 'pid') = %d, want 15", result.PermanentBonuses.RAM)
		}
		// "mem" JSON key should map to FD field
		if result.PermanentBonuses.FD != 8 {
			t.Errorf("FD (from 'mem') = %d, want 8", result.PermanentBonuses.FD)
		}
		if result.PermanentBonuses.CPU != 3 {
			t.Errorf("CPU = %d, want 3", result.PermanentBonuses.CPU)
		}
		if result.PermanentBonuses.NICE != -1 {
			t.Errorf("NICE = %d, want -1", result.PermanentBonuses.NICE)
		}
	})
}
