package ui

import (
	"fmt"
	"strings"
	"time"

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
	for i := len(log) - 1; i >= 0 && i > len(log)-50; i-- {
		e := log[i]
		ts := time.Unix(e.Time, 0).Format("15:04:05")
		line := fmt.Sprintf("  %s  %s", DimStyle.Render(ts), CategoryStyle(e.Category).Render(e.Text))
		lines = append(lines, line)
	}
	return PanelStyle.Width(100).Render(strings.Join(lines, "\n"))
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
		"",
		HeaderStyle.Render(i18n.T("help.global")),
		KeyHint.Render(i18n.T("help.g.space")),
		KeyHint.Render(i18n.T("help.g.save")),
		KeyHint.Render(i18n.T("help.g.pump")),
		KeyHint.Render(i18n.T("help.g.lang")),
		KeyHint.Render(i18n.T("help.g.vent")),
		KeyHint.Render(i18n.T("help.g.quit")),
		"",
		HeaderStyle.Render(i18n.T("help.defense")),
		DimStyle.Render(i18n.T("help.defense_row")),
		"",
		DimStyle.Render(i18n.T("help.tip.idle")),
		DimStyle.Render(i18n.T("help.tip.offline")),
	}
	return PanelStyle.Width(70).Render(strings.Join(lines, "\n"))
}
