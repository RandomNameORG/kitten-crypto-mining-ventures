package ui

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

// renderMasteryView shows the mastery ladder — infinite-ish TP sinks with
// multiplicative bonuses. Shows current level, next-level cost, and the
// effective multiplier so the player can plan TP spending.
func (a App) renderMasteryView() string {
	tracks := data.MasteryTracks()
	header := TitleStyle.Render(i18n.T("mastery.title")) +
		"   " + DimStyle.Render(i18n.T("skills.tp_count", a.state.TechPoint))
	help := DimStyle.Render(i18n.T("mastery.help"))

	lines := []string{header, help, ""}

	for i, t := range tracks {
		cur := a.state.MasteryLevel(t.ID)
		cost := t.CostFor(cur)
		mult := math.Pow(1.0+t.PerLevel, float64(cur))

		cursor := "  "
		if i == a.masteryCursor {
			cursor = TitleStyle.Render("▶ ")
		}

		nameLine := cursor + lipgloss.NewStyle().Foreground(KittenPink).Bold(true).
			Render(fmt.Sprintf("%s  %s", t.Emoji, t.LocalName()))
		nameLine += "   " + DimStyle.Render(fmt.Sprintf("Lv %d / %d", cur, t.MaxLevel))

		// Show current effective multiplier; for power that's < 1.0 (a discount).
		var multStr string
		if t.PerLevel < 0 {
			multStr = fmt.Sprintf("%.0f%%", mult*100)
		} else {
			multStr = fmt.Sprintf("×%.3f", mult)
		}

		costStr := i18n.T("mastery.maxed")
		if cost > 0 {
			costStr = i18n.T("mastery.next_cost", cost)
		}

		statLine := DimStyle.Render(fmt.Sprintf("    %s   ·   now %s   ·   %s",
			t.LocalDesc(), multStr, costStr))

		lines = append(lines, nameLine, statLine, "")
	}

	lines = append(lines,
		DimStyle.Render(i18n.T("mastery.alchemy_note")))

	return PanelStyle.Width(fitWidth(90, a.w)).Render(strings.Join(lines, "\n"))
}

func (a App) handleMasteryKey(key string) (tea.Model, tea.Cmd) {
	tracks := data.MasteryTracks()
	switch key {
	case "up", "k":
		if a.masteryCursor > 0 {
			a.masteryCursor--
		}
	case "down", "j":
		if a.masteryCursor < len(tracks)-1 {
			a.masteryCursor++
		}
	case "u", "enter":
		if a.masteryCursor < len(tracks) {
			t := tracks[a.masteryCursor]
			if _, err := a.state.LevelUpMastery(t.ID); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.mastery_up", t.LocalName()))
			}
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
