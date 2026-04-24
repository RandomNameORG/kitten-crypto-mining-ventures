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

func (a App) renderGPUsView() string {
	gpus := a.state.GPUs
	lines := []string{TitleStyle.Render(i18n.T("gpus.title"))}
	lines = append(lines, DimStyle.Render(i18n.T("gpus.help")))
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
		rate := a.state.GPUEarnRatePerSec(g)
		rateCell := ""
		if rate > 0 {
			rateCell = lipgloss.NewStyle().Foreground(OppGreen).Render(fmt.Sprintf("₿%.3f/s", rate))
		} else {
			rateCell = DimStyle.Render("—")
		}
		line := fmt.Sprintf("%s#%-3d %-36s%s  %-12s  %-18s  durab %5.1fh  %s",
			marker,
			g.InstanceID,
			gpuDisplayName(a.state, g),
			upMark,
			statusDecor,
			roomName,
			g.HoursLeft,
			rateCell,
		)
		lines = append(lines, line)
	}
	return PanelStyle.Width(fitWidth(110, a.w)).Render(strings.Join(lines, "\n"))
}

func (a App) handleGPUsKey(key string) (tea.Model, tea.Cmd) {
	gpus := a.state.GPUs
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
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
