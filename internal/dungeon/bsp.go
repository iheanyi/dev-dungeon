package dungeon

import (
	"math/rand"

	"github.com/iheanyi/devdungeon/internal/types"
)

// BSPNode represents a node in the Binary Space Partition tree.
type BSPNode struct {
	X, Y          int
	Width, Height int
	Left, Right   *BSPNode
	Room          *types.Room
}

// IsLeaf returns true if this node has no children.
func (n *BSPNode) IsLeaf() bool {
	return n.Left == nil && n.Right == nil
}

// Split recursively divides the node into smaller cells.
func (n *BSPNode) Split(rng *rand.Rand, cfg Config) {
	// Don't split if already small enough
	if n.Width <= cfg.MaxCellSize && n.Height <= cfg.MaxCellSize {
		// Chance to not split even if we could
		if rng.Intn(100) >= cfg.SplitChance {
			return
		}
	}

	// Must split if too large
	mustSplit := n.Width > cfg.MaxCellSize || n.Height > cfg.MaxCellSize

	// Can't split if too small
	canSplitHorizontal := n.Height >= cfg.MinCellSize*2
	canSplitVertical := n.Width >= cfg.MinCellSize*2

	if !canSplitHorizontal && !canSplitVertical {
		return
	}

	// Random split unless forced by size constraints
	if !mustSplit && rng.Intn(100) >= cfg.SplitChance {
		return
	}

	// Decide split direction
	var splitHorizontal bool
	if canSplitHorizontal && canSplitVertical {
		// Prefer splitting the longer dimension
		if n.Width > n.Height {
			splitHorizontal = rng.Float32() < 0.3
		} else if n.Height > n.Width {
			splitHorizontal = rng.Float32() < 0.7
		} else {
			splitHorizontal = rng.Intn(2) == 0
		}
	} else {
		splitHorizontal = canSplitHorizontal
	}

	// Calculate split position
	var splitPos int
	if splitHorizontal {
		// Split between MinCellSize and Height-MinCellSize
		minPos := cfg.MinCellSize
		maxPos := n.Height - cfg.MinCellSize
		if maxPos <= minPos {
			return
		}
		splitPos = minPos + rng.Intn(maxPos-minPos+1)

		n.Left = &BSPNode{
			X:      n.X,
			Y:      n.Y,
			Width:  n.Width,
			Height: splitPos,
		}
		n.Right = &BSPNode{
			X:      n.X,
			Y:      n.Y + splitPos,
			Width:  n.Width,
			Height: n.Height - splitPos,
		}
	} else {
		// Vertical split
		minPos := cfg.MinCellSize
		maxPos := n.Width - cfg.MinCellSize
		if maxPos <= minPos {
			return
		}
		splitPos = minPos + rng.Intn(maxPos-minPos+1)

		n.Left = &BSPNode{
			X:      n.X,
			Y:      n.Y,
			Width:  splitPos,
			Height: n.Height,
		}
		n.Right = &BSPNode{
			X:      n.X + splitPos,
			Y:      n.Y,
			Width:  n.Width - splitPos,
			Height: n.Height,
		}
	}

	// Recursively split children
	n.Left.Split(rng, cfg)
	n.Right.Split(rng, cfg)
}

// GetLeaves returns all leaf nodes in the BSP tree.
func (n *BSPNode) GetLeaves() []*BSPNode {
	if n.IsLeaf() {
		return []*BSPNode{n}
	}

	var leaves []*BSPNode
	if n.Left != nil {
		leaves = append(leaves, n.Left.GetLeaves()...)
	}
	if n.Right != nil {
		leaves = append(leaves, n.Right.GetLeaves()...)
	}
	return leaves
}

// GetRoom returns the room in this node or a child's room if this is not a leaf.
func (n *BSPNode) GetRoom() *types.Room {
	if n.Room != nil {
		return n.Room
	}

	// Get room from children
	if n.Left != nil {
		if room := n.Left.GetRoom(); room != nil {
			return room
		}
	}
	if n.Right != nil {
		if room := n.Right.GetRoom(); room != nil {
			return room
		}
	}

	return nil
}

// createRoomInCell creates a room within a BSP cell.
func createRoomInCell(rng *rand.Rand, cell *BSPNode, cfg Config) *types.Room {
	// Calculate available space for room
	maxWidth := cell.Width - cfg.RoomPadding*2
	maxHeight := cell.Height - cfg.RoomPadding*2

	// Check if there's enough space
	if maxWidth < cfg.MinRoomSize || maxHeight < cfg.MinRoomSize {
		return nil
	}

	// Clamp max room size to available space
	roomMaxW := min(maxWidth, cfg.MaxRoomSize)
	roomMaxH := min(maxHeight, cfg.MaxRoomSize)

	// Generate random room size
	width := cfg.MinRoomSize
	if roomMaxW > cfg.MinRoomSize {
		width = cfg.MinRoomSize + rng.Intn(roomMaxW-cfg.MinRoomSize+1)
	}

	height := cfg.MinRoomSize
	if roomMaxH > cfg.MinRoomSize {
		height = cfg.MinRoomSize + rng.Intn(roomMaxH-cfg.MinRoomSize+1)
	}

	// Generate random position within cell (with padding)
	maxX := cell.X + cell.Width - width - cfg.RoomPadding
	maxY := cell.Y + cell.Height - height - cfg.RoomPadding
	minX := cell.X + cfg.RoomPadding
	minY := cell.Y + cfg.RoomPadding

	x := minX
	if maxX > minX {
		x = minX + rng.Intn(maxX-minX+1)
	}

	y := minY
	if maxY > minY {
		y = minY + rng.Intn(maxY-minY+1)
	}

	room := &types.Room{
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		Connected: false,
	}

	// Store room in cell
	cell.Room = room

	return room
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
