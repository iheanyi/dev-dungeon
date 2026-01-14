package dungeon

import (
	"math/rand"

	"github.com/iheanyi/devdungeon/internal/types"
)

// connectRooms connects all rooms in the BSP tree using corridors.
func connectRooms(rng *rand.Rand, floor *types.Floor, root *BSPNode) {
	// Connect rooms through the BSP tree
	connectBSPNode(rng, floor, root)

	// Mark all rooms as connected after connecting through BSP
	for i := range floor.Rooms {
		floor.Rooms[i].Connected = true
	}
}

// connectBSPNode recursively connects rooms in sibling nodes.
func connectBSPNode(rng *rand.Rand, floor *types.Floor, node *BSPNode) {
	if node == nil || node.IsLeaf() {
		return
	}

	// Recursively connect children first
	connectBSPNode(rng, floor, node.Left)
	connectBSPNode(rng, floor, node.Right)

	// Connect left and right subtrees
	if node.Left != nil && node.Right != nil {
		leftRoom := node.Left.GetRoom()
		rightRoom := node.Right.GetRoom()

		if leftRoom != nil && rightRoom != nil {
			connectTwoRooms(rng, floor, leftRoom, rightRoom)
		}
	}
}

// connectTwoRooms creates a corridor between two rooms.
func connectTwoRooms(rng *rand.Rand, floor *types.Floor, room1, room2 *types.Room) {
	// Get center points of each room
	center1 := room1.Center()
	center2 := room2.Center()

	// Randomly choose L-shaped corridor direction
	if rng.Intn(2) == 0 {
		// Horizontal then vertical
		carveHorizontalCorridor(floor, center1.X, center2.X, center1.Y)
		carveVerticalCorridor(floor, center1.Y, center2.Y, center2.X)
	} else {
		// Vertical then horizontal
		carveVerticalCorridor(floor, center1.Y, center2.Y, center1.X)
		carveHorizontalCorridor(floor, center1.X, center2.X, center2.Y)
	}
}

// carveHorizontalCorridor carves a horizontal corridor from x1 to x2 at y.
func carveHorizontalCorridor(floor *types.Floor, x1, x2, y int) {
	// Ensure x1 <= x2
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	for x := x1; x <= x2; x++ {
		pos := types.Position{X: x, Y: y}
		if floor.InBounds(pos) {
			tile := floor.GetTile(pos)
			// Only carve if it's a wall (don't overwrite rooms or stairs)
			if tile != nil && tile.Type == types.TileWall {
				floor.SetTile(pos, types.NewTile(types.TileFloor))
			}
		}
	}
}

// carveVerticalCorridor carves a vertical corridor from y1 to y2 at x.
func carveVerticalCorridor(floor *types.Floor, y1, y2, x int) {
	// Ensure y1 <= y2
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	for y := y1; y <= y2; y++ {
		pos := types.Position{X: x, Y: y}
		if floor.InBounds(pos) {
			tile := floor.GetTile(pos)
			// Only carve if it's a wall (don't overwrite rooms or stairs)
			if tile != nil && tile.Type == types.TileWall {
				floor.SetTile(pos, types.NewTile(types.TileFloor))
			}
		}
	}
}

// CarveDoor places a door at a position (optional enhancement).
func carveDoor(floor *types.Floor, pos types.Position) {
	if floor.InBounds(pos) {
		floor.SetTile(pos, types.NewTile(types.TileDoor))
	}
}
