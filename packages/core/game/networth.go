package game

import (
	"math"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
)

// Gas-fee + NetWorth constants. See docs/GAME_DESIGN.md §8 / §11.2.
const (
	// BaseGasRate is the per-cashout proportional fee at zero congestion
	// (§11.2: 0.5%). Effective rate scales with NetworkCongestion via
	// EffectiveGasFeeRate.
	BaseGasRate = 0.005
	// GasFlatFloor is the flat surcharge applied to every cashout in
	// addition to the proportional rate (§11.2: $5).
	GasFlatFloor = 5.0
	// PSUNetWorthMult is the paper-resale factor applied to a PSU's
	// catalog Price when it's counted as an asset on the NetWorth panel.
	// §8.4: 70% of original — PSUs are durable but secondhand-unfriendly.
	PSUNetWorthMult = 0.7

	// congestionMin / congestionMax bound the NetworkCongestion drift so
	// cashouts always have a finite worst-case fee. See advanceCongestion.
	congestionMin = 0.05
	congestionMax = 0.95
	// congestionPeriodSec is the period of the deterministic sin-wave
	// drift — ~30 minutes so the congestion bar feels alive without
	// asking the player to re-plan every minute.
	congestionPeriodSec = 1800
)

// advanceCongestion updates NetworkCongestion via a deterministic sin wave
// mapped to [congestionMin, congestionMax]. RNG-FREE on purpose: any
// rand.* call here would shift the global sequence and break the
// byte-for-byte sim determinism that events_test.go relies on.
func (s *State) advanceCongestion(now int64) {
	phase := 2.0 * math.Pi * float64(now) / float64(congestionPeriodSec)
	// sin -> [-1, 1]; map to [congestionMin, congestionMax]
	mid := (congestionMin + congestionMax) / 2.0
	half := (congestionMax - congestionMin) / 2.0
	s.NetworkCongestion = mid + half*math.Sin(phase)
	s.LastCongestionTickUnix = now
}

// EffectiveGasFeeRate returns the proportional gas rate applied to a
// cashout's gross value, scaled by the current network congestion.
// At neutral congestion (0) → BaseGasRate; at full congestion (1.0) → 2×.
func (s *State) EffectiveGasFeeRate() float64 {
	return BaseGasRate * (1.0 + s.NetworkCongestion)
}

// GasFeeFor returns the absolute gas fee charged on a cashout of `gross`.
// Clamps to ≤ gross so a tiny sell can never drive the player into negative
// BTC — the player just nets zero on dust.
func (s *State) GasFeeFor(gross float64) float64 {
	if gross <= 0 {
		return 0
	}
	fee := gross*s.EffectiveGasFeeRate() + GasFlatFloor
	if fee > gross {
		fee = gross
	}
	return fee
}

// GPUResalePrice computes a GPU's BTC-linked secondhand value (§8.1).
// Catalog GPUs use the data-driven BaseResaleRatio + BtcSensitivity; a
// player-built MEOWCore inherits a tier-scaled inherent base with
// modest BTC sensitivity since custom hardware tracks the market loosely.
func (s *State) GPUResalePrice(g *GPU) float64 {
	if g == nil {
		return 0
	}
	if g.BlueprintID != "" {
		tier := s.blueprintTier(g.BlueprintID)
		base := 2000.0 + float64(tier-1)*2000.0
		sens := 0.2 + 0.05*float64(tier-1) - s.BtcSensitivityBonus()
		if sens < 0 {
			sens = 0
		}
		// MEOWCore base_resale_ratio is implicitly 1.0 — the inherent
		// base already accounts for handcrafted-asset pricing.
		return base * (1.0 + (s.MarketPrice-1.0)*sens)
	}
	def, ok := data.GPUByID(g.DefID)
	if !ok {
		return 0
	}
	sens := def.BtcSensitivity - s.BtcSensitivityBonus()
	if sens < 0 {
		sens = 0
	}
	mult := 1.0 + (s.MarketPrice-1.0)*sens
	return float64(def.Price) * def.BaseResaleRatio * mult
}

// NetWorth sums the player's liquid + paper assets for the dashboard
// (§8.4). Currency is unified (BTC = cash), so any "btc_held × price" term
// from older specs collapses to zero. PSUs count at PSUNetWorthMult of
// catalog price; the builtin (price 0) contributes nothing. Debt is wired
// in concept for a future loan system but currently always 0.
func (s *State) NetWorth() float64 {
	total := s.BTC
	for _, g := range s.GPUs {
		if g == nil || g.Status == "stolen" {
			continue
		}
		total += s.GPUResalePrice(g)
	}
	for _, room := range s.Rooms {
		if room == nil {
			continue
		}
		for _, p := range room.PSUUnits {
			if p == nil || p.Status != "running" {
				continue
			}
			def, ok := data.PSUByID(p.DefID)
			if !ok || def.Price == 0 {
				continue
			}
			total += float64(def.Price) * PSUNetWorthMult
		}
	}
	// Future: subtract outstanding loan principal here. No debt system today.
	return total
}
