package game

import (
	"fmt"
	"math"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

// PrestigeThreshold is the lifetime-earned cash needed before Retire unlocks.
// Tuned for the recalibrated economy: a dedicated session with mid-tier gear
// hits this in roughly 4-8 hours, which is the typical first-prestige window
// for the genre.
const PrestigeThreshold = 250_000.0

// PrestigeTPCarryFraction is the slice of unspent TP banked into the
// fresh run on Retire (floor-rounded). 0.25 keeps prestige worth doing
// without making it a TP printer.
const PrestigeTPCarryFraction = 0.25

// PrestigeTPCarryCap clamps the carryover so a player who hoards TP for a
// dozen runs can't enter the next one with a four-digit head start.
const PrestigeTPCarryCap = 200

// LegacyPerk defines a purchasable prestige bonus.
type LegacyPerk struct {
	ID          string
	Name        string
	Desc        string
	Cost        int
	ApplyOnBuy  func(*LegacyStore) // mutates legacy immediately
	Available   func(*LegacyStore) bool
}

var legacyPerks = []LegacyPerk{
	{
		ID: "starter_cash_500", Name: "Seed Capital", Cost: 10,
		Desc: "Start new runs with an extra ₿500.",
		ApplyOnBuy: func(l *LegacyStore) { l.StarterCash += 500 },
		Available:  func(l *LegacyStore) bool { return l.StarterCash < 5000 },
	},
	{
		ID: "unlock_university", Name: "Alumni Privileges", Cost: 50,
		Desc: "New runs begin with the University Server Room pre-unlocked.",
		ApplyOnBuy: func(l *LegacyStore) { l.UnlockedUniversity = true },
		Available:  func(l *LegacyStore) bool { return !l.UnlockedUniversity },
	},
	{
		ID: "efficiency_5pct", Name: "Muscle Memory", Cost: 200,
		Desc: "Permanent +5% GPU efficiency across all runs.",
		ApplyOnBuy: func(l *LegacyStore) { l.EfficiencyBoost += 0.05 },
		Available:  func(l *LegacyStore) bool { return l.EfficiencyBoost < 0.50 },
	},
}

func LegacyPerks() []LegacyPerk { return legacyPerks }

// CanRetire returns whether the player can Retire right now.
func (s *State) CanRetire() bool {
	return s.HasUnlock("prestige") && s.LifetimeEarned >= PrestigeThreshold
}

// RetireReward returns the LP the player would earn if they retired now.
func (s *State) RetireReward() int {
	if s.LifetimeEarned < 1 {
		return 0
	}
	return int(math.Floor(math.Sqrt(s.LifetimeEarned / 10000.0)))
}

// Retire ends the current run, banks LP + blueprints, returns a fresh State.
// Returns the NEW state (to swap in) and the LP awarded.
func (s *State) Retire() (*State, int, error) {
	if !s.CanRetire() {
		return nil, 0, fmt.Errorf("need %s lifetime earnings; have %s", FmtBTC(PrestigeThreshold), FmtBTC(s.LifetimeEarned))
	}
	lp := s.RetireReward()
	legacy := LoadLegacy()
	legacy.TotalEarned += s.LifetimeEarned
	legacy.TotalLP += lp
	s.grantAchievement("first_retire")
	// Carry blueprints.
	for _, bp := range s.Blueprints {
		legacy.Blueprints = append(legacy.Blueprints, bp)
	}
	carry := int(math.Floor(float64(s.TechPoint) * PrestigeTPCarryFraction))
	if carry > PrestigeTPCarryCap {
		carry = PrestigeTPCarryCap
	}
	if carry < 0 {
		carry = 0
	}
	legacy.CarriedTP = carry
	_ = legacy.Save()

	fresh := newStateWithLegacy(s.KittenName, legacy)
	fresh.appendLog("opportunity", i18n.T("log.prestige.retired", lp))
	return fresh, lp, nil
}

// BuyLegacyPerk spends LP on a perk.
func BuyLegacyPerk(perkID string) error {
	legacy := LoadLegacy()
	var chosen *LegacyPerk
	for i := range legacyPerks {
		if legacyPerks[i].ID == perkID {
			chosen = &legacyPerks[i]
			break
		}
	}
	if chosen == nil {
		return fmt.Errorf("no such perk")
	}
	if !chosen.Available(legacy) {
		return fmt.Errorf("not available (maxed or already owned)")
	}
	if legacy.LPAvailable() < chosen.Cost {
		return fmt.Errorf("need %d LP, have %d", chosen.Cost, legacy.LPAvailable())
	}
	legacy.SpentLP += chosen.Cost
	chosen.ApplyOnBuy(legacy)
	return legacy.Save()
}
