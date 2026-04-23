package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
)

func (a App) renderGPUsView() string {
	gpus := a.state.GPUs
	lines := []string{TitleStyle.Render("🖥  Your GPUs")}
	lines = append(lines, DimStyle.Render("↑/↓ select   [u] upgrade   [r] repair   [s] scrap   [esc]/[1] back"))
	lines = append(lines, "")

	if len(gpus) == 0 {
		lines = append(lines, DimStyle.Render("  (no GPUs yet — visit the store)"))
	}
	for i, g := range gpus {
		marker := "  "
		if i == a.gpusCursor {
			marker = TitleStyle.Render("▶ ")
		}
		def, _ := data.GPUByID(g.DefID)
		roomDef, _ := data.RoomByID(g.Room)
		statusDecor := g.Status
		if g.Status == "shipping" {
			statusDecor = "shipping…"
		}
		upMark := ""
		if g.UpgradeLevel > 0 {
			upMark = fmt.Sprintf(" +%d", g.UpgradeLevel)
		}
		line := fmt.Sprintf("%s#%-3d %-30s%s  %-12s  %-18s  durab %.1fh",
			marker,
			g.InstanceID,
			def.Name,
			upMark,
			statusDecor,
			roomDef.Name,
			g.HoursLeft,
		)
		lines = append(lines, line)
	}
	return PanelStyle.Width(100).Render(strings.Join(lines, "\n"))
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
				a = a.withStatus("❌ " + err.Error())
			} else {
				a = a.withStatus("⚙️  upgrade attempted")
			}
		}
	case "r":
		if a.gpusCursor < len(gpus) {
			if err := a.state.RepairGPU(gpus[a.gpusCursor].InstanceID); err != nil {
				a = a.withStatus("❌ " + err.Error())
			} else {
				a = a.withStatus("🔧 repaired")
			}
		}
	case "s":
		if a.gpusCursor < len(gpus) {
			if err := a.state.SellGPU(gpus[a.gpusCursor].InstanceID); err != nil {
				a = a.withStatus("❌ " + err.Error())
			} else {
				a = a.withStatus("💵 sold")
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
