package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/i18n"
)

func (a App) renderDashboard() string {
	roomDef, _ := data.RoomByID(a.state.CurrentRoom)
	left := a.renderRoomPanel(roomDef)
	right := a.renderLogPanel(10)
	cols := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	return lipgloss.NewStyle().Padding(0, 1).Render(cols)
}

func (a App) renderRoomPanel(def data.RoomDef) string {
	var heat float64
	if rs := a.state.Rooms[a.state.CurrentRoom]; rs != nil {
		heat = rs.Heat
	}

	gpus := a.state.GPUsInRoom(def.ID)
	lines := []string{}

	lines = append(lines, TitleStyle.Render(i18n.T("dash.location", def.LocalName())))
	lines = append(lines, DimStyle.Render(def.LocalFlavor()))
	lines = append(lines, "")

	var volt float64
	for _, g := range gpus {
		if g.Status != "running" {
			continue
		}
		_, pow, _, _ := a.state.GPUStats(g)
		volt += pow
	}

	roomID := def.ID
	bill := a.state.RoomBillRatePerSec(roomID)
	earn := a.state.RoomEarnRatePerSec(roomID)
	net := earn - bill
	heatDelta, heatTickSec := a.state.RoomHeatDeltaPerTick(roomID)
	heatTickIn := a.state.SecondsUntilNextHeatTick(roomID)
	nextBill := a.state.SecondsUntilNextBill()

	var maxHeat float64 = 90
	if rs := a.state.Rooms[roomID]; rs != nil {
		maxHeat = rs.MaxHeat
	}

	netStyle := lipgloss.NewStyle().Foreground(OppGreen)
	if net < 0 {
		netStyle = lipgloss.NewStyle().Foreground(CrisisRed)
	}

	// Colour the heat line by danger band:
	//   >=95%  RED    — durability wears 8× normal
	//   >=80%  AMBER  — efficiency halved + 3× wear
	heatStyle := HeatStyle
	heatSuffix := ""
	if heat >= 0.95*maxHeat {
		heatStyle = lipgloss.NewStyle().Foreground(CrisisRed).Bold(true)
		heatSuffix = " " + i18n.T("dash.heat.critical")
	} else if heat >= 0.80*maxHeat {
		heatStyle = lipgloss.NewStyle().Foreground(ThreatOrange).Bold(true)
		heatSuffix = " " + i18n.T("dash.heat.warning")
	}

	lines = append(lines, fmt.Sprintf("%s   %s",
		VoltStyle.Render(i18n.T("dash.line.volt", volt, bill, nextBill)),
		DimStyle.Render(i18n.T("dash.slots_of", len(gpus), def.Slots))))
	lines = append(lines, heatStyle.Render(i18n.T("dash.line.heat", heat, maxHeat, heatDelta, heatTickSec, heatTickIn)+heatSuffix))
	lines = append(lines, netStyle.Render(i18n.T("dash.line.cash", earn, net)))
	lines = append(lines, "")

	lines = append(lines, HeaderStyle.Render(i18n.T("dash.rack")))
	if len(gpus) == 0 {
		lines = append(lines, DimStyle.Render(i18n.T("dash.empty_hint")))
	}
	for i := 0; i < def.Slots; i++ {
		if i < len(gpus) {
			g := gpus[i]
			statusIcon := "●"
			statusColor := OppGreen
			statusText := g.Status
			switch g.Status {
			case "shipping":
				statusIcon = "📦"
				statusColor = SocialCyan
				statusText = "shipping"
			case "broken":
				statusIcon = "✕"
				statusColor = CrisisRed
				statusText = "broken"
			case "stolen":
				statusIcon = "?"
				statusColor = MutedGrey
				statusText = "stolen"
			}
			indicator := lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon)
			upMark := ""
			if g.UpgradeLevel > 0 {
				upMark = lipgloss.NewStyle().Foreground(AccentPurple).Render(fmt.Sprintf(" +%d", g.UpgradeLevel))
			}
			line := fmt.Sprintf("  %d. %s %s%s  %s", i+1, indicator, gpuDisplayName(a.state, g), upMark, DimStyle.Render(statusText))
			lines = append(lines, line)
		} else {
			lines = append(lines, DimStyle.Render(fmt.Sprintf(i18n.T("dash.slot_empty"), i+1)))
		}
	}
	return PanelStyle.Width(52).Render(strings.Join(lines, "\n"))
}

func (a App) renderLogPanel(maxLines int) string {
	log := a.state.Log
	lines := []string{TitleStyle.Render(i18n.T("dash.log_title"))}

	start := 0
	if len(log) > maxLines {
		start = len(log) - maxLines
	}
	if start == len(log) {
		lines = append(lines, DimStyle.Render(i18n.T("dash.log_quiet")))
	}
	for i := start; i < len(log); i++ {
		entry := log[i]
		style := CategoryStyle(entry.Category)
		lines = append(lines, "  "+style.Render(truncate(entry.Text, 44)))
	}
	return PanelStyle.Width(50).Render(strings.Join(lines, "\n"))
}

func (a App) overlayEvent(content string) string {
	if a.showEventPopup == nil {
		return content
	}
	e := a.showEventPopup
	box := PanelStyle.
		Width(52).
		BorderForeground(KittenPink).
		Render(strings.Join([]string{
			TitleStyle.Render(fmt.Sprintf("%s  %s", e.Emoji, e.LocalName())),
			"",
			wrap(e.LocalText(), 48),
			"",
			DimStyle.Render(i18n.T("event.dismiss")),
		}, "\n"))

	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		lipgloss.NewStyle().Padding(1, 2).Render(box),
	)
}

func truncate(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n-1]) + "…"
}

func wrap(s string, width int) string {
	words := strings.Fields(s)
	var line strings.Builder
	var out []string
	for _, w := range words {
		if line.Len()+len(w)+1 > width {
			out = append(out, line.String())
			line.Reset()
		}
		if line.Len() > 0 {
			line.WriteByte(' ')
		}
		line.WriteString(w)
	}
	if line.Len() > 0 {
		out = append(out, line.String())
	}
	return strings.Join(out, "\n")
}
