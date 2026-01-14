package entity

import (
	"testing"

	"github.com/iheanyi/devdungeon/internal/types"
)

func TestNewPlayer(t *testing.T) {
	classes := []PlayerClass{ClassInit, ClassCron, ClassBash, ClassVim, ClassSudo}

	for _, class := range classes {
		t.Run(string(class), func(t *testing.T) {
			player := NewPlayer(class)

			if player == nil {
				t.Fatal("NewPlayer returned nil")
			}

			if player.Class != class {
				t.Errorf("expected class %s, got %s", class, player.Class)
			}

			if player.Level != 1 {
				t.Errorf("expected level 1, got %d", player.Level)
			}

			if player.Stats.RAM <= 0 {
				t.Error("player should have positive RAM")
			}

			if player.Inventory == nil {
				t.Error("player should have inventory")
			}

			if player.Equipment == nil {
				t.Error("player should have equipment")
			}

			// All classes should have at least the basic attack
			if len(player.Skills) == 0 {
				t.Error("player should have at least one skill")
			}
		})
	}
}

func TestPlayerStartingGear(t *testing.T) {
	// Each class should start with some items
	classes := []PlayerClass{ClassInit, ClassCron, ClassBash, ClassVim, ClassSudo}

	for _, class := range classes {
		t.Run(string(class), func(t *testing.T) {
			player := NewPlayer(class)

			// Everyone starts with malloc items
			hasHealing := false
			for _, item := range player.Inventory.Items {
				if item.TemplateID == "malloc" {
					hasHealing = true
					break
				}
			}
			if !hasHealing {
				t.Error("all classes should start with healing items")
			}
		})
	}
}

func TestPlayerTakeDamage(t *testing.T) {
	player := NewPlayer(ClassInit)
	initialRAM := player.Stats.RAM

	// Take some damage
	died := player.TakeDamage(10)
	if died {
		t.Error("player should not die from 10 damage")
	}
	if player.Stats.RAM != initialRAM-10 {
		t.Errorf("expected RAM %d, got %d", initialRAM-10, player.Stats.RAM)
	}

	// Take lethal damage
	died = player.TakeDamage(player.Stats.RAM + 100)
	if !died {
		t.Error("player should die from lethal damage")
	}
	if player.Stats.RAM != 0 {
		t.Error("RAM should be 0 after death")
	}
}

func TestPlayerTakeDamageWithInvincibility(t *testing.T) {
	player := NewPlayer(ClassSudo)
	initialRAM := player.Stats.RAM

	// Add invincibility buff
	player.AddBuff(Buff{
		Type:     BuffInvincible,
		Name:     "Sudo Mode",
		Duration: 3,
	})

	// Take damage
	died := player.TakeDamage(100)

	if died {
		t.Error("invincible player should not die")
	}
	if player.Stats.RAM != initialRAM {
		t.Errorf("invincible player should not lose RAM: expected %d, got %d", initialRAM, player.Stats.RAM)
	}
}

func TestPlayerHeal(t *testing.T) {
	player := NewPlayer(ClassInit)
	player.Stats.RAM = 50

	player.Heal(30)
	if player.Stats.RAM != 80 {
		t.Errorf("expected RAM 80, got %d", player.Stats.RAM)
	}

	// Heal should not exceed max
	player.Heal(1000)
	if player.Stats.RAM != player.MaxStats.MaxRAM {
		t.Errorf("RAM should be capped at MaxRAM %d, got %d", player.MaxStats.MaxRAM, player.Stats.RAM)
	}
}

func TestPlayerUseFD(t *testing.T) {
	player := NewPlayer(ClassVim) // Vim has more FD
	initialFD := player.Stats.FD

	// Use some FD
	if !player.UseFD(5) {
		t.Error("should be able to use 5 FD")
	}
	if player.Stats.FD != initialFD-5 {
		t.Errorf("expected FD %d, got %d", initialFD-5, player.Stats.FD)
	}

	// Try to use more FD than available
	if player.UseFD(player.Stats.FD + 10) {
		t.Error("should not be able to use more FD than available")
	}
}

func TestPlayerRestoreFD(t *testing.T) {
	player := NewPlayer(ClassInit)
	player.Stats.FD = 5

	player.RestoreFD(5)
	if player.Stats.FD != 10 {
		t.Errorf("expected FD 10, got %d", player.Stats.FD)
	}

	// Restore should not exceed max
	player.RestoreFD(1000)
	if player.Stats.FD != player.MaxStats.MaxFD {
		t.Errorf("FD should be capped at MaxFD %d, got %d", player.MaxStats.MaxFD, player.Stats.FD)
	}
}

func TestPlayerGainXP(t *testing.T) {
	player := NewPlayer(ClassInit)
	initialLevel := player.Level
	initialXPToLevel := player.XPToLevel

	// Gain some XP (not enough to level)
	leveled := player.GainXP(10)
	if leveled {
		t.Error("should not level up from 10 XP")
	}
	if player.XP != 10 {
		t.Errorf("expected XP 10, got %d", player.XP)
	}

	// Gain enough XP to level
	leveled = player.GainXP(player.XPToLevel)
	if !leveled {
		t.Error("should level up")
	}
	if player.Level != initialLevel+1 {
		t.Errorf("expected level %d, got %d", initialLevel+1, player.Level)
	}
	if player.XPToLevel <= initialXPToLevel {
		t.Error("XP to next level should increase")
	}
}

func TestPlayerLevelUp(t *testing.T) {
	player := NewPlayer(ClassInit)
	initialMaxRAM := player.MaxStats.MaxRAM
	initialMaxFD := player.MaxStats.MaxFD
	initialCPU := player.Stats.CPU

	player.LevelUp()

	if player.Level != 2 {
		t.Errorf("expected level 2, got %d", player.Level)
	}
	if player.MaxStats.MaxRAM != initialMaxRAM+10 {
		t.Errorf("expected MaxRAM %d, got %d", initialMaxRAM+10, player.MaxStats.MaxRAM)
	}
	if player.MaxStats.MaxFD != initialMaxFD+2 {
		t.Errorf("expected MaxFD %d, got %d", initialMaxFD+2, player.MaxStats.MaxFD)
	}
	if player.Stats.CPU != initialCPU+2 {
		t.Errorf("expected CPU %d, got %d", initialCPU+2, player.Stats.CPU)
	}
	// Full heal on level up
	if player.Stats.RAM != player.MaxStats.MaxRAM {
		t.Error("should be fully healed on level up")
	}
	if player.Stats.FD != player.MaxStats.MaxFD {
		t.Error("FD should be fully restored on level up")
	}
}

func TestPlayerBuffs(t *testing.T) {
	player := NewPlayer(ClassInit)

	// Add a buff
	player.AddBuff(Buff{
		Type:     BuffStrength,
		Name:     "Test Strength",
		Duration: 3,
		Value:    5,
	})

	if !player.HasBuff(BuffStrength) {
		t.Error("player should have strength buff")
	}

	buff := player.GetBuff(BuffStrength)
	if buff == nil {
		t.Fatal("GetBuff returned nil")
	}
	if buff.Value != 5 {
		t.Errorf("expected buff value 5, got %d", buff.Value)
	}

	// Stacking strength buffs should add value
	player.AddBuff(Buff{
		Type:     BuffStrength,
		Name:     "More Strength",
		Duration: 2,
		Value:    3,
	})

	buff = player.GetBuff(BuffStrength)
	if buff.Value != 8 {
		t.Errorf("stacked buff should have value 8, got %d", buff.Value)
	}

	// Remove buff
	player.RemoveBuff(BuffStrength)
	if player.HasBuff(BuffStrength) {
		t.Error("buff should be removed")
	}
}

func TestPlayerTickBuffs(t *testing.T) {
	player := NewPlayer(ClassInit)
	player.Stats.RAM = 50 // Lower health to test regen

	player.AddBuff(Buff{
		Type:     BuffRegenRAM,
		Name:     "Regen",
		Duration: 2,
		Value:    10,
	})

	// First tick
	messages := player.TickBuffs()
	if len(messages) == 0 {
		t.Error("expected tick messages")
	}
	if player.Stats.RAM != 60 {
		t.Errorf("expected RAM 60 after regen, got %d", player.Stats.RAM)
	}

	// Buff should still be active (1 turn left)
	if !player.HasBuff(BuffRegenRAM) {
		t.Error("buff should still be active")
	}

	// Second tick - buff should expire
	player.TickBuffs()
	if player.HasBuff(BuffRegenRAM) {
		t.Error("buff should have expired")
	}
}

func TestPlayerEffectiveStats(t *testing.T) {
	player := NewPlayer(ClassInit)
	baseCPU := player.Stats.CPU
	baseNICE := player.Stats.NICE

	// Without buffs, effective stats equal base stats
	if player.GetEffectiveCPU() != baseCPU {
		t.Errorf("expected effective CPU %d, got %d", baseCPU, player.GetEffectiveCPU())
	}
	if player.GetEffectiveNICE() != baseNICE {
		t.Errorf("expected effective NICE %d, got %d", baseNICE, player.GetEffectiveNICE())
	}

	// Add strength buff
	player.AddBuff(Buff{Type: BuffStrength, Duration: 3, Value: 5})
	if player.GetEffectiveCPU() != baseCPU+5 {
		t.Errorf("expected effective CPU %d, got %d", baseCPU+5, player.GetEffectiveCPU())
	}

	// Add haste buff
	player.AddBuff(Buff{Type: BuffHaste, Duration: 3, Value: 3})
	if player.GetEffectiveNICE() != baseNICE-3 {
		t.Errorf("expected effective NICE %d, got %d", baseNICE-3, player.GetEffectiveNICE())
	}
}

func TestPlayerPosition(t *testing.T) {
	player := NewPlayer(ClassInit)

	pos := types.Position{X: 5, Y: 10}
	player.SetPosition(pos)

	if player.Position() != pos {
		t.Errorf("expected position %v, got %v", pos, player.Position())
	}
}

func TestNewPlayerWithBonuses(t *testing.T) {
	bonuses := PermanentBonuses{
		RAM:  20,
		CPU:  5,
		FD:   10,
		NICE: 2,
	}

	player := NewPlayerWithBonuses(ClassInit, bonuses)

	// Base init stats: RAM=100, CPU=10, FD=16, NICE=10
	if player.Stats.RAM != 120 {
		t.Errorf("expected RAM 120 (100 + 20 bonus), got %d", player.Stats.RAM)
	}
	if player.MaxStats.MaxRAM != 120 {
		t.Errorf("expected MaxRAM 120 (100 + 20 bonus), got %d", player.MaxStats.MaxRAM)
	}
	if player.Stats.CPU != 15 {
		t.Errorf("expected CPU 15 (10 + 5 bonus), got %d", player.Stats.CPU)
	}
	if player.Stats.FD != 26 {
		t.Errorf("expected FD 26 (16 + 10 bonus), got %d", player.Stats.FD)
	}
	if player.MaxStats.MaxFD != 26 {
		t.Errorf("expected MaxFD 26 (16 + 10 bonus), got %d", player.MaxStats.MaxFD)
	}
	if player.Stats.NICE != 8 {
		t.Errorf("expected NICE 8 (10 - 2 bonus), got %d", player.Stats.NICE)
	}
}

func TestNewPlayerWithBonusesNICEMinimum(t *testing.T) {
	// Test that NICE doesn't go below 1
	bonuses := PermanentBonuses{
		NICE: 100, // Should not make NICE go below 1
	}

	player := NewPlayerWithBonuses(ClassCron, bonuses) // Cron has NICE=5

	if player.Stats.NICE < 1 {
		t.Errorf("NICE should not go below 1, got %d", player.Stats.NICE)
	}
	if player.Stats.NICE != 1 {
		t.Errorf("expected NICE to be capped at 1, got %d", player.Stats.NICE)
	}
}

func TestPermanentBonusesDefaultNoEffect(t *testing.T) {
	// Zero bonuses should give same result as NewPlayer
	player1 := NewPlayer(ClassInit)
	player2 := NewPlayerWithBonuses(ClassInit, PermanentBonuses{})

	if player1.Stats.RAM != player2.Stats.RAM {
		t.Error("zero bonuses should not affect RAM")
	}
	if player1.Stats.CPU != player2.Stats.CPU {
		t.Error("zero bonuses should not affect CPU")
	}
	if player1.Stats.FD != player2.Stats.FD {
		t.Error("zero bonuses should not affect FD")
	}
	if player1.Stats.NICE != player2.Stats.NICE {
		t.Error("zero bonuses should not affect NICE")
	}
}
