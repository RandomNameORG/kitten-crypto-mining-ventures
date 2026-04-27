package game

import (
	"encoding/json"
	"math"
	"testing"
)

// approxEqual accepts tiny float drift from multiplication; any real bug
// produces a percentage-point mismatch, not a rounding hiccup.
func approxEqual(a, b, eps float64) bool {
	return math.Abs(a-b) <= eps*math.Max(1.0, math.Abs(b))
}

// TestOCMultipliersScaleGPUStats locks in the owner-decided tradeoff table:
// at OC 1 earn scales 1.25× but power and heat scale 1.40×; at OC 2 earn is
// 1.50× while power/heat are 1.90×. If someone nudges a constant and the
// non-earn factors drop below the earn factor, the feature stops being a
// tradeoff — this test is the guardrail.
func TestOCMultipliersScaleGPUStats(t *testing.T) {
	withTempHome(t)
	s := NewState("oc-math")
	if len(s.GPUs) == 0 {
		t.Fatalf("expected starter GPU")
	}
	g := s.GPUs[0]

	g.OCLevel = 0
	eff0, pow0, heat0, _ := s.GPUStats(g)

	for _, c := range []struct {
		level                  int
		earnK, powK, heatK     float64
	}{
		{1, 1.25, 1.40, 1.40},
		{2, 1.50, 1.90, 1.90},
	} {
		g.OCLevel = c.level
		eff, pow, heat, _ := s.GPUStats(g)
		if !approxEqual(eff, eff0*c.earnK, 1e-9) {
			t.Errorf("OC %d: eff = %v, want %v (×%.2f of base %v)", c.level, eff, eff0*c.earnK, c.earnK, eff0)
		}
		if !approxEqual(pow, pow0*c.powK, 1e-9) {
			t.Errorf("OC %d: pow = %v, want %v (×%.2f of base %v)", c.level, pow, pow0*c.powK, c.powK, pow0)
		}
		if !approxEqual(heat, heat0*c.heatK, 1e-9) {
			t.Errorf("OC %d: heat = %v, want %v (×%.2f of base %v)", c.level, heat, heat0*c.heatK, c.heatK, heat0)
		}
		// Invariant the feature hinges on: non-earn factors must be >= earn.
		if c.powK < c.earnK {
			t.Errorf("OC %d: power factor %.2f is below earn factor %.2f — OC is no longer a tradeoff", c.level, c.powK, c.earnK)
		}
		if c.heatK < c.earnK {
			t.Errorf("OC %d: heat factor %.2f is below earn factor %.2f — OC is no longer a tradeoff", c.level, c.heatK, c.earnK)
		}
	}
}

// TestOCEarnRateScales verifies the UI-facing BTC/s rate tracks the OC
// multiplier, since the dashboard derives it from GPUEarnRatePerSec rather
// than running a full Tick. Uses a fresh state so market price is a clean
// 1.0× and the room isn't in the half-efficiency hot zone.
func TestOCEarnRateScales(t *testing.T) {
	withTempHome(t)
	s := NewState("oc-rate")
	g := s.GPUs[0]

	g.OCLevel = 0
	base := s.GPUEarnRatePerSec(g)
	if base <= 0 {
		t.Fatalf("base earn rate non-positive: %v", base)
	}
	g.OCLevel = 1
	r1 := s.GPUEarnRatePerSec(g)
	if !approxEqual(r1, base*1.25, 1e-9) {
		t.Errorf("OC 1 earn = %v, want %v (1.25×%v)", r1, base*1.25, base)
	}
	g.OCLevel = 2
	r2 := s.GPUEarnRatePerSec(g)
	if !approxEqual(r2, base*1.50, 1e-9) {
		t.Errorf("OC 2 earn = %v, want %v (1.50×%v)", r2, base*1.50, base)
	}
}

// TestCycleGPUOCRolls exercises the 0→1→2→0 wrap-around and checks that a
// log line lands every step. Unknown IDs and non-running GPUs have to error
// cleanly so the UI can surface a message instead of silently no-op'ing.
func TestCycleGPUOCRolls(t *testing.T) {
	withTempHome(t)
	s := NewState("oc-cycle")
	g := s.GPUs[0]
	startLogs := len(s.Log)

	for _, want := range []int{1, 2, 0} {
		if err := s.CycleGPUOC(g.InstanceID); err != nil {
			t.Fatalf("CycleGPUOC: %v", err)
		}
		if g.OCLevel != want {
			t.Errorf("after cycle: OCLevel = %d, want %d", g.OCLevel, want)
		}
	}
	if len(s.Log)-startLogs != 3 {
		t.Errorf("expected 3 new log entries from 3 cycles, got %d", len(s.Log)-startLogs)
	}

	if err := s.CycleGPUOC(99999); err == nil {
		t.Error("expected error for unknown instance ID")
	}

	g.Status = "broken"
	if err := s.CycleGPUOC(g.InstanceID); err == nil {
		t.Error("expected error when GPU is not running")
	}
	if g.OCLevel != 0 {
		t.Errorf("OCLevel should be unchanged when cycle fails, got %d", g.OCLevel)
	}
}

// TestOCSaveLoadRoundTrip proves the OCLevel sticks through the JSON save
// cycle — a load that silently dropped the field would reset the player's
// overclock choices on every relaunch.
func TestOCSaveLoadRoundTrip(t *testing.T) {
	withTempHome(t)
	s := NewState("oc-save")
	s.GPUs[0].OCLevel = 2
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	loaded, err := LoadFrom(b)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if len(loaded.GPUs) == 0 {
		t.Fatal("no GPUs after reload")
	}
	if loaded.GPUs[0].OCLevel != 2 {
		t.Errorf("OCLevel after roundtrip = %d, want 2", loaded.GPUs[0].OCLevel)
	}
}

// TestOCLegacyLoadDefaultsToZero confirms that saves predating the OC field
// deserialize to the sane default (off, 1× everything) — not a bogus index
// into the multiplier tables.
func TestOCLegacyLoadDefaultsToZero(t *testing.T) {
	withTempHome(t)
	legacy := []byte(`{
		"version": 1,
		"kitten_name": "Legacy",
		"btc": 100.0,
		"difficulty": "normal",
		"gpus": [
			{"instance_id": 1, "def_id": "gtx1060", "status": "running", "hours_left": 10.0, "room": "alley"}
		]
	}`)
	s, err := LoadFrom(legacy)
	if err != nil {
		t.Fatalf("LoadFrom legacy: %v", err)
	}
	if len(s.GPUs) != 1 {
		t.Fatalf("expected 1 GPU, got %d", len(s.GPUs))
	}
	if s.GPUs[0].OCLevel != 0 {
		t.Errorf("legacy GPU OCLevel = %d, want 0", s.GPUs[0].OCLevel)
	}
}

// TestOCClampCorruptSave ensures ensureInit scrubs an out-of-table OCLevel
// so a hand-edited save can't crash the sim by indexing past the end of
// ocEarnMult.
func TestOCClampCorruptSave(t *testing.T) {
	withTempHome(t)
	corrupt := []byte(`{
		"version": 1,
		"kitten_name": "Corrupt",
		"btc": 100.0,
		"difficulty": "normal",
		"gpus": [
			{"instance_id": 1, "def_id": "gtx1060", "status": "running", "hours_left": 10.0, "room": "alley", "oc_level": 99},
			{"instance_id": 2, "def_id": "gtx1060", "status": "running", "hours_left": 10.0, "room": "alley", "oc_level": -1}
		]
	}`)
	s, err := LoadFrom(corrupt)
	if err != nil {
		t.Fatalf("LoadFrom corrupt: %v", err)
	}
	for _, g := range s.GPUs {
		if g.OCLevel != 0 {
			t.Errorf("corrupt GPU OCLevel = %d, want clamped to 0", g.OCLevel)
		}
	}
}
