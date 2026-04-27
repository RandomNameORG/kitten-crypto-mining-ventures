package game

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
)

// TestStaleRateNewGameSeedsFromDef: NewState seeds the alley RoomState
// with the catalog StaleRate so a fresh game has the §7.2 fraction
// already wired without requiring migration.
func TestStaleRateNewGameSeedsFromDef(t *testing.T) {
	withTempHome(t)
	s := NewState("kit")
	rs, ok := s.Rooms["alley"]
	if !ok {
		t.Fatal("alley not unlocked")
	}
	def, _ := data.RoomByID("alley")
	if def.StaleRate <= 0 {
		t.Fatalf("alley def StaleRate non-positive: %v", def.StaleRate)
	}
	if math.Abs(rs.StaleRate-def.StaleRate) > 1e-9 {
		t.Errorf("alley RoomState.StaleRate=%v, want %v", rs.StaleRate, def.StaleRate)
	}
}

// TestStaleRateLegacyMigrationCopiesDefault: a save written before the
// stale system has RoomState.StaleRate=0. After LoadFrom it must be
// backfilled from the room's catalog default. Mirrors the shape of
// TestPoolMigrationLegacySave / TestPSUMigrationLegacySave.
func TestStaleRateLegacyMigrationCopiesDefault(t *testing.T) {
	withTempHome(t)
	legacy := &State{
		Version:     1,
		KittenName:  "legacy",
		BTC:         500,
		CurrentRoom: "alley",
		Rooms: map[string]*RoomState{
			"alley": {DefID: "alley", Heat: 20, MaxHeat: 80}, // StaleRate intentionally 0
		},
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
	def, _ := data.RoomByID("alley")
	rs := loaded.Rooms["alley"]
	if rs == nil {
		t.Fatal("alley dropped during migration")
	}
	if math.Abs(rs.StaleRate-def.StaleRate) > 1e-9 {
		t.Errorf("legacy migration left StaleRate=%v, want %v", rs.StaleRate, def.StaleRate)
	}
}

// TestEffectiveStaleRateLowRiskPoolNoModifier: scratch_pool is risk=low
// → 0 modifier, so EffectiveStaleRate returns the room baseline as-is.
// Anchors the conservative starter balance.
func TestEffectiveStaleRateLowRiskPoolNoModifier(t *testing.T) {
	withTempHome(t)
	s := NewState("kit")
	if got := s.CurrentPool().Risk; got != "low" {
		t.Fatalf("default pool risk=%q, want low", got)
	}
	got := s.EffectiveStaleRate("alley")
	want := s.Rooms["alley"].StaleRate
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("EffectiveStaleRate(alley)=%v, want %v (low-risk pool, no modifier)", got, want)
	}
}

// TestEffectiveStaleRateRiskAdditive: switching to a high-risk pool
// (whisker_fi) bumps the effective rate by the high-risk modifier
// (1pp). Tick past PoolSwitchSec so the post-transition convention
// kicks in (mid-switch we deliberately return the baseline).
func TestEffectiveStaleRateRiskAdditive(t *testing.T) {
	withTempHome(t)
	s := NewState("kit")
	const now int64 = 1_700_000_000
	if err := s.SwitchPool("whisker_fi", now); err != nil {
		t.Fatalf("SwitchPool: %v", err)
	}
	if s.CurrentPool().Risk != "high" {
		t.Fatalf("whisker_fi risk=%q, want high", s.CurrentPool().Risk)
	}
	// Mid-switch: baseline only.
	mid := s.EffectiveStaleRate("alley")
	if math.Abs(mid-s.Rooms["alley"].StaleRate) > 1e-9 {
		t.Errorf("mid-switch EffectiveStaleRate=%v, want baseline %v", mid, s.Rooms["alley"].StaleRate)
	}
	// Walk past the transition window.
	s.PoolSwitchAt = 0
	got := s.EffectiveStaleRate("alley")
	want := s.Rooms["alley"].StaleRate + staleRiskHigh
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("EffectiveStaleRate(alley) on whisker_fi=%v, want %v (base+high modifier)", got, want)
	}
}

// staleSimRun runs a single advanceMining-only sim path with a fixed
// alley StaleRate so the two-state comparison tests stay in sync. We
// deliberately skip MaybeFireEvent to keep the result purely mechanical
// — events would inject seed-driven noise that the comparison can't
// cleanly subtract out.
func staleSimRun(t *testing.T, seed int64, ticks int, staleRate float64) *State {
	t.Helper()
	SeedRNG(seed)
	s := NewState("stale-test")
	s.SetDifficulty("normal")
	s.LastTickUnix = simTestBaseUnix
	s.LastBillUnix = simTestBaseUnix
	s.LastWagesUnix = simTestBaseUnix
	s.LastMarketTickUnix = simTestBaseUnix
	s.StartedUnix = simTestBaseUnix
	if rs, ok := s.Rooms["alley"]; ok {
		rs.StaleRate = staleRate
	}
	for i := 1; i <= ticks; i++ {
		s.Tick(simTestBaseUnix + int64(i))
	}
	return s
}

// TestStaleRateAppliedToEarnings: two parallel sims on the same seed,
// one with StaleRate=0 in alley and one with StaleRate=0.10. The second
// should earn approximately 90% of the first. Tolerance is wide because
// the heat/efficiency-factor cliff at 80% MaxHeat can still introduce
// sub-percent drift between runs.
func TestStaleRateAppliedToEarnings(t *testing.T) {
	withTempHome(t)
	clean := staleSimRun(t, 1, 600, 0)
	withTempHome(t)
	stale := staleSimRun(t, 1, 600, 0.10)

	if clean.LifetimeEarned <= 0 {
		t.Fatalf("clean run earned nothing: %v", clean.LifetimeEarned)
	}
	if stale.LifetimeEarned >= clean.LifetimeEarned {
		t.Fatalf("stale run should earn less: clean=%v stale=%v", clean.LifetimeEarned, stale.LifetimeEarned)
	}
	ratio := stale.LifetimeEarned / clean.LifetimeEarned
	const want = 0.90
	if math.Abs(ratio-want) > 0.02 {
		t.Errorf("stale/clean earnings ratio=%v, want ~%v (±0.02): clean=%v stale=%v",
			ratio, want, clean.LifetimeEarned, stale.LifetimeEarned)
	}
}

// TestStaleRateAppliedToPPLNSShares: same shape as the earnings test
// but on the PPLNS share accumulator. Stale work doesn't earn shares
// either — they should drop in lockstep with earnings.
func TestStaleRateAppliedToPPLNSShares(t *testing.T) {
	withTempHome(t)
	clean := staleSimRun(t, 1, 600, 0)
	withTempHome(t)
	stale := staleSimRun(t, 1, 600, 0.10)

	if clean.PoolShares <= 0 {
		t.Fatalf("clean run accumulated no shares: %v", clean.PoolShares)
	}
	if stale.PoolShares >= clean.PoolShares {
		t.Fatalf("stale run should accumulate fewer shares: clean=%v stale=%v",
			clean.PoolShares, stale.PoolShares)
	}
	ratio := stale.PoolShares / clean.PoolShares
	const want = 0.90
	if math.Abs(ratio-want) > 0.02 {
		t.Errorf("stale/clean shares ratio=%v, want ~%v (±0.02): clean=%v stale=%v",
			ratio, want, clean.PoolShares, stale.PoolShares)
	}
}
