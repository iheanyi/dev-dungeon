package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// submitToLeaderboard submits the current run to the multiplayer leaderboard.
func (m *Model) submitToLeaderboard(victory bool) {
	if m.leaderboardSubmitter == nil || !m.isMultiplayer {
		return // No submitter set or not in multiplayer mode
	}

	if m.player == nil || m.engine == nil {
		return // No game state
	}

	// Calculate score: base on floors cleared + level + bonus for victory
	floorsCleared := m.engine.CurrentDepth()
	score := floorsCleared * 100 // Base score from depth
	score += m.player.Level * 50 // Bonus for level
	if victory {
		score += 1000 // Victory bonus
	}

	// Get total play time for this run
	timeSeconds := m.GetTotalPlayTime()

	// Get game info
	class := string(m.player.Class)
	seed := m.engine.MasterSeed()
	runType := m.currentRunType
	if runType == "" {
		runType = "standard"
	}

	// Submit asynchronously (don't block the UI)
	go func() {
		// Error is logged in the server callback, don't disrupt the game
		_ = m.leaderboardSubmitter(score, floorsCleared, timeSeconds, class, seed, runType, victory)
	}()
}

// openLeaderboard initializes and opens the leaderboard view.
func (m *Model) openLeaderboard() {
	m.leaderboardCursor = 0
	m.leaderboardRunType = "all"
	m.leaderboardError = ""
	m.leaderboardEntries = nil

	// Fetch leaderboard data if we have a fetcher
	if m.leaderboardFetcher != nil {
		entries, err := m.leaderboardFetcher(m.leaderboardRunType, 20)
		if err != nil {
			m.leaderboardError = "Failed to load leaderboard"
		} else {
			m.leaderboardEntries = entries
		}
	}

	m.currentView = ViewLeaderboard
}

// updateLeaderboard handles leaderboard view input.
func (m *Model) updateLeaderboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.currentView = ViewMainMenu
	case "up", "k", "w":
		if len(m.leaderboardEntries) > 0 {
			m.leaderboardCursor--
			if m.leaderboardCursor < 0 {
				m.leaderboardCursor = len(m.leaderboardEntries) - 1
			}
		}
	case "down", "j", "s":
		if len(m.leaderboardEntries) > 0 {
			m.leaderboardCursor++
			if m.leaderboardCursor >= len(m.leaderboardEntries) {
				m.leaderboardCursor = 0
			}
		}
	case "left", "h", "a":
		// Cycle through run types: all -> standard -> daily -> all
		switch m.leaderboardRunType {
		case "all":
			m.leaderboardRunType = "daily"
		case "daily":
			m.leaderboardRunType = "standard"
		case "standard":
			m.leaderboardRunType = "all"
		}
		m.refreshLeaderboard()
	case "right", "l", "d":
		// Cycle through run types: all -> daily -> standard -> all
		switch m.leaderboardRunType {
		case "all":
			m.leaderboardRunType = "standard"
		case "standard":
			m.leaderboardRunType = "daily"
		case "daily":
			m.leaderboardRunType = "all"
		}
		m.refreshLeaderboard()
	case "r":
		// Refresh leaderboard
		m.refreshLeaderboard()
	}
	return m, nil
}

// refreshLeaderboard reloads the leaderboard data.
func (m *Model) refreshLeaderboard() {
	m.leaderboardCursor = 0
	m.leaderboardError = ""

	if m.leaderboardFetcher != nil {
		entries, err := m.leaderboardFetcher(m.leaderboardRunType, 20)
		if err != nil {
			m.leaderboardError = "Failed to load leaderboard"
		} else {
			m.leaderboardEntries = entries
		}
	}
}

// showDailyLeaderboard initializes and shows the daily leaderboard view.
func (m *Model) showDailyLeaderboard() {
	m.dailyLeaderboardDate = time.Now().UTC().Truncate(24 * time.Hour)
	m.refreshDailyLeaderboard()
	m.currentView = ViewDailyLeaderboard
}

// updateDailyLeaderboard handles daily leaderboard view input.
func (m *Model) updateDailyLeaderboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.currentView = ViewMainMenu
	case "left", "h", "a":
		// Navigate to previous day (max 7 days back)
		minDate := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -6)
		newDate := m.dailyLeaderboardDate.AddDate(0, 0, -1)
		if !newDate.Before(minDate) {
			m.dailyLeaderboardDate = newDate
			m.refreshDailyLeaderboard()
		}
	case "right", "l", "d":
		// Navigate to next day (can't go past today)
		today := time.Now().UTC().Truncate(24 * time.Hour)
		newDate := m.dailyLeaderboardDate.AddDate(0, 0, 1)
		if !newDate.After(today) {
			m.dailyLeaderboardDate = newDate
			m.refreshDailyLeaderboard()
		}
	case "r":
		// Refresh leaderboard
		m.refreshDailyLeaderboard()
	}
	return m, nil
}

// refreshDailyLeaderboard reloads the daily leaderboard data for the selected date.
func (m *Model) refreshDailyLeaderboard() {
	m.dailyLeaderboardError = ""
	m.dailyLeaderboardEntries = nil
	m.dailyPlayerRank = 0
	m.dailyPlayerEntry = nil

	if m.dailyLeaderboardFetcher != nil {
		entries, rank, playerEntry, err := m.dailyLeaderboardFetcher(m.dailyLeaderboardDate, 10, 0)
		if err != nil {
			m.dailyLeaderboardError = "Failed to load leaderboard"
		} else {
			m.dailyLeaderboardEntries = entries
			m.dailyPlayerRank = rank
			m.dailyPlayerEntry = playerEntry
		}
	}
}
