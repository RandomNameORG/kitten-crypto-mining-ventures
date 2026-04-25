package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/game"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

func (a App) renderLogView() string {
	log := a.state.Log
	lines := []string{TitleStyle.Render(i18n.T("log.title"))}
	lines = append(lines, DimStyle.Render(i18n.T("log.help")))
	lines = append(lines, "")
	if len(log) == 0 {
		lines = append(lines, DimStyle.Render(i18n.T("log.empty")))
	}
	// Show the most recent N entries that can fit. The body-clip in
	// View() handles terminal height; here we just cap the visible window
	// well above the GPU page size so the log feels chronological.
	limit := 30
	if limit > len(log) {
		limit = len(log)
	}
	for n := 0; n < limit; n++ {
		e := log[len(log)-1-n]
		ts := time.Unix(e.Time, 0).Format("15:04:05")
		line := fmt.Sprintf("  %s  %s", DimStyle.Render(ts), CategoryStyle(e.Category).Render(e.Text))
		lines = append(lines, line)
	}
	if len(log) > limit {
		lines = append(lines, "", DimStyle.Render(fmt.Sprintf(i18n.T("log.older"), len(log)-limit)))
	}
	return PanelStyle.Width(fitWidth(100, a.w)).Render(strings.Join(lines, "\n"))
}

func (a App) renderHelpView() string {
	lines := []string{
		TitleStyle.Render(i18n.T("help.title")),
		"",
		HeaderStyle.Render(i18n.T("help.views")),
		KeyHint.Render(i18n.T("help.view.1")),
		KeyHint.Render(i18n.T("help.view.2")),
		KeyHint.Render(i18n.T("help.view.3")),
		KeyHint.Render(i18n.T("help.view.4")),
		KeyHint.Render(i18n.T("help.view.5")),
		KeyHint.Render(i18n.T("help.view.6")),
		KeyHint.Render(i18n.T("help.view.7")),
		KeyHint.Render(i18n.T("help.view.8")),
		KeyHint.Render(i18n.T("help.view.9")),
		KeyHint.Render(i18n.T("help.view.0")),
		"",
		HeaderStyle.Render(i18n.T("help.global")),
		KeyHint.Render(i18n.T("help.g.space")),
		KeyHint.Render(i18n.T("help.g.save")),
		KeyHint.Render(i18n.T("help.g.pump")),
		KeyHint.Render(i18n.T("help.g.lang")),
		KeyHint.Render(i18n.T("help.g.vent", "-"+game.FmtBTCInt(game.EmergencyVentCost))),
		KeyHint.Render(i18n.T("help.g.quit")),
		"",
		HeaderStyle.Render(i18n.T("help.defense")),
		DimStyle.Render(i18n.T("help.defense_row")),
		"",
		HeaderStyle.Render(i18n.T("help.mechanics")),
		lipgloss.NewStyle().Foreground(HeatRed).Render(i18n.T("help.mech.heat")),
		lipgloss.NewStyle().Foreground(OppGreen).Render(i18n.T("help.mech.heat.z1")),
		lipgloss.NewStyle().Foreground(ThreatOrange).Render(i18n.T("help.mech.heat.z2")),
		lipgloss.NewStyle().Foreground(CrisisRed).Render(i18n.T("help.mech.heat.z3")),
		DimStyle.Render(i18n.T("help.mech.heat.act")),
		"",
		VoltStyle.Render(i18n.T("help.mech.power")),
		DimStyle.Render(i18n.T("help.mech.power.2")),
		lipgloss.NewStyle().Foreground(CrisisRed).Render(i18n.T("help.mech.power.3")),
		DimStyle.Render(i18n.T("help.mech.power.act")),
		"",
		lipgloss.NewStyle().Foreground(OppGreen).Render(i18n.T("help.mech.earn")),
		DimStyle.Render(i18n.T("help.mech.earn.2")),
		"",
		lipgloss.NewStyle().Foreground(AccentPurple).Render(i18n.T("help.mech.market")),
		DimStyle.Render(i18n.T("help.mech.market.2")),
		"",
		DimStyle.Render(i18n.T("help.tip.idle")),
		DimStyle.Render(i18n.T("help.tip.offline")),
	}
	return PanelStyle.Width(fitWidth(70, a.w)).Render(strings.Join(lines, "\n"))
}
