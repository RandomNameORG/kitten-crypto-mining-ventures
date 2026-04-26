package ui

import "testing"

// TestSparklineShape pins the contract the Stats view depends on: one rune
// per sample, the lowest input maps to ▁, the highest to █, and the
// degenerate empty / all-equal inputs render usefully without panicking.
func TestSparklineShape(t *testing.T) {
	// Empty input → empty output, no padding glyph injected.
	if got := Sparkline(nil); got != "" {
		t.Errorf("nil input: got %q, want empty", got)
	}
	if got := Sparkline([]float64{}); got != "" {
		t.Errorf("empty slice: got %q, want empty", got)
	}

	// Ten ascending values: must produce exactly 10 runes, with extremes
	// pinned to the lowest/highest ramp glyphs.
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	out := Sparkline(values)
	runes := []rune(out)
	if len(runes) != 10 {
		t.Errorf("ascending input: got %d runes, want 10 (output=%q)", len(runes), out)
	}
	if len(runes) > 0 {
		if runes[0] != '▁' {
			t.Errorf("first rune = %q, want ▁ (lowest bucket)", runes[0])
		}
		if runes[len(runes)-1] != '█' {
			t.Errorf("last rune = %q, want █ (highest bucket)", runes[len(runes)-1])
		}
	}

	// All-equal inputs: must render N visible runes, never collapse to "".
	flat := []float64{2.5, 2.5, 2.5, 2.5}
	flatOut := Sparkline(flat)
	if flatOut == "" {
		t.Error("all-equal input rendered empty; want a non-empty repeated glyph")
	}
	if got := len([]rune(flatOut)); got != len(flat) {
		t.Errorf("all-equal input: got %d runes, want %d", got, len(flat))
	}
}
