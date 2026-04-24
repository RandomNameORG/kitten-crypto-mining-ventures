package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

var defenseDims = []struct {
	Key   string
	Dim   string
	I18n  string
}{
	{"l", "lock", "rooms.dim.lock"},
	{"c", "cctv", "rooms.dim.cctv"},
	{"w", "wiring", "rooms.dim.wiring"},
	{"o", "cooling", "rooms.dim.cooling"},
	{"a", "armor", "rooms.dim.armor"},
}

func (a App) renderRoomsView() string {
	rooms := data.Rooms()
	lines := []string{TitleStyle.Render(i18n.T("rooms.title"))}
	lines = append(lines, DimStyle.Render(i18n.T("rooms.help")))
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
			state = BTCStyle.Render(i18n.T("rooms.here"))
		case unlocked:
			state = DimStyle.Render(i18n.T("rooms.unlocked"))
		default:
			state = BTCStyle.Render(i18n.T("rooms.to_unlock", r.UnlockCost))
		}
		lines = append(lines, fmt.Sprintf("%s%-32s  %-10s  slots %-2d  rent ₿%d/h",
			marker, r.LocalName(), state, r.Slots, r.RentPerHour,
		))
	}

	lines = append(lines, "")
	if a.roomsCursor < len(rooms) {
		sel := rooms[a.roomsCursor]
		lines = append(lines, lipgloss.NewStyle().Foreground(AccentPurple).Italic(true).Render("  "+sel.LocalFlavor()))
		lines = append(lines, DimStyle.Render(i18n.T("rooms.stats", sel.BaseCooling, sel.ElectricCostMult, sel.ThreatBase)))
	}

	lines = append(lines, "")
	curRoomDef, _ := data.RoomByID(a.state.CurrentRoom)
	lines = append(lines, HeaderStyle.Render(i18n.T("rooms.defense", curRoomDef.LocalName())))
	if cur := a.state.Rooms[a.state.CurrentRoom]; cur != nil {
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
			lines = append(lines, style.Render(fmt.Sprintf("  [%s] %-8s  lv %d/5   next ₿%d",
				d.Key, i18n.T(d.I18n), lvl, next)))
		}
	}
	return PanelStyle.Width(fitWidth(100, a.w)).Render(strings.Join(lines, "\n"))
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
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.unlocked", sel.LocalName()))
			}
		}
	case "enter":
		if a.roomsCursor < len(rooms) {
			sel := rooms[a.roomsCursor]
			if err := a.state.SwitchRoom(sel.ID); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.now_in", sel.LocalName()))
			}
		}
	case "l", "c", "w", "o", "a":
		dim := ""
		label := ""
		for _, d := range defenseDims {
			if d.Key == key {
				dim = d.Dim
				label = i18n.T(d.I18n)
			}
		}
		if dim != "" {
			if err := a.state.UpgradeDefense(dim); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.defense_up", label))
			}
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
