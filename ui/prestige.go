package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/game"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

func (a App) renderPrestigeView() string {
	legacy := game.LoadLegacy()
	header := TitleStyle.Render(i18n.T("prestige.title"))
	if !a.state.HasUnlock("prestige") {
		return strings.Join([]string{
			header, "",
			DimStyle.Render(i18n.T("prestige.locked")),
		}, "\n")
	}
	help := DimStyle.Render(i18n.T("prestige.help"))

	lines := []string{HeaderStyle.Render(i18n.T("prestige.status"))}
	lines = append(lines, i18n.T("prestige.lifetime", a.state.LifetimeEarned, game.PrestigeThreshold))
	reward := a.state.RetireReward()
	canRetire := a.state.CanRetire()
	status := DimStyle.Render(i18n.T("prestige.eligible_no"))
	if canRetire {
		status = lipgloss.NewStyle().Foreground(OppGreen).Render(i18n.T("prestige.eligible_yes"))
	}
	lines = append(lines, i18n.T("prestige.eligible_row", status))
	lines = append(lines, i18n.T("prestige.reward", reward))
	lines = append(lines, "")
	lines = append(lines, i18n.T("prestige.bank", legacy.TotalLP, legacy.SpentLP, legacy.LPAvailable()))
	statusPanel := PanelStyle.Width(fitWidth(90, a.w)).Render(strings.Join(lines, "\n"))

	perks := game.LegacyPerks()
	perkLines := []string{HeaderStyle.Render(i18n.T("prestige.perks"))}
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
			meta = i18n.T("prestige.perk_owned")
		} else if legacy.LPAvailable() >= p.Cost {
			label = MoneyStyle.Render(label)
		}
		perkLines = append(perkLines, fmt.Sprintf("%s%-26s  %s", cursor, label, meta))
		perkLines = append(perkLines, DimStyle.Render("    "+p.Desc))
	}
	perksPanel := PanelStyle.Width(fitWidth(90, a.w)).Render(strings.Join(perkLines, "\n"))

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
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.perk_bought"))
			}
		}
	case "R":
		if !a.state.CanRetire() {
			a = a.withStatus(i18n.T("status.retire_deny"))
			return a, nil
		}
		if a.retireArmedUntil.IsZero() || time.Now().After(a.retireArmedUntil) {
			a.retireArmedUntil = time.Now().Add(5 * time.Second)
			a = a.withStatus(i18n.T("status.retire_arm"))
			return a, nil
		}
		a.retireArmedUntil = time.Time{}
		fresh, lp, err := a.state.Retire()
		if err != nil {
			a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
		} else {
			a.state = fresh
			_ = a.saveNow()
			a = a.withStatus(i18n.T("status.retired", lp))
			a.view = viewDashboard
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
