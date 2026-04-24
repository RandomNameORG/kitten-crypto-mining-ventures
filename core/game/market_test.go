package game

import (
	"encoding/json"
	"math"
	"testing"
)

// advanceMarketTick drives advanceMarket forward by stepping LastTickUnix
// and calling the internal helper directly. Lets us test drift without
// booting the whole Tick pipeline (which also earns, spends, etc.).
func advanceMarketSteps(s *State, startUnix int64, steps int) int64 {
	now := startUnix
	for i := 0; i < steps; i++ {
		now += MarketTickSec
		s.advanceMarket(now)
	}
	return now
}

func TestMarketStartsAtOne(t *testing.T) {
	withTempHome(t)
	s := NewState("market-test")
	if s.MarketPrice != 1.0 {
		t.Errorf("MarketPrice = %v, want 1.0", s.MarketPrice)
	}
	if s.PrevMarketPrice != 1.0 {
		t.Errorf("PrevMarketPrice = %v, want 1.0", s.PrevMarketPrice)
	}
}

func TestMarketDriftsAfterManyTicks(t *testing.T) {
	withTempHome(t)
	SeedRNG(7)
	s := NewState("market-test")
	s.LastMarketTickUnix = simTestBaseUnix
	advanceMarketSteps(s, simTestBaseUnix, 120)
	// After 120 discrete steps the walk should have moved off 1.0. A Gaussian
	// kick with sigma 0.03 and 120 samples practically never stays exactly at
	// 1.0; this guards against "drift isn't firing at all" regressions.
	if s.MarketPrice == 1.0 {
		t.Errorf("MarketPrice stuck at 1.0 after 120 drift steps — advanceMarket not running?")
	}
}

func TestMarketStaysBounded(t *testing.T) {
	withTempHome(t)
	// Sweep several seeds so we're confident the clamp holds across noise
	// realisations, not just a lucky one.
	for _, seed := range []int64{1, 2, 3, 42, 99} {
		SeedRNG(seed)
		s := NewState("market-test")
		s.LastMarketTickUnix = simTestBaseUnix
		now := simTestBaseUnix
		for i := 0; i < 10000; i++ {
			now += MarketTickSec
			s.advanceMarket(now)
			if math.IsNaN(s.MarketPrice) || math.IsInf(s.MarketPrice, 0) {
				t.Fatalf("seed=%d step=%d: MarketPrice non-finite: %v", seed, i, s.MarketPrice)
			}
			if s.MarketPrice < marketPriceMin || s.MarketPrice > marketPriceMax {
				t.Fatalf("seed=%d step=%d: MarketPrice %.4f out of [%.2f, %.2f]", seed, i, s.MarketPrice, marketPriceMin, marketPriceMax)
			}
		}
	}
}

func TestMarketSaveLoadRoundTrip(t *testing.T) {
	withTempHome(t)
	SeedRNG(5)
	s := NewState("market-test")
	s.LastMarketTickUnix = simTestBaseUnix
	advanceMarketSteps(s, simTestBaseUnix, 50)
	wantPrice := s.MarketPrice
	wantPrev := s.PrevMarketPrice
	wantAnchor := s.LastMarketTickUnix

	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	loaded, err := LoadFrom(b)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if loaded.MarketPrice != wantPrice {
		t.Errorf("MarketPrice round-trip drift: got %v, want %v", loaded.MarketPrice, wantPrice)
	}
	if loaded.PrevMarketPrice != wantPrev {
		t.Errorf("PrevMarketPrice round-trip drift: got %v, want %v", loaded.PrevMarketPrice, wantPrev)
	}
	if loaded.LastMarketTickUnix != wantAnchor {
		t.Errorf("LastMarketTickUnix round-trip drift: got %v, want %v", loaded.LastMarketTickUnix, wantAnchor)
	}
}

func TestMarketLegacySaveBackfills(t *testing.T) {
	withTempHome(t)
	// JSON that predates the market fields entirely — missing keys should
	// decode to zero values, then ensureInit should lift them back to 1.0.
	legacy := []byte(`{
		"version": 1,
		"kitten_name": "Legacy",
		"btc": 100.0,
		"difficulty": "normal"
	}`)
	s, err := LoadFrom(legacy)
	if err != nil {
		t.Fatalf("LoadFrom legacy: %v", err)
	}
	if s.MarketPrice != 1.0 {
		t.Errorf("legacy MarketPrice = %v, want 1.0", s.MarketPrice)
	}
	if s.PrevMarketPrice != 1.0 {
		t.Errorf("legacy PrevMarketPrice = %v, want 1.0", s.PrevMarketPrice)
	}
}

func TestMarketSameSeedSameTrajectory(t *testing.T) {
	withTempHome(t)
	run := func() []float64 {
		SeedRNG(13)
		s := NewState("market-test")
		s.LastMarketTickUnix = simTestBaseUnix
		series := make([]float64, 0, 1000)
		now := simTestBaseUnix
		for i := 0; i < 1000; i++ {
			now += MarketTickSec
			s.advanceMarket(now)
			series = append(series, s.MarketPrice)
		}
		return series
	}
	a := run()
	withTempHome(t)
	b := run()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("trajectory diverged at step %d: %v vs %v — RNG is leaking non-determinism", i, a[i], b[i])
		}
	}
}
