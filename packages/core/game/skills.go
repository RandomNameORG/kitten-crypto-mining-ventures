package game

import (
	"fmt"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"
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
	// Pool Infiltration carries a one-shot Karma hit at unlock — the spec
	// frames it as the moral cost of compromising another pool's worker
	// pipeline. Fires once here, not per-tick, so the player pays for the
	// decision rather than the ongoing 2% earn boost.
	if def.Effect.Kind == "pool_infiltrate" {
		s.Karma -= 5
		s.appendLog("threat", "Pool infiltration online — Karma −5")
	}
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

// RepairCostMult is the multiplier applied to a repair's base cost. PCB
// Surgery now stacks as a 50% discount instead of zeroing the cost.
// Multiple discount-kind skills (if added later) compound multiplicatively.
func (s *State) RepairCostMult() float64 {
	mult := 1.0
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "repair_discount" {
			v := def.Effect.Value
			if v <= 0 || v >= 1 {
				v = 0.5 // sane default for misconfigured catalogs
			}
			mult *= v
		}
	}
	return mult
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

// PSUOverloadToleranceBonus is the additive bonus applied to the weakest
// running PSU's overload_tolerance — Wiring Optimization (engineer T1)
// gives the room a wider safe band before the explosion roll fires.
// Sums over multiple "wiring_opt" skills so future stacks compose cleanly.
func (s *State) PSUOverloadToleranceBonus() float64 {
	bonus := 0.0
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "wiring_opt" {
			bonus += def.Effect.Value
		}
	}
	return bonus
}

// PSUHeatMult scales PSU heat output. Wiring Optimization shaves 20% off
// every running PSU's heat contribution; flat multiplier rather than a
// stack so the spec's "−20%" stays the ceiling.
func (s *State) PSUHeatMult() float64 {
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "wiring_opt" {
			return 0.80
		}
	}
	return 1.0
}

// PoolSwitchDurationSec is the transition window opened by SwitchPool.
// Pool Hopping (mogul T2) shortens it from PoolSwitchSec (600s) to the
// Effect.Value tucked in the catalog (180s), turning pool-shopping from
// a 10-minute commitment into a 3-minute one.
func (s *State) PoolSwitchDurationSec() int64 {
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "pool_hop" {
			return int64(def.Effect.Value)
		}
	}
	return PoolSwitchSec
}

// PoolHoppingShareRetention is the fraction of PPLNS shares preserved when
// the player walks away from a PPLNS pool. Default 0 (the spec'd "shares
// evaporate" rule); Pool Hopping bumps it to 0.5.
func (s *State) PoolHoppingShareRetention() float64 {
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "pool_hop" {
			return 0.5
		}
	}
	return 0.0
}

// BtcSensitivityBonus is the subtractive reduction applied to a GPU's
// BtcSensitivity when computing resale price. Asset Hedging (mogul T3)
// flattens the BTC swing in resale value; callers clamp the result so the
// effective sensitivity floors at 0 rather than going negative.
func (s *State) BtcSensitivityBonus() float64 {
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "asset_hedge" {
			return def.Effect.Value
		}
	}
	return 0.0
}

// StaleRateBonus is the additive reduction applied to every room's stale
// rate. Network Optimization (hacker T1) trims 3 percentage points; sums
// over multiple "stale_reduce" skills so future stacks compose. The
// EffectiveStaleRate caller subtracts this and lets clampStale floor at 0.
func (s *State) StaleRateBonus() float64 {
	bonus := 0.0
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "stale_reduce" {
			bonus += def.Effect.Value
		}
	}
	return bonus
}

// PoolInfiltrationEarnMult multiplies mining earn. Pool Infiltration
// (hacker T3) folds in a 1.02 factor — the upside that pays for the
// one-shot Karma hit applied at unlock.
func (s *State) PoolInfiltrationEarnMult() float64 {
	for id := range s.UnlockedSkills {
		if def, ok := data.SkillByID(id); ok && def.Effect.Kind == "pool_infiltrate" {
			return 1.0 + def.Effect.Value
		}
	}
	return 1.0
}
