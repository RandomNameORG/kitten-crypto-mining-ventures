package game

import (
	"math"
	"time"
)

// BTCBasePrice is the centre of the random walk.
const BTCBasePrice = 30000.0

// ElectricPerVoltMin is the per-V per-minute price of electricity. Rooms
// multiply this by their electric_cost_mult.
const ElectricPerVoltMin = 0.10

// BTCPriceAt computes the BTC price at a given unix second using a seeded
// double-sine oscillator + any active event multipliers. Deterministic per
// seed+time so the UI can render a graph without storing history.
func (s *State) BTCPriceAt(t int64) float64 {
	// Long wave: ~10 min period, ±10%
	long := math.Sin(float64(t-s.StartedUnix)*2.0*math.Pi/600.0) * 0.10
	// Short wave: ~30 sec period, ±3%
	short := math.Sin(float64(t-s.StartedUnix)*2.0*math.Pi/30.0) * 0.03
	// Seeded medium wave (personalises the curve per save).
	mseed := float64(s.BTCPriceSeed%1_000_000) / 1_000_000.0
	medium := math.Sin(float64(t-s.StartedUnix)*2.0*math.Pi/180.0+mseed*2*math.Pi) * 0.05
	price := BTCBasePrice * (1.0 + long + short + medium)
	// News multiplier.
	for _, m := range s.Modifiers {
		if m.Kind == "btc_mult" && m.ExpiresAt > t {
			price *= m.Factor
		}
	}
	if price < 500 {
		price = 500
	}
	return price
}

// CurrentBTCPrice is convenience for BTCPriceAt(now).
func (s *State) CurrentBTCPrice() float64 {
	return s.BTCPriceAt(time.Now().Unix())
}

// earnMultiplier returns the aggregate positive-production multiplier from
// active modifiers.
func (s *State) earnMultiplier(now int64) float64 {
	mult := 1.0
	for _, m := range s.Modifiers {
		if m.Kind == "earn_mult" && m.ExpiresAt > now {
			mult *= m.Factor
		}
	}
	return mult
}

// IsMiningPaused returns true if a pause_mining modifier is active.
func (s *State) IsMiningPaused(now int64) bool {
	for _, m := range s.Modifiers {
		if m.Kind == "pause_mining" && m.ExpiresAt > now {
			return true
		}
	}
	return false
}

// pruneModifiers removes expired modifiers.
func (s *State) pruneModifiers(now int64) {
	alive := s.Modifiers[:0]
	for _, m := range s.Modifiers {
		if m.ExpiresAt > now {
			alive = append(alive, m)
		}
	}
	s.Modifiers = alive
}
