package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/i18n"
)

func (a App) renderMercsView() string {
	owned := a.state.Mercs
	hireable := data.Mercs()

	header := TitleStyle.Render(i18n.T("mercs.title"))
	help := DimStyle.Render(i18n.T("mercs.help"))

	ownedLines := []string{HeaderStyle.Render(i18n.T("mercs.yours"))}
	if len(owned) == 0 {
		ownedLines = append(ownedLines, DimStyle.Render(i18n.T("mercs.empty")))
	}
	for i, m := range owned {
		cursor := "  "
		if a.mercsTab == 0 && i == a.mercsOwnedCur {
			cursor = TitleStyle.Render("▶ ")
		}
		def, _ := data.MercByID(m.DefID)
		roomDef, _ := data.RoomByID(m.RoomID)
		roomName := roomDef.LocalName()
		if roomName == "" {
			roomName = m.RoomID
		}
		loyStyle := DimStyle
		switch {
		case m.Loyalty < 20:
			loyStyle = lipgloss.NewStyle().Foreground(CrisisRed)
		case m.Loyalty < 50:
			loyStyle = lipgloss.NewStyle().Foreground(ThreatOrange)
		case m.Loyalty >= 80:
			loyStyle = lipgloss.NewStyle().Foreground(OppGreen)
		}
		line := fmt.Sprintf("%s#%-3d %-28s  %s",
			cursor, m.InstanceID, def.LocalName(),
			loyStyle.Render(i18n.T("mercs.owned_line", roomName, def.WeeklyWage, m.Loyalty)),
		)
		ownedLines = append(ownedLines, line)
	}

	hireLines := []string{HeaderStyle.Render(i18n.T("mercs.hire"))}
	for i, d := range hireable {
		cursor := "  "
		if a.mercsTab == 1 && i == a.mercsCursor {
			cursor = TitleStyle.Render("▶ ")
		}
		priceStyle := MoneyStyle
		if a.state.Money < float64(d.HireCost) {
			priceStyle = DimStyle
		}
		line := fmt.Sprintf("%s%-24s  %s  wage $%d/wk  %s",
			cursor, d.LocalName(),
			priceStyle.Render(i18n.T("mercs.hire_line", d.HireCost)),
			d.WeeklyWage, i18n.T("mercs.defbonus", d.DefenseBonus*100),
		)
		hireLines = append(hireLines, line)
		hireLines = append(hireLines, DimStyle.Render("   "+d.LocalFlavor()))
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
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.hired", sel.LocalName()))
			}
		}
	case "f":
		if a.mercsOwnedCur < len(owned) {
			sel := owned[a.mercsOwnedCur]
			if err := a.state.FireMerc(sel.InstanceID); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.dismissed"))
				if a.mercsOwnedCur > 0 {
					a.mercsOwnedCur--
				}
			}
		}
	case "b":
		if a.mercsOwnedCur < len(owned) {
			sel := owned[a.mercsOwnedCur]
			if err := a.state.BribeMerc(sel.InstanceID); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.bribed"))
			}
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
