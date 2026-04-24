package data

import "testing"

// TestDifficultyCryptoWinterRegistered guards the Sprint 10 contract: the new
// fourth tier must be discoverable via DifficultyByID with bilingual copy and
// knob values strictly harsher than hard. StarterCash matches hard (100) on
// purpose — pain comes from multipliers, not wallet size.
func TestDifficultyCryptoWinterRegistered(t *testing.T) {
	d := DifficultyByID("crypto_winter")
	if d.ID != "crypto_winter" {
		t.Fatalf("DifficultyByID(crypto_winter) returned ID=%q — registry wiring missing", d.ID)
	}
	if d.LabelEN == "" || d.LabelZH == "" {
		t.Errorf("crypto_winter missing bilingual label: EN=%q ZH=%q", d.LabelEN, d.LabelZH)
	}
	if d.DescEN == "" || d.DescZH == "" {
		t.Errorf("crypto_winter missing bilingual description: EN=%q ZH=%q", d.DescEN, d.DescZH)
	}
	if d.EarnMult >= 0.75 {
		t.Errorf("crypto_winter EarnMult=%v should be harsher than hard (0.75)", d.EarnMult)
	}
	if d.BillMult <= 1.25 {
		t.Errorf("crypto_winter BillMult=%v should be harsher than hard (1.25)", d.BillMult)
	}
	if d.ThreatMult <= 1.5 {
		t.Errorf("crypto_winter ThreatMult=%v should be harsher than hard (1.5)", d.ThreatMult)
	}
	if d.StarterCash != 100 {
		t.Errorf("crypto_winter StarterCash=%v, want 100", d.StarterCash)
	}
	if d.MarketVolatilityMult != 2.0 {
		t.Errorf("crypto_winter MarketVolatilityMult=%v, want 2.0", d.MarketVolatilityMult)
	}
	if d.EventFreqMult != 1.5 {
		t.Errorf("crypto_winter EventFreqMult=%v, want 1.5", d.EventFreqMult)
	}
}

// TestDifficultyDefaultsUnchanged guards against accidental rebalancing of the
// existing tiers when somebody tweaks the new knobs. easy/normal/hard must
// keep MarketVolatilityMult=EventFreqMult=1.0 so pre-Sprint-10 behavior is
// preserved exactly for current saves.
func TestDifficultyDefaultsUnchanged(t *testing.T) {
	for _, id := range []string{"easy", "normal", "hard"} {
		d := DifficultyByID(id)
		if d.MarketVolatilityMult != 1.0 {
			t.Errorf("difficulty %q MarketVolatilityMult=%v, want 1.0 (default must be identity)", id, d.MarketVolatilityMult)
		}
		if d.EventFreqMult != 1.0 {
			t.Errorf("difficulty %q EventFreqMult=%v, want 1.0 (default must be identity)", id, d.EventFreqMult)
		}
	}
}

// TestDifficultiesSliceOrder pins the slice order the splash picker renders.
// Easiest → hardest keeps the UI list intuitive; a reorder should be a
// deliberate, reviewed change.
func TestDifficultiesSliceOrder(t *testing.T) {
	want := []string{"easy", "normal", "hard", "crypto_winter"}
	got := Difficulties()
	if len(got) != len(want) {
		t.Fatalf("Difficulties() len=%d, want %d (%v)", len(got), len(want), want)
	}
	for i, id := range want {
		if got[i].ID != id {
			t.Errorf("Difficulties()[%d].ID = %q, want %q", i, got[i].ID, id)
		}
	}
}
