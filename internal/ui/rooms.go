package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
)

func (a App) renderRoomsView() string {
	rooms := data.Rooms()
	lines := []string{TitleStyle.Render("🏠 Rooms")}
	lines = append(lines, DimStyle.Render("↑/↓ select   [u] unlock   [enter] switch   [esc]/[1] back"))
	lines = append(lines, "")

	for i, r := range rooms {
		marker := "  "
		if i == a.roomsCursor {
			marker = TitleStyle.Render("▶ ")
		}
		unlocked := a.state.Rooms[r.ID] != nil
		current := r.ID == a.state.CurrentRoom
		state := ""
		switch {
		case current:
			state = BTCStyle.Render("● here")
		case unlocked:
			state = DimStyle.Render("unlocked")
		default:
			state = MoneyStyle.Render(fmt.Sprintf("$%d to unlock", r.UnlockCost))
		}

		lines = append(lines, fmt.Sprintf("%s%-32s  %-10s  slots %-2d  rent $%d/h",
			marker, r.Name, state, r.Slots, r.RentPerHour,
		))
	}

	lines = append(lines, "")
	if a.roomsCursor < len(rooms) {
		sel := rooms[a.roomsCursor]
		lines = append(lines, DimStyle.Italic(true).Render("  "+sel.Flavor))
		lines = append(lines, DimStyle.Render(fmt.Sprintf("  cooling %.1f · elec ×%.2f · threat base %.2f",
			sel.BaseCooling, sel.ElectricCostMult, sel.ThreatBase)))
	}
	return PanelStyle.Width(90).Render(strings.Join(lines, "\n"))
}

func (a App) handleRoomsKey(key string) (tea.Model, tea.Cmd) {
	rooms := data.Rooms()
	switch key {
	case "up", "k":
		if a.roomsCursor > 0 {
			a.roomsCursor--
		}
	case "down", "j":
		if a.roomsCursor < len(rooms)-1 {
			a.roomsCursor++
		}
	case "u":
		if a.roomsCursor < len(rooms) {
			sel := rooms[a.roomsCursor]
			if err := a.state.UnlockRoom(sel.ID); err != nil {
				a = a.withStatus("❌ " + err.Error())
			} else {
				a = a.withStatus(fmt.Sprintf("🔓 %s unlocked", sel.Name))
			}
		}
	case "enter":
		if a.roomsCursor < len(rooms) {
			sel := rooms[a.roomsCursor]
			if err := a.state.SwitchRoom(sel.ID); err != nil {
				a = a.withStatus("❌ " + err.Error())
			} else {
				a = a.withStatus(fmt.Sprintf("📍 now in %s", sel.Name))
			}
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
