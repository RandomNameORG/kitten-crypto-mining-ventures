package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/game"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"
)

var defenseDims = []struct {
	Key  string
	Dim  string
	I18n string
}{
	{"l", "lock", "rooms.dim.lock"},
	{"c", "cctv", "rooms.dim.cctv"},
	{"w", "wiring", "rooms.dim.wiring"},
	{"o", "cooling", "rooms.dim.cooling"},
	{"a", "armor", "rooms.dim.armor"},
}

// installablePSUs returns the catalog filtered to PSUs the player can buy
// (every entry except the implicit psu_builtin). Used by both the picker
// renderer and the install/replace key handlers.
func installablePSUs() []data.PSUDef {
	all := data.PSUs()
	out := make([]data.PSUDef, 0, len(all))
	for _, p := range all {
		if p.ID == "psu_builtin" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func (a App) renderRoomsView() string {
	rooms := data.Rooms()
	lines := []string{TitleStyle.Render(i18n.T("rooms.title"))}
	lines = append(lines, DimStyle.Render(i18n.T("rooms.help")))
	lines = append(lines, "")

	for i, r := range rooms {
		marker := "  "
		if i == a.roomsCursor && a.roomsFocus == 0 {
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
			state = BTCStyle.Render(i18n.T("rooms.to_unlock", game.FmtBTCInt(r.UnlockCost)))
		}
		lines = append(lines, fmt.Sprintf("%s%-32s  %-10s  slots %-2d  rent %s/h",
			marker, r.LocalName(), state, r.Slots, game.FmtBTCInt(r.RentPerHour),
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
			frags := game.DefenseFragsForLevel(lvl)
			style := DimStyle
			if lvl >= game.MaxDefenseLevel {
				style = lipgloss.NewStyle().Foreground(OppGreen)
			}
			fragSuffix := ""
			if frags > 0 {
				fragSuffix = fmt.Sprintf(" + %d frags", frags)
			}
			lines = append(lines, style.Render(fmt.Sprintf("  [%s] %-8s  lv %d/%d   next %s%s",
				d.Key, i18n.T(d.I18n), lvl, game.MaxDefenseLevel, game.FmtBTCInt(next), fragSuffix)))
		}
	}

	// PSU section — Sprint 1.
	lines = append(lines, "")
	lines = append(lines, HeaderStyle.Render(i18n.T("rooms.psu_section", curRoomDef.LocalName())))
	lines = append(lines, DimStyle.Render(i18n.T("rooms.psu_help")))
	if cur := a.state.Rooms[a.state.CurrentRoom]; cur != nil {
		now := time.Now().Unix()
		load := a.state.RoomPSULoad(a.state.CurrentRoom)
		capa := a.state.RoomPSUCapacity(a.state.CurrentRoom)
		eff := a.state.RoomPSUEfficiency(a.state.CurrentRoom)
		over := a.state.RoomPSUOverloadFactor(a.state.CurrentRoom)
		lines = append(lines, DimStyle.Render(i18n.T("rooms.psu_aggregates", load, capa, eff, over)))
		if a.state.IsRoomPSUPaused(a.state.CurrentRoom, now) {
			remain := cur.PSUResumeAt - now
			if remain < 0 {
				remain = 0
			}
			lines = append(lines, DimStyle.Render(i18n.T("rooms.psu_paused", remain)))
		}
		if len(cur.PSUUnits) == 0 {
			lines = append(lines, DimStyle.Render(i18n.T("rooms.psu_empty")))
		}
		for i, p := range cur.PSUUnits {
			def, _ := data.PSUByID(p.DefID)
			marker := "  "
			if i == a.psuCursor && a.roomsFocus == 1 {
				marker = TitleStyle.Render("▶ ")
			}
			row := i18n.T("rooms.psu_row", def.LocalName(), def.Efficiency, def.HeatOutput, p.Status, def.RatedPower)
			style := lipgloss.NewStyle()
			if p.Status == "broken" {
				style = lipgloss.NewStyle().Foreground(CrisisRed)
			} else if p.DefID == "psu_builtin" {
				style = DimStyle
			}
			lines = append(lines, marker+style.Render(row))
		}
	}

	// Inline PSU picker — appended below the panel body when active.
	if a.psuPickerActive {
		lines = append(lines, "")
		lines = append(lines, a.renderPSUPicker()...)
	}

	return PanelStyle.Width(fitWidth(100, a.w)).Render(strings.Join(lines, "\n"))
}

// renderPSUPicker returns the inline picker rows. Cursor highlights the
// selected row; flavor text shows on the highlighted row only. Unaffordable
// rows render dim so the player can see what they can't buy.
func (a App) renderPSUPicker() []string {
	titleKey := "psu.picker.title.install"
	if a.psuPickerMode == "replace" {
		titleKey = "psu.picker.title.replace"
	}
	out := []string{
		HeaderStyle.Render(i18n.T(titleKey)),
		DimStyle.Render(i18n.T("psu.picker.help")),
	}
	cat := installablePSUs()
	for i, def := range cat {
		marker := "  "
		if i == a.psuPickerCursor {
			marker = TitleStyle.Render("▶ ")
		}
		priceCell := game.FmtBTCInt(def.Price)
		row := i18n.T("psu.picker.row",
			def.LocalName(), priceCell, def.Efficiency, def.RatedPower, def.HeatOutput, def.OverloadTolerance)
		style := lipgloss.NewStyle()
		if a.state.BTC < float64(def.Price) {
			style = DimStyle
		}
		out = append(out, marker+style.Render(row))
		if i == a.psuPickerCursor {
			flavor := lipgloss.NewStyle().Foreground(AccentPurple).Italic(true).Render("    " + def.LocalFlavor())
			out = append(out, flavor)
		}
	}
	return out
}

// currentRoomPSUs returns the PSUs installed in the player's current room,
// or nil if the room has no state.
func (a App) currentRoomPSUs() []*game.PSU {
	if cur := a.state.Rooms[a.state.CurrentRoom]; cur != nil {
		return cur.PSUUnits
	}
	return nil
}

func (a App) handleRoomsKey(key string) (tea.Model, tea.Cmd) {
	rooms := data.Rooms()

	// PSU picker intercepts up/down/enter/esc while it's active. Universal
	// nav keys (1–9, ?, etc.) already short-circuited in handleKey.
	if a.psuPickerActive {
		cat := installablePSUs()
		switch key {
		case "up", "k":
			if a.psuPickerCursor > 0 {
				a.psuPickerCursor--
			}
			return a, nil
		case "down", "j":
			if a.psuPickerCursor < len(cat)-1 {
				a.psuPickerCursor++
			}
			return a, nil
		case "esc":
			a.psuPickerActive = false
			return a, nil
		case "enter":
			if a.psuPickerCursor >= len(cat) {
				a.psuPickerActive = false
				return a, nil
			}
			def := cat[a.psuPickerCursor]
			switch a.psuPickerMode {
			case "install":
				if err := a.state.InstallPSU(a.state.CurrentRoom, def.ID); err != nil {
					a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
				} else {
					a = a.withStatus(i18n.T("status.psu_installed", def.LocalName()))
					a.psuPickerActive = false
				}
			case "replace":
				psus := a.currentRoomPSUs()
				if a.psuCursor >= len(psus) {
					a = a.withStatus(i18n.T("status.error_prefix") + "no PSU selected")
					return a, nil
				}
				inst := psus[a.psuCursor]
				if err := a.state.ReplacePSU(a.state.CurrentRoom, inst.InstanceID, def.ID); err != nil {
					a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
				} else {
					a = a.withStatus(i18n.T("status.psu_replaced", def.LocalName()))
					a.psuPickerActive = false
				}
			}
			return a, nil
		}
		// Other keys are ignored while the picker holds focus.
		return a, nil
	}

	switch key {
	case "tab":
		// Toggle focus between the rooms list and the PSU list. PSU focus
		// is only meaningful if the current room has at least one PSU
		// instance — fall back to rooms-list focus otherwise so cursor keys
		// don't silently no-op.
		if a.roomsFocus == 0 && len(a.currentRoomPSUs()) > 0 {
			a.roomsFocus = 1
			if a.psuCursor >= len(a.currentRoomPSUs()) {
				a.psuCursor = 0
			}
		} else {
			a.roomsFocus = 0
		}
	case "up", "k":
		if a.roomsFocus == 1 {
			if a.psuCursor > 0 {
				a.psuCursor--
			}
		} else if a.roomsCursor > 0 {
			a.roomsCursor--
		}
	case "down", "j":
		if a.roomsFocus == 1 {
			if n := len(a.currentRoomPSUs()); a.psuCursor < n-1 {
				a.psuCursor++
			}
		} else if a.roomsCursor < len(rooms)-1 {
			a.roomsCursor++
		}
	case "u":
		if a.roomsFocus == 0 && a.roomsCursor < len(rooms) {
			sel := rooms[a.roomsCursor]
			if err := a.state.UnlockRoom(sel.ID); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.unlocked", sel.LocalName()))
			}
		}
	case "enter":
		if a.roomsFocus == 0 && a.roomsCursor < len(rooms) {
			sel := rooms[a.roomsCursor]
			if err := a.state.SwitchRoom(sel.ID); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.now_in", sel.LocalName()))
				// New room: collapse PSU focus so cursor doesn't dangle on
				// stale instances from the previous room.
				a.roomsFocus = 0
				a.psuCursor = 0
			}
		}
	case "i":
		// Install picker — available regardless of focus.
		a.psuPickerActive = true
		a.psuPickerMode = "install"
		a.psuPickerCursor = 0
	case "r":
		// Replace picker — needs a PSU under the psuCursor in the current
		// room. The engine itself accepts replacing psu_builtin, so we
		// don't filter here.
		psus := a.currentRoomPSUs()
		if len(psus) == 0 {
			a = a.withStatus(i18n.T("status.error_prefix") + "no PSU to replace")
			return a, nil
		}
		if a.psuCursor >= len(psus) {
			a.psuCursor = 0
		}
		a.psuPickerActive = true
		a.psuPickerMode = "replace"
		a.psuPickerCursor = 0
	case "x":
		psus := a.currentRoomPSUs()
		if len(psus) == 0 {
			a = a.withStatus(i18n.T("status.error_prefix") + "no PSU to remove")
			return a, nil
		}
		if a.psuCursor >= len(psus) {
			a.psuCursor = 0
		}
		inst := psus[a.psuCursor]
		refund, err := a.state.RemovePSU(a.state.CurrentRoom, inst.InstanceID)
		if err != nil {
			a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
		} else {
			a = a.withStatus(i18n.T("status.psu_removed", game.FmtBTCInt(refund)))
			// Clamp cursor in case we removed the last entry.
			if remaining := len(a.currentRoomPSUs()); a.psuCursor >= remaining {
				if remaining == 0 {
					a.psuCursor = 0
					a.roomsFocus = 0
				} else {
					a.psuCursor = remaining - 1
				}
			}
		}
	case "l", "c", "w", "o", "a":
		if a.roomsFocus != 0 {
			break
		}
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
