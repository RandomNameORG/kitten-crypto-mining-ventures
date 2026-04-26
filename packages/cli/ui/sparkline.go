package ui

import "strings"

// sparklineRamp is the 8-level block ramp used by Sparkline. Ordered low→high
// so the bucket index doubles as a brightness/height cue.
var sparklineRamp = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Sparkline renders a slice of values into a fixed-width block-glyph string,
// one rune per sample. Each value is bucketed to one of the 8 ramp glyphs
// based on its position within the [min, max] window. Empty input yields
// "" so callers can detect "no data" without an extra branch. When every
// value is identical (max == min, including a single-element slice) we emit
// the mid glyph repeated — the height is meaningless but readers still see
// the sample count.
func Sparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}
	min, max := values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	mid := sparklineRamp[len(sparklineRamp)/2]
	if max == min {
		return strings.Repeat(string(mid), len(values))
	}
	span := max - min
	last := len(sparklineRamp) - 1
	var b strings.Builder
	for _, v := range values {
		idx := int(((v - min) / span) * float64(last))
		if idx < 0 {
			idx = 0
		}
		if idx > last {
			idx = last
		}
		b.WriteRune(sparklineRamp[idx])
	}
	return b.String()
}
