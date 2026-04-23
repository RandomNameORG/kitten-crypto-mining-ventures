package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
)

func (a App) renderMercsView() string {
	owned := a.state.Mercs
	hireable := data.Mercs()

	header := TitleStyle.Render("🐾 Mercenaries")
	help := DimStyle.Render("[tab] switch tab   ↑/↓ select   [h] hire   [f] fire   [b] bribe (+15 loyalty, $200)   [esc]/[1] back")

	// Owned column.
	ownedLines := []string{HeaderStyle.Render("Your Mercs")}
	if len(owned) == 0 {
		ownedLines = append(ownedLines, DimStyle.Render("  (none — switch to Hire tab)"))
	}
	for i, m := range owned {
		cursor := "  "
		if a.mercsTab == 0 && i == a.mercsOwnedCur {
			cursor = TitleStyle.Render("▶ ")
		}
		def, _ := data.MercByID(m.DefID)
		loyStyle := DimStyle
		switch {
		case m.Loyalty < 20:
			loyStyle = lipgloss.NewStyle().Foreground(CrisisRed)
		case m.Loyalty < 50:
			loyStyle = lipgloss.NewStyle().Foreground(ThreatOrange)
		case m.Loyalty >= 80:
			loyStyle = lipgloss.NewStyle().Foreground(OppGreen)
		}
		line := fmt.Sprintf("%s#%-3d %-28s  room %-12s  wage $%d/wk  %s",
			cursor, m.InstanceID, def.Name, m.RoomID, def.WeeklyWage,
			loyStyle.Render(fmt.Sprintf("loyalty %d", m.Loyalty)),
		)
		ownedLines = append(ownedLines, line)
	}

	// Hireable column.
	hireLines := []string{HeaderStyle.Render("Hire")}
	for i, d := range hireable {
		cursor := "  "
		if a.mercsTab == 1 && i == a.mercsCursor {
			cursor = TitleStyle.Render("▶ ")
		}
		priceStyle := MoneyStyle
		if a.state.Money < float64(d.HireCost) {
			priceStyle = DimStyle
		}
		line := fmt.Sprintf("%s%-24s  %s  wage $%d/wk  def +%.0f%%",
			cursor, d.Name,
			priceStyle.Render(fmt.Sprintf("hire $%d", d.HireCost)),
			d.WeeklyWage, d.DefenseBonus*100,
		)
		hireLines = append(hireLines, line)
		hireLines = append(hireLines, DimStyle.Render("   "+d.Flavor))
	}

	left := PanelStyle.Width(70).Render(strings.Join(ownedLines, "\n"))
	right := PanelStyle.Width(70).Render(strings.Join(hireLines, "\n"))

	var body string
	if a.mercsTab == 0 {
		body = lipgloss.JoinVertical(lipgloss.Left, left, right)
	} else {
		body = lipgloss.JoinVertical(lipgloss.Left, right, left)
	}
	return strings.Join([]string{header, help, "", body}, "\n")
}

func (a App) handleMercsKey(key string) (tea.Model, tea.Cmd) {
	owned := a.state.Mercs
	hireable := data.Mercs()
	switch key {
	case "tab":
		a.mercsTab = 1 - a.mercsTab
	case "up", "k":
		if a.mercsTab == 0 {
			if a.mercsOwnedCur > 0 {
				a.mercsOwnedCur--
			}
		} else {
			if a.mercsCursor > 0 {
				a.mercsCursor--
			}
		}
	case "down", "j":
		if a.mercsTab == 0 {
			if a.mercsOwnedCur < len(owned)-1 {
				a.mercsOwnedCur++
			}
		} else {
			if a.mercsCursor < len(hireable)-1 {
				a.mercsCursor++
			}
		}
	case "h":
		if a.mercsCursor < len(hireable) {
			sel := hireable[a.mercsCursor]
			if err := a.state.HireMerc(sel.ID); err != nil {
				a = a.withStatus("❌ " + err.Error())
			} else {
				a = a.withStatus("🐾 hired " + sel.Name)
			}
		}
	case "f":
		if a.mercsOwnedCur < len(owned) {
			sel := owned[a.mercsOwnedCur]
			if err := a.state.FireMerc(sel.InstanceID); err != nil {
				a = a.withStatus("❌ " + err.Error())
			} else {
				a = a.withStatus("dismissed")
				if a.mercsOwnedCur > 0 {
					a.mercsOwnedCur--
				}
			}
		}
	case "b":
		if a.mercsOwnedCur < len(owned) {
			sel := owned[a.mercsOwnedCur]
			if err := a.state.BribeMerc(sel.InstanceID); err != nil {
				a = a.withStatus("❌ " + err.Error())
			} else {
				a = a.withStatus("🎁 loyalty boosted")
			}
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
