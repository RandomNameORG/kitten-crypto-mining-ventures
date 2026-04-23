package ui

import (
	"fmt"
	"strings"
	"time"
)

func (a App) renderLogView() string {
	log := a.state.Log
	lines := []string{TitleStyle.Render("📜 Full Event Log")}
	lines = append(lines, DimStyle.Render("[esc]/[1] back"))
	lines = append(lines, "")
	if len(log) == 0 {
		lines = append(lines, DimStyle.Render("  (empty)"))
	}
	// Print most recent first.
	for i := len(log) - 1; i >= 0 && i > len(log)-50; i-- {
		e := log[i]
		ts := time.Unix(e.Time, 0).Format("15:04:05")
		line := fmt.Sprintf("  %s  %s", DimStyle.Render(ts), CategoryStyle(e.Category).Render(e.Text))
		lines = append(lines, line)
	}
	return PanelStyle.Width(100).Render(strings.Join(lines, "\n"))
}

func (a App) renderHelpView() string {
	lines := []string{
		TitleStyle.Render("🐾 Help"),
		"",
		KeyHint.Render("[1]") + "  dashboard",
		KeyHint.Render("[2]") + "  store — buy new GPUs (shipping delay)",
		KeyHint.Render("[3]") + "  your GPUs — upgrade, repair, scrap",
		KeyHint.Render("[4]") + "  rooms — unlock & switch biomes",
		KeyHint.Render("[5]") + "  skills — plan your build",
		KeyHint.Render("[6]") + "  log — full history",
		"",
		KeyHint.Render("[space]") + "  pause / resume",
		KeyHint.Render("[s]") + "       save now",
		KeyHint.Render("[q]") + "       quit (auto-saves)",
		"",
		DimStyle.Render("Tip: it's an incremental game. Feel free to leave it running."),
		DimStyle.Render("Offline progress accumulates while the save is idle."),
	}
	return PanelStyle.Width(60).Render(strings.Join(lines, "\n"))
}
