package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
)

// catalog returns the list of GPUs visible in the store — simple tier gate
// by player money (don't show a $60k card when you have $200).
func storeCatalog(money float64) []data.GPUDef {
	all := data.GPUs()
	out := []data.GPUDef{}
	for _, g := range all {
		// Reveal a card once the player has at least 15% of its price.
		if money >= float64(g.Price)*0.15 || g.Tier == "trash" || g.Tier == "common" {
			out = append(out, g)
		}
	}
	return out
}

func (a App) renderStore() string {
	cat := storeCatalog(a.state.Money)
	lines := []string{TitleStyle.Render("🛒 Store  ·  Shipping: ~30–180s")}
	lines = append(lines, DimStyle.Render("↑/↓ select   [b] buy   [esc]/[1] back"))
	lines = append(lines, "")

	for i, g := range cat {
		marker := "  "
		if i == a.storeCursor {
			marker = TitleStyle.Render("▶ ")
		}
		affordable := a.state.Money >= float64(g.Price)
		priceStyle := MoneyStyle
		if !affordable {
			priceStyle = DimStyle
		}
		name := g.Name
		if !affordable {
			name = DimStyle.Render(name)
		}
		lines = append(lines, fmt.Sprintf("%s%-32s  %s  %s",
			marker,
			name,
			priceStyle.Render(fmt.Sprintf("$%-6d", g.Price)),
			DimStyle.Render(fmt.Sprintf("eff %.4f ₿/s   %.0fV   %dh", g.Efficiency, g.PowerDraw, g.DurabilityHours)),
		))
	}

	lines = append(lines, "")
	if a.storeCursor < len(cat) {
		sel := cat[a.storeCursor]
		lines = append(lines, lipgloss.NewStyle().Foreground(AccentPurple).Italic(true).Render("  "+sel.Flavor))
	}
	return PanelStyle.Width(90).Render(strings.Join(lines, "\n"))
}

func (a App) handleStoreKey(key string) (tea.Model, tea.Cmd) {
	cat := storeCatalog(a.state.Money)
	switch key {
	case "up", "k":
		if a.storeCursor > 0 {
			a.storeCursor--
		}
	case "down", "j":
		if a.storeCursor < len(cat)-1 {
			a.storeCursor++
		}
	case "b", "enter":
		if a.storeCursor < len(cat) {
			sel := cat[a.storeCursor]
			if err := a.state.BuyGPU(sel.ID); err != nil {
				a = a.withStatus("❌ " + err.Error())
			} else {
				a = a.withStatus(fmt.Sprintf("📦 Ordered %s", sel.Name))
			}
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
