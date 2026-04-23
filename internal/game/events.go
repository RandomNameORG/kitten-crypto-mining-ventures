package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
)

// EventTickInterval is how often (in simulated seconds) we roll for events.
const EventTickInterval = 30

// MaybeFireEvent rolls once to decide if an event should fire now, and if so
// resolves it. Called by the UI tick.
//
// Pacing target: one event every 5-10 minutes on average.
// Each call has ~5% chance to fire if enough time has passed.
func (s *State) MaybeFireEvent() *data.EventDef {
	now := time.Now().Unix()
	roomDef, ok := data.RoomByID(s.CurrentRoom)
	if !ok {
		return nil
	}
	// Weight roll against base threat level — idle-friendly cap.
	pool := roomDef.ThreatPool
	// Include global pool (opportunity + social events that don't care about room).
	globalPool := []string{"tech_share", "extra_delivery", "btc_pump", "lucky_fish"}
	all := append([]string{}, pool...)
	all = append(all, globalPool...)

	// Filter out events on cooldown.
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

	// Per-call fire probability tied to room base threat rate but capped.
	baseFire := 0.04 + roomDef.ThreatBase*0.4
	if baseFire > 0.12 {
		baseFire = 0.12
	}
	if rand.Float64() > baseFire {
		return nil
	}

	// Weighted pick.
	roll := rand.Intn(totalWeight)
	var chosen data.EventDef
	for _, def := range eligible {
		roll -= def.Weight
		if roll < 0 {
			chosen = def
			break
		}
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
			s.Modifiers = append(s.Modifiers, Modifier{
				Kind:      "pause_mining",
				ExpiresAt: now + int64(eff.Seconds),
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
				s.appendLog("info", "…but there was no room to install it. Sold it for cash.")
				if def, ok := data.GPUByID(candidate); ok {
					s.Money += float64(def.ScrapValue)
				}
			}
		case "btc_multiplier":
			s.Modifiers = append(s.Modifiers, Modifier{
				Kind:      "btc_mult",
				Factor:    eff.Factor,
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
			if rand.Float64() < eff.Chance {
				s.burnCurrentRoom()
			} else {
				s.appendLog("opportunity", "Somehow, nothing caught fire. Lucky.")
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
	defense := float64(room.LockLvl)*0.03 + float64(room.CCTVLvl)*0.02
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
			s.appendLog("opportunity", "🛡  Defense held. Nothing stolen.")
			continue
		}
		idx := rand.Intn(len(candidates))
		target := candidates[idx]
		target.Status = "stolen"
		if def, ok := data.GPUByID(target.DefID); ok {
			s.appendLog("threat", fmt.Sprintf("🐀 They took your %s. Gone.", def.Name))
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
	if def, ok := data.GPUByID(victim.DefID); ok {
		victim.HoursLeft -= float64(def.DurabilityHours) * amount
		if victim.HoursLeft <= 0 {
			victim.HoursLeft = 0
			victim.Status = "broken"
			s.appendLog("threat", fmt.Sprintf("💥 %s took too much damage — broken.", def.Name))
		} else {
			s.appendLog("threat", fmt.Sprintf("⚠️  %s damaged.", def.Name))
		}
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

// RepairGPU repairs a broken GPU for 30% of price.
func (s *State) RepairGPU(instanceID int) error {
	for _, g := range s.GPUs {
		if g.InstanceID != instanceID {
			continue
		}
		if g.Status != "broken" {
			return fmt.Errorf("not broken")
		}
		def, _ := data.GPUByID(g.DefID)
		cost := def.Price * 3 / 10
		if s.Money < float64(cost) {
			return fmt.Errorf("need $%d to repair", cost)
		}
		s.Money -= float64(cost)
		g.Status = "running"
		g.HoursLeft = float64(def.DurabilityHours) * 0.6
		s.appendLog("info", fmt.Sprintf("🔧 Repaired %s for $%d.", def.Name, cost))
		return nil
	}
	return fmt.Errorf("no such GPU")
}
