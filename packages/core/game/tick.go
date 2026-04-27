package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"
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
	s.TotalTicks += int64(dt)

	// OC time bookkeeping: add dt to T1/T2 buckets if any running GPU sits at
	// that level this tick. Counted once per tick, not per GPU — the metric
	// is "wall-time spent overclocking" rather than GPU-seconds.
	hasT1, hasT2 := false, false
	for _, g := range s.GPUs {
		if g.Status != "running" {
			continue
		}
		switch g.OCLevel {
		case 1:
			hasT1 = true
		case 2:
			hasT2 = true
		}
	}
	if hasT1 {
		s.OCTimeT1Sec += int64(dt)
	}
	if hasT2 {
		s.OCTimeT2Sec += int64(dt)
	}

	s.pruneModifiers(now)
	s.advanceShipping(now)
	s.advanceMarket(now)
	s.advanceMining(now, dt)
	s.advanceBilling(now)
	s.advancePSUOverload(now, dt)
	s.advanceSyndicate(now)
	s.advanceResearch(now)
	s.payWages(now)
	s.advanceAutoRepair(now)
	s.CheckAchievements()
}

// advanceAutoRepair handles the Auto-Repair Loop skill chain. Costs run
// through RepairGPU which honours PCB Surgery's 50% discount — the player
// still pays for cycles, just cheaply when the chain is fully built.
//
//	auto_repair      — 1 broken GPU / 60s
//	auto_repair_ii   — 1 broken GPU / 30s
//	auto_repair_iii  — all broken GPUs each cycle (still 30s with II,
//	                   60s without)
func (s *State) advanceAutoRepair(now int64) {
	if !s.HasSkill("auto_repair") {
		return
	}
	if s.LastAutoRepairUnix == 0 {
		s.LastAutoRepairUnix = now
		return
	}
	interval := int64(60)
	if s.HasSkill("auto_repair_ii") {
		interval = 30
	}
	if now-s.LastAutoRepairUnix < interval {
		return
	}
	s.LastAutoRepairUnix = now
	burst := s.HasSkill("auto_repair_iii")
	for _, g := range s.GPUs {
		if g.Status != "broken" {
			continue
		}
		_ = s.RepairGPU(g.InstanceID)
		if !burst {
			return
		}
	}
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

// Overclock tradeoff tables, indexed by GPU.OCLevel (0..2). The non-earn
// factors are intentionally ≥ the earn factor — OC must feel like a real
// choice, not a free dial. Applied in GPUStats (eff/pow/heat) and folded
// into advanceMining's wearMult (durability decay rate).
var (
	ocEarnMult  = [3]float64{1.00, 1.25, 1.50}
	ocPowerMult = [3]float64{1.00, 1.40, 1.90}
	ocHeatMult  = [3]float64{1.00, 1.40, 1.90}
	ocWearMult  = [3]float64{1.00, 1.75, 3.00}
)

// ocIndex returns a valid index into the OC multiplier tables for g, even if
// the save was hand-edited past the bounds.
func ocIndex(g *GPU) int {
	if g.OCLevel < 0 || g.OCLevel >= len(ocEarnMult) {
		return 0
	}
	return g.OCLevel
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
	upBonus := upgradeEffMult(g.UpgradeLevel)
	upPow := upgradePowerMult(g.UpgradeLevel)
	upHeat := upgradeHeatMult(g.UpgradeLevel)
	eff *= upBonus * s.EfficiencyMult()
	pow *= upPow * s.PowerDrawMult()
	heat *= upHeat * s.HeatMult()
	oc := ocIndex(g)
	eff *= ocEarnMult[oc]
	pow *= ocPowerMult[oc]
	heat *= ocHeatMult[oc]
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
		coolingBonus := (1.0 + 0.25*float64(room.CoolingLvl)) * s.MasteryCoolingMult()
		// PSU swap downtime: while a replacement is in progress every GPU
		// in the room sits idle (no earn, no wear from heat). Mirrors the
		// design's 2-minute "rebooting the rack" feel.
		roomPSUPaused := room.PSUResumeAt > now

		for _, g := range s.GPUs {
			if g.Room != roomID || g.Status != "running" {
				continue
			}
			eff, _, _, dur := s.GPUStats(g)
			efficiencyFactor := 1.0
			if room.Heat > 0.8*room.MaxHeat {
				efficiencyFactor = 0.5
			}
			if !miningPaused && !roomPSUPaused {
				earned := eff * dt * earnMult * efficiencyFactor * s.DifficultyEarnMult() * s.MarketPrice * MiningScale * s.MasteryEarnMult()
				// Syndicate cut: divert the agreed fraction into the
				// contribution pool before crediting BTC so the player
				// only sees (1-cut) of each GPU's raw earn. Proportional
				// to hashpower falls out of the per-GPU loop naturally.
				if s.SyndicateJoined && earned > 0 {
					cut := earned * SyndicateCutRate
					s.SyndicateContribution += cut
					earned -= cut
				}
				s.BTC += earned
				s.LifetimeEarned += earned
			}

			// Durability decay — GPUs wear out faster when the room is hot.
			//   heat > 80% max: 3× normal wear
			//   heat > 95% max: 8× wear (real danger zone)
			// Overclock multiplies on top: a +50% OC in a critical room is
			// 8×3 = 24× baseline wear. That's the intended compounding cost.
			if dur > 0 {
				wearMult := 1.0
				switch {
				case room.Heat > 0.95*room.MaxHeat:
					wearMult = 8.0
				case room.Heat > 0.80*room.MaxHeat:
					wearMult = 3.0
				}
				wearMult *= ocWearMult[ocIndex(g)]
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
			// PSU(next-sprint): RoomPSUEfficiency / RoomPSUHeat ready to multiply in
			// once balance retune is scheduled.
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

	billMult := s.BillMult() * s.DifficultyBillMult() * s.MasteryBillMult()
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
		// PSU(next-sprint): RoomPSUEfficiency / RoomPSUHeat ready to multiply in
		// once balance retune is scheduled.
		totalBill += volt * ElectricPerVoltMin * roomDef.ElectricCostMult * minutes * billMult
		totalRent += float64(roomDef.RentPerHour) * (minutes / 60.0) * s.DifficultyBillMult()
	}
	s.BTC -= totalBill
	s.BTC -= totalRent
	if totalBill+totalRent > 0 {
		s.appendLog("info", i18n.T("log.bills.settled", FmtBTC(totalBill), FmtBTC(totalRent)))
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

// MaxGPUUpgradeLevel caps the per-card upgrade ladder. Levels 1-5 are
// money-only; levels 6-10 also burn research fragments (a sink for the
// frag economy that otherwise hits ceiling fast).
const MaxGPUUpgradeLevel = 10

// upgradeEffMult: +15%/level for L1-5, +10%/level for L6-10.
// At L5 = 1.75 (unchanged). At L10 = 2.25.
func upgradeEffMult(level int) float64 {
	if level <= 5 {
		return 1.0 + 0.15*float64(level)
	}
	return 1.75 + 0.10*float64(level-5)
}

// upgradePowerMult: +10%/level for L1-5, +5%/level for L6-10.
// Late levels hurt less so high-tier upgrades aren't pure power-scaling.
func upgradePowerMult(level int) float64 {
	if level <= 5 {
		return 1.0 + 0.10*float64(level)
	}
	return 1.50 + 0.05*float64(level-5)
}

// upgradeHeatMult: +20%/level for L1-5, +10%/level for L6-10. Same idea.
func upgradeHeatMult(level int) float64 {
	if level <= 5 {
		return 1.0 + 0.20*float64(level)
	}
	return 2.00 + 0.10*float64(level-5)
}

// UpgradeFragsForLevel returns the research-fragment cost to advance FROM
// the given level (i.e. to reach level+1). Levels 1-5 are free (money
// only); L5→L10 ramp 3/5/8/12/20 — softer than first draft so frags
// remain a consistent earn rather than a grind gate.
func UpgradeFragsForLevel(currentLevel int) int {
	switch currentLevel {
	case 5:
		return 3
	case 6:
		return 5
	case 7:
		return 8
	case 8:
		return 12
	case 9:
		return 20
	}
	return 0
}

// UpgradeMoneyMult bumps the money cost for high-level upgrades — late
// levels are a serious investment.
func UpgradeMoneyMult(currentLevel int) float64 {
	if currentLevel <= 4 {
		return 1.0
	}
	// L5→6 = 1.3x, L6→7 = 1.6x, ..., L9→10 = 2.5x base
	return 1.0 + 0.30*float64(currentLevel-4)
}

// UpgradeGPU upgrades a GPU instance one level. Levels 1-5 cost money only;
// levels 6-10 also consume research fragments. Fail chance ramps with level.
func (s *State) UpgradeGPU(instanceID int) error {
	for _, g := range s.GPUs {
		if g.InstanceID != instanceID {
			continue
		}
		if g.UpgradeLevel >= MaxGPUUpgradeLevel {
			return fmt.Errorf("already at max level")
		}
		var price int
		if def, ok := data.GPUByID(g.DefID); ok {
			price = def.Price
		} else {
			price = 3000 // MEOWCore default upgrade base
		}
		cost := int(float64(price) / 3.0 * UpgradeMoneyMult(g.UpgradeLevel))
		fragsNeeded := UpgradeFragsForLevel(g.UpgradeLevel)
		if s.BTC < float64(cost) {
			return fmt.Errorf("need %s, have %s", FmtBTCInt(cost), FmtBTC(s.BTC))
		}
		if fragsNeeded > 0 && s.ResearchFrags < fragsNeeded {
			return fmt.Errorf("need %d research fragments, have %d", fragsNeeded, s.ResearchFrags)
		}
		s.BTC -= float64(cost)
		s.ResearchFrags -= fragsNeeded
		failChance := 0.05 + 0.05*float64(g.UpgradeLevel)
		if failChance > 0.45 {
			failChance = 0.45 // cap at 45% so L9→10 isn't a coin flip
		}
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
		return fmt.Errorf("need %s", FmtBTCInt(EmergencyVentCost))
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
	s.appendLog("info", i18n.T("log.room.vent", FmtBTCInt(EmergencyVentCost)))
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

// TriggerPumpDump invokes the Hacker "Pump & Dump" ability. Default 30-min
// cooldown; Pump & Dump II cuts that in half.
func (s *State) TriggerPumpDump() error {
	if !s.HasUnlock("pump_dump_action") {
		return fmt.Errorf("requires Pump & Dump skill")
	}
	cooldown := int64(1800)
	if s.HasSkill("pump_dump_ii") {
		cooldown = 900
	}
	last := s.EventCooldown["pump_dump"]
	now := time.Now().Unix()
	if now-last < cooldown {
		return fmt.Errorf("on cooldown for %d more minutes", (cooldown-(now-last))/60)
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

// MaxDefenseLevel is the cap for any defense dimension. Levels 1-3 are
// money-only; levels 4-8 also burn research fragments.
const MaxDefenseLevel = 8

// DefenseFragsForLevel returns how many fragments are needed to advance
// FROM the given level. 0 for L1-3, then ramps 2/4/6/8/10 for L4-8.
func DefenseFragsForLevel(currentLevel int) int {
	if currentLevel < 3 {
		return 0
	}
	return (currentLevel - 2) * 2
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
	if *lvl >= MaxDefenseLevel {
		return fmt.Errorf("%s already maxed", label)
	}
	cost := (*lvl + 1) * 250
	fragsNeeded := DefenseFragsForLevel(*lvl)
	if s.BTC < float64(cost) {
		return fmt.Errorf("need %s, have %s", FmtBTCInt(cost), FmtBTC(s.BTC))
	}
	if fragsNeeded > 0 && s.ResearchFrags < fragsNeeded {
		return fmt.Errorf("need %d research fragments, have %d", fragsNeeded, s.ResearchFrags)
	}
	s.BTC -= float64(cost)
	s.ResearchFrags -= fragsNeeded
	*lvl++
	s.appendLog("info", i18n.T("log.defense.upgraded", label, *lvl))
	return nil
}
