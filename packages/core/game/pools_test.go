package game

import (
	"encoding/json"
	"math"
	"testing"
)

// TestPoolMigrationLegacySave: a save that predates the pool system has
// PoolID == "" on disk. After LoadFrom it must default to scratch_pool
// so the rest of the engine has a real PoolDef to consult.
func TestPoolMigrationLegacySave(t *testing.T) {
	withTempHome(t)
	legacy := &State{
		Version:        1,
		KittenName:     "legacy",
		BTC:            500,
		CurrentRoom:    "alley",
		Rooms:          map[string]*RoomState{"alley": {DefID: "alley", Heat: 20, MaxHeat: 80}},
		GPUs:           []*GPU{},
		NextGPUID:      1,
		Modifiers:      []Modifier{},
		EventCooldown:  EventCooldowns{},
		UnlockedSkills: map[string]bool{},
		Mercs:          []*Merc{},
		Blueprints:     []*Blueprint{},
		Log:            []LogEntry{},
		Difficulty:     "normal",
		MarketPrice:    1.0,
	}
	b, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("marshal legacy: %v", err)
	}
	loaded, err := LoadFrom(b)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if loaded.PoolID != "scratch_pool" {
		t.Errorf("legacy save migrated to PoolID=%q, want scratch_pool", loaded.PoolID)
	}
	if loaded.PoolSwitchAt != 0 {
		t.Errorf("legacy save loaded with PoolSwitchAt=%d, want 0 (stable)", loaded.PoolSwitchAt)
	}
}

// TestPoolNewGameDefaultsScratch: NewState seeds scratch_pool so a fresh
// game starts on the safe mainstream pool (matches §5 onboarding flow).
func TestPoolNewGameDefaultsScratch(t *testing.T) {
	withTempHome(t)
	s := NewState("kit")
	if s.PoolID != "scratch_pool" {
		t.Errorf("new game PoolID=%q, want scratch_pool", s.PoolID)
	}
	if s.IsPoolSwitching(s.LastTickUnix) {
		t.Error("new game should not be mid-switch")
	}
	if got := s.PoolSettlementMode(); got != "pplns" {
		t.Errorf("scratch_pool settlement = %q, want pplns", got)
	}
}

// TestSwitchPoolStartsTransition: SwitchPool sets the switch metadata
// and opens a 10-minute window (PoolSwitchSec).
func TestSwitchPoolStartsTransition(t *testing.T) {
	withTempHome(t)
	s := NewState("kit")
	const now int64 = 1_700_000_000

	if err := s.SwitchPool("kitten_hash", now); err != nil {
		t.Fatalf("SwitchPool: %v", err)
	}
	if s.PoolID != "kitten_hash" {
		t.Errorf("PoolID=%q, want kitten_hash", s.PoolID)
	}
	if s.PoolSwitchFrom != "scratch_pool" {
		t.Errorf("PoolSwitchFrom=%q, want scratch_pool", s.PoolSwitchFrom)
	}
	if s.PoolSwitchAt != now+PoolSwitchSec {
		t.Errorf("PoolSwitchAt=%d, want %d", s.PoolSwitchAt, now+PoolSwitchSec)
	}
	if !s.IsPoolSwitching(now + 1) {
		t.Error("IsPoolSwitching(now+1) should be true mid-window")
	}
	if s.IsPoolSwitching(now + PoolSwitchSec + 1) {
		t.Error("IsPoolSwitching after the window should be false")
	}
}

// TestSwitchPoolPausesMiningDuringTransition: confirm the structural
// pause actually blocks earnings while the window is open and that
// earnings resume cleanly once it closes.
func TestSwitchPoolPausesMiningDuringTransition(t *testing.T) {
	withTempHome(t)
	SeedRNG(1)
	s := NewState("kit")

	const baseEpoch int64 = 1_700_000_000
	s.LastTickUnix = baseEpoch
	s.LastBillUnix = baseEpoch
	s.LastWagesUnix = baseEpoch
	s.Rooms["alley"].LastHeatTickUnix = baseEpoch

	// Run 5s before switching so earnings are flowing at a known cadence.
	for i := 1; i <= 5; i++ {
		s.Tick(baseEpoch + int64(i))
	}
	earnedBeforeSwitch := s.LifetimeEarned
	if earnedBeforeSwitch <= 0 {
		t.Fatalf("starter should be earning before switch, got %v", earnedBeforeSwitch)
	}

	switchAt := baseEpoch + 5
	if err := s.SwitchPool("kitten_hash", switchAt); err != nil {
		t.Fatalf("SwitchPool: %v", err)
	}

	// Tick forward 300s — well inside the 600s pause window.
	for i := 6; i <= 305; i++ {
		s.Tick(baseEpoch + int64(i))
	}
	if s.LifetimeEarned != earnedBeforeSwitch {
		t.Errorf("earnings should be frozen during pool switch: before=%v after=%v",
			earnedBeforeSwitch, s.LifetimeEarned)
	}

	// Tick past the pause window.
	for i := 306; i <= 700; i++ {
		s.Tick(baseEpoch + int64(i))
	}
	if s.IsPoolSwitching(baseEpoch + 700) {
		t.Fatal("pool should have resumed after PoolSwitchSec")
	}
	if s.LifetimeEarned <= earnedBeforeSwitch {
		t.Errorf("earnings should resume after window: %v -> %v", earnedBeforeSwitch, s.LifetimeEarned)
	}
}

// TestSwitchOutOfPPLNSVoidsShares: §5.5 — leaving a PPLNS pool wipes any
// unsettled shares. Switching scratch_pool (PPLNS) → kitten_hash (PPS)
// must zero PoolShares.
func TestSwitchOutOfPPLNSVoidsShares(t *testing.T) {
	withTempHome(t)
	s := NewState("kit")
	s.PoolShares = 1234
	s.PoolID = "scratch_pool" // PPLNS
	const now int64 = 1_700_000_000
	if err := s.SwitchPool("kitten_hash", now); err != nil {
		t.Fatalf("SwitchPool: %v", err)
	}
	if s.PoolShares != 0 {
		t.Errorf("PoolShares=%v after leaving PPLNS, want 0", s.PoolShares)
	}
}

// TestSwitchOutOfPPSKeepsShares: PPS / PPS+ / Solo settlement is
// per-share-paid-immediately, so shares are an irrelevant accumulator
// at the time of departure. Leaving them alone is correct (next sprint
// the field will only ever be non-zero on PPLNS pools anyway, but the
// structural rule still has to hold today).
func TestSwitchOutOfPPSKeepsShares(t *testing.T) {
	withTempHome(t)
	s := NewState("kit")
	s.PoolShares = 1234
	s.PoolID = "kitten_hash" // PPS
	const now int64 = 1_700_000_000
	if err := s.SwitchPool("scratch_pool", now); err != nil {
		t.Fatalf("SwitchPool: %v", err)
	}
	if s.PoolShares != 1234 {
		t.Errorf("PoolShares=%v after leaving PPS, want 1234", s.PoolShares)
	}
}

// TestSwitchPoolRejectsInvalidAndCurrent: bad pool id and no-op switch
// both error out and leave state untouched.
func TestSwitchPoolRejectsInvalidAndCurrent(t *testing.T) {
	withTempHome(t)
	s := NewState("kit")
	const now int64 = 1_700_000_000

	priorID := s.PoolID
	priorAt := s.PoolSwitchAt
	if err := s.SwitchPool("does_not_exist", now); err == nil {
		t.Error("SwitchPool to unknown id should fail")
	}
	if s.PoolID != priorID || s.PoolSwitchAt != priorAt {
		t.Errorf("state mutated on bad-id switch: PoolID=%q PoolSwitchAt=%d", s.PoolID, s.PoolSwitchAt)
	}

	if err := s.SwitchPool(s.PoolID, now); err == nil {
		t.Error("SwitchPool to current pool should fail")
	}
	if s.PoolID != priorID || s.PoolSwitchAt != priorAt {
		t.Errorf("state mutated on no-op switch: PoolID=%q PoolSwitchAt=%d", s.PoolID, s.PoolSwitchAt)
	}
}

// TestPoolDefaultPathSimSurvives: a 1h sim on the default path must keep
// earning, stay finite, and end on scratch_pool — proves the structural
// wiring didn't break the migration's balance-neutral promise.
func TestPoolDefaultPathSimSurvives(t *testing.T) {
	withTempHome(t)
	s := runSim(t, 1, 3600)

	if s.PoolID != "scratch_pool" {
		t.Errorf("default sim ended on PoolID=%q, want scratch_pool", s.PoolID)
	}
	if math.IsNaN(s.LifetimeEarned) || math.IsInf(s.LifetimeEarned, 0) {
		t.Fatalf("LifetimeEarned non-finite: %v", s.LifetimeEarned)
	}
	if s.LifetimeEarned <= 0 {
		t.Errorf("LifetimeEarned=%v after 1h, expected positive", s.LifetimeEarned)
	}
	// Structural accumulator must be advancing on the default PPLNS pool
	// — if it's still 0 after an hour the wiring's broken.
	if s.PoolShares <= 0 {
		t.Errorf("PoolShares=%v after 1h on PPLNS, expected positive", s.PoolShares)
	}
}
