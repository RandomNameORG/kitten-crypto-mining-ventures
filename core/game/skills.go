package game

import (
	"fmt"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

// UnlockSkill purchases a skill if the player can afford it and the prereq is met.
func (s *State) UnlockSkill(id string) error {
	def, ok := data.SkillByID(id)
	if !ok {
		return fmt.Errorf("no such skill")
	}
	if s.UnlockedSkills[id] {
		return fmt.Errorf("already unlocked")
	}
	if def.Prereq != "" && !s.UnlockedSkills[def.Prereq] {
		return fmt.Errorf("requires prerequisite skill")
	}
	if s.TechPoint < def.Cost {
		return fmt.Errorf("need %d TP, have %d", def.Cost, s.TechPoint)
	}
	s.TechPoint -= def.Cost
	if s.UnlockedSkills == nil {
		s.UnlockedSkills = map[string]bool{}
	}
	s.UnlockedSkills[id] = true
	s.appendLog("opportunity", i18n.T("log.skill.learned", def.LocalName()))
	return nil
}

// HasSkill reports whether the given skill ID is unlocked.
func (s *State) HasSkill(id string) bool { return s.UnlockedSkills[id] }

// HasUnlock reports whether any unlocked skill grants the given feature.
func (s *State) HasUnlock(feature string) bool {
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok {
			if def.Effect.Kind == "unlock" && def.Effect.Unlocks == feature {
				return true
			}
		}
	}
	return false
}

// PowerDrawMult is the product of all power-reduction skills.
func (s *State) PowerDrawMult() float64 {
	mult := 1.0
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "power_mult" {
			mult *= def.Effect.Value
		}
	}
	return mult
}

// BillMult is the product of all bill-reduction skills.
func (s *State) BillMult() float64 {
	mult := 1.0
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "bill_mult" {
			mult *= def.Effect.Value
		}
	}
	return mult
}

// ScrapValueMult boosts scrap/sell value.
func (s *State) ScrapValueMult() float64 {
	mult := 1.0
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "scrap_mult" {
			mult *= def.Effect.Value
		}
	}
	// Legacy: efficiency boost doesn't apply here; this is scrap only.
	return mult
}

// EfficiencyMult combines overclock skill + legacy bonus.
func (s *State) EfficiencyMult() float64 {
	mult := 1.0
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "overclock" {
			mult *= 1.0 + def.Effect.Value
		}
	}
	// LegacyStore.EfficiencyBoost is read at NewState but not stored on State,
	// so we derive from the saved LegacyAvailable indirectly. Simpler: read
	// legacy fresh each call.
	if legacy := LoadLegacy(); legacy != nil {
		mult *= 1.0 + legacy.EfficiencyBoost
	}
	return mult
}

// HeatMult boosts heat output (overclock raises heat).
func (s *State) HeatMult() float64 {
	mult := 1.0
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "overclock" {
			mult *= 1.0 + 0.15 // fixed per the design
		}
	}
	return mult
}

// RepairFree reports whether repairs cost nothing.
func (s *State) RepairFree() bool {
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "repair_free" {
			return true
		}
	}
	return false
}

// MercLoyaltyFloor is a flat loyalty bonus from skills.
func (s *State) MercLoyaltyFloor() int {
	floor := 0
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "merc_loyalty" {
			floor += int(def.Effect.Value)
		}
	}
	return floor
}

// EarnVolatilityDamp returns a 0..1 factor that dampens earn-multiplier
// swings (1.0 = no damp). Hedged Wallet cuts event-driven volatility in half.
func (s *State) EarnVolatilityDamp() float64 {
	damp := 1.0
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "earn_damp" {
			damp *= def.Effect.Value
		}
	}
	return damp
}
