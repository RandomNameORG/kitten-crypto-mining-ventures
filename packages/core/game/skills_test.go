package game

import (
	"math"
	"testing"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
)

func TestUnlockSkillDeductsAndPrereqs(t *testing.T) {
	withTempHome(t)
	s := NewState("Skill")
	s.TechPoint = 10

	if err := s.UnlockSkill("undervolt_ii"); err == nil {
		t.Error("should require prereq undervolt_i")
	}
	if err := s.UnlockSkill("undervolt_i"); err != nil {
		t.Fatalf("unlock undervolt_i: %v", err)
	}
	if !s.HasSkill("undervolt_i") {
		t.Error("skill should be flagged unlocked")
	}
	if s.TechPoint != 7 {
		t.Errorf("expected 7 TP after spending 3, got %d", s.TechPoint)
	}
	if err := s.UnlockSkill("undervolt_i"); err == nil {
		t.Error("should refuse to unlock already-owned skill")
	}
}

func TestPowerDrawMultStacks(t *testing.T) {
	withTempHome(t)
	s := NewState("Power")
	s.TechPoint = 99
	_ = s.UnlockSkill("undervolt_i")
	_ = s.UnlockSkill("undervolt_ii")
	// Two 0.9 multipliers → 0.81.
	got := s.PowerDrawMult()
	if got < 0.80 || got > 0.82 {
		t.Errorf("expected PowerDrawMult ≈ 0.81, got %.3f", got)
	}
}

func TestHasUnlockGatesResearch(t *testing.T) {
	withTempHome(t)
	s := NewState("Gate")
	if s.HasUnlock("rd") {
		t.Error("rd should not be unlocked by default")
	}
	s.TechPoint = 99
	// Unlock entire engineer prerequisite chain up to rd_unlock.
	_ = s.UnlockSkill("undervolt_i")
	_ = s.UnlockSkill("undervolt_ii")
	_ = s.UnlockSkill("rd_unlock")
	if !s.HasUnlock("rd") {
		t.Error("rd should now be unlocked")
	}
}

func TestHireMercRequiresMoney(t *testing.T) {
	withTempHome(t)
	s := NewState("Merc")
	s.BTC = 50
	if err := s.HireMerc("tabby_guard"); err == nil {
		t.Error("should refuse hire without enough money")
	}
	s.BTC = 2000
	if err := s.HireMerc("tabby_guard"); err != nil {
		t.Fatalf("hire: %v", err)
	}
	if len(s.Mercs) != 1 {
		t.Errorf("expected 1 merc hired, got %d", len(s.Mercs))
	}
}

func TestBribeMercRaisesLoyalty(t *testing.T) {
	withTempHome(t)
	s := NewState("Bribe")
	s.BTC = 5000
	_ = s.HireMerc("tabby_guard")
	m := s.Mercs[0]
	m.Loyalty = 30
	if err := s.BribeMerc(m.InstanceID); err != nil {
		t.Fatalf("bribe: %v", err)
	}
	if m.Loyalty != 45 {
		t.Errorf("expected loyalty 45 after +15 bribe, got %d", m.Loyalty)
	}
}

func TestFireMercLowersPeerLoyalty(t *testing.T) {
	withTempHome(t)
	s := NewState("Fire")
	s.BTC = 10000
	_ = s.HireMerc("tabby_guard")
	_ = s.HireMerc("siamese_it")
	peer := s.Mercs[1]
	peer.Loyalty = 60
	victim := s.Mercs[0]
	if err := s.FireMerc(victim.InstanceID); err != nil {
		t.Fatalf("fire: %v", err)
	}
	if peer.Loyalty != 55 {
		t.Errorf("peer loyalty should drop 5 after a firing, got %d", peer.Loyalty)
	}
}

func TestUpgradeDefenseIncrementsLevel(t *testing.T) {
	withTempHome(t)
	s := NewState("Defense")
	s.BTC = 5000
	if err := s.UpgradeDefense("lock"); err != nil {
		t.Fatalf("upgrade lock: %v", err)
	}
	if s.Rooms["alley"].LockLvl != 1 {
		t.Errorf("expected lock lvl 1, got %d", s.Rooms["alley"].LockLvl)
	}
}

func TestUpgradeDefenseRejectsBadDim(t *testing.T) {
	withTempHome(t)
	s := NewState("DefenseBad")
	s.BTC = 99999
	if err := s.UpgradeDefense("nonsense"); err == nil {
		t.Error("should reject unknown dim")
	}
}

// --- Sprint 7 §9 skill nodes ---------------------------------------------

// TestWiringOptimizationApplies: unlocking the engineer T1 wiring node
// bumps the PSU overload tolerance by +0.10 and clamps PSU heat output at
// 0.80 of catalog. RoomPSUHeat must drop the full 20% on a non-builtin PSU.
func TestWiringOptimizationApplies(t *testing.T) {
	withTempHome(t)
	s := NewState("Wiring")
	s.TechPoint = 99
	s.BTC = 10_000

	if got := s.PSUOverloadToleranceBonus(); got != 0 {
		t.Errorf("baseline tolerance bonus = %v, want 0", got)
	}
	if got := s.PSUHeatMult(); got != 1.0 {
		t.Errorf("baseline heat mult = %v, want 1.0", got)
	}

	if err := s.UnlockSkill("wiring_optimization"); err != nil {
		t.Fatalf("unlock wiring_optimization: %v", err)
	}
	if got := s.PSUOverloadToleranceBonus(); math.Abs(got-0.10) > 1e-9 {
		t.Errorf("PSUOverloadToleranceBonus = %v, want 0.10", got)
	}
	if got := s.PSUHeatMult(); math.Abs(got-0.80) > 1e-9 {
		t.Errorf("PSUHeatMult = %v, want 0.80", got)
	}

	// Install psu_silver650 (heat_output=2) into alley alongside the
	// builtin (heat_output=0). Total catalog heat = 2; with the multiplier
	// active the room should report 1.6.
	if err := s.InstallPSU("alley", "psu_silver650"); err != nil {
		t.Fatalf("install psu_silver650: %v", err)
	}
	got := s.RoomPSUHeat("alley")
	want := 2.0 * 0.80
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("RoomPSUHeat with wiring opt = %v, want %v", got, want)
	}
}

// TestPoolHoppingShortensWindowAndKeepsShares: the mogul T2 pool-hop
// node drops the SwitchPool transition window from 600s → 180s, and a
// SwitchPool out of a PPLNS pool now keeps half the share accumulator
// instead of zeroing it.
func TestPoolHoppingShortensWindowAndKeepsShares(t *testing.T) {
	withTempHome(t)
	s := NewState("PoolHop")
	s.TechPoint = 99

	// Walk the prereq chain.
	if err := s.UnlockSkill("pool_hopping"); err == nil {
		t.Error("pool_hopping should require smart_invoicing")
	}
	if err := s.UnlockSkill("smart_invoicing"); err != nil {
		t.Fatalf("unlock smart_invoicing: %v", err)
	}
	if err := s.UnlockSkill("pool_hopping"); err != nil {
		t.Fatalf("unlock pool_hopping: %v", err)
	}

	if got := s.PoolSwitchDurationSec(); got != 180 {
		t.Errorf("PoolSwitchDurationSec = %d, want 180", got)
	}
	if got := s.PoolHoppingShareRetention(); got != 0.5 {
		t.Errorf("PoolHoppingShareRetention = %v, want 0.5", got)
	}

	// Pre-seed shares on the PPLNS scratch_pool, switch out, confirm the
	// 50% retention applied (no rounding loss on round numbers).
	s.PoolID = "scratch_pool"
	s.PoolShares = 1234
	const now int64 = 1_700_000_000
	if err := s.SwitchPool("kitten_hash", now); err != nil {
		t.Fatalf("SwitchPool: %v", err)
	}
	if math.Abs(s.PoolShares-617) > 1e-9 {
		t.Errorf("PoolShares after pool-hop leave = %v, want 617 (50%% of 1234)", s.PoolShares)
	}
	if s.PoolSwitchAt != now+180 {
		t.Errorf("PoolSwitchAt = %d, want %d", s.PoolSwitchAt, now+180)
	}
}

// TestAssetHedgingFlattensResale: unlocking the mogul T3 hedging node
// trims a GPU's effective BtcSensitivity by 0.20, and the floor clamp
// keeps a low-sensitivity card at 0 instead of going negative (which
// would invert the bull/bear relationship).
func TestAssetHedgingFlattensResale(t *testing.T) {
	withTempHome(t)
	s := NewState("Hedge")
	s.TechPoint = 99
	// Walk prereq chain.
	if err := s.UnlockSkill("smart_invoicing"); err != nil {
		t.Fatalf("smart_invoicing: %v", err)
	}
	if err := s.UnlockSkill("hedged_wallet"); err != nil {
		t.Fatalf("hedged_wallet: %v", err)
	}

	// Use rtx3080 — catalog sensitivity 0.8. At MarketPrice=1.5 the
	// 20% reduction translates to a measurable drop in resale.
	g := &GPU{InstanceID: 999, DefID: "rtx3080", Status: "running"}
	def, _ := data.GPUByID("rtx3080")
	s.MarketPrice = 1.5
	before := s.GPUResalePrice(g)

	if err := s.UnlockSkill("asset_hedging"); err != nil {
		t.Fatalf("unlock asset_hedging: %v", err)
	}
	if got := s.BtcSensitivityBonus(); math.Abs(got-0.20) > 1e-9 {
		t.Errorf("BtcSensitivityBonus = %v, want 0.20", got)
	}
	after := s.GPUResalePrice(g)

	// The exact analytical drop: 0.20 (the bonus) × (MarketPrice-1) ×
	// price × baseResaleRatio. before > after on a bull market; the
	// difference must equal that quantity.
	expectedDrop := 0.20 * (s.MarketPrice - 1.0) * float64(def.Price) * def.BaseResaleRatio
	if math.Abs((before-after)-expectedDrop) > 1e-6 {
		t.Errorf("hedging drop = %v, want %v (before=%v after=%v)",
			before-after, expectedDrop, before, after)
	}

	// Sensitivity floor: gtx1060 catalog sensitivity is 0.3. A 0.20
	// reduction lands at 0.10 — still positive, so resale tracks BTC
	// faintly. Push the player into multi-stack territory by injecting
	// a fake low-sensitivity catalog probe via meowcore_v1 instead.
	bp := &Blueprint{ID: "bp_hedge", Tier: 1, Boosts: nil}
	s.Blueprints = append(s.Blueprints, bp)
	core := &GPU{InstanceID: 1000, DefID: "meowcore_v1", Status: "running", BlueprintID: "bp_hedge"}
	// Tier 1 inherent sens = 0.2 → after −0.20 it lands at 0; resale
	// should sit flat at the inherent base regardless of MarketPrice.
	s.MarketPrice = 1.5
	bull := s.GPUResalePrice(core)
	s.MarketPrice = 0.5
	bear := s.GPUResalePrice(core)
	if math.Abs(bull-bear) > 1e-9 {
		t.Errorf("clamped sensitivity should make bull/bear equal: bull=%v bear=%v", bull, bear)
	}
	if math.Abs(bull-2000.0) > 1e-9 {
		t.Errorf("clamped MEOWCore tier-1 resale = %v, want 2000 (flat at inherent base)", bull)
	}
}

// TestNetworkOptimizationCutsStaleRate: hacker T1 net-opt subtracts 3pp
// from EffectiveStaleRate on every room. The clampStale floor at 0
// guarantees a tiny baseline can't go negative.
func TestNetworkOptimizationCutsStaleRate(t *testing.T) {
	withTempHome(t)
	s := NewState("NetOpt")
	s.TechPoint = 99

	rs := s.Rooms["alley"]
	rs.StaleRate = 0.05

	before := s.EffectiveStaleRate("alley")
	if err := s.UnlockSkill("network_optimization"); err != nil {
		t.Fatalf("unlock network_optimization: %v", err)
	}
	if got := s.StaleRateBonus(); math.Abs(got-0.03) > 1e-9 {
		t.Errorf("StaleRateBonus = %v, want 0.03", got)
	}
	after := s.EffectiveStaleRate("alley")
	if math.Abs((before-after)-0.03) > 1e-9 {
		t.Errorf("EffectiveStaleRate drop = %v, want 0.03 (before=%v after=%v)",
			before-after, before, after)
	}

	// Floor: a baseline below the bonus must clamp at 0.
	rs.StaleRate = 0.01
	if got := s.EffectiveStaleRate("alley"); got != 0 {
		t.Errorf("EffectiveStaleRate with baseline 0.01 < bonus 0.03 = %v, want 0", got)
	}
}

// TestPoolInfiltrationEarnsAndKarma: hacker T3 pool-infiltrate folds a
// 1.02 multiplier into mining earn and burns 5 Karma at unlock (one-shot,
// not per-tick).
func TestPoolInfiltrationEarnsAndKarma(t *testing.T) {
	withTempHome(t)
	s := NewState("Infil")
	s.TechPoint = 99

	if err := s.UnlockSkill("pool_infiltration"); err == nil {
		t.Error("pool_infiltration should require pump_dump")
	}
	if err := s.UnlockSkill("pump_dump"); err != nil {
		t.Fatalf("unlock pump_dump: %v", err)
	}

	karmaBefore := s.Karma
	if err := s.UnlockSkill("pool_infiltration"); err != nil {
		t.Fatalf("unlock pool_infiltration: %v", err)
	}
	if got := s.PoolInfiltrationEarnMult(); math.Abs(got-1.02) > 1e-9 {
		t.Errorf("PoolInfiltrationEarnMult = %v, want 1.02", got)
	}
	if delta := s.Karma - karmaBefore; delta != -5 {
		t.Errorf("Karma delta on unlock = %d, want -5", delta)
	}

	// Karma penalty fires once: bumping the player back to a fresh state
	// without re-unlocking should leave Karma untouched on subsequent
	// HasSkill checks (the multiplier still reads, no extra hit).
	karmaAfter := s.Karma
	_ = s.PoolInfiltrationEarnMult()
	if s.Karma != karmaAfter {
		t.Errorf("Karma changed outside of UnlockSkill: %d -> %d", karmaAfter, s.Karma)
	}
}
