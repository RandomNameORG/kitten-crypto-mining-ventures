package game

import "math/rand"

// SeedRNG seeds the global math/rand source used throughout the simulation
// (shipping delays, scrap fragments, upgrade failures, event rolls). Call this
// before touching state if you want reproducible runs — matches the seeding
// pattern already used in core/game tests.
func SeedRNG(seed int64) { rand.Seed(seed) }
