package game

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
)

// TestPSUMigrationLegacySave: a save that predates the PSU system has nil
// PSUUnits on every room. After LoadFrom, each room must have exactly one
// running psu_builtin so the rest of the engine has something to aggregate.
func TestPSUMigrationLegacySave(t *testing.T) {
	withTempHome(t)
	// Construct a legacy state (rooms set, PSUUnits intentionally nil).
	legacy := &State{
		Version:       1,
		KittenName:    "legacy",
		BTC:           500,
		CurrentRoom:   "alley",
		Rooms:         map[string]*RoomState{
			"alley":     {DefID: "alley", Heat: 20, MaxHeat: 80},
			"warehouse": {DefID: "warehouse", Heat: 20, MaxHeat: 95},
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
	if len(loaded.Rooms) != 2 {
		t.Fatalf("expected 2 rooms after load, got %d", len(loaded.Rooms))
	}
	for id, rs := range loaded.Rooms {
		if len(rs.PSUUnits) != 1 {
			t.Fatalf("room %s: expected 1 PSU after migration, got %d", id, len(rs.PSUUnits))
		}
		p := rs.PSUUnits[0]
		if p.DefID != "psu_builtin" {
			t.Errorf("room %s: expected psu_builtin, got %q", id, p.DefID)
		}
		if p.Status != "running" {
			t.Errorf("room %s: expected running, got %q", id, p.Status)
		}
	}
}

// TestPSUNewGameHasBuiltin: NewState seeds psu_builtin in alley via the
// default room-unlock loop.
func TestPSUNewGameHasBuiltin(t *testing.T) {
	withTempHome(t)
	s := NewState("kit")
	rs, ok := s.Rooms["alley"]
	if !ok {
		t.Fatal("alley not unlocked")
	}
	if len(rs.PSUUnits) != 1 {
		t.Fatalf("alley should have 1 builtin PSU, got %d", len(rs.PSUUnits))
	}
	if rs.PSUUnits[0].DefID != "psu_builtin" {
		t.Errorf("expected psu_builtin, got %q", rs.PSUUnits[0].DefID)
	}
	if rs.PSUUnits[0].Status != "running" {
		t.Errorf("expected running, got %q", rs.PSUUnits[0].Status)
	}
}

// TestPSUCapacityBlocksBuy: with only psu_trash (300W) installed, buying an
// rtx4090 must fail with a capacity error and BTC must stay put.
//
// To make the assertion meaningful inside the engine's scaled power_draw
// units (rtx4090 ≈ 12 in catalog) we pre-fill alley with hand-built
// rtx4090 instances directly on s.GPUs (bypassing slot/PSU checks) until
// the room is at capacity, then call BuyGPU to confirm the gate fires
// before BTC is debited.
func TestPSUCapacityBlocksBuy(t *testing.T) {
	withTempHome(t)
	s := NewState("kit")
	s.BTC = 100_000 // plenty to cover the 4090

	// Replace the builtin with psu_trash so the capacity ceiling is finite.
	rs := s.Rooms["alley"]
	if len(rs.PSUUnits) != 1 {
		t.Fatalf("expected starter builtin, got %d PSUs", len(rs.PSUUnits))
	}
	if err := s.InstallPSU("alley", "psu_trash"); err != nil {
		t.Fatalf("install psu_trash: %v", err)
	}
	if _, err := s.RemovePSU("alley", rs.PSUUnits[0].InstanceID); err != nil {
		t.Fatalf("remove builtin: %v", err)
	}
	if got := s.RoomPSUCapacity("alley"); got != 300 {
		t.Fatalf("after swap, capacity = %v, want 300", got)
	}

	// Pre-fill load. With rtx4090 base power_draw ≈ 12, 25 instances
	// already sit at 300 (cap) — adding one more must trip RoomCanFitGPU.
	for i := 0; i < 25; i++ {
		s.GPUs = append(s.GPUs, &GPU{
			InstanceID: s.NextGPUID,
			DefID:      "rtx4090",
			Status:     "running",
			Room:       "alley",
			HoursLeft:  100,
		})
		s.NextGPUID++
	}

	btcBefore := s.BTC
	err := s.BuyGPU("rtx4090")
	if err == nil {
		t.Fatal("BuyGPU should fail when PSU capacity is exceeded")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "psu") &&
		!strings.Contains(strings.ToLower(err.Error()), "capacity") {
		t.Errorf("expected PSU/capacity error, got: %v", err)
	}
	if s.BTC != btcBefore {
		t.Errorf("BTC changed despite failed buy: before=%v after=%v", btcBefore, s.BTC)
	}
}

// TestPSUOverloadExplodes: with psu_trash gating the room and a heavy
// pre-filled load, a deterministic seeded run must blow the PSU within
// the 1-hour window and break at least one GPU as collateral.
func TestPSUOverloadExplodes(t *testing.T) {
	withTempHome(t)
	SeedRNG(1)
	s := NewState("kit")
	s.BTC = 100_000

	// Same swap dance: install psu_trash, then drop the builtin.
	rs := s.Rooms["alley"]
	builtinID := rs.PSUUnits[0].InstanceID
	if err := s.InstallPSU("alley", "psu_trash"); err != nil {
		t.Fatalf("install psu_trash: %v", err)
	}
	if _, err := s.RemovePSU("alley", builtinID); err != nil {
		t.Fatalf("remove builtin: %v", err)
	}

	// Stuff alley with enough running rtx4090s that load > capacity ×
	// (1 + tolerance). psu_trash: 300W cap, 5% tol → trip line 315.
	// 33 × 12 ≈ 396 → factor 1.32, well past 1.05.
	for i := 0; i < 33; i++ {
		s.GPUs = append(s.GPUs, &GPU{
			InstanceID: s.NextGPUID,
			DefID:      "rtx4090",
			Status:     "running",
			Room:       "alley",
			HoursLeft:  10000, // huge so room.Heat doesn't wear them out first
		})
		s.NextGPUID++
	}

	// Cool the room enough that heat-driven wear doesn't break GPUs first.
	rs.Heat = 20

	// Walk forward 3600 virtual seconds, breaking out once the explosion has
	// fired. Tick is the only legitimate way to advance the per-second
	// overload roll.
	const baseEpoch int64 = 1_700_000_000
	s.LastTickUnix = baseEpoch
	s.LastBillUnix = baseEpoch
	s.LastWagesUnix = baseEpoch
	rs.LastHeatTickUnix = baseEpoch

	psuBlew := false
	gpuBroke := 0
	for i := 1; i <= 3600; i++ {
		s.Tick(baseEpoch + int64(i))
		if rs.PSUUnits[0].Status == "broken" {
			psuBlew = true
			for _, g := range s.GPUs {
				if g.Room == "alley" && g.Status == "broken" {
					gpuBroke++
				}
			}
			break
		}
	}
	if !psuBlew {
		t.Fatal("psu_trash never exploded under sustained heavy overload across 3600 ticks")
	}
	if gpuBroke < 1 {
		t.Errorf("explosion fired but no GPU was bricked (psu_trash explosion_damage=2)")
	}
}

// TestPSUReplacePauses: ReplacePSU must pause earnings in the affected
// room for psuReplacePauseSec (120s), and earnings must resume cleanly
// once that window passes.
func TestPSUReplacePauses(t *testing.T) {
	withTempHome(t)
	SeedRNG(1)
	s := NewState("kit")
	s.BTC = 100_000

	rs := s.Rooms["alley"]
	builtinID := rs.PSUUnits[0].InstanceID

	// Step the simulation briefly with the builtin so earnings are flowing
	// at a known cadence.
	const baseEpoch int64 = 1_700_000_000
	s.LastTickUnix = baseEpoch
	s.LastBillUnix = baseEpoch
	s.LastWagesUnix = baseEpoch
	rs.LastHeatTickUnix = baseEpoch
	s.Tick(baseEpoch + 5)
	if s.LifetimeEarned <= 0 {
		t.Fatalf("starter should be earning before replace, got %v", s.LifetimeEarned)
	}

	// Replace builtin with psu_silver650.
	if err := s.ReplacePSU("alley", builtinID, "psu_silver650"); err != nil {
		t.Fatalf("ReplacePSU: %v", err)
	}
	if !s.IsRoomPSUPaused("alley", baseEpoch+10) {
		t.Fatal("room should be paused immediately after ReplacePSU")
	}

	// Tick forward 60s — still inside the 120s pause window.
	earnedBefore := s.LifetimeEarned
	for i := 6; i <= 65; i++ {
		s.Tick(baseEpoch + int64(i))
	}
	if s.LifetimeEarned != earnedBefore {
		t.Errorf("earnings should be frozen during PSU replace pause: before=%v after=%v",
			earnedBefore, s.LifetimeEarned)
	}

	// Tick past the pause window.
	for i := 66; i <= 200; i++ {
		s.Tick(baseEpoch + int64(i))
	}
	if s.IsRoomPSUPaused("alley", baseEpoch+200) {
		t.Fatal("room should have resumed after 120s pause window")
	}
	if s.LifetimeEarned <= earnedBefore {
		t.Errorf("earnings should resume after pause: %v -> %v", earnedBefore, s.LifetimeEarned)
	}
}

// TestPSURemoveRefund: RemovePSU credits exactly 30% of original price.
func TestPSURemoveRefund(t *testing.T) {
	withTempHome(t)
	s := NewState("kit")
	s.BTC = 100_000

	if err := s.InstallPSU("alley", "psu_silver650"); err != nil {
		t.Fatalf("install psu_silver650: %v", err)
	}
	silver, _ := data.PSUByID("psu_silver650")
	expectedRefund := int(float64(silver.Price) * psuRefundFactor)

	// Find the silver instance.
	rs := s.Rooms["alley"]
	var silverID int
	for _, p := range rs.PSUUnits {
		if p.DefID == "psu_silver650" {
			silverID = p.InstanceID
		}
	}
	if silverID == 0 {
		t.Fatal("silver PSU not found after install")
	}

	btcBefore := s.BTC
	refund, err := s.RemovePSU("alley", silverID)
	if err != nil {
		t.Fatalf("RemovePSU: %v", err)
	}
	if refund != expectedRefund {
		t.Errorf("refund = %d, want %d (30%% of %d)", refund, expectedRefund, silver.Price)
	}
	if delta := s.BTC - btcBefore; delta != float64(expectedRefund) {
		t.Errorf("BTC delta = %v, want %d", delta, expectedRefund)
	}
}

// TestPSULegacySaveStaysBalanced: a migrated legacy save running through
// a 10-minute sim must not fire the PSU explosion or leave any PSU
// broken — the migration is balance-neutral by design (builtin tolerance
// 1.0 + capacity 100k means overload is structurally unreachable).
func TestPSULegacySaveStaysBalanced(t *testing.T) {
	withTempHome(t)
	s := runSim(t, 1, 600)

	for roomID, rs := range s.Rooms {
		for _, p := range rs.PSUUnits {
			if p.Status != "running" {
				t.Errorf("room %s: PSU %d status=%q after migration sim — overload should never fire on builtin",
					roomID, p.InstanceID, p.Status)
			}
		}
	}
	for _, entry := range s.Log {
		// crisis events are the bucket the explosion log lands in. Any
		// "PSU exploded" line means the overload roll fired against a
		// builtin — a regression in the migration's balance promise.
		if strings.Contains(strings.ToLower(entry.Text), "psu exploded") {
			t.Errorf("unexpected PSU explosion in legacy/migrated save: %q", entry.Text)
		}
	}
}
