// Package entity provides game entity types for /dev/dungeon.
package entity

import "github.com/iheanyi/devdungeon/internal/types"

// Entity is the base interface for all game entities.
type Entity interface {
	ID() string
	Name() string
	Position() types.Position
	SetPosition(pos types.Position)
	Glyph() rune
	IsBlocking() bool
}

// BaseEntity provides common entity functionality.
type BaseEntity struct {
	id       string
	name     string
	position types.Position
	glyph    rune
	blocking bool
}

// ID returns the entity's unique identifier.
func (e *BaseEntity) ID() string { return e.id }

// Name returns the entity's display name.
func (e *BaseEntity) Name() string { return e.name }

// Position returns the entity's current position.
func (e *BaseEntity) Position() types.Position { return e.position }

// SetPosition sets the entity's position.
func (e *BaseEntity) SetPosition(pos types.Position) { e.position = pos }

// Glyph returns the ASCII character representing this entity.
func (e *BaseEntity) Glyph() rune { return e.glyph }

// IsBlocking returns whether this entity blocks movement.
func (e *BaseEntity) IsBlocking() bool { return e.blocking }

// NewBaseEntity creates a new base entity.
func NewBaseEntity(id, name string, pos types.Position, glyph rune, blocking bool) *BaseEntity {
	return &BaseEntity{
		id:       id,
		name:     name,
		position: pos,
		glyph:    glyph,
		blocking: blocking,
	}
}
