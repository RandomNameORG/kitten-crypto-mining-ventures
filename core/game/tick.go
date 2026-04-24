package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

// OfflineCapSeconds is the maximum gap the offline catch-up will simulate.
// Beyond this, we clamp the timestamps so bills/wages don't drain months of
// accumulated costs in one blow.
const OfflineCapSeconds int64 = 8 * 3600

// RunOfflineCatchup advances the sim from the last tick to `now` in a single
// catch-up pass and leaves an OfflineSummary on the state for the UI to
// display on startup. Small gaps (< 60s) are skipped — no summary.
func (s *State) RunOfflineCatchup(now int64) {
	gap := now - s.LastTickUnix
	if gap < 60 {
		s.Tick(now)
		return
	}
	capped := false
	if gap > OfflineCapSeconds {
		gap = OfflineCapSeconds
		capped = true
		s.LastTickUnix = now - gap
		s.LastBillUnix = now - gap
		s.LastWagesUnix = now - gap
	}
	btcBefore := s.BTC
	s.Tick(now)
	s.OfflineSummary = &OfflineSummary{
		GapSeconds: gap,
		BTCGained:  s.BTC - btcBefore,
		Capped:     capped,
	}
}

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
				s.appendLog("info", i18n.T("log.gpu.arrived", def.LocalName()))
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

	for roomID, room := range s.Rooms {
		roomDef, ok := data.RoomByID(roomID)
		if !ok {
			continue
		}
		coolingBonus := 1.0 + 0.25*float64(room.CoolingLvl)

		for _, g := range s.GPUs {
			if g.Room != roomID || g.Status != "running" {
				continue
			}
			eff, _, _, dur := s.GPUStats(g)
			efficiencyFactor := 1.0
			if room.Heat > 0.8*room.MaxHeat {
				efficiencyFactor = 0.5
			}
			if !miningPaused {
				earned := eff * dt * earnMult * efficiencyFactor * s.DifficultyEarnMult() * MiningScale
				s.BTC += earned
				s.LifetimeEarned += earned
			}

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
					s.appendLog("threat", i18n.T("log.gpu.failed", name))
				}
			}
		}
		// Heat updates at a room-specific interval (see HeatTickSec in
		// rooms.json). Good-cooling rooms update rarely (chunky jumps),
		// bad rooms update often. Between ticks heat is flat so the
		// player sees discrete thermal events, not a per-second crawl.
		tickInterval := int64(roomDef.HeatTickSec)
		if tickInterval <= 0 {
			tickInterval = 10
		}
		if room.LastHeatTickUnix == 0 {
			room.LastHeatTickUnix = now
		}
		elapsedSinceHeatTick := now - room.LastHeatTickUnix
		if elapsedSinceHeatTick >= tickInterval {
			ticks := elapsedSinceHeatTick / tickInterval
			room.LastHeatTickUnix += ticks * tickInterval
			// heatDelta accumulator above is in per-second units; convert
			// to per-tick by multiplying by the GPU-side heat rate directly.
			// Recompute instead of rescaling heatDelta to keep units clear.
			var heatPerTick float64
			for _, g := range s.GPUs {
				if g.Room != roomID || g.Status != "running" {
					continue
				}
				_, _, hOut, _ := s.GPUStats(g)
				heatPerTick += hOut
			}
			netPerTick := heatPerTick - roomDef.BaseCooling*coolingBonus
			room.Heat += netPerTick * float64(ticks)
		}
		if room.Heat < 20 {
			room.Heat = 20
		}
		if room.Heat > room.MaxHeat {
			room.Heat = room.MaxHeat
		}
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
	s.BTC -= totalBill
	s.BTC -= totalRent
	if totalBill+totalRent > 0 {
		s.appendLog("info", i18n.T("log.bills.settled", totalBill, totalRent))
	}
	if s.BTC < 0 {
		s.BTC = 0
		s.Modifiers = append(s.Modifiers, Modifier{
			Kind:      "pause_mining",
			ExpiresAt: now + 60,
		})
		s.appendLog("threat", i18n.T("log.bills.blackout"))
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
		if s.BTC < float64(cost) {
			return fmt.Errorf("need ₿%d, have ₿%.0f", cost, s.BTC)
		}
		s.BTC -= float64(cost)
		failChance := 0.05 + 0.05*float64(g.UpgradeLevel)
		if rand.Float64() < failChance {
			g.Status = "broken"
			s.appendLog("threat", i18n.T("log.gpu.upgrade.bricked"))
			return nil
		}
		g.UpgradeLevel++
		s.appendLog("info", i18n.T("log.gpu.upgrade.success", g.UpgradeLevel))
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
	if s.BTC < EmergencyVentCost {
		return fmt.Errorf("need ₿%d", EmergencyVentCost)
	}
	now := time.Now().Unix()
	last := s.EventCooldown["vent:"+s.CurrentRoom]
	if now-last < EmergencyVentCooldownSec {
		return fmt.Errorf("cooldown: %ds left", EmergencyVentCooldownSec-(now-last))
	}
	s.BTC -= EmergencyVentCost
	room.Heat = 20
	room.LastHeatTickUnix = now // give the player a fresh interval after cooling
	s.EventCooldown["vent:"+s.CurrentRoom] = now
	s.Modifiers = append(s.Modifiers, Modifier{
		Kind:      "pause_mining",
		ExpiresAt: now + 30,
	})
	s.appendLog("info", i18n.T("log.room.vent", EmergencyVentCost))
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
		Kind:      "earn_mult",
		Factor:    1.5,
		ExpiresAt: now + 300,
	})
	s.appendLog("opportunity", i18n.T("log.pump.fired"))
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
		lvl, label = &room.LockLvl, i18n.T("defense.lock")
	case "cctv":
		lvl, label = &room.CCTVLvl, i18n.T("defense.cctv")
	case "wiring":
		lvl, label = &room.WiringLvl, i18n.T("defense.wiring")
	case "cooling":
		lvl, label = &room.CoolingLvl, i18n.T("defense.cooling")
	case "armor":
		lvl, label = &room.ArmorLvl, i18n.T("defense.armor")
	default:
		return fmt.Errorf("bad dim %q", dim)
	}
	if *lvl >= 5 {
		return fmt.Errorf("%s already maxed", label)
	}
	cost := (*lvl + 1) * 250
	if s.BTC < float64(cost) {
		return fmt.Errorf("need ₿%d, have ₿%.0f", cost, s.BTC)
	}
	s.BTC -= float64(cost)
	*lvl++
	s.appendLog("info", i18n.T("log.defense.upgraded", label, *lvl))
	return nil
}
