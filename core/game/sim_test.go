package game

import (
	"math"
	"testing"
)

// simTestBaseUnix matches the fixed epoch used by cmd/meowmine-sim, so tests
// and the binary exercise the same pinned-time progression.
const simTestBaseUnix int64 = 1_700_000_000

// runSim mirrors cmd/meowmine-sim's inner loop: seed the RNG, pin every
// timestamp to a fixed epoch, advance N virtual seconds via Tick + event
// rolls. Kept package-private so the test file is the only caller.
func runSim(t *testing.T, seed int64, ticks int) *State {
	t.Helper()
	SeedRNG(seed)
	s := NewState("sim-test")
	s.SetDifficulty("normal")
	s.LastTickUnix = simTestBaseUnix
	s.LastBillUnix = simTestBaseUnix
	s.LastWagesUnix = simTestBaseUnix
	s.StartedUnix = simTestBaseUnix
	for i := 1; i <= ticks; i++ {
		s.Tick(simTestBaseUnix + int64(i))
		_ = s.MaybeFireEvent()
	}
	return s
}

// TestSimLongRunSanity catches the class of regression the simulator exists
// for: a broken tick path that silently produces NaN earnings, leaves the
// clock stuck, or skips the billing subsystem entirely.
func TestSimLongRunSanity(t *testing.T) {
	withTempHome(t)
	s := runSim(t, 1, 3600) // 1 virtual hour

	if math.IsNaN(s.BTC) || math.IsInf(s.BTC, 0) {
		t.Fatalf("BTC became non-finite: %v", s.BTC)
	}
	if math.IsNaN(s.LifetimeEarned) || s.LifetimeEarned < 0 {
		t.Fatalf("LifetimeEarned invalid: %v", s.LifetimeEarned)
	}
	wantTick := simTestBaseUnix + 3600
	if s.LastTickUnix != wantTick {
		t.Fatalf("LastTickUnix = %d, want %d", s.LastTickUnix, wantTick)
	}
	// Billing fires every 60s. After an hour we expect LastBillUnix to have
	// moved off its starting value — if it hasn't, advanceBilling isn't
	// being reached.
	if s.LastBillUnix <= simTestBaseUnix {
		t.Errorf("billing never advanced in 1h of ticks (LastBillUnix=%d)", s.LastBillUnix)
	}
	// Starter GPU should still exist (may be running or broken, but not gone).
	if len(s.GPUs) == 0 {
		t.Error("GPU list emptied itself during tick loop")
	}
}

// TestSimDeterministicGameState asserts that two fresh runs with the same
// seed produce identical *game* fields. Timestamp fields stamped via
// time.Now() inside appendLog/ShipsAt drift by milliseconds between runs and
// are intentionally excluded — this test is about game logic determinism,
// not wall-clock bookkeeping.
func TestSimDeterministicGameState(t *testing.T) {
	withTempHome(t)
	a := runSim(t, 42, 1800)
	withTempHome(t)
	b := runSim(t, 42, 1800)

	if a.BTC != b.BTC {
		t.Errorf("BTC drift: %v vs %v", a.BTC, b.BTC)
	}
	if a.LifetimeEarned != b.LifetimeEarned {
		t.Errorf("LifetimeEarned drift: %v vs %v", a.LifetimeEarned, b.LifetimeEarned)
	}
	if a.TechPoint != b.TechPoint {
		t.Errorf("TechPoint drift: %d vs %d", a.TechPoint, b.TechPoint)
	}
	if a.Reputation != b.Reputation {
		t.Errorf("Reputation drift: %d vs %d", a.Reputation, b.Reputation)
	}
	if a.Karma != b.Karma {
		t.Errorf("Karma drift: %d vs %d", a.Karma, b.Karma)
	}
	if len(a.GPUs) != len(b.GPUs) {
		t.Errorf("GPU count drift: %d vs %d", len(a.GPUs), len(b.GPUs))
	}
	if len(a.Modifiers) != len(b.Modifiers) {
		t.Errorf("Modifier count drift: %d vs %d", len(a.Modifiers), len(b.Modifiers))
	}
	if len(a.Achievements) != len(b.Achievements) {
		t.Errorf("Achievement count drift: %d vs %d", len(a.Achievements), len(b.Achievements))
	}
}

// TestSimSeedsDiverge proves that seeding actually threads into the game —
// if somebody refactored RNG calls to use a non-seeded source, this catches
// it by observing that two seeds produce at least one different outcome.
func TestSimSeedsDiverge(t *testing.T) {
	withTempHome(t)
	a := runSim(t, 1, 1800)
	withTempHome(t)
	b := runSim(t, 2, 1800)

	same := a.BTC == b.BTC &&
		a.Reputation == b.Reputation &&
		a.TechPoint == b.TechPoint &&
		len(a.Log) == len(b.Log) &&
		len(a.Modifiers) == len(b.Modifiers)
	if same {
		t.Fatal("seed=1 and seed=2 produced identical observable state — RNG is probably not being threaded through")
	}
}
