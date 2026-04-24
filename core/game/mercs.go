package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

// HireMerc hires a merc by def id into the current room.
func (s *State) HireMerc(defID string) error {
	def, ok := data.MercByID(defID)
	if !ok {
		return fmt.Errorf("unknown merc")
	}
	if s.BTC < float64(def.HireCost) {
		return fmt.Errorf("need %s, have %s", FmtBTCInt(def.HireCost), FmtBTC(s.BTC))
	}
	s.BTC -= float64(def.HireCost)
	loyalty := def.LoyaltyBase + s.MercLoyaltyFloor()
	if loyalty > 100 {
		loyalty = 100
	}
	m := &Merc{
		InstanceID: s.NextMercID,
		DefID:      defID,
		Loyalty:    loyalty,
		HiredAt:    time.Now().Unix(),
		RoomID:     s.CurrentRoom,
	}
	s.NextMercID++
	s.Mercs = append(s.Mercs, m)
	s.appendLog("info", i18n.T("log.merc.hired", def.LocalName(), FmtBTCInt(def.HireCost)))
	return nil
}

// FireMerc dismisses a merc (no refund, loyalty hit for remaining mercs).
func (s *State) FireMerc(instanceID int) error {
	for i, m := range s.Mercs {
		if m.InstanceID == instanceID {
			def, _ := data.MercByID(m.DefID)
			s.Mercs = append(s.Mercs[:i], s.Mercs[i+1:]...)
			s.appendLog("info", i18n.T("log.merc.dismissed", def.LocalName()))
			for _, other := range s.Mercs {
				other.Loyalty -= 5
			}
			return nil
		}
	}
	return fmt.Errorf("no such merc")
}

// BribeMerc spends ₿200 for +15 loyalty.
func (s *State) BribeMerc(instanceID int) error {
	const cost = 200
	if s.BTC < cost {
		return fmt.Errorf("need %s", FmtBTCInt(cost))
	}
	for _, m := range s.Mercs {
		if m.InstanceID == instanceID {
			s.BTC -= cost
			m.Loyalty += 15
			if m.Loyalty > 100 {
				m.Loyalty = 100
			}
			def, _ := data.MercByID(m.DefID)
			s.appendLog("info", i18n.T("log.merc.bribed", def.LocalName(), m.Loyalty))
			return nil
		}
	}
	return fmt.Errorf("no such merc")
}

// payWages is called from tick; pays all mercs at the weekly cadence
// (1 game week = 60 sim minutes).
func (s *State) payWages(now int64) {
	if now-s.LastWagesUnix < 3600 {
		return
	}
	weeks := float64(now-s.LastWagesUnix) / 3600.0
	s.LastWagesUnix = now

	totalWage := 0.0
	for _, m := range s.Mercs {
		def, ok := data.MercByID(m.DefID)
		if !ok {
			continue
		}
		wage := float64(def.WeeklyWage) * weeks
		totalWage += wage
		// Loyalty drift: +1 per week on time, or −10 if we can't afford.
		if s.BTC >= wage {
			s.BTC -= wage
			m.Loyalty++
		} else {
			// Missed wages — loyalty tanks, wage still banked as debt-reduction.
			s.BTC = 0
			m.Loyalty -= 10
		}
		if m.Loyalty > 100 {
			m.Loyalty = 100
		}
		if m.Loyalty < 0 {
			m.Loyalty = 0
		}
	}
	if totalWage > 0 {
		s.appendLog("info", i18n.T("log.merc.wages", FmtBTC(totalWage)))
	}

	// Random betrayal check — once per wage tick, one low-loyalty merc might flip.
	for _, m := range s.Mercs {
		if m.Loyalty < 20 && rand.Float64() < 0.15 {
			s.triggerBetrayal(m)
			return
		}
	}
}

// triggerBetrayal is a minor crisis event produced by the merc system itself.
func (s *State) triggerBetrayal(m *Merc) {
	def, _ := data.MercByID(m.DefID)
	switch def.Specialty {
	case "guard":
		// Let the nearest thief in.
		s.appendLog("crisis", i18n.T("log.merc.betray.unlock", def.LocalName()))
		if room := s.Rooms[m.RoomID]; room != nil {
			room.LockLvl = 0
		}
	case "tech":
		// Sabotage: damage a random GPU.
		s.damageRandomGPU(0.5)
		s.appendLog("crisis", i18n.T("log.merc.betray.sabotage", def.LocalName()))
	case "social":
		// Leak to competitors: rep hit.
		s.Reputation -= 10
		s.appendLog("crisis", i18n.T("log.merc.betray.sold_story", def.LocalName()))
	case "combat":
		// Runs off with the biggest GPU.
		s.stealMostValuable()
		s.appendLog("crisis", i18n.T("log.merc.betray.stole_gpu", def.LocalName()))
	case "sea":
		// Tips off pirates.
		s.appendLog("crisis", i18n.T("log.merc.betray.pirate_crew", def.LocalName()))
		s.Modifiers = append(s.Modifiers, Modifier{
			Kind:      "pirate_warning",
			ExpiresAt: time.Now().Unix() + 600,
		})
	default:
		s.appendLog("crisis", i18n.T("log.merc.betray.generic", def.LocalName()))
	}
	// Fire the traitor.
	for i, x := range s.Mercs {
		if x.InstanceID == m.InstanceID {
			s.Mercs = append(s.Mercs[:i], s.Mercs[i+1:]...)
			break
		}
	}
}

func (s *State) stealMostValuable() {
	var best *GPU
	bestVal := -1
	for _, g := range s.GPUs {
		if g.Room != s.CurrentRoom || g.Status != "running" {
			continue
		}
		def, ok := data.GPUByID(g.DefID)
		if !ok {
			continue
		}
		if def.Price > bestVal {
			bestVal = def.Price
			best = g
		}
	}
	if best != nil {
		s.removeGPU(best.InstanceID)
	}
}

// MercDefenseBonus is the aggregate defense contribution of mercs in a room.
func (s *State) MercDefenseBonus(roomID string) float64 {
	bonus := 0.0
	for _, m := range s.Mercs {
		if m.RoomID != roomID {
			continue
		}
		def, ok := data.MercByID(m.DefID)
		if !ok {
			continue
		}
		// Loyalty scales effectiveness linearly.
		bonus += def.DefenseBonus * (float64(m.Loyalty) / 100.0)
	}
	return bonus
}
