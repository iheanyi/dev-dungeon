package ui

import (
	"github.com/iheanyi/devdungeon/internal/config"
	"github.com/iheanyi/devdungeon/internal/entity"
	"github.com/iheanyi/devdungeon/internal/game"
)

func newTestModel() *Model {
	cfg := config.DefaultConfig()
	m := New(cfg)
	return m
}

// Helper to create a model with game engine ready
func newTestModelWithEngine(seed int64) *Model {
	m := newTestModel()
	cfg := config.DefaultConfig()
	engine := game.NewEngine(cfg, seed)
	if err := engine.StartNewGame(entity.ClassInit); err != nil {
		panic(err)
	}
	m.engine = engine
	m.player = engine.Player()
	return m
}
