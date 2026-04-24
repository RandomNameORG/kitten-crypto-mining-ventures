package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/game"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

// ocLevelPercent maps a GPU's OC level to the earn-boost percentage shown in
// UI markers. Zero means the OC is off; the UI then suppresses the marker.
func ocLevelPercent(level int) int {
	switch level {
	case 1:
		return 25
	case 2:
		return 50
	default:
		return 0
	}
}

// gpuDisplayName returns the localized name for any GPU (catalog or MEOWCore).
func gpuDisplayName(s *game.State, g *game.GPU) string {
	if g.BlueprintID != "" {
		if bp := s.BlueprintByID(g.BlueprintID); bp != nil {
			return fmt.Sprintf("MEOWCore v%d [%s]", bp.Tier, strings.Join(bp.Boosts, "+"))
		}
		return "MEOWCore"
	}
	if def, ok := data.GPUByID(g.DefID); ok {
		return def.LocalName()
	}
	return g.DefID
}

// sortModeLabel translates a gpuSortMode to its localized display name.
func sortModeLabel(m gpuSortMode) string {
	switch m {
	case gpuSortEarnDesc:
		return i18n.T("gpus.sort_earn")
	case gpuSortEffDesc:
		return i18n.T("gpus.sort_eff")
	case gpuSortDurAsc:
		return i18n.T("gpus.sort_dur")
	}
	return i18n.T("gpus.sort_default")
}

func (a App) renderGPUsView() string {
	gpus, metrics, ranks := prepareGPUView(a.state, a.state.GPUs, a.gpusSortMode)
	lines := []string{TitleStyle.Render(i18n.T("gpus.title"))}
	lines = append(lines, DimStyle.Render(i18n.T("gpus.help")))
	lines = append(lines, DimStyle.Render(i18n.T("gpus.sort_label", sortModeLabel(a.gpusSortMode))))
	lines = append(lines, "")

	if len(gpus) == 0 {
		lines = append(lines, DimStyle.Render(i18n.T("gpus.empty")))
	}
	for i, g := range gpus {
		marker := "  "
		if i == a.gpusCursor {
			marker = TitleStyle.Render("▶ ")
		}
		roomDef, _ := data.RoomByID(g.Room)
		roomName := roomDef.LocalName()
		if roomName == "" {
			roomName = g.Room
		}
		statusDecor := g.Status
		if g.Status == "shipping" {
			statusDecor = "shipping…"
		}
		upMark := ""
		if g.UpgradeLevel > 0 {
			upMark = fmt.Sprintf(" +%d", g.UpgradeLevel)
		}
		ocMark := ""
		if pct := ocLevelPercent(g.OCLevel); pct > 0 {
			ocMark = lipgloss.NewStyle().Foreground(ThreatOrange).Render(fmt.Sprintf(i18n.T("gpus.oc_mark"), pct))
		}
		m := metrics[g.InstanceID]
		tier := ranks[g.InstanceID]

		rateCell := DimStyle.Render("—")
		if m.running && m.earn > 0 {
			rateCell = lipgloss.NewStyle().Foreground(rankColour(tier)).Render(game.FmtBTC(m.earn) + "/s")
		}
		powCell := DimStyle.Render("  —  ")
		if m.running {
			powCell = fmt.Sprintf("%4.0fW", m.power)
		}
		effCell := DimStyle.Render("    —    ")
		if m.running && m.power > 0 {
			effCell = DimStyle.Render(game.FmtBTC(m.eff) + "/W")
		}

		line := fmt.Sprintf("%s#%-3d %-30s%s%s  %-11s  %-12s  %5.1fh  %5s  %s  %s",
			marker,
			g.InstanceID,
			gpuDisplayName(a.state, g),
			upMark,
			ocMark,
			statusDecor,
			roomName,
			g.HoursLeft,
			powCell,
			rateCell,
			effCell,
		)
		lines = append(lines, line)
	}
	return PanelStyle.Width(fitWidth(110, a.w)).Render(strings.Join(lines, "\n"))
}

func (a App) handleGPUsKey(key string) (tea.Model, tea.Cmd) {
	// Work from the currently displayed (sorted) slice so cursor positions
	// and `gpus[cursor]` pick the right GPU even with a non-default sort.
	gpus, _, _ := prepareGPUView(a.state, a.state.GPUs, a.gpusSortMode)
	switch key {
	case "up", "k":
		if a.gpusCursor > 0 {
			a.gpusCursor--
		}
	case "down", "j":
		if a.gpusCursor < len(gpus)-1 {
			a.gpusCursor++
		}
	case "u":
		if a.gpusCursor < len(gpus) {
			if err := a.state.UpgradeGPU(gpus[a.gpusCursor].InstanceID); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.upgrade"))
			}
		}
	case "o":
		if a.gpusCursor < len(gpus) {
			if err := a.state.CycleGPUOC(gpus[a.gpusCursor].InstanceID); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			}
		}
	case "r":
		if a.gpusCursor < len(gpus) {
			if err := a.state.RepairGPU(gpus[a.gpusCursor].InstanceID); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.repaired"))
			}
		}
	case "s":
		if a.gpusCursor < len(gpus) {
			if err := a.state.SellGPU(gpus[a.gpusCursor].InstanceID); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.sold"))
				if a.gpusCursor > 0 {
					a.gpusCursor--
				}
			}
		}
	case "b":
		// Remember the anchor GPU under the cursor so the new sort keeps
		// it highlighted instead of jumping to whatever lands at the same
		// index. If the row is out of bounds (empty list), skip the anchor.
		var anchorID int = -1
		if a.gpusCursor >= 0 && a.gpusCursor < len(gpus) {
			anchorID = gpus[a.gpusCursor].InstanceID
		}
		a.gpusSortMode = cycleGPUSortMode(a.gpusSortMode)
		if anchorID >= 0 {
			resorted, _, _ := prepareGPUView(a.state, a.state.GPUs, a.gpusSortMode)
			if idx := indexOfGPU(resorted, anchorID); idx >= 0 {
				a.gpusCursor = idx
			}
		}
		a = a.withStatus(i18n.T("gpus.sort_label", sortModeLabel(a.gpusSortMode)))
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
