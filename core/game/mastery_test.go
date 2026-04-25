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
	got, err := s.ConvertFragsToBTC(10)
	if err != nil {
		t.Fatalf("ConvertFragsToBTC: %v", err)
	}
	if got <= 0 {
		t.Errorf("expected positive BTC, got %.4f", got)
	}
	if s.BTC <= beforeBTC {
		t.Error("BTC should have increased")
	}
	if s.ResearchFrags != 40 {
		t.Errorf("expected 40 frags after spend, got %d", s.ResearchFrags)
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
