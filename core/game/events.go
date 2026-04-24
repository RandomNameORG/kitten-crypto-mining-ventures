package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

// MaybeFireEvent rolls once to decide if an event should fire now, and if so
// resolves it. Called by the UI tick. Target: one event every 5-10 minutes.
func (s *State) MaybeFireEvent() *data.EventDef {
	now := time.Now().Unix()
	roomDef, ok := data.RoomByID(s.CurrentRoom)
	if !ok {
		return nil
	}
	pool := roomDef.ThreatPool
	globalPool := []string{
		"tech_share", "extra_delivery", "btc_pump", "lucky_fish",
		"group_chat_sos", "celeb_interview", "halving", "police_visit",
		"voltage_dip",
		"tax_audit", "power_surge", "market_crash",
	}
	all := append([]string{}, pool...)
	all = append(all, globalPool...)

	eligible := []data.EventDef{}
	totalWeight := 0
	for _, id := range all {
		def, ok := data.EventByID(id)
		if !ok {
			continue
		}
		last := s.EventCooldown[id]
		if now-last < int64(def.CooldownSec) {
			continue
		}
		if !s.eventGatePasses(id) {
			continue
		}
		eligible = append(eligible, def)
		totalWeight += def.Weight
	}
	if len(eligible) == 0 || totalWeight == 0 {
		return nil
	}

	baseFire := 0.04 + roomDef.ThreatBase*0.4
	if baseFire > 0.12 {
		baseFire = 0.12
	}
	baseFire *= s.DifficultyThreatMult()
	// Rich-cat tax: flush players attract more attention.
	baseFire *= 1.0 + s.GreedScore()
	if rand.Float64() > baseFire {
		return nil
	}

	roll := rand.Intn(totalWeight)
	var chosen data.EventDef
	for _, def := range eligible {
		roll -= def.Weight
		if roll < 0 {
			chosen = def
			break
		}
	}

	// Chain Ghost skill auto-handles threats rarely.
	if chosen.Category == "threat" && s.HasSkill("chain_ghost") && rand.Float64() < 0.25 {
		s.appendLog("opportunity", i18n.T("log.event.chain_ghost", chosen.Emoji))
		s.EventCooldown[chosen.ID] = now
		return nil
	}

	s.EventCooldown[chosen.ID] = now
	s.applyEvent(chosen)
	return &chosen
}

// eventGatePasses returns false when the event should NOT fire under the
// current state (e.g. Police only shows up when Karma/Rep are low, Celeb
// interviews only happen once you have some assets to show off).
func (s *State) eventGatePasses(id string) bool {
	switch id {
	case "police_visit":
		return s.Karma < -30 || s.Reputation < -50
	case "celeb_interview":
		return s.LifetimeEarned > 50_000
	case "group_chat_sos":
		return s.Reputation >= 0
	case "tax_audit":
		return s.LifetimeEarned > 100_000
	case "market_crash":
		return s.LifetimeEarned > 50_000
	case "power_surge":
		// Only meaningful when there's an overclocked, running GPU in the
		// current room — otherwise the effect fizzles on apply and just
		// burns a roll. Gate it out so it never even enters the pool.
		for _, g := range s.GPUs {
			if g.Room == s.CurrentRoom && g.Status == "running" && g.OCLevel > 0 {
				return true
			}
		}
		return false
	}
	return true
}

func (s *State) applyEvent(e data.EventDef) {
	s.appendLog(e.Category, fmt.Sprintf("%s %s — %s", e.Emoji, e.LocalName(), e.LocalText()))
	now := time.Now().Unix()

	for _, eff := range e.Effects {
		switch eff.Kind {
		case "steal_gpu":
			s.tryStealGPUs(eff)
		case "pause_mining":
			// Wiring reduces outage duration.
			secs := eff.Seconds
			if room := s.Rooms[s.CurrentRoom]; room != nil {
				reduction := room.WiringLvl * 10
				secs -= reduction
				if secs < 10 {
					secs = 10
				}
			}
			s.Modifiers = append(s.Modifiers, Modifier{
				Kind:      "pause_mining",
				ExpiresAt: now + int64(secs),
			})
		case "rep_change":
			s.Reputation += eff.Delta
		case "tech_point":
			s.TechPoint += eff.Delta
			s.appendLog("opportunity", i18n.T("log.event.tp_gained", eff.Delta))
		case "gift_gpu":
			candidate := "gtx1060"
			switch eff.Tier {
			case "common":
				pool := []string{"gtx1060", "gtx1060ti", "rx580"}
				candidate = pool[rand.Intn(len(pool))]
			case "rare":
				candidate = "gtx1080ti"
			}
			if s.RoomHasFreeSlot(s.CurrentRoom) {
				s.addGPU(candidate, s.CurrentRoom, false)
				if def, ok := data.GPUByID(candidate); ok {
					s.appendLog("opportunity", i18n.T("log.event.gift.installed", def.LocalName()))
				}
			} else {
				if def, ok := data.GPUByID(candidate); ok {
					s.BTC += float64(def.ScrapValue) * s.ScrapValueMult()
					s.appendLog("info", i18n.T("log.event.gift.sold", def.LocalName()))
				}
			}
		case "earn_multiplier":
			// Hedged Wallet softens earn-rate swings (both positive and
			// negative) back toward 1.0.
			factor := eff.Factor
			if damp := s.EarnVolatilityDamp(); damp < 1.0 {
				factor = 1.0 + (factor-1.0)*damp
			}
			s.Modifiers = append(s.Modifiers, Modifier{
				Kind:      "earn_mult",
				Factor:    factor,
				ExpiresAt: now + int64(eff.Seconds),
			})
		case "damage_gpu":
			s.damageRandomGPU(eff.Amount)
		case "burn_room_chance":
			// Armor protects vs crisis fires.
			chance := eff.Chance
			if room := s.Rooms[s.CurrentRoom]; room != nil {
				chance -= float64(room.ArmorLvl) * 0.08
			}
			if chance < 0.05 {
				chance = 0.05
			}
			if rand.Float64() < chance {
				s.burnCurrentRoom()
			} else {
				s.appendLog("opportunity", i18n.T("log.event.fire.averted"))
			}
		case "eviction_warning":
			s.Reputation -= 5
			s.appendLog("threat", i18n.T("log.event.fire.warning"))
		case "money_loss":
			frac := eff.Amount
			if frac <= 0 {
				frac = 0.1
			}
			loss := s.BTC * frac
			s.BTC -= loss
			if s.BTC < 0 {
				s.BTC = 0
			}
			s.appendLog("threat", i18n.T("log.event.fire.money", FmtBTC(loss), frac*100))
		case "tax_audit":
			threshold := eff.ReserveFactor * s.LifetimeEarned
			if s.BTC >= threshold {
				s.appendLog("info", i18n.T("log.event.tax.clean"))
			} else {
				frac := eff.Amount
				if frac <= 0 {
					frac = 0.1
				}
				loss := s.BTC * frac
				s.BTC -= loss
				if s.BTC < 0 {
					s.BTC = 0
				}
				s.Reputation -= 5
				s.appendLog("threat", i18n.T("log.event.tax.hit", FmtBTC(loss), frac*100))
			}
		case "damage_oc_gpu":
			s.damageRandomOCGPU(eff.Amount)
		case "market_pin":
			s.Modifiers = append(s.Modifiers, Modifier{
				Kind:      "market_pin",
				Factor:    eff.Factor,
				ExpiresAt: now + int64(eff.Seconds),
			})
			s.MarketPrice = eff.Factor
			s.PrevMarketPrice = s.MarketPrice
			s.appendLog("crisis", i18n.T("log.event.crash.fired", eff.Factor, eff.Seconds))
		}
	}
}

func (s *State) tryStealGPUs(eff data.EventEffect) {
	room := s.Rooms[s.CurrentRoom]
	if room == nil {
		return
	}
	defense := float64(room.LockLvl)*0.03 +
		float64(room.CCTVLvl)*0.02 +
		float64(room.ArmorLvl)*0.025 +
		s.MercDefenseBonus(s.CurrentRoom)
	count := eff.Count
	if count == 0 {
		count = 1
	}
	chance := eff.ChanceIfNoDefense - defense
	if chance < 0.05 {
		chance = 0.05
	}
	candidates := []*GPU{}
	for _, g := range s.GPUs {
		if g.Room == s.CurrentRoom && g.Status == "running" {
			candidates = append(candidates, g)
		}
	}
	if len(candidates) == 0 {
		s.appendLog("opportunity", i18n.T("log.event.thief.empty"))
		return
	}
	for i := 0; i < count && len(candidates) > 0; i++ {
		if rand.Float64() > chance {
			s.appendLog("opportunity", i18n.T("log.event.thief.defended"))
			continue
		}
		idx := rand.Intn(len(candidates))
		target := candidates[idx]
		if def, ok := data.GPUByID(target.DefID); ok {
			s.appendLog("threat", i18n.T("log.event.thief.took_gpu", def.LocalName()))
		} else {
			s.appendLog("threat", i18n.T("log.event.thief.took_bp"))
		}
		s.removeGPU(target.InstanceID)
		candidates = append(candidates[:idx], candidates[idx+1:]...)
	}
}

func (s *State) damageRandomGPU(amount float64) {
	if amount <= 0 {
		amount = 0.1
	}
	candidates := []*GPU{}
	for _, g := range s.GPUs {
		if g.Room == s.CurrentRoom && g.Status == "running" {
			candidates = append(candidates, g)
		}
	}
	if len(candidates) == 0 {
		return
	}
	victim := candidates[rand.Intn(len(candidates))]
	_, _, _, dur := s.GPUStats(victim)
	victim.HoursLeft -= dur * amount
	if victim.HoursLeft <= 0 {
		victim.HoursLeft = 0
		victim.Status = "broken"
		s.appendLog("threat", i18n.T("log.event.gpu.broken"))
	} else {
		s.appendLog("threat", i18n.T("log.event.gpu.damaged"))
	}
}

// damageRandomOCGPU is damageRandomGPU's overclock-only sibling used by the
// power_surge event: only running GPUs with OCLevel > 0 are eligible. If no
// OC'd GPU is present at apply time (the gate may have passed earlier this
// tick but the set can change), the surge fizzles harmlessly.
func (s *State) damageRandomOCGPU(amount float64) {
	if amount <= 0 {
		amount = 0.1
	}
	candidates := []*GPU{}
	for _, g := range s.GPUs {
		if g.Room == s.CurrentRoom && g.Status == "running" && g.OCLevel > 0 {
			candidates = append(candidates, g)
		}
	}
	if len(candidates) == 0 {
		s.appendLog("info", i18n.T("log.event.surge.fizzle"))
		return
	}
	victim := candidates[rand.Intn(len(candidates))]
	_, _, _, dur := s.GPUStats(victim)
	victim.HoursLeft -= dur * amount
	if victim.HoursLeft <= 0 {
		victim.HoursLeft = 0
		victim.Status = "broken"
		s.appendLog("threat", i18n.T("log.event.gpu.broken"))
	} else {
		s.appendLog("threat", i18n.T("log.event.surge.damaged"))
	}
}

func (s *State) burnCurrentRoom() {
	destroyed := 0
	for _, g := range s.GPUs {
		if g.Room == s.CurrentRoom && g.Status == "running" {
			g.Status = "broken"
			g.HoursLeft = 0
			destroyed++
		}
	}
	roomName := s.CurrentRoom
	if def, ok := data.RoomByID(s.CurrentRoom); ok {
		roomName = def.LocalName()
	}
	s.appendLog("crisis", i18n.T("log.event.fire.destroyed", destroyed, roomName))
}

// RepairGPU repairs a broken GPU. Free if PCB Surgery is unlocked.
func (s *State) RepairGPU(instanceID int) error {
	for _, g := range s.GPUs {
		if g.InstanceID != instanceID {
			continue
		}
		if g.Status != "broken" {
			return fmt.Errorf("not broken")
		}
		var price int
		if def, ok := data.GPUByID(g.DefID); ok {
			price = def.Price
		} else {
			price = 3000
		}
		cost := price * 3 / 10
		if s.RepairFree() {
			cost = 0
		}
		if s.BTC < float64(cost) {
			return fmt.Errorf("need %s to repair", FmtBTCInt(cost))
		}
		s.BTC -= float64(cost)
		g.Status = "running"
		_, _, _, dur := s.GPUStats(g)
		g.HoursLeft = dur * 0.6
		if cost == 0 {
			s.appendLog("info", i18n.T("log.event.repair.free"))
		} else {
			s.appendLog("info", i18n.T("log.event.repair.paid", FmtBTCInt(cost)))
		}
		return nil
	}
	return fmt.Errorf("no such GPU")
}
