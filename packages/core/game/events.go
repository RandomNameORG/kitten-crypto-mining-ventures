package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"
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
		// Sprint 6 §10.1 — PSU/pool/BTC-linked event family. Gates in
		// eventGatePasses keep these out of the fresh-game baseline.
		"psu_explode", "psu_smoking", "mining_disaster", "pool_runaway",
		"solo_block_hit", "psu_chain_explode", "share_dilution", "fire_sale",
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
	baseFire *= s.DifficultyEventFreqMult()
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
	case "psu_explode":
		// E21: a non-builtin PSU in the current room is past its
		// overload tolerance band — same condition the silent
		// per-tick path watches, surfaced as an event banner here.
		return s.roomHasOverloadedNonBuiltinPSU(s.CurrentRoom)
	case "psu_smoking":
		// E22: a psu_trash in current room older than 5h.
		return s.roomHasOldTrashPSU(s.CurrentRoom, 18000)
	case "mining_disaster":
		// E23: late-game crash event — pinned behind LE so the 1h
		// fresh-game baseline can't trigger it even in a deep dip.
		return s.MarketPrice < 0.4 && s.LifetimeEarned > 1000
	case "pool_runaway":
		// E24: only the rug-pull pool can rug-pull.
		return s.PoolID == "whisker_fi"
	case "solo_block_hit":
		// E25: only matters when actually solo mining.
		return s.PoolID == "solo"
	case "psu_chain_explode":
		// E26: ≥2 running psu_trash AND room past 1.0× capacity.
		return s.roomTrashPSUCount(s.CurrentRoom) >= 2 &&
			s.RoomPSUOverloadFactor(s.CurrentRoom) > 1.0
	case "share_dilution":
		// E27: PPLNS / PPS+ pools only, late-game gated.
		mode := s.PoolSettlementMode()
		if mode != "pplns" && mode != "pps_plus" {
			return false
		}
		if s.LifetimeEarned <= 1000 {
			return false
		}
		// Don't pile on while the player is mid-pool-switch — the
		// transition pause already blocks earnings.
		return s.PoolSwitchAt == 0
	case "fire_sale":
		// E28: the buyer's market only opens in the wake of a recent
		// mining_disaster. Use s.LastTickUnix as the time anchor (sim
		// determinism) and require an actual cooldown entry — a fresh
		// game's zero-valued map never trips this.
		last, ok := s.EventCooldown["mining_disaster"]
		if !ok {
			return false
		}
		return s.LastTickUnix-last < 600
	}
	return true
}

func (s *State) applyEvent(e data.EventDef) {
	s.appendLog(e.Category, fmt.Sprintf("%s %s — %s", e.Emoji, e.LocalName(), e.LocalText()))
	if s.EventsByCategory == nil {
		s.EventsByCategory = map[string]int{}
	}
	s.EventsByCategory[e.Category]++
	if e.ID == "market_crash" {
		s.MarketCrashCount++
	}
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
					// No free slot → auto-scrap. Gas (§11.2) applies the
					// same way as a manual SellGPU so the player can't
					// dodge cashout fees by waiting on a no-slot gift.
					gross := float64(def.ScrapValue) * s.ScrapValueMult()
					net := gross - s.GasFeeFor(gross)
					if net < 0 {
						net = 0
					}
					s.BTC += net
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
				s.grantAchievement("tax_survivor")
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
		case "psu_explode":
			// E21 §10.1 — pop the weakest non-builtin PSU in the
			// current room (the same "blow-first" candidate the per-
			// tick overload path picks). Reuses psuExplode for the
			// damage roll so the bricked-GPU cadence stays uniform.
			if id, ok := s.weakestNonBuiltinRunningPSU(s.CurrentRoom); ok {
				s.psuExplode(s.CurrentRoom, id)
			} else {
				s.appendLog("info", "PSU jolt — but no real PSU to fail. Lucky.")
			}
		case "psu_smoking_chain":
			// E22 §10.1 — 5% chance the smoking PSU also explodes.
			// The earn_mult portion is wired via the catalog effect
			// list; this branch only handles the chain-explosion
			// roll so the dice stay together.
			if rand.Float64() < 0.05 {
				if id, ok := s.weakestNonBuiltinRunningPSU(s.CurrentRoom); ok {
					s.psuExplode(s.CurrentRoom, id)
				}
			}
		case "pool_runaway":
			// E24 §10.1 — the pool rugged. Unsettled PPLNS shares
			// vanish and we snap the player back to scratch_pool
			// with no transition window (no time to settle when the
			// pool literally vanished).
			s.PoolShares = 0
			s.PoolID = "scratch_pool"
			s.PoolSwitchFrom = ""
			s.PoolSwitchAt = 0
			s.appendLog("crisis", "Pool runaway! Unsettled shares lost — back to ScratchPool.")
		case "solo_block_hit":
			// E25 §10.1 — solo lottery hit. Effect Amount × 1000
			// → 0.5 spec amount → ₿500 lump-sum payout.
			reward := eff.Amount * 1000
			s.BTC += reward
			s.appendLog("opportunity", fmt.Sprintf("Solo block hit! +%s lump sum.", FmtBTC(reward)))
		case "psu_chain_explode":
			// E26 §10.1 — every running psu_trash in the room
			// breaks; half (rounded down, min 1 if any GPUs were
			// running) of running GPUs in the room go down too.
			rs := s.Rooms[s.CurrentRoom]
			if rs == nil {
				break
			}
			brokenPSUs := 0
			for _, p := range rs.PSUUnits {
				if p.Status != "running" {
					continue
				}
				if p.DefID == "psu_trash" {
					p.Status = "broken"
					brokenPSUs++
				}
			}
			candidates := []*GPU{}
			for _, g := range s.GPUs {
				if g.Room == s.CurrentRoom && g.Status == "running" {
					candidates = append(candidates, g)
				}
			}
			toBreak := len(candidates) / 2
			if toBreak == 0 && len(candidates) > 0 {
				toBreak = 1
			}
			brokenGPUs := 0
			for i := 0; i < toBreak && len(candidates) > 0; i++ {
				idx := rand.Intn(len(candidates))
				victim := candidates[idx]
				victim.Status = "broken"
				victim.HoursLeft = 0
				candidates = append(candidates[:idx], candidates[idx+1:]...)
				brokenGPUs++
			}
			s.appendLog("crisis", fmt.Sprintf("PSU chain explosion — %d trash PSU(s) and %d GPU(s) bricked.", brokenPSUs, brokenGPUs))
		case "fire_sale":
			// E28 §10.1 — opportunity log only.
			// future sprint: wire shop-discount modifier here.
			s.appendLog("opportunity", "Fire sale on used hardware — bargains everywhere.")
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

// roomHasOverloadedNonBuiltinPSU returns true when the room is past its
// weakest PSU's safe band AND that weakest PSU isn't the freebie psu_builtin.
// Mirrors the silent overload roll's gating condition (advancePSUOverload)
// so E21 can ride the same surface as a player-visible event.
func (s *State) roomHasOverloadedNonBuiltinPSU(roomID string) bool {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return false
	}
	hasNonBuiltin := false
	for _, p := range rs.PSUUnits {
		if p.Status != "running" {
			continue
		}
		if p.DefID != "psu_builtin" {
			hasNonBuiltin = true
			break
		}
	}
	if !hasNonBuiltin {
		return false
	}
	tol := s.roomMinOverloadTolerance(roomID)
	return s.RoomPSUOverloadFactor(roomID) > 1.0+tol
}

// roomHasOldTrashPSU returns true when any running psu_trash in the room
// has been installed for at least olderThanSec seconds (anchored on
// s.LastTickUnix to keep the headless sim deterministic). Empty / freshly
// installed rooms naturally return false.
func (s *State) roomHasOldTrashPSU(roomID string, olderThanSec int64) bool {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return false
	}
	for _, p := range rs.PSUUnits {
		if p.Status != "running" || p.DefID != "psu_trash" {
			continue
		}
		if s.LastTickUnix-p.InstalledAt > olderThanSec {
			return true
		}
	}
	return false
}

// roomTrashPSUCount counts running psu_trash instances in the room.
func (s *State) roomTrashPSUCount(roomID string) int {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return 0
	}
	n := 0
	for _, p := range rs.PSUUnits {
		if p.Status == "running" && p.DefID == "psu_trash" {
			n++
		}
	}
	return n
}

// weakestNonBuiltinRunningPSU picks the lowest-tolerance non-builtin running
// PSU instance — the canonical "blow-first" candidate the random-roll
// PSU-explosion events should target. Returns (instanceID, true) on hit;
// (0, false) when no eligible PSU exists in the room.
func (s *State) weakestNonBuiltinRunningPSU(roomID string) (int, bool) {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return 0, false
	}
	pickID := 0
	pickTol := -1.0
	for _, p := range rs.PSUUnits {
		if p.Status != "running" || p.DefID == "psu_builtin" {
			continue
		}
		def, ok := data.PSUByID(p.DefID)
		if !ok {
			continue
		}
		if pickTol < 0 || def.OverloadTolerance < pickTol {
			pickTol = def.OverloadTolerance
			pickID = p.InstanceID
		}
	}
	if pickID == 0 {
		return 0, false
	}
	return pickID, true
}

// RepairAllBroken repairs every broken GPU the player can afford in cost
// order (cheapest first), so a partial budget at least gets some cards
// back online. Returns the number of GPUs repaired and the total cost.
// Honors the PCB Surgery free-repair effect.
func (s *State) RepairAllBroken() (int, int) {
	type candidate struct {
		id   int
		cost int
	}
	cands := []candidate{}
	for _, g := range s.GPUs {
		if g.Status != "broken" {
			continue
		}
		price := 3000
		if def, ok := data.GPUByID(g.DefID); ok {
			price = def.Price
		}
		baseCost := price * 3 / 10
		cost := int(float64(baseCost) * s.RepairCostMult())
		cands = append(cands, candidate{id: g.InstanceID, cost: cost})
	}
	if len(cands) == 0 {
		return 0, 0
	}
	// Sort cheapest first so a constrained budget covers as many as possible.
	for i := range cands {
		for j := i + 1; j < len(cands); j++ {
			if cands[j].cost < cands[i].cost {
				cands[i], cands[j] = cands[j], cands[i]
			}
		}
	}
	repaired := 0
	totalCost := 0
	for _, c := range cands {
		if err := s.RepairGPU(c.id); err != nil {
			break // out of money mid-loop — stop cleanly
		}
		repaired++
		totalCost += c.cost
	}
	return repaired, totalCost
}

// RepairGPU repairs a broken GPU. PCB Surgery cuts the cost in half.
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
		baseCost := price * 3 / 10
		cost := int(float64(baseCost) * s.RepairCostMult())
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
