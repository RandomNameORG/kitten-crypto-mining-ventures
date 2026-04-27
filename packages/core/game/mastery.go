package game

import (
	"fmt"
	"math"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"
)

// LevelUpMastery advances the named mastery track by one level. Returns the
// new level on success, or an error describing why the purchase failed.
func (s *State) LevelUpMastery(id string) (int, error) {
	track, ok := data.MasteryByID(id)
	if !ok {
		return 0, fmt.Errorf("no such mastery track")
	}
	if s.MasteryLevels == nil {
		s.MasteryLevels = map[string]int{}
	}
	cur := s.MasteryLevels[id]
	if cur >= track.MaxLevel {
		return cur, fmt.Errorf("mastery already at max")
	}
	cost := track.CostFor(cur)
	if cost < 0 {
		return cur, fmt.Errorf("mastery already at max")
	}
	if s.TechPoint < cost {
		return cur, fmt.Errorf("need %d TP, have %d", cost, s.TechPoint)
	}
	s.TechPoint -= cost
	s.MasteryLevels[id] = cur + 1
	s.appendLog("opportunity",
		i18n.T("log.mastery.leveled", track.LocalName(), cur+1))
	return cur + 1, nil
}

// MasteryLevel returns the current level of a track (0 if untouched).
func (s *State) MasteryLevel(id string) int {
	if s.MasteryLevels == nil {
		return 0
	}
	return s.MasteryLevels[id]
}

// masteryMult returns (1+per_level)^level for the given track effect — the
// multiplicative bonus that level N grants. Called by the per-effect
// helpers below.
func (s *State) masteryMult(effect string) float64 {
	mult := 1.0
	for _, t := range data.MasteryTracks() {
		if t.Effect != effect {
			continue
		}
		lvl := s.MasteryLevel(t.ID)
		if lvl <= 0 {
			continue
		}
		mult *= math.Pow(1.0+t.PerLevel, float64(lvl))
	}
	return mult
}

// MasteryEarnMult is the mining-mastery multiplier applied to BTC earn rate.
func (s *State) MasteryEarnMult() float64 { return s.masteryMult("mining") }

// MasteryBillMult is the power-engineering multiplier applied to bills.
// Note: PerLevel is negative for "power" so the math returns < 1.0 when
// levels are unlocked (the player's bills shrink).
func (s *State) MasteryBillMult() float64 { return s.masteryMult("power") }

// MasteryCoolingMult multiplies the room's passive cooling rate.
func (s *State) MasteryCoolingMult() float64 { return s.masteryMult("cooling") }

// MasteryFragMult scales fragments earned when scrapping a GPU.
func (s *State) MasteryFragMult() float64 { return s.masteryMult("frags") }

// MasteryScrapMult stacks with skill-based scrap multipliers on sell value.
func (s *State) MasteryScrapMult() float64 { return s.masteryMult("scrap") }

// ConvertFragsToBTC is the lab-side alchemy: trade research fragments for
// raw BTC at a lossy rate. Provides a sink when frags pile up faster than
// research can absorb them. Requires R&D to be unlocked.
//
//	frags is the count to spend; rate is the per-frag BTC payout.
//
// Returns the BTC credited or an error.
func (s *State) ConvertFragsToBTC(frags int) (float64, error) {
	if !s.HasUnlock("rd") {
		return 0, fmt.Errorf("requires R&D unlock")
	}
	if frags <= 0 {
		return 0, fmt.Errorf("frags must be positive")
	}
	if s.ResearchFrags < frags {
		return 0, fmt.Errorf("need %d frags, have %d", frags, s.ResearchFrags)
	}
	const ratePerFrag = 0.5 // 1 frag → ₿0.5; tune if frags overflow worse
	gained := float64(frags) * ratePerFrag
	s.ResearchFrags -= frags
	s.BTC += gained
	s.appendLog("info", i18n.T("log.alchemy.frags", frags, FmtBTC(gained)))
	return gained, nil
}
