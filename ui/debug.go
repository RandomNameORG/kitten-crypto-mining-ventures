package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// debugState holds runtime-debug affordances. It's only populated when
// cmd/meowmine is started with --debug. Everything here is intentionally
// local-only: the SSH server never calls EnableDebug, so cheats and time
// acceleration can't reach remote sessions.
type debugState struct {
	enabled       bool
	hudVisible    bool
	timeMult      int   // 1, 4, 16, 64 — applied to virtual-time advance per tick
	virtualOffset int64 // accumulated extra seconds beyond wall clock
	lastDumpPath  string
}

// timeMultCycle returns the next value in the rotation.
func timeMultCycle(cur int) int {
	switch cur {
	case 1:
		return 4
	case 4:
		return 16
	case 16:
		return 64
	default:
		return 1
	}
}

// EnableDebug flips the app into debug mode. Idempotent.
func (a *App) EnableDebug() {
	a.debug.enabled = true
	a.debug.hudVisible = true
	if a.debug.timeMult == 0 {
		a.debug.timeMult = 1
	}
}

// handleDebugKey routes a keypress to a debug action. Returns (updatedModel,
// cmd, handled) — when handled=false, the caller should fall through to the
// normal key pipeline. Only called when debug mode is enabled.
func (a App) handleDebugKey(k tea.KeyMsg) (App, tea.Cmd, bool) {
	switch k.String() {
	case "ctrl+f":
		a.debug.timeMult = timeMultCycle(a.debug.timeMult)
		a = a.withStatus(fmt.Sprintf("⏩ debug: sim speed ×%d", a.debug.timeMult))
		return a, nil, true
	case "ctrl+d":
		path, err := a.dumpDebugState()
		if err != nil {
			a = a.withStatus("debug dump failed: " + err.Error())
		} else {
			a.debug.lastDumpPath = path
			a = a.withStatus("📸 debug dump → " + path)
		}
		return a, nil, true
	case "ctrl+b":
		a.debug.hudVisible = !a.debug.hudVisible
		return a, nil, true
	case "ctrl+y":
		// Cheat: add 1 BTC. `ctrl+m` is equivalent to Enter in many
		// terminals, which would conflict with normal input — use ctrl+y
		// instead ("yarn").
		a.state.BTC += 1
		a = a.withStatus("🐾 debug: +₿1")
		return a, nil, true
	case "ctrl+t":
		a.state.TechPoint += 10
		a = a.withStatus("🧠 debug: +10 TP")
		return a, nil, true
	}
	return a, nil, false
}

// dumpDebugState writes the full state as JSON to /tmp with a timestamped
// filename and returns the path. Uses SaveAs under the hood so it matches the
// on-disk save format exactly.
func (a App) dumpDebugState() (string, error) {
	path := filepath.Join(os.TempDir(), fmt.Sprintf("meowmine-debug-%d.json", time.Now().Unix()))
	if err := a.state.SaveAs(path); err != nil {
		return "", err
	}
	return path, nil
}

// debugHUDLine renders the one-line debug HUD. Returns "" when disabled or
// hidden so callers can concat unconditionally.
func (a App) debugHUDLine() string {
	if !a.debug.enabled || !a.debug.hudVisible {
		return ""
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff8800"))
	info := fmt.Sprintf(
		"🛠 debug  tick=%d  mult=×%d  offset=+%ds  BTC=%.2f  TP=%d  mods=%d",
		a.state.LastTickUnix,
		a.debug.timeMult,
		a.debug.virtualOffset,
		a.state.BTC,
		a.state.TechPoint,
		len(a.state.Modifiers),
	)
	if a.debug.lastDumpPath != "" {
		info += "  dump=" + a.debug.lastDumpPath
	}
	return style.Render(info)
}

