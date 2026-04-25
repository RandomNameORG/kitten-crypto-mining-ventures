package game

import (
	"testing"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/data"
)

// marketCrashRealEvent builds an event whose ID matches the catalog entry so
// applyEvent's crash-counter branch fires. The shared shim in testshim_test.go
// hardcodes ID="shim", which is deliberate there but not usable here.
func marketCrashRealEvent() data.EventDef {
	return data.EventDef{
		ID:       "market_crash",
		Name:     "Crash",
		Category: "crisis",
		Emoji:    "📉",
		Text:     "(test)",
		Weight:   1,
		Effects: []data.EventEffect{
			{Kind: "market_pin", Factor: 0.3, Seconds: 60},
		},
	}
}

func TestAchievement_MarketTiming(t *testing.T) {
	withTempHome(t)
	s := NewState("T")
	s.BTC = 10_000

	s.MarketPrice = 1.0
	if err := s.BuyGPU("gtx1060"); err != nil {
		t.Fatalf("BuyGPU: %v", err)
	}
	if s.HasAchievement("market_timing") {
		t.Fatal("market_timing should not be granted at neutral MarketPrice")
	}

	s.MarketPrice = 0.5
	if err := s.BuyGPU("gtx1060"); err != nil {
		t.Fatalf("BuyGPU: %v", err)
	}
	if !s.HasAchievement("market_timing") {
		t.Fatal("market_timing should be granted buying at MarketPrice < 0.7")
	}
}

func TestAchievement_OCMastery(t *testing.T) {
	withTempHome(t)
	s := NewState("T")

	s.OCTimeT1Sec = 1800
	s.OCTimeT2Sec = 1799
	s.CheckAchievements()
	if s.HasAchievement("oc_mastery") {
		t.Fatal("oc_mastery should not be granted below 3600 OC-seconds")
	}

	s.OCTimeT2Sec = 1800
	s.CheckAchievements()
	if !s.HasAchievement("oc_mastery") {
		t.Fatal("oc_mastery should be granted at 3600 cumulative OC-seconds")
	}
}

func TestAchievement_TaxSurvivor(t *testing.T) {
	withTempHome(t)
	s := NewState("T")

	// Hit branch: BTC well below reserve threshold.
	s.BTC = 100_000
	s.LifetimeEarned = 5_000_000
	s.applyEvent(taxAuditEvent(0.20, 0.10))
	if s.HasAchievement("tax_survivor") {
		t.Fatal("tax_survivor should not be granted on the audit-hit branch")
	}

	// Clean branch: reserves cover the threshold.
	s.BTC = 1_000_000
	s.applyEvent(taxAuditEvent(0.20, 0.10))
	if !s.HasAchievement("tax_survivor") {
		t.Fatal("tax_survivor should be granted when reserves cover the audit")
	}
}

func TestAchievement_Overdrive(t *testing.T) {
	withTempHome(t)
	s := NewState("T")

	// Starter GPU is OCLevel 0 — universal-max quantifier is false.
	s.CheckAchievements()
	if s.HasAchievement("overdrive") {
		t.Fatal("overdrive should not be granted while any GPU sits below OC level 2")
	}

	for _, g := range s.GPUs {
		g.OCLevel = 2
	}
	s.CheckAchievements()
	if !s.HasAchievement("overdrive") {
		t.Fatal("overdrive should be granted when every installed GPU is at OC level 2")
	}
}

func TestAchievement_PeakSell(t *testing.T) {
	withTempHome(t)
	s := NewState("T")
	// Give ourselves a second GPU so we can sell twice — once at neutral,
	// once at peak.
	s.addGPU("gtx1060", s.CurrentRoom, false)
	if len(s.GPUs) < 2 {
		t.Fatalf("expected 2 GPUs, got %d", len(s.GPUs))
	}
	first := s.GPUs[0]

	s.MarketPrice = 1.0
	if err := s.SellGPU(first.InstanceID); err != nil {
		t.Fatalf("SellGPU: %v", err)
	}
	if s.HasAchievement("peak_sell") {
		t.Fatal("peak_sell should not be granted at neutral MarketPrice")
	}

	second := s.GPUs[0]
	s.MarketPrice = 2.0
	if err := s.SellGPU(second.InstanceID); err != nil {
		t.Fatalf("SellGPU: %v", err)
	}
	if !s.HasAchievement("peak_sell") {
		t.Fatal("peak_sell should be granted selling at MarketPrice > 1.5")
	}
}

func TestAchievement_EventVeteran(t *testing.T) {
	withTempHome(t)
	s := NewState("T")

	s.EventsByCategory = map[string]int{}
	s.EventsByCategory["info"] = 10
	s.EventsByCategory["threat"] = 20
	s.EventsByCategory["opportunity"] = 19
	s.CheckAchievements()
	if s.HasAchievement("event_veteran") {
		t.Fatal("event_veteran should not be granted below 50 total events")
	}

	s.EventsByCategory["opportunity"] = 20
	s.CheckAchievements()
	if !s.HasAchievement("event_veteran") {
		t.Fatal("event_veteran should be granted when EventsByCategory sums to 50")
	}
}

func TestAchievement_Marathon(t *testing.T) {
	withTempHome(t)
	s := NewState("T")

	s.TotalTicks = 99_999
	s.CheckAchievements()
	if s.HasAchievement("marathon") {
		t.Fatal("marathon should not be granted below 100,000 ticks")
	}

	s.TotalTicks = 100_000
	s.CheckAchievements()
	if !s.HasAchievement("marathon") {
		t.Fatal("marathon should be granted at 100,000 ticks")
	}
}

func TestAchievement_CrisisManager(t *testing.T) {
	withTempHome(t)
	s := NewState("T")

	s.applyEvent(marketCrashRealEvent())
	s.applyEvent(marketCrashRealEvent())
	s.CheckAchievements()
	if s.HasAchievement("crisis_manager") {
		t.Fatalf("crisis_manager should not be granted after 2 crashes (count=%d)", s.MarketCrashCount)
	}

	s.applyEvent(marketCrashRealEvent())
	s.CheckAchievements()
	if !s.HasAchievement("crisis_manager") {
		t.Fatalf("crisis_manager should be granted after 3 crashes (count=%d)", s.MarketCrashCount)
	}
}

// TestAchievementTPReward — granting an achievement with a non-zero
// TPReward credits exactly that many TP and is idempotent across repeated
// grant calls. Targets the new TP-faucet path in grantAchievement.
func TestAchievementTPReward(t *testing.T) {
	withTempHome(t)
	s := NewState("TPReward")
	def, ok := data.AchievementByID("first_million")
	if !ok {
		t.Fatal("first_million not in catalog")
	}
	if def.TPReward != 5 {
		t.Fatalf("expected first_million TPReward=5, got %d (sprint-2 backfill regressed)", def.TPReward)
	}

	tpBefore := s.TechPoint
	s.grantAchievement("first_million")
	gotGain := s.TechPoint - tpBefore
	if gotGain != def.TPReward {
		t.Errorf("TP gain on grant = %d, want %d", gotGain, def.TPReward)
	}
	// Idempotency: second grant of an already-owned achievement must not
	// re-credit TP. This is the load-bearing invariant — without it, every
	// CheckAchievements() call would mint TP forever.
	tpAfterFirst := s.TechPoint
	s.grantAchievement("first_million")
	if s.TechPoint != tpAfterFirst {
		t.Errorf("re-granting an owned achievement double-credited TP: %d → %d",
			tpAfterFirst, s.TechPoint)
	}
}

// TestLifetimeMilestonePaysOnce — every tier in lifetimeMilestones fires
// exactly once across multiple CheckAchievements calls, and the high-water
// counter advances tier-by-tier as LifetimeEarned crosses thresholds. The
// test pre-grants the LE-bound achievements (first_drop / first_ten_k /
// first_million) so their TPRewards don't pollute the milestone delta we
// want to measure.
func TestLifetimeMilestonePaysOnce(t *testing.T) {
	withTempHome(t)
	s := NewState("Milestones")
	// Stash achievements that would otherwise fire on these LE thresholds
	// and credit TP via grantAchievement. We're isolating the milestone
	// faucet here — the achievement faucet has its own dedicated test.
	s.Achievements = append(s.Achievements,
		"first_drop", "first_ten_k", "first_million")

	s.TechPoint = 0
	s.LifetimeEarned = 0
	s.LifetimeMilestonesPaid = 0

	// Cross tier 1 (10K). Expect +5 TP exactly, regardless of how many
	// times CheckAchievements is called.
	s.LifetimeEarned = 10_001
	tpBefore := s.TechPoint
	s.CheckAchievements()
	s.CheckAchievements()
	s.CheckAchievements()
	if s.TechPoint-tpBefore != 5 {
		t.Errorf("after crossing tier 1, TP delta=%d want 5", s.TechPoint-tpBefore)
	}
	if s.LifetimeMilestonesPaid != 1 {
		t.Errorf("LifetimeMilestonesPaid=%d want 1", s.LifetimeMilestonesPaid)
	}

	// Cross tier 2 (100K). +15 more TP; cumulative=20.
	s.LifetimeEarned = 100_001
	tpBefore = s.TechPoint
	s.CheckAchievements()
	s.CheckAchievements()
	if s.TechPoint-tpBefore != 15 {
		t.Errorf("after crossing tier 2, TP delta=%d want 15", s.TechPoint-tpBefore)
	}
	if s.LifetimeMilestonesPaid != 2 {
		t.Errorf("LifetimeMilestonesPaid=%d want 2", s.LifetimeMilestonesPaid)
	}

	// Big jump skipping a tier — crossing 10M should pay tiers 3+4
	// (30+60=90 more) in a single call, advancing the counter twice.
	s.LifetimeEarned = 10_000_001
	tpBefore = s.TechPoint
	s.CheckAchievements()
	if s.TechPoint-tpBefore != 30+60 {
		t.Errorf("after crossing tiers 3+4, TP delta=%d want %d",
			s.TechPoint-tpBefore, 30+60)
	}
	if s.LifetimeMilestonesPaid != 4 {
		t.Errorf("LifetimeMilestonesPaid=%d want 4", s.LifetimeMilestonesPaid)
	}
}
