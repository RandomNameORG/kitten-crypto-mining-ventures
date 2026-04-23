package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
)

var defenseDims = []struct {
	Key  string
	Dim  string
	Name string
}{
	{"l", "lock", "Lock"},
	{"c", "cctv", "CCTV"},
	{"w", "wiring", "Wiring"},
	{"o", "cooling", "Cooling"},
	{"a", "armor", "Armor"},
}

func (a App) renderRoomsView() string {
	rooms := data.Rooms()
	lines := []string{TitleStyle.Render("🏠 Rooms")}
	lines = append(lines, DimStyle.Render("↑/↓ room   [u] unlock   [enter] switch   [l/c/w/o/a] upgrade defense on current room   [esc]/[1] back"))
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
		lines = append(lines, lipgloss.NewStyle().Foreground(AccentPurple).Italic(true).Render("  "+sel.Flavor))
		lines = append(lines, DimStyle.Render(fmt.Sprintf("  cooling %.1f · elec ×%.2f · threat base %.2f",
			sel.BaseCooling, sel.ElectricCostMult, sel.ThreatBase)))
	}

	lines = append(lines, "")
	lines = append(lines, HeaderStyle.Render(fmt.Sprintf("🛡  Defense — current room (%s)", a.state.CurrentRoom)))
	if cur := a.state.Rooms[a.state.CurrentRoom]; cur != nil {
		rowFmt := "  [%s] %-8s  lv %d/5   next $%d"
		levels := map[string]int{
			"lock":    cur.LockLvl,
			"cctv":    cur.CCTVLvl,
			"wiring":  cur.WiringLvl,
			"cooling": cur.CoolingLvl,
			"armor":   cur.ArmorLvl,
		}
		for _, d := range defenseDims {
			lvl := levels[d.Dim]
			next := (lvl + 1) * 250
			style := DimStyle
			if lvl >= 5 {
				style = lipgloss.NewStyle().Foreground(OppGreen)
			}
			lines = append(lines, style.Render(fmt.Sprintf(rowFmt, d.Key, d.Name, lvl, next)))
		}
	}
	return PanelStyle.Width(100).Render(strings.Join(lines, "\n"))
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
	case "l", "c", "w", "o", "a":
		dim := ""
		for _, d := range defenseDims {
			if d.Key == key {
				dim = d.Dim
			}
		}
		if dim != "" {
			if err := a.state.UpgradeDefense(dim); err != nil {
				a = a.withStatus("❌ " + err.Error())
			} else {
				a = a.withStatus(fmt.Sprintf("🛡 %s upgraded", dim))
			}
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
