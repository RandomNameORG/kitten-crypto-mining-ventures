package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

// storeBuyCooldown stops auto-repeat on a held `b` key from buying dozens of
// GPUs by accident. 400ms is long enough to block keyboard repeat (~20/s)
// but short enough that intentional repeated presses still feel snappy.
const storeBuyCooldown = 400 * time.Millisecond

func storeCatalog(money float64) []data.GPUDef {
	all := data.GPUs()
	out := []data.GPUDef{}
	for _, g := range all {
		if money >= float64(g.Price)*0.15 || g.Tier == "trash" || g.Tier == "common" {
			out = append(out, g)
		}
	}
	return out
}

func (a App) renderStore() string {
	cat := storeCatalog(a.state.Money)
	lines := []string{TitleStyle.Render(i18n.T("store.title"))}
	lines = append(lines, DimStyle.Render(i18n.T("store.help")))
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
		name := g.LocalName()
		if !affordable {
			name = DimStyle.Render(name)
		}
		lines = append(lines, fmt.Sprintf("%s%-32s  %s  %s",
			marker,
			name,
			priceStyle.Render(fmt.Sprintf("$%-6d", g.Price)),
			DimStyle.Render(fmt.Sprintf("%s   %.0fV   %dh",
				i18n.T("label.eff", g.Efficiency), g.PowerDraw, g.DurabilityHours)),
		))
	}

	lines = append(lines, "")
	if a.storeCursor < len(cat) {
		sel := cat[a.storeCursor]
		lines = append(lines, lipgloss.NewStyle().Foreground(AccentPurple).Italic(true).Render("  "+sel.LocalFlavor()))
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
			if time.Since(a.lastBuyAt) < storeBuyCooldown {
				return a, nil
			}
			a.lastBuyAt = time.Now()
			sel := cat[a.storeCursor]
			if err := a.state.BuyGPU(sel.ID); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.order", sel.LocalName()))
			}
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
