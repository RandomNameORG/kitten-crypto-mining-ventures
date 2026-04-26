package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	BTCGreen    = lipgloss.Color("#7EE787")
	MoneyGold   = lipgloss.Color("#F9E79F")
	VoltBlue    = lipgloss.Color("#7FDBFF")
	HeatRed     = lipgloss.Color("#FF7A7A")
	KittenPink  = lipgloss.Color("#F5A9B8")
	MutedGrey   = lipgloss.Color("#808080")
	AccentPurple = lipgloss.Color("#BD93F9")
	BorderDim   = lipgloss.Color("#444")
	CrisisRed   = lipgloss.Color("#FF4444")
	OppGreen    = lipgloss.Color("#44FF88")
	SocialCyan  = lipgloss.Color("#44CCFF")
	ThreatOrange = lipgloss.Color("#FFAA44")
	OCWarm1     = lipgloss.Color("#FFD467") // amber/gold — +25% OC (modest boost)
	OCWarm2     = lipgloss.Color("#FF8A5B") // coral — +50% OC (bigger boost)

	TitleStyle = lipgloss.NewStyle().
			Foreground(KittenPink).
			Bold(true)

	DimStyle = lipgloss.NewStyle().Foreground(MutedGrey)

	MoneyStyle = lipgloss.NewStyle().Foreground(MoneyGold).Bold(true)
	BTCStyle   = lipgloss.NewStyle().Foreground(BTCGreen).Bold(true)
	VoltStyle  = lipgloss.NewStyle().Foreground(VoltBlue)
	HeatStyle  = lipgloss.NewStyle().Foreground(HeatRed)

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderDim).
			Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(KittenPink).
			Bold(true).
			Padding(0, 1)

	FooterStyle = lipgloss.NewStyle().
			Foreground(MutedGrey).
			Italic(true).
			Padding(0, 1)

	KeyHint = lipgloss.NewStyle().Foreground(AccentPurple).Bold(true)
)

// fitWidth clamps an ideal panel width to what the terminal can actually
// show, with a sensible floor so we don't collapse past readability.
// `availW` should be the current terminal width (App.w).
func fitWidth(ideal, availW int) int {
	max := availW - 2 // outer padding around panels
	if max < 30 {
		max = 30
	}
	if ideal < max {
		return ideal
	}
	return max
}

// innerFromOuter converts an outer (bordered) panel height into the inner
// content+padding height, which is what lipgloss's Style.Height expects.
// Subtracts 2 for the top/bottom RoundedBorder rows. Clamped to 1.
func innerFromOuter(outerH int) int {
	n := outerH - 2
	if n < 1 {
		n = 1
	}
	return n
}

// renderHeatBar draws a zoned horizontal gauge for heat. Cells 0..80% of the
// bar are tinted green, 80..95% orange, 95..100% red — so the danger zones
// are visible even when the current fill is low. Filled cells use the zone
// colour at full intensity; unfilled cells use a faint variant so the bar
// still reads as "where could this go" rather than "where is it now only".
func renderHeatBar(frac float64, width int) string {
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	if width < 4 {
		width = 4
	}
	filled := int(frac*float64(width) + 0.5)
	var b strings.Builder
	for i := 0; i < width; i++ {
		cellFrac := float64(i) / float64(width)
		var col lipgloss.Color
		switch {
		case cellFrac < 0.80:
			col = OppGreen
		case cellFrac < 0.95:
			col = ThreatOrange
		default:
			col = CrisisRed
		}
		if i < filled {
			b.WriteString(lipgloss.NewStyle().Foreground(col).Render("█"))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(col).Faint(true).Render("░"))
		}
	}
	return b.String()
}

// CategoryStyle returns the log style for an event category.
func CategoryStyle(category string) lipgloss.Style {
	switch category {
	case "crisis":
		return lipgloss.NewStyle().Foreground(CrisisRed).Bold(true)
	case "threat":
		return lipgloss.NewStyle().Foreground(ThreatOrange)
	case "opportunity":
		return lipgloss.NewStyle().Foreground(OppGreen)
	case "social":
		return lipgloss.NewStyle().Foreground(SocialCyan)
	default:
		return lipgloss.NewStyle().Foreground(MutedGrey)
	}
}

// OCLevelStyle returns the style for an OC marker given the GPU's OC level.
// 1 → OCWarm1, 2 → OCWarm2. Callers should gate on ocLevelPercent > 0.
func OCLevelStyle(level int) lipgloss.Style {
	switch level {
	case 1:
		return lipgloss.NewStyle().Foreground(OCWarm1)
	case 2:
		return lipgloss.NewStyle().Foreground(OCWarm2).Bold(true)
	}
	return lipgloss.NewStyle()
}
