package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/game"
)

func (a App) renderPrestigeView() string {
	legacy := game.LoadLegacy()
	header := TitleStyle.Render("🎓 Prestige — Retire & Restart")
	if !a.state.HasUnlock("prestige") {
		return strings.Join([]string{
			header, "",
			DimStyle.Render("Prestige is locked. Unlock 'Venture Capital' in the Mogul skill lane."),
		}, "\n")
	}
	help := DimStyle.Render("[↑/↓] select perk   [p] buy perk   [R] RETIRE (only when eligible)   [esc]/[1] back")

	// Status.
	lines := []string{HeaderStyle.Render("Status")}
	lines = append(lines, fmt.Sprintf("  lifetime earned: $%.0f / $%.0f", a.state.LifetimeEarned, game.PrestigeThreshold))
	reward := a.state.RetireReward()
	canRetire := a.state.CanRetire()
	status := DimStyle.Render("not eligible")
	if canRetire {
		status = lipgloss.NewStyle().Foreground(OppGreen).Render("ELIGIBLE")
	}
	lines = append(lines, fmt.Sprintf("  retirement status: %s", status))
	lines = append(lines, fmt.Sprintf("  retire reward: %d LP", reward))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  bank balance: %d LP total · %d spent · %d available",
		legacy.TotalLP, legacy.SpentLP, legacy.LPAvailable()))
	statusPanel := PanelStyle.Width(90).Render(strings.Join(lines, "\n"))

	// Perks.
	perks := game.LegacyPerks()
	perkLines := []string{HeaderStyle.Render("Legacy Perks")}
	for i, p := range perks {
		cursor := "  "
		if i == a.prestigeCursor {
			cursor = TitleStyle.Render("▶ ")
		}
		available := p.Available(legacy)
		label := p.Name
		meta := fmt.Sprintf("%d LP", p.Cost)
		if !available {
			label = DimStyle.Render("✓ " + label)
			meta = "owned / maxed"
		} else if legacy.LPAvailable() >= p.Cost {
			label = MoneyStyle.Render(label)
		}
		perkLines = append(perkLines, fmt.Sprintf("%s%-26s  %s", cursor, label, meta))
		perkLines = append(perkLines, DimStyle.Render("    "+p.Desc))
	}
	perksPanel := PanelStyle.Width(90).Render(strings.Join(perkLines, "\n"))

	return strings.Join([]string{
		header, help, "",
		lipgloss.JoinVertical(lipgloss.Left, statusPanel, perksPanel),
	}, "\n")
}

func (a App) handlePrestigeKey(key string) (tea.Model, tea.Cmd) {
	perks := game.LegacyPerks()
	switch key {
	case "up", "k":
		if a.prestigeCursor > 0 {
			a.prestigeCursor--
		}
	case "down", "j":
		if a.prestigeCursor < len(perks)-1 {
			a.prestigeCursor++
		}
	case "p":
		if a.prestigeCursor < len(perks) {
			sel := perks[a.prestigeCursor]
			if err := game.BuyLegacyPerk(sel.ID); err != nil {
				a = a.withStatus("❌ " + err.Error())
			} else {
				a = a.withStatus("🎁 perk purchased")
			}
		}
	case "R":
		fresh, lp, err := a.state.Retire()
		if err != nil {
			a = a.withStatus("❌ " + err.Error())
		} else {
			// Swap in fresh state. Save both legacy + new run immediately.
			a.state = fresh
			_ = a.saveNow()
			a = a.withStatus(fmt.Sprintf("🐾 retired. +%d LP banked. New run begins.", lp))
			a.view = viewDashboard
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
