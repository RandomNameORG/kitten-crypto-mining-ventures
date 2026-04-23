package game

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
)

// Tick advances the simulation forward to `now`. It's safe to call every
// frame — it only operates on the delta since LastTickUnix.
func (s *State) Tick(now int64) {
	if s.Paused {
		// Paused: just bump the bill anchor so a long pause doesn't spike
		// the next tick's electricity bill.
		s.LastTickUnix = now
		s.LastBillUnix = now
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
		var heatDelta float64
		for _, g := range s.GPUs {
			if g.Room != roomID || g.Status != "running" {
				continue
			}
			gDef, ok := data.GPUByID(g.DefID)
			if !ok {
				continue
			}
			upgradeMult := 1.0 + 0.15*float64(g.UpgradeLevel)
			upgradeHeatMult := 1.0 + 0.20*float64(g.UpgradeLevel)

			// Overheat debuff: if heat exceeds 80% of max, mining efficiency drops.
			efficiencyFactor := 1.0
			if room.Heat > 0.8*room.MaxHeat {
				efficiencyFactor = 0.5
			}

			if !miningPaused {
				btcEarned := gDef.Efficiency * upgradeMult * dt * earnMult * efficiencyFactor
				s.BTC += btcEarned
			}
			heatDelta += gDef.HeatOutput * upgradeHeatMult * dt

			// Durability decay — GPUs wear out over time.
			g.HoursLeft -= dt / 3600.0
			if g.HoursLeft <= 0 {
				g.Status = "broken"
				g.HoursLeft = 0
				s.appendLog("threat", fmt.Sprintf("💥 %s failed. It needs repair or scrapping.", gDef.Name))
			}
		}
		// Passive cooling every second.
		room.Heat += heatDelta - roomDef.BaseCooling*dt
		if room.Heat < 20 {
			room.Heat = 20
		}
		if room.Heat > room.MaxHeat {
			room.Heat = room.MaxHeat
		}
	}

	// Auto-cash-out BTC: sell a small trickle every tick so money moves too.
	// Players can HODL later via a skill; for v0 we stream 5%/sec of balance.
	if s.BTC > 0 {
		sell := s.BTC * (1 - math.Pow(0.95, dt))
		s.Money += sell * price
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

	totalBill := 0.0
	totalRent := 0.0
	for roomID, _ := range s.Rooms {
		roomDef, ok := data.RoomByID(roomID)
		if !ok {
			continue
		}
		var volt float64
		for _, g := range s.GPUs {
			if g.Room != roomID || g.Status != "running" {
				continue
			}
			gDef, ok := data.GPUByID(g.DefID)
			if !ok {
				continue
			}
			upMult := 1.0 + 0.10*float64(g.UpgradeLevel)
			volt += gDef.PowerDraw * upMult
		}
		totalBill += volt * ElectricPerVoltMin * roomDef.ElectricCostMult * minutes
		totalRent += float64(roomDef.RentPerHour) * (minutes / 60.0)
	}
	s.Money -= totalBill
	s.Money -= totalRent
	if totalBill+totalRent > 0 {
		s.appendLog("info", fmt.Sprintf("💸 Bills settled: $%.2f electricity, $%.2f rent.", totalBill, totalRent))
	}
	if s.Money < 0 {
		s.Money = 0
		// If broke, pause all GPUs by blackout.
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
		def, _ := data.GPUByID(g.DefID)
		cost := def.Price / 3
		if s.Money < float64(cost) {
			return fmt.Errorf("need $%d, have $%.0f", cost, s.Money)
		}
		s.Money -= float64(cost)
		// Failure chance grows per level.
		failChance := 0.05 + 0.05*float64(g.UpgradeLevel)
		if rand.Float64() < failChance {
			g.Status = "broken"
			s.appendLog("threat", fmt.Sprintf("🔥 Upgrade failed — %s is bricked.", def.Name))
			return nil
		}
		g.UpgradeLevel++
		s.appendLog("info", fmt.Sprintf("⚙️  %s upgraded to level %d.", def.Name, g.UpgradeLevel))
		return nil
	}
	return fmt.Errorf("no such GPU")
}

// TogglePause flips the Paused flag.
func (s *State) TogglePause() {
	s.Paused = !s.Paused
	now := time.Now().Unix()
	s.LastTickUnix = now
	s.LastBillUnix = now
	if s.Paused {
		s.appendLog("info", "⏸  Paused.")
	} else {
		s.appendLog("info", "▶️  Resumed.")
	}
}

