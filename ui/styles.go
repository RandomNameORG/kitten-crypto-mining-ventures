package ui

import "github.com/charmbracelet/lipgloss"

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
