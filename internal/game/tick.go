package game

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/i18n"
)

// Tick advances the simulation forward to `now`. It's safe to call every
// frame — it only operates on the delta since LastTickUnix.
func (s *State) Tick(now int64) {
	if s.Paused {
		s.LastTickUnix = now
		s.LastBillUnix = now
		s.LastWagesUnix = now
		return
	}
	if now <= s.LastTickUnix {
		return
	}
	dt := float64(now - s.LastTickUnix)
	s.LastTickUnix = now

	s.pruneModifiers(now)
	s.advanceShipping(now)
	s.advanceMining(now, dt)
	s.advanceBilling(now)
	s.advanceResearch(now)
	s.payWages(now)
	s.CheckAchievements()
}

// advanceShipping transitions shipping GPUs to running when their ETA passes.
func (s *State) advanceShipping(now int64) {
	for _, g := range s.GPUs {
		if g.Status == "shipping" && now >= g.ShipsAt {
			g.Status = "running"
			if def, ok := data.GPUByID(g.DefID); ok {
				s.appendLog("info", fmt.Sprintf("📦 %s arrived and is online.", def.Name))
			}
		}
	}
}

// GPUStats returns the effective (efficiency, power, heat, durability) for a
// GPU instance, honoring blueprint overrides and skill multipliers.
func (s *State) GPUStats(g *GPU) (eff, pow, heat, dur float64) {
	if g.BlueprintID != "" {
		if bp := s.BlueprintByID(g.BlueprintID); bp != nil {
			eff, pow, heat, dur = BlueprintStats(bp)
		}
	} else {
		def, ok := data.GPUByID(g.DefID)
		if !ok {
			return 0, 0, 0, 0
		}
		eff, pow, heat, dur = def.Efficiency, def.PowerDraw, def.HeatOutput, float64(def.DurabilityHours)
	}
	upBonus := 1.0 + 0.15*float64(g.UpgradeLevel)
	upPow := 1.0 + 0.10*float64(g.UpgradeLevel)
	upHeat := 1.0 + 0.20*float64(g.UpgradeLevel)
	eff *= upBonus * s.EfficiencyMult()
	pow *= upPow * s.PowerDrawMult()
	heat *= upHeat * s.HeatMult()
	return
}

// advanceMining advances BTC earnings, volt draw, heat.
func (s *State) advanceMining(now int64, dt float64) {
	miningPaused := s.IsMiningPaused(now)
	earnMult := s.earnMultiplier(now)
	price := s.BTCPriceAt(now)

	for roomID, room := range s.Rooms {
		roomDef, ok := data.RoomByID(roomID)
		if !ok {
			continue
		}
		coolingBonus := 1.0 + 0.25*float64(room.CoolingLvl)

		var heatDelta float64
		for _, g := range s.GPUs {
			if g.Room != roomID || g.Status != "running" {
				continue
			}
			eff, _, hOut, dur := s.GPUStats(g)
			efficiencyFactor := 1.0
			if room.Heat > 0.8*room.MaxHeat {
				efficiencyFactor = 0.5
			}
			if !miningPaused {
				btcEarned := eff * dt * earnMult * efficiencyFactor * s.DifficultyEarnMult()
				s.BTC += btcEarned
			}
			heatDelta += hOut * dt

			// Durability decay — GPUs wear out faster when the room is hot.
			//   heat > 80% max: 3× normal wear
			//   heat > 95% max: 8× wear (real danger zone)
			if dur > 0 {
				wearMult := 1.0
				switch {
				case room.Heat > 0.95*room.MaxHeat:
					wearMult = 8.0
				case room.Heat > 0.80*room.MaxHeat:
					wearMult = 3.0
				}
				g.HoursLeft -= (dt / 3600.0) * wearMult
				if g.HoursLeft <= 0 {
					g.Status = "broken"
					g.HoursLeft = 0
					name := g.DefID
					if def, ok := data.GPUByID(g.DefID); ok {
						name = def.LocalName()
					} else if g.BlueprintID != "" {
						name = "MEOWCore"
					}
					s.appendLog("threat", fmt.Sprintf("💥 %s failed. It needs repair or scrapping.", name))
				}
			}
		}
		room.Heat += heatDelta - roomDef.BaseCooling*coolingBonus*dt
		if room.Heat < 20 {
			room.Heat = 20
		}
		if room.Heat > room.MaxHeat {
			room.Heat = room.MaxHeat
		}
	}

	// Auto-cash-out BTC — small trickle so money moves too. Accumulate to
	// LifetimeEarned for prestige tracking.
	if s.BTC > 0 {
		sell := s.BTC * (1 - math.Pow(0.95, dt))
		cashIn := sell * price
		s.Money += cashIn
		s.LifetimeEarned += cashIn
		s.BTC -= sell
	}
}

// advanceBilling deducts electricity bill + rent every 60s.
func (s *State) advanceBilling(now int64) {
	if now-s.LastBillUnix < 60 {
		return
	}
	minutes := float64(now-s.LastBillUnix) / 60.0
	s.LastBillUnix = now

	billMult := s.BillMult() * s.DifficultyBillMult()
	totalBill := 0.0
	totalRent := 0.0
	for roomID := range s.Rooms {
		roomDef, ok := data.RoomByID(roomID)
		if !ok {
			continue
		}
		var volt float64
		for _, g := range s.GPUs {
			if g.Room != roomID || g.Status != "running" {
				continue
			}
			_, pow, _, _ := s.GPUStats(g)
			volt += pow
		}
		totalBill += volt * ElectricPerVoltMin * roomDef.ElectricCostMult * minutes * billMult
		totalRent += float64(roomDef.RentPerHour) * (minutes / 60.0) * s.DifficultyBillMult()
	}
	s.Money -= totalBill
	s.Money -= totalRent
	if totalBill+totalRent > 0 {
		s.appendLog("info", fmt.Sprintf("💸 Bills settled: $%.2f electricity, $%.2f rent.", totalBill, totalRent))
	}
	if s.Money < 0 {
		s.Money = 0
		s.Modifiers = append(s.Modifiers, Modifier{
			Kind:      "pause_mining",
			ExpiresAt: now + 60,
		})
		s.appendLog("threat", "🔌 Couldn't pay the bill. Blackout for 60s.")
	}
}

// UpgradeGPU upgrades a GPU instance one level. 10% chance to brick per level.
func (s *State) UpgradeGPU(instanceID int) error {
	for _, g := range s.GPUs {
		if g.InstanceID != instanceID {
			continue
		}
		if g.UpgradeLevel >= 5 {
			return fmt.Errorf("already at max level")
		}
		var price int
		if def, ok := data.GPUByID(g.DefID); ok {
			price = def.Price
		} else {
			price = 3000 // MEOWCore default upgrade base
		}
		cost := price / 3
		if s.Money < float64(cost) {
			return fmt.Errorf("need $%d, have $%.0f", cost, s.Money)
		}
		s.Money -= float64(cost)
		failChance := 0.05 + 0.05*float64(g.UpgradeLevel)
		if rand.Float64() < failChance {
			g.Status = "broken"
			s.appendLog("threat", "🔥 Upgrade failed — GPU is bricked.")
			return nil
		}
		g.UpgradeLevel++
		s.appendLog("info", fmt.Sprintf("⚙️  GPU upgraded to level %d.", g.UpgradeLevel))
		return nil
	}
	return fmt.Errorf("no such GPU")
}

// EmergencyVentCost is the cash price of a single vent action.
const EmergencyVentCost = 100

// EmergencyVentCooldownSec is the minimum gap between two vents in the same
// room. Prevents spamming it away; you still pay both in cash and in a
// 30-second mining pause.
const EmergencyVentCooldownSec = 120

// EmergencyVent drops the current room's heat to 20°C immediately. Costs
// cash and pauses mining for 30 seconds while the rack reboots.
func (s *State) EmergencyVent() error {
	room := s.Rooms[s.CurrentRoom]
	if room == nil {
		return fmt.Errorf("no room")
	}
	if s.Money < EmergencyVentCost {
		return fmt.Errorf("need $%d", EmergencyVentCost)
	}
	now := time.Now().Unix()
	last := s.EventCooldown["vent:"+s.CurrentRoom]
	if now-last < EmergencyVentCooldownSec {
		return fmt.Errorf("cooldown: %ds left", EmergencyVentCooldownSec-(now-last))
	}
	s.Money -= EmergencyVentCost
	room.Heat = 20
	s.EventCooldown["vent:"+s.CurrentRoom] = now
	s.Modifiers = append(s.Modifiers, Modifier{
		Kind:      "pause_mining",
		ExpiresAt: now + 30,
	})
	s.appendLog("info", fmt.Sprintf("🧊 Emergency vent — heat reset, 30s power cycle, -$%d.", EmergencyVentCost))
	return nil
}

// TogglePause flips the Paused flag.
func (s *State) TogglePause() {
	s.Paused = !s.Paused
	now := time.Now().Unix()
	s.LastTickUnix = now
	s.LastBillUnix = now
	s.LastWagesUnix = now
	if s.Paused {
		s.appendLog("info", i18n.T("game.paused"))
	} else {
		s.appendLog("info", i18n.T("game.resumed"))
	}
}

// TriggerPumpDump invokes the Hacker "Pump & Dump" ability; 30min cooldown.
func (s *State) TriggerPumpDump() error {
	if !s.HasUnlock("pump_dump_action") {
		return fmt.Errorf("requires Pump & Dump skill")
	}
	last := s.EventCooldown["pump_dump"]
	now := time.Now().Unix()
	if now-last < 1800 {
		return fmt.Errorf("on cooldown for %d more minutes", (1800-(now-last))/60)
	}
	s.EventCooldown["pump_dump"] = now
	s.Modifiers = append(s.Modifiers, Modifier{
		Kind:      "btc_mult",
		Factor:    1.5,
		ExpiresAt: now + 300,
	})
	s.appendLog("opportunity", "📈 Pump & Dump — BTC price ×1.5 for 5 minutes.")
	return nil
}

// UpgradeDefense bumps a single defense dimension on the current room.
// dim: "lock" | "cctv" | "wiring" | "cooling" | "armor"
func (s *State) UpgradeDefense(dim string) error {
	room, ok := s.Rooms[s.CurrentRoom]
	if !ok {
		return fmt.Errorf("no current room")
	}
	var lvl *int
	var label string
	switch dim {
	case "lock":
		lvl, label = &room.LockLvl, "Lock"
	case "cctv":
		lvl, label = &room.CCTVLvl, "CCTV"
	case "wiring":
		lvl, label = &room.WiringLvl, "Wiring"
	case "cooling":
		lvl, label = &room.CoolingLvl, "Cooling"
	case "armor":
		lvl, label = &room.ArmorLvl, "Armor"
	default:
		return fmt.Errorf("bad dim %q", dim)
	}
	if *lvl >= 5 {
		return fmt.Errorf("%s already maxed", label)
	}
	cost := (*lvl + 1) * 250
	if s.Money < float64(cost) {
		return fmt.Errorf("need $%d, have $%.0f", cost, s.Money)
	}
	s.Money -= float64(cost)
	*lvl++
	s.appendLog("info", fmt.Sprintf("🛡 %s upgraded to level %d.", label, *lvl))
	return nil
}
