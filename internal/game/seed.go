package game

import "hash/fnv"

// DeriveFloorSeed generates a deterministic seed for a floor based on the master seed and depth.
// This ensures that the same master seed always generates the same floor at a given depth.
func DeriveFloorSeed(masterSeed int64, depth int) int64 {
	h := fnv.New64a()
	// Combine master seed and depth into a unique hash
	h.Write([]byte{
		byte(masterSeed >> 56),
		byte(masterSeed >> 48),
		byte(masterSeed >> 40),
		byte(masterSeed >> 32),
		byte(masterSeed >> 24),
		byte(masterSeed >> 16),
		byte(masterSeed >> 8),
		byte(masterSeed),
		byte(depth >> 24),
		byte(depth >> 16),
		byte(depth >> 8),
		byte(depth),
	})
	return int64(h.Sum64())
}

// DeriveEntitySeed generates a deterministic seed for entity placement on a floor.
func DeriveEntitySeed(floorSeed int64, entityIndex int) int64 {
	return DeriveFloorSeed(floorSeed, entityIndex+10000)
}

// DeriveItemSeed generates a deterministic seed for item placement on a floor.
func DeriveItemSeed(floorSeed int64, itemIndex int) int64 {
	return DeriveFloorSeed(floorSeed, itemIndex+20000)
}
