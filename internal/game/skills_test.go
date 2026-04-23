package game

import (
	"testing"
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
	s.Money = 50
	if err := s.HireMerc("tabby_guard"); err == nil {
		t.Error("should refuse hire without enough money")
	}
	s.Money = 2000
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
	s.Money = 5000
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
	s.Money = 10000
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
	s.Money = 5000
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
	s.Money = 99999
	if err := s.UpgradeDefense("nonsense"); err == nil {
		t.Error("should reject unknown dim")
	}
}
