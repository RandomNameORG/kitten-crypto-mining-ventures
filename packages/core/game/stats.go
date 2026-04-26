package game

import (
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
)

// This file exposes read-only, point-in-time rate computations that mirror
// the tick loop's math. Used by the UI to render live "bill/s", "net/s",
// heat trend, etc. without inventing numbers that would drift from what
// the simulation actually does.

// GPUEarnRatePerSec returns the BTC-per-second rate for a single running GPU
// under the current active modifiers + difficulty + room heat.
func (s *State) GPUEarnRatePerSec(g *GPU) float64 {
	if g.Status != "running" {
		return 0
	}
	now := time.Now().Unix()
	if s.IsMiningPaused(now) {
		return 0
	}
	eff, _, _, _ := s.GPUStats(g)
	effFactor := 1.0
	if room := s.Rooms[g.Room]; room != nil && room.Heat > 0.8*room.MaxHeat {
		effFactor = 0.5
	}
	earnMult := s.earnMultiplier(now)
	rate := eff * earnMult * effFactor * s.DifficultyEarnMult() * s.MarketPrice * MiningScale
	// Display-only haircut: the tick loop actually diverts SyndicateCutRate
	// of each GPU's earn into the contribution pool. Mirror that here so
	// the dashboard's "earn +₿/s" matches what the player sees hit BTC.
	if s.SyndicateJoined {
		rate *= (1.0 - SyndicateCutRate)
	}
	return rate
}

// RoomEarnRatePerSec is the sum of per-GPU earn rates for every running GPU
// in the given room.
func (s *State) RoomEarnRatePerSec(roomID string) float64 {
	var total float64
	for _, g := range s.GPUs {
		if g.Room != roomID {
			continue
		}
		total += s.GPUEarnRatePerSec(g)
	}
	return total
}

// RoomBillRatePerSec is the per-second cost of electricity + rent for the
// given room, respecting skill + difficulty multipliers.
func (s *State) RoomBillRatePerSec(roomID string) float64 {
	roomDef, ok := data.RoomByID(roomID)
	if !ok {
		return 0
	}
	billMult := s.BillMult() * s.DifficultyBillMult()

	var volt float64
	for _, g := range s.GPUs {
		if g.Room != roomID || g.Status != "running" {
			continue
		}
		_, pow, _, _ := s.GPUStats(g)
		volt += pow
	}
	// ElectricPerVoltMin is per minute — divide by 60 for per-second.
	elec := volt * ElectricPerVoltMin * roomDef.ElectricCostMult * billMult / 60.0
	rent := float64(roomDef.RentPerHour) * s.DifficultyBillMult() / 3600.0
	return elec + rent
}

// RoomHeatDeltaPerTick returns (deltaPerTick, tickSec) — the degrees applied
// at each heat tick, plus the room's configured tick interval. The
// dashboard uses this to show "⚡ +2.5 /15s" style hints.
func (s *State) RoomHeatDeltaPerTick(roomID string) (float64, int) {
	roomDef, ok := data.RoomByID(roomID)
	if !ok {
		return 0, 0
	}
	room := s.Rooms[roomID]
	if room == nil {
		return 0, 0
	}
	coolingBonus := 1.0 + 0.25*float64(room.CoolingLvl)
	var heatIn float64
	for _, g := range s.GPUs {
		if g.Room != roomID || g.Status != "running" {
			continue
		}
		_, _, hOut, _ := s.GPUStats(g)
		heatIn += hOut
	}
	delta := heatIn - roomDef.BaseCooling*coolingBonus
	if room.Heat <= 20 && delta < 0 {
		delta = 0
	}
	tickSec := roomDef.HeatTickSec
	if tickSec <= 0 {
		tickSec = 10
	}
	return delta, tickSec
}

// SecondsUntilNextHeatTick reports how long until the current room's heat
// updates next — used to render a countdown on the dashboard.
func (s *State) SecondsUntilNextHeatTick(roomID string) int {
	roomDef, ok := data.RoomByID(roomID)
	if !ok {
		return 0
	}
	room := s.Rooms[roomID]
	if room == nil {
		return 0
	}
	tickSec := roomDef.HeatTickSec
	if tickSec <= 0 {
		tickSec = 10
	}
	if room.LastHeatTickUnix == 0 {
		return tickSec
	}
	remaining := int64(tickSec) - (time.Now().Unix() - room.LastHeatTickUnix)
	if remaining < 0 {
		return 0
	}
	if remaining > int64(tickSec) {
		return tickSec
	}
	return int(remaining)
}

// SecondsUntilNextBill is the countdown (in whole seconds, 0..60) until
// the next billing cycle fires.
func (s *State) SecondsUntilNextBill() int {
	remaining := 60 - (time.Now().Unix() - s.LastBillUnix)
	if remaining < 0 {
		return 0
	}
	if remaining > 60 {
		return 60
	}
	return int(remaining)
}

// NetRatePerSec is the room's earn rate minus its bill rate. Positive is good.
func (s *State) RoomNetRatePerSec(roomID string) float64 {
	return s.RoomEarnRatePerSec(roomID) - s.RoomBillRatePerSec(roomID)
}
