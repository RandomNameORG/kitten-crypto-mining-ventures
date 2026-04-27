package game

import (
	"math"
	"testing"
)

// floatEq tolerates floating-point fuzz for assertions on derived values.
const floatEq = 1e-6

func nearly(a, b float64) bool { return math.Abs(a-b) <= floatEq }

func TestGasFeeBasicMath(t *testing.T) {
	withTempHome(t)
	s := NewState("gas")
	s.NetworkCongestion = 0.0
	got := s.GasFeeFor(1000)
	want := 1000.0*BaseGasRate + GasFlatFloor // 5 + 5 = 10
	if !nearly(got, want) {
		t.Fatalf("GasFeeFor(1000) at cong=0 = %v, want %v", got, want)
	}
	if rate := s.EffectiveGasFeeRate(); !nearly(rate, BaseGasRate) {
		t.Fatalf("EffectiveGasFeeRate at cong=0 = %v, want %v", rate, BaseGasRate)
	}
}

func TestGasFeeScalesWithCongestion(t *testing.T) {
	withTempHome(t)
	s := NewState("gas")
	s.NetworkCongestion = 1.0
	if rate := s.EffectiveGasFeeRate(); !nearly(rate, 2*BaseGasRate) {
		t.Fatalf("EffectiveGasFeeRate at cong=1 = %v, want %v", rate, 2*BaseGasRate)
	}
	got := s.GasFeeFor(1000)
	want := 1000.0*0.01 + GasFlatFloor // 10 + 5 = 15
	if !nearly(got, want) {
		t.Fatalf("GasFeeFor(1000) at cong=1 = %v, want %v", got, want)
	}
}

func TestGasFeeClampsAtGross(t *testing.T) {
	withTempHome(t)
	s := NewState("gas")
	s.NetworkCongestion = 0.0
	// Tiny sell: raw fee (1*0.005 + 5 = 5.005) far exceeds gross — must clamp.
	got := s.GasFeeFor(1)
	if !nearly(got, 1.0) {
		t.Fatalf("GasFeeFor(1) clamp = %v, want 1.0", got)
	}
	// Net (gross - gas) is what SellGPU credits — should be exactly zero.
	if net := 1.0 - got; net != 0 {
		t.Fatalf("net of dust sell = %v, want 0", net)
	}
}

func TestGPUResalePriceAtNeutralMarket(t *testing.T) {
	withTempHome(t)
	s := NewState("nw")
	s.MarketPrice = 1.0
	g := &GPU{InstanceID: 999, DefID: "rtx3080", Status: "running"}
	got := s.GPUResalePrice(g)
	want := 5500.0 * 0.45 // 2475
	if !nearly(got, want) {
		t.Fatalf("rtx3080 resale at neutral = %v, want %v", got, want)
	}
}

func TestGPUResalePriceTracksBTC(t *testing.T) {
	withTempHome(t)
	s := NewState("nw")
	gpu := &GPU{InstanceID: 999, DefID: "rtx3080", Status: "running"}

	s.MarketPrice = 1.0
	neutral := s.GPUResalePrice(gpu)
	s.MarketPrice = 1.5
	bull := s.GPUResalePrice(gpu)
	s.MarketPrice = 0.5
	bear := s.GPUResalePrice(gpu)

	if !(bull > neutral && neutral > bear) {
		t.Fatalf("expected bull > neutral > bear, got %v / %v / %v", bull, neutral, bear)
	}
	// Catalog says rtx3080 sens=0.8. Bull (Δ=+0.5) → +40% over neutral.
	wantBull := neutral * (1.0 + 0.5*0.8)
	if !nearly(bull, wantBull) {
		t.Fatalf("bull resale = %v, want %v", bull, wantBull)
	}

	// MEOWCore tier 1: sens = 0.2 — barely moves. Build a blueprint and
	// register it so blueprintTier resolves.
	bp := &Blueprint{ID: "bp_test", Tier: 1, Boosts: nil}
	s.Blueprints = append(s.Blueprints, bp)
	core := &GPU{InstanceID: 1000, DefID: "meowcore_v1", Status: "running", BlueprintID: "bp_test"}

	s.MarketPrice = 1.0
	coreNeutral := s.GPUResalePrice(core)
	s.MarketPrice = 1.5
	coreBull := s.GPUResalePrice(core)

	if coreNeutral != 2000.0 {
		t.Fatalf("MEOWCore tier-1 neutral resale = %v, want 2000", coreNeutral)
	}
	wantCoreBull := 2000.0 * (1.0 + 0.5*0.2) // 2200
	if !nearly(coreBull, wantCoreBull) {
		t.Fatalf("MEOWCore tier-1 bull resale = %v, want %v", coreBull, wantCoreBull)
	}
	// Sanity: catalog GPU swing dwarfs MEOWCore swing at the same Δprice.
	catalogSwing := bull - neutral
	coreSwing := coreBull - coreNeutral
	if catalogSwing/neutral <= coreSwing/coreNeutral {
		t.Fatalf("expected MEOWCore to swing less proportionally; catalog %v/%v vs core %v/%v",
			catalogSwing, neutral, coreSwing, coreNeutral)
	}
}

func TestNetWorthIncludesAllAssets(t *testing.T) {
	withTempHome(t)
	s := NewState("nw")
	s.MarketPrice = 1.0
	// Wipe the lingering legacy modifiers so resale math is plain.
	s.BTC = 1000
	if err := s.BuyGPU("gtx1060"); err != nil {
		t.Fatalf("buy: %v", err)
	}
	if err := s.InstallPSU(s.CurrentRoom, "psu_silver650"); err != nil {
		t.Fatalf("install psu: %v", err)
	}
	// Expected: cash + 2× resale(GTX 1060) + 0.7 × 200 (psu_silver650).
	// Builtin (price 0) contributes nothing.
	cash := s.BTC
	gpuResale := 120.0 * 0.35 // 42 per card
	want := cash + 2*gpuResale + 0.7*200.0
	got := s.NetWorth()
	if !nearly(got, want) {
		t.Fatalf("NetWorth = %v, want %v (cash=%v)", got, want, cash)
	}
}

func TestSellGPUDeductsGas(t *testing.T) {
	withTempHome(t)
	s := NewState("sell")
	s.MarketPrice = 1.0
	s.NetworkCongestion = 0.0
	// Pin scrap/mastery mults to 1.0 by leaving the fresh state alone — no
	// skills unlocked, no mastery levels purchased.
	starter := s.GPUs[0]
	if starter.DefID != "gtx1060" {
		t.Fatalf("expected starter to be gtx1060, got %s", starter.DefID)
	}
	gross := 120.0 * 0.35 // 42
	gas := gross*BaseGasRate + GasFlatFloor
	if gas > gross {
		gas = gross
	}
	wantNet := gross - gas
	if wantNet < 0 {
		wantNet = 0
	}

	before := s.BTC
	if err := s.SellGPU(starter.InstanceID); err != nil {
		t.Fatalf("sell: %v", err)
	}
	delta := s.BTC - before
	if !nearly(delta, wantNet) {
		t.Fatalf("SellGPU credited %v, want %v (gross=%v gas=%v)", delta, wantNet, gross, gas)
	}
	if gas <= 0 || delta <= 0 {
		t.Fatalf("expected both gas and net positive, got gas=%v delta=%v", gas, delta)
	}
}

func TestCongestionStaysInRange(t *testing.T) {
	withTempHome(t)
	s := NewState("cong")
	// Walk a virtual day in 30s steps — ~2880 samples cover ~48 sin periods.
	for offset := int64(0); offset < 86400; offset += 30 {
		s.advanceCongestion(simTestBaseUnix + offset)
		if s.NetworkCongestion < congestionMin-floatEq {
			t.Fatalf("congestion below min at offset %d: %v", offset, s.NetworkCongestion)
		}
		if s.NetworkCongestion > congestionMax+floatEq {
			t.Fatalf("congestion above max at offset %d: %v", offset, s.NetworkCongestion)
		}
	}
}
