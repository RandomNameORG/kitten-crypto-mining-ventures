package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"
)

// ResearchBoosts are the three axes players pick 2 of when starting a research.
var ResearchBoosts = []string{"efficiency", "undervolt", "durability"}

// ResearchTierInfo describes the gate and cost per MEOWCore tier.
type ResearchTierInfo struct {
	Tier      int
	Name      string
	Duration  int // seconds of real-time research
	Frags     int
	Money     int
	MinLvl    int // TP total required to access
}

var researchTiers = []ResearchTierInfo{
	{Tier: 1, Name: "MEOWCore v1", Duration: 900, Frags: 20, Money: 2000, MinLvl: 0},
	{Tier: 2, Name: "MEOWCore v2", Duration: 1800, Frags: 50, Money: 8000, MinLvl: 8},
	{Tier: 3, Name: "MEOWCore Purrfect", Duration: 3600, Frags: 120, Money: 25000, MinLvl: 16},
}

func ResearchTiers() []ResearchTierInfo { return researchTiers }

// StartResearch begins a research project. Boosts must be 2 strings from ResearchBoosts.
func (s *State) StartResearch(tier int, boosts []string) error {
	if !s.HasUnlock("rd") {
		return fmt.Errorf("R&D is locked — unlock MEOWCore Blueprint in Engineer skills")
	}
	if s.ActiveResearch != nil {
		return fmt.Errorf("already researching %s", s.ActiveResearch.Boosts)
	}
	if len(boosts) != 2 {
		return fmt.Errorf("must pick exactly 2 boosts")
	}
	for _, b := range boosts {
		if !validBoost(b) {
			return fmt.Errorf("bad boost %q", b)
		}
	}
	var info *ResearchTierInfo
	for i := range researchTiers {
		if researchTiers[i].Tier == tier {
			info = &researchTiers[i]
			break
		}
	}
	if info == nil {
		return fmt.Errorf("bad tier")
	}
	if s.ResearchFrags < info.Frags {
		return fmt.Errorf("need %d research fragments, have %d", info.Frags, s.ResearchFrags)
	}
	if s.BTC < float64(info.Money) {
		return fmt.Errorf("need %s, have %s", FmtBTCInt(info.Money), FmtBTC(s.BTC))
	}
	s.ResearchFrags -= info.Frags
	s.BTC -= float64(info.Money)
	s.ActiveResearch = &Research{
		BlueprintTier: tier,
		Boosts:        append([]string{}, boosts...),
		StartedAt:     time.Now().Unix(),
		DurationSec:   info.Duration,
	}
	s.appendLog("info", i18n.T("log.research.started", info.Name, info.Duration))
	return nil
}

// ResearchProgress returns the [0..1] completion fraction of active research.
func (s *State) ResearchProgress() float64 {
	if s.ActiveResearch == nil {
		return 0
	}
	elapsed := time.Now().Unix() - s.ActiveResearch.StartedAt
	if elapsed >= int64(s.ActiveResearch.DurationSec) {
		return 1.0
	}
	return float64(elapsed) / float64(s.ActiveResearch.DurationSec)
}

// advanceResearch finalises research when its duration elapses.
func (s *State) advanceResearch(now int64) {
	if s.ActiveResearch == nil {
		return
	}
	if now < s.ActiveResearch.StartedAt+int64(s.ActiveResearch.DurationSec) {
		return
	}
	bp := &Blueprint{
		ID:        fmt.Sprintf("bp_%d", s.NextBlueprintN),
		Tier:      s.ActiveResearch.BlueprintTier,
		Boosts:    append([]string{}, s.ActiveResearch.Boosts...),
		CreatedAt: now,
	}
	s.NextBlueprintN++
	// 10% chance of a random bonus boost.
	if rand.Float64() < 0.10 {
		for _, b := range ResearchBoosts {
			hit := false
			for _, x := range bp.Boosts {
				if x == b {
					hit = true
					break
				}
			}
			if !hit {
				bp.Boosts = append(bp.Boosts, b)
				s.appendLog("opportunity", i18n.T("log.research.breakthrough"))
				break
			}
		}
	}
	s.Blueprints = append(s.Blueprints, bp)
	s.ActiveResearch = nil
	s.appendLog("opportunity", i18n.T("log.research.complete", bpName(bp.Tier), joinStrs(bp.Boosts)))
}

// PrintMEOWCore instantiates a MEOWCore GPU from a researched blueprint.
// Costs: 30% of the original research money + 20% frags, and a room slot.
func (s *State) PrintMEOWCore(blueprintID string) error {
	bp := s.BlueprintByID(blueprintID)
	if bp == nil {
		return fmt.Errorf("no such blueprint")
	}
	var info *ResearchTierInfo
	for i := range researchTiers {
		if researchTiers[i].Tier == bp.Tier {
			info = &researchTiers[i]
			break
		}
	}
	if info == nil {
		return fmt.Errorf("bad blueprint tier")
	}
	cost := info.Money * 3 / 10
	frags := info.Frags / 5
	if s.BTC < float64(cost) {
		return fmt.Errorf("need %s to print", FmtBTCInt(cost))
	}
	if s.ResearchFrags < frags {
		return fmt.Errorf("need %d fragments to print", frags)
	}
	if !s.RoomHasFreeSlot(s.CurrentRoom) {
		return fmt.Errorf("no free slot in this room")
	}
	s.BTC -= float64(cost)
	s.ResearchFrags -= frags
	s.addMEOWCore(bp, s.CurrentRoom)
	s.appendLog("opportunity", i18n.T("log.research.printed", bpName(bp.Tier), joinStrs(bp.Boosts)))
	return nil
}

// BlueprintStats resolves effective stats for a MEOWCore GPU at runtime.
// Returns efficiency, power_draw, heat_output, durability.
func BlueprintStats(bp *Blueprint) (float64, float64, float64, float64) {
	// Base by tier.
	base := struct{ Eff, Pow, Heat, Dur float64 }{0.050, 8, 4, 80}
	switch bp.Tier {
	case 2:
		base = struct{ Eff, Pow, Heat, Dur float64 }{0.080, 10, 5, 100}
	case 3:
		base = struct{ Eff, Pow, Heat, Dur float64 }{0.130, 12, 3, 200}
	}
	for _, b := range bp.Boosts {
		switch b {
		case "efficiency":
			base.Eff *= 1.40
			base.Pow *= 1.20
		case "undervolt":
			base.Pow *= 0.60
			base.Eff *= 0.90
		case "durability":
			base.Dur *= 2.0
			base.Eff *= 0.95
		}
	}
	return base.Eff, base.Pow, base.Heat, base.Dur
}

func bpName(tier int) string {
	switch tier {
	case 1:
		return "MEOWCore v1"
	case 2:
		return "MEOWCore v2"
	case 3:
		return "MEOWCore Purrfect"
	}
	return fmt.Sprintf("MEOWCore v%d", tier)
}

func validBoost(b string) bool {
	for _, x := range ResearchBoosts {
		if x == b {
			return true
		}
	}
	return false
}

func joinStrs(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += " + "
		}
		out += s
	}
	return out
}
