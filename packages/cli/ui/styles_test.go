package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestCategoryStyleKeys asserts each known log category resolves to a style
// with a populated foreground, and that the unknown/info default also paints
// (so "info" events render in the MutedGrey fallback rather than bare).
func TestCategoryStyleKeys(t *testing.T) {
	cases := []string{"crisis", "threat", "opportunity", "social", "info"}
	for _, c := range cases {
		fg := CategoryStyle(c).GetForeground()
		if fg == nil {
			t.Fatalf("CategoryStyle(%q) returned nil foreground", c)
		}
		if s, ok := fg.(lipgloss.Color); ok && string(s) == "" {
			t.Fatalf("CategoryStyle(%q) returned empty foreground color", c)
		}
	}
	// Unknown key should fall through to the same default arm.
	fg := CategoryStyle("made-up").GetForeground()
	if s, ok := fg.(lipgloss.Color); ok && string(s) == "" {
		t.Fatalf("CategoryStyle default arm returned empty foreground color")
	}
}

// TestOCLevelStyleKeys asserts the two active tiers have non-empty, distinct
// foregrounds and level 0 returns the zero style (no colour applied).
func TestOCLevelStyleKeys(t *testing.T) {
	fg1 := OCLevelStyle(1).GetForeground()
	fg2 := OCLevelStyle(2).GetForeground()
	c1, ok1 := fg1.(lipgloss.Color)
	c2, ok2 := fg2.(lipgloss.Color)
	if !ok1 || string(c1) == "" {
		t.Fatalf("OCLevelStyle(1) foreground empty: %#v", fg1)
	}
	if !ok2 || string(c2) == "" {
		t.Fatalf("OCLevelStyle(2) foreground empty: %#v", fg2)
	}
	if string(c1) == string(c2) {
		t.Fatalf("OCLevelStyle(1) and OCLevelStyle(2) share the same colour %q", string(c1))
	}

	fg0 := OCLevelStyle(0).GetForeground()
	if c0, ok := fg0.(lipgloss.Color); ok && string(c0) != "" {
		t.Fatalf("OCLevelStyle(0) should be the zero style; got foreground %q", string(c0))
	}
}

// TestRankColourDistinct asserts all three ranked tiers paint distinct,
// non-empty colours — a regression here would collapse the rank ramp.
func TestRankColourDistinct(t *testing.T) {
	top := string(rankColour(rankTop))
	mid := string(rankColour(rankMid))
	low := string(rankColour(rankLow))
	if top == "" || mid == "" || low == "" {
		t.Fatalf("empty rank colour: top=%q mid=%q low=%q", top, mid, low)
	}
	if top == mid || mid == low || top == low {
		t.Fatalf("rank colours not pairwise distinct: top=%q mid=%q low=%q", top, mid, low)
	}
}

// TestPaletteNonEmpty guards against an accidental `lipgloss.Color("")`
// landing in the palette — that would silently fall back to terminal default
// in every place the named colour is used.
func TestPaletteNonEmpty(t *testing.T) {
	palette := map[string]lipgloss.Color{
		"BTCGreen":     BTCGreen,
		"MoneyGold":    MoneyGold,
		"VoltBlue":     VoltBlue,
		"HeatRed":      HeatRed,
		"CrisisRed":    CrisisRed,
		"OppGreen":     OppGreen,
		"SocialCyan":   SocialCyan,
		"ThreatOrange": ThreatOrange,
		"OCWarm1":      OCWarm1,
		"OCWarm2":      OCWarm2,
		"AccentPurple": AccentPurple,
		"MutedGrey":    MutedGrey,
		"BorderDim":    BorderDim,
		"KittenPink":   KittenPink,
	}
	for name, c := range palette {
		if string(c) == "" {
			t.Fatalf("palette colour %s is empty", name)
		}
	}
}
