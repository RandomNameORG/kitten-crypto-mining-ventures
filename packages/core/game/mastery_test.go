package game

import (
	"math"
	"testing"
)

// TestLevelUpMasteryDeductsTP verifies the basic TP-pay-and-advance loop.
func TestLevelUpMasteryDeductsTP(t *testing.T) {
	withTempHome(t)
	s := NewState("Mastery")
	s.TechPoint = 100

	beforeTP := s.TechPoint
	lvl, err := s.LevelUpMastery("mining")
	if err != nil {
		t.Fatalf("LevelUpMastery: %v", err)
	}
	if lvl != 1 {
		t.Errorf("expected level 1 after first up, got %d", lvl)
	}
	if s.TechPoint >= beforeTP {
		t.Error("TP should have been deducted")
	}
	if s.MasteryLevel("mining") != 1 {
		t.Errorf("MasteryLevel mismatch: %d", s.MasteryLevel("mining"))
	}
}

// TestMasteryMultStacks verifies multiplicative compounding.
func TestMasteryMultStacks(t *testing.T) {
	withTempHome(t)
	s := NewState("Stack")
	s.TechPoint = 10000
	for i := 0; i < 10; i++ {
		if _, err := s.LevelUpMastery("mining"); err != nil {
			t.Fatalf("level %d: %v", i+1, err)
		}
	}
	got := s.MasteryEarnMult()
	want := math.Pow(1.01, 10)
	if math.Abs(got-want) > 0.0001 {
		t.Errorf("expected %.4f, got %.4f", want, got)
	}
}

// TestPowerMasteryReducesBills verifies the negative-PerLevel track shrinks
// the multiplier (bills go down, not up).
func TestPowerMasteryReducesBills(t *testing.T) {
	withTempHome(t)
	s := NewState("Power")
	s.TechPoint = 5000
	for i := 0; i < 20; i++ {
		if _, err := s.LevelUpMastery("power"); err != nil {
			t.Fatalf("level %d: %v", i+1, err)
		}
	}
	got := s.MasteryBillMult()
	if got >= 1.0 {
		t.Errorf("expected discount < 1.0, got %.4f", got)
	}
}

// TestConvertFragsRequiresRD ensures the alchemy gate fires correctly.
func TestConvertFragsRequiresRD(t *testing.T) {
	withTempHome(t)
	s := NewState("Alchemy")
	s.ResearchFrags = 50

	if _, err := s.ConvertFragsToBTC(10); err == nil {
		t.Error("expected error without R&D unlocked")
	}

	// Walk the engineer chain to unlock rd.
	s.TechPoint = 100
	_ = s.UnlockSkill("undervolt_i")
	_ = s.UnlockSkill("undervolt_ii")
	_ = s.UnlockSkill("rd_unlock")
	if !s.HasUnlock("rd") {
		t.Fatal("rd unlock should be set after walking the prereq chain")
	}

	beforeBTC := s.BTC
	// Convert enough frags that the cashout gas (§11.2) leaves a positive
	// net — at the 0.5 BTC/frag rate, 10 frags grosses 5 BTC and the
	// flat-floor gas surcharge alone would clamp the net to zero. 50
	// frags grosses 25 BTC, well past the gas threshold.
	got, err := s.ConvertFragsToBTC(50)
	if err != nil {
		t.Fatalf("ConvertFragsToBTC: %v", err)
	}
	if got <= 0 {
		t.Errorf("expected positive BTC, got %.4f", got)
	}
	if s.BTC <= beforeBTC {
		t.Error("BTC should have increased")
	}
	if s.ResearchFrags != 0 {
		t.Errorf("expected 0 frags after spend, got %d", s.ResearchFrags)
	}
}

// TestConvertFragsLowCountClampsToZero: at a frag count where the gross
// (frags * 0.5 BTC) is smaller than the gas surcharge, the cashout must
// clamp to zero net BTC rather than driving the player negative. The
// frags are still consumed — that's the rate the player accepts when
// they elect to convert dust.
func TestConvertFragsLowCountClampsToZero(t *testing.T) {
	withTempHome(t)
	s := NewState("Dust")
	s.TechPoint = 100
	_ = s.UnlockSkill("undervolt_i")
	_ = s.UnlockSkill("undervolt_ii")
	_ = s.UnlockSkill("rd_unlock")
	if !s.HasUnlock("rd") {
		t.Fatal("rd unlock should be set after walking the prereq chain")
	}
	s.ResearchFrags = 5
	beforeBTC := s.BTC

	// 5 frags at rate 0.5 = 2.5 BTC gross. GasFlatFloor alone is 5.0,
	// so net would be deeply negative — must clamp to zero.
	got, err := s.ConvertFragsToBTC(5)
	if err != nil {
		t.Fatalf("ConvertFragsToBTC: %v", err)
	}
	if got != 0 {
		t.Errorf("expected net 0 on dust trade, got %.4f", got)
	}
	if got < 0 {
		t.Errorf("net BTC went negative: %.4f", got)
	}
	if s.BTC < beforeBTC {
		t.Errorf("BTC dropped on a clamp-to-zero conversion: before=%v after=%v", beforeBTC, s.BTC)
	}
	if s.ResearchFrags != 0 {
		t.Errorf("frags should still be consumed (5 → 0), got %d", s.ResearchFrags)
	}
}

// TestConvertFragsRejectsOverdraft prevents spending more frags than held.
func TestConvertFragsRejectsOverdraft(t *testing.T) {
	withTempHome(t)
	s := NewState("Bust")
	s.TechPoint = 100
	_ = s.UnlockSkill("undervolt_i")
	_ = s.UnlockSkill("undervolt_ii")
	_ = s.UnlockSkill("rd_unlock")
	s.ResearchFrags = 5
	if _, err := s.ConvertFragsToBTC(10); err == nil {
		t.Error("expected error when spending more frags than held")
	}
}
