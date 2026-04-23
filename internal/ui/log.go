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
		HeaderStyle.Render("Views"),
		KeyHint.Render("[1]") + "  dashboard — GPU rack + live event log",
		KeyHint.Render("[2]") + "  store — buy new GPUs (shipping delay)",
		KeyHint.Render("[3]") + "  your GPUs — upgrade · repair · scrap",
		KeyHint.Render("[4]") + "  rooms — unlock · switch · defense upgrades",
		KeyHint.Render("[5]") + "  skills — spend TechPoints",
		KeyHint.Render("[6]") + "  log — full history",
		KeyHint.Render("[7]") + "  mercs — hire · fire · bribe",
		KeyHint.Render("[8]") + "  lab — research custom MEOWCore GPUs",
		KeyHint.Render("[9]") + "  prestige — retire & buy legacy perks",
		"",
		HeaderStyle.Render("Global"),
		KeyHint.Render("[space]") + "  pause / resume",
		KeyHint.Render("[s]") + "       save (dashboard only — other views reuse 's')",
		KeyHint.Render("[p]") + "       Pump & Dump ability (dashboard, if unlocked)",
		KeyHint.Render("[q]") + "       quit (auto-saves)",
		"",
		HeaderStyle.Render("Room defense (from rooms view)"),
		DimStyle.Render("[l] lock · [c] CCTV · [w] wiring · [o] cooling · [a] armor"),
		"",
		DimStyle.Render("Tip: it's an incremental game — feel free to leave it running in tmux."),
		DimStyle.Render("Offline progress catches up on relaunch (capped at 8h)."),
	}
	return PanelStyle.Width(70).Render(strings.Join(lines, "\n"))
}
