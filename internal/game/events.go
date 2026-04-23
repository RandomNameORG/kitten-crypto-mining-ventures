package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
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
	globalPool := []string{"tech_share", "extra_delivery", "btc_pump", "lucky_fish"}
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
		s.appendLog("opportunity", fmt.Sprintf("%s A threat was averted silently by Chain Ghost.", chosen.Emoji))
		s.EventCooldown[chosen.ID] = now
		return nil
	}

	s.EventCooldown[chosen.ID] = now
	s.applyEvent(chosen)
	return &chosen
}

func (s *State) applyEvent(e data.EventDef) {
	s.appendLog(e.Category, fmt.Sprintf("%s %s — %s", e.Emoji, e.Name, e.Text))
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
			s.appendLog("opportunity", fmt.Sprintf("🧠 +%d TechPoint.", eff.Delta))
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
					s.appendLog("opportunity", fmt.Sprintf("🎁 Free %s installed.", def.Name))
				}
			} else {
				if def, ok := data.GPUByID(candidate); ok {
					s.Money += float64(def.ScrapValue) * s.ScrapValueMult()
					s.appendLog("info", fmt.Sprintf("…but no room. Sold %s for cash.", def.Name))
				}
			}
		case "btc_multiplier":
			// Hedged wallet dampens the swing toward 1.0.
			factor := eff.Factor
			if damp := s.BTCVolatilityDamp(); damp < 1.0 {
				factor = 1.0 + (factor-1.0)*damp
			}
			s.Modifiers = append(s.Modifiers, Modifier{
				Kind:      "btc_mult",
				Factor:    factor,
				ExpiresAt: now + int64(eff.Seconds),
			})
		case "earn_multiplier":
			s.Modifiers = append(s.Modifiers, Modifier{
				Kind:      "earn_mult",
				Factor:    eff.Factor,
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
				s.appendLog("opportunity", "🛡 Armor held. Nothing caught fire.")
			}
		case "eviction_warning":
			s.Reputation -= 5
			s.appendLog("threat", "You've been warned. One more incident and the room is gone.")
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
		s.appendLog("opportunity", "Thief found nothing worth taking. Huh.")
		return
	}
	for i := 0; i < count && len(candidates) > 0; i++ {
		if rand.Float64() > chance {
			s.appendLog("opportunity", "🛡 Defense held. Nothing stolen.")
			continue
		}
		idx := rand.Intn(len(candidates))
		target := candidates[idx]
		target.Status = "stolen"
		if def, ok := data.GPUByID(target.DefID); ok {
			s.appendLog("threat", fmt.Sprintf("🐀 They took your %s. Gone.", def.Name))
		} else {
			s.appendLog("threat", "🐀 They took one of your MEOWCores. Devastating.")
		}
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
		s.appendLog("threat", "💥 A GPU took too much damage — broken.")
	} else {
		s.appendLog("threat", "⚠️  A GPU is damaged.")
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
	s.appendLog("crisis", fmt.Sprintf("🔥 Fire! %d GPUs destroyed in %s.", destroyed, s.CurrentRoom))
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
		if s.Money < float64(cost) {
			return fmt.Errorf("need $%d to repair", cost)
		}
		s.Money -= float64(cost)
		g.Status = "running"
		_, _, _, dur := s.GPUStats(g)
		g.HoursLeft = dur * 0.6
		if cost == 0 {
			s.appendLog("info", "🔧 PCB surgery — free repair.")
		} else {
			s.appendLog("info", fmt.Sprintf("🔧 Repaired for $%d.", cost))
		}
		return nil
	}
	return fmt.Errorf("no such GPU")
}
