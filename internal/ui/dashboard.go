package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
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

	header := TitleStyle.Render(fmt.Sprintf("📍 %s", def.Name))
	lines = append(lines, header)
	lines = append(lines, DimStyle.Render(def.Flavor))
	lines = append(lines, "")

	var volt float64
	for _, g := range gpus {
		if g.Status != "running" {
			continue
		}
		_, pow, _, _ := a.state.GPUStats(g)
		volt += pow
	}
	meters := fmt.Sprintf("%s   %s   %s",
		VoltStyle.Render(fmt.Sprintf("⚡ %.0f V/s draw", volt)),
		HeatStyle.Render(fmt.Sprintf("🌡  %.0f°C", heat)),
		DimStyle.Render(fmt.Sprintf("slots %d/%d", len(gpus), def.Slots)),
	)
	lines = append(lines, meters)
	lines = append(lines, "")

	lines = append(lines, HeaderStyle.Render("GPU Rack"))
	if len(gpus) == 0 {
		lines = append(lines, DimStyle.Render("  (empty — press [2] to go to the store)"))
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
			lines = append(lines, DimStyle.Render(fmt.Sprintf("  %d. (empty)", i+1)))
		}
	}

	width := 52
	return PanelStyle.Width(width).Render(strings.Join(lines, "\n"))
}

func (a App) renderLogPanel(maxLines int) string {
	log := a.state.Log
	lines := []string{TitleStyle.Render("📜 Event Log")}

	start := 0
	if len(log) > maxLines {
		start = len(log) - maxLines
	}
	if start == len(log) {
		lines = append(lines, DimStyle.Render("  (quiet so far)"))
	}
	for i := start; i < len(log); i++ {
		entry := log[i]
		style := CategoryStyle(entry.Category)
		lines = append(lines, "  "+style.Render(truncate(entry.Text, 44)))
	}
	width := 50
	return PanelStyle.Width(width).Render(strings.Join(lines, "\n"))
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
			TitleStyle.Render(fmt.Sprintf("%s  %s", e.Emoji, e.Name)),
			"",
			wrap(e.Text, 48),
			"",
			DimStyle.Render("[press any key to dismiss]"),
		}, "\n"))

	// Simple centered overlay approximation: stack below content with a blank.
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
