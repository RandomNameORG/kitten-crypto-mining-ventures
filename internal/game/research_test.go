package game

import (
	"testing"
	"time"
)

func TestStartResearchLockedByDefault(t *testing.T) {
	withTempHome(t)
	s := NewState("NoRD")
	s.Money = 99999
	s.ResearchFrags = 999
	if err := s.StartResearch(1, []string{"efficiency", "undervolt"}); err == nil {
		t.Error("should be gated behind rd unlock")
	}
}

func TestStartResearchRejectsDuplicateBoosts(t *testing.T) {
	withTempHome(t)
	s := unlockRDFixture(t)
	if err := s.StartResearch(1, []string{"efficiency"}); err == nil {
		t.Error("should reject fewer than 2 boosts")
	}
	if err := s.StartResearch(1, []string{"bad", "efficiency"}); err == nil {
		t.Error("should reject unknown boost")
	}
}

func TestResearchCompletionCreatesBlueprint(t *testing.T) {
	withTempHome(t)
	s := unlockRDFixture(t)
	if err := s.StartResearch(1, []string{"efficiency", "undervolt"}); err != nil {
		t.Fatalf("start: %v", err)
	}
	// Force completion by rewinding StartedAt.
	s.ActiveResearch.StartedAt = time.Now().Unix() - int64(s.ActiveResearch.DurationSec) - 5
	s.advanceResearch(time.Now().Unix())
	if s.ActiveResearch != nil {
		t.Error("research should have completed")
	}
	if len(s.Blueprints) != 1 {
		t.Fatalf("expected 1 blueprint, got %d", len(s.Blueprints))
	}
	bp := s.Blueprints[0]
	if bp.Tier != 1 {
		t.Errorf("unexpected tier %d", bp.Tier)
	}
	// Boosts always contain both picked axes (plus up to 1 bonus).
	has := map[string]bool{}
	for _, b := range bp.Boosts {
		has[b] = true
	}
	if !has["efficiency"] || !has["undervolt"] {
		t.Errorf("blueprint missing requested boosts: %v", bp.Boosts)
	}
}

func TestPrintMEOWCoreConsumesResources(t *testing.T) {
	withTempHome(t)
	s := unlockRDFixture(t)
	// Hand-craft a blueprint.
	s.Blueprints = append(s.Blueprints, &Blueprint{
		ID: "bp_test", Tier: 1, Boosts: []string{"efficiency", "undervolt"},
	})
	// Give resources; ensure alley has room (start has 1 GPU, 4 slots).
	s.Money = 10000
	s.ResearchFrags = 100
	before := len(s.GPUs)
	if err := s.PrintMEOWCore("bp_test"); err != nil {
		t.Fatalf("print: %v", err)
	}
	if len(s.GPUs) != before+1 {
		t.Errorf("expected +1 GPU after print, got %d → %d", before, len(s.GPUs))
	}
	// The new GPU should have a BlueprintID set.
	last := s.GPUs[len(s.GPUs)-1]
	if last.BlueprintID != "bp_test" {
		t.Errorf("printed GPU missing BlueprintID: %+v", last)
	}
}

func TestBlueprintStatsChangeWithBoosts(t *testing.T) {
	base := &Blueprint{Tier: 1, Boosts: []string{"efficiency", "undervolt"}}
	plain := &Blueprint{Tier: 1}
	eBase, pBase, _, _ := BlueprintStats(plain)
	eBoosted, pBoosted, _, _ := BlueprintStats(base)
	if eBoosted <= eBase {
		t.Error("efficiency boost should raise efficiency")
	}
	if pBoosted >= pBase {
		t.Error("undervolt boost should lower power draw")
	}
}

func TestCanRetireRequiresThreshold(t *testing.T) {
	withTempHome(t)
	s := NewState("Retire")
	s.TechPoint = 99
	// Walk full Mogul prereq chain to unlock prestige.
	_ = s.UnlockSkill("smart_invoicing")
	_ = s.UnlockSkill("hedged_wallet")
	_ = s.UnlockSkill("venture_cap")
	if s.CanRetire() {
		t.Error("should not be able to retire without lifetime earnings")
	}
	s.LifetimeEarned = PrestigeThreshold + 1
	if !s.CanRetire() {
		t.Error("should be able to retire past threshold")
	}
}

func TestRetireProducesFreshStateAndLP(t *testing.T) {
	withTempHome(t)
	s := NewState("Cycle")
	s.TechPoint = 99
	_ = s.UnlockSkill("smart_invoicing")
	_ = s.UnlockSkill("hedged_wallet")
	_ = s.UnlockSkill("venture_cap")
	s.LifetimeEarned = 4_000_000
	s.Money = 5000
	fresh, lp, err := s.Retire()
	if err != nil {
		t.Fatalf("retire: %v", err)
	}
	if lp <= 0 {
		t.Error("expected positive LP")
	}
	if fresh.LifetimeEarned != 0 {
		t.Error("fresh state should reset lifetime earnings")
	}
	if len(fresh.UnlockedSkills) != 0 {
		t.Error("fresh state should reset skills")
	}
	// Legacy should be persisted and reflect the LP.
	legacy := LoadLegacy()
	if legacy.TotalLP < lp {
		t.Errorf("legacy TotalLP (%d) should include at least the retire LP (%d)", legacy.TotalLP, lp)
	}
}

// unlockRDFixture returns a state with RD + enough TP/money/frags to start.
func unlockRDFixture(t *testing.T) *State {
	t.Helper()
	withTempHome(t)
	s := NewState("RD")
	s.TechPoint = 99
	_ = s.UnlockSkill("undervolt_i")
	_ = s.UnlockSkill("undervolt_ii")
	_ = s.UnlockSkill("rd_unlock")
	s.Money = 99999
	s.ResearchFrags = 999
	return s
}
