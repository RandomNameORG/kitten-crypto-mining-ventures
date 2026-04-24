package game

import "math/rand"

// MarketTickSec is the cadence at which the BTC market price drifts. The
// price only moves on this interval, so one long Tick advances the price by
// an integer number of discrete steps rather than continuously.
const MarketTickSec int64 = 30

// Market clamps. The drift equation is unbounded in principle, but
// implausibly large swings make balance untunable — clamp the effective
// multiplier to the window below so earnings stay within about 3× of
// nominal in either direction.
const (
	marketPriceMin = 0.3
	marketPriceMax = 3.0
)

// MarketHistoryCap bounds the rolling window of market prices recorded for
// the Stats sparkline. Sized for "recent past" eyeballing, not analytics.
const MarketHistoryCap = 60

// recordMarketPrice appends p to the rolling history slice and trims back to
// MarketHistoryCap. Lazy-inits on nil so bare struct literals stay safe.
func (s *State) recordMarketPrice(p float64) {
	s.MarketPriceHistory = append(s.MarketPriceHistory, p)
	if len(s.MarketPriceHistory) > MarketHistoryCap {
		s.MarketPriceHistory = s.MarketPriceHistory[len(s.MarketPriceHistory)-MarketHistoryCap:]
	}
}

// advanceMarket ticks the BTC market-price multiplier toward 1.0 with a
// small Gaussian kick each step. Mean-reversion keeps the series from
// wandering off; the kick keeps it interesting. Uses the global math/rand
// source so SeedRNG makes the trajectory reproducible.
//
// Called once per Tick. Integer-divides the elapsed virtual time into
// MarketTickSec-sized steps so a long (e.g. offline catch-up) Tick still
// gets the right number of market updates.
func (s *State) advanceMarket(now int64) {
	if s.MarketPrice == 0 {
		// Belt-and-braces — ensureInit already seeds this, but bare struct
		// literals in tests can skip it and we don't want a NaN/zero leak
		// into earn rates.
		s.MarketPrice = 1.0
		s.PrevMarketPrice = 1.0
	}
	if s.LastMarketTickUnix == 0 {
		s.LastMarketTickUnix = now
		return
	}
	// A market_pin modifier (from the market_crash event) holds the price at
	// a fixed floor for its duration. Advance the anchor so no drift backlog
	// accumulates while pinned — mean-reversion resumes naturally from the
	// pinned value once the modifier expires.
	if pinned, factor := s.MarketPinned(now); pinned {
		s.LastMarketTickUnix = now
		s.PrevMarketPrice = factor
		s.MarketPrice = factor
		s.recordMarketPrice(factor)
		return
	}
	elapsed := now - s.LastMarketTickUnix
	if elapsed < MarketTickSec {
		return
	}
	steps := elapsed / MarketTickSec
	s.LastMarketTickUnix += steps * MarketTickSec
	s.PrevMarketPrice = s.MarketPrice
	price := s.MarketPrice
	for i := int64(0); i < steps; i++ {
		price += (1.0-price)*0.02 + rand.NormFloat64()*0.03
		if price < marketPriceMin {
			price = marketPriceMin
		}
		if price > marketPriceMax {
			price = marketPriceMax
		}
	}
	s.MarketPrice = price
	s.recordMarketPrice(price)
}

// MarketPinned reports whether a market_pin modifier is active and, if so,
// the factor (pinned price) it's holding at. The first active pin wins if
// multiple are somehow stacked.
func (s *State) MarketPinned(now int64) (bool, float64) {
	for _, m := range s.Modifiers {
		if m.Kind == "market_pin" && m.ExpiresAt > now {
			return true, m.Factor
		}
	}
	return false, 0
}

// MarketTrend returns +1 / 0 / -1 for up/flat/down vs PrevMarketPrice. The
// dashboard uses this to pick an arrow glyph without leaking the exact
// delta (which the player doesn't care about — the magnitude itself is the
// signal).
func (s *State) MarketTrend() int {
	switch {
	case s.MarketPrice > s.PrevMarketPrice:
		return 1
	case s.MarketPrice < s.PrevMarketPrice:
		return -1
	default:
		return 0
	}
}
