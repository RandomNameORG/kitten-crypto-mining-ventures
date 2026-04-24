package game

import "fmt"

// DisplayScale is applied to raw BTC values before they are shown to the
// player. Internal math (state.BTC, JSON prices, MiningScale) operates on
// unscaled units; this keeps save files compatible while letting the UI
// feel like actual BTC — a GTX 1060 displays as ₿0.4, earns ~₿0.0012/s.
const DisplayScale = 300.0

// FmtBTC formats a raw balance amount with adaptive precision suited to
// post-rescale magnitudes. Handles negatives with a leading minus.
func FmtBTC(raw float64) string {
	v := raw / DisplayScale
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	switch {
	case v < 0.001:
		return fmt.Sprintf("%s₿%.5f", sign, v)
	case v < 1:
		return fmt.Sprintf("%s₿%.4f", sign, v)
	case v < 100:
		return fmt.Sprintf("%s₿%.3f", sign, v)
	case v < 10000:
		return fmt.Sprintf("%s₿%.2f", sign, v)
	default:
		return fmt.Sprintf("%s₿%.0f", sign, v)
	}
}

// FmtBTCInt formats an integer BTC amount (typically JSON-sourced prices).
func FmtBTCInt(raw int) string { return FmtBTC(float64(raw)) }

// FmtBTCSigned always emits an explicit +/- so "net" readouts read as
// deltas rather than absolute numbers. Zero renders as "+₿…".
func FmtBTCSigned(raw float64) string {
	v := raw / DisplayScale
	sign := "+"
	if v < 0 {
		sign = "-"
		v = -v
	}
	switch {
	case v < 0.001:
		return fmt.Sprintf("%s₿%.5f", sign, v)
	case v < 1:
		return fmt.Sprintf("%s₿%.4f", sign, v)
	case v < 100:
		return fmt.Sprintf("%s₿%.3f", sign, v)
	case v < 10000:
		return fmt.Sprintf("%s₿%.2f", sign, v)
	default:
		return fmt.Sprintf("%s₿%.0f", sign, v)
	}
}
