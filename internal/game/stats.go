package game

import (
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
)

// This file exposes read-only, point-in-time rate computations that mirror
// the tick loop's math. Used by the UI to render live "bill/s", "net/s",
// heat trend, etc. without inventing numbers that would drift from what
// the simulation actually does.

// EarnRatePerSec returns the dollar earn rate for a single running GPU at
// the current BTC price + active modifiers + active difficulty.
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
	btcPerSec := eff * earnMult * effFactor * s.DifficultyEarnMult()
	return btcPerSec * s.BTCPriceAt(now)
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

// RoomHeatDeltaPerSec is the net heat change per second: GPU output minus
// passive cooling (room base × cooling-upgrade bonus).
func (s *State) RoomHeatDeltaPerSec(roomID string) float64 {
	roomDef, ok := data.RoomByID(roomID)
	if !ok {
		return 0
	}
	room := s.Rooms[roomID]
	if room == nil {
		return 0
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
	// Cap at floor temp (20°C) so trend matches sim (Heat clamps low).
	if room.Heat <= 20 && delta < 0 {
		delta = 0
	}
	return delta
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
