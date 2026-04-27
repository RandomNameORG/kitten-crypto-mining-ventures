package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/game"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"
)

// statsCategoryOrder pins the display order of event categories so the panel
// doesn't reshuffle on every render. Matches CategoryStyle's switch arms.
var statsCategoryOrder = []string{"info", "opportunity", "social", "threat", "crisis"}

// settlementTag maps the engine's settlement_mode string to the uppercase
// glanceable tag rendered in the Pool panel and picker rows.
func settlementTag(mode string) string {
	switch mode {
	case "pps":
		return "PPS"
	case "pplns":
		return "PPLNS"
	case "pps_plus":
		return "PPS+"
	case "solo":
		return "SOLO"
	}
	return strings.ToUpper(mode)
}

func (a App) renderStatsView() string {
	s := a.state
	lines := []string{
		TitleStyle.Render(i18n.T("stats.title")),
		DimStyle.Render(i18n.T("stats.help")),
		"",
	}

	row := func(label, value string) string {
		return fmt.Sprintf("  %-26s  %s", label, value)
	}

	lines = append(lines,
		row(i18n.T("stats.row.lifetime"), BTCStyle.Render(game.FmtBTC(s.LifetimeEarned))),
		row(i18n.T("stats.row.ticks"), fmt.Sprintf("%d", s.TotalTicks)),
		row(i18n.T("stats.row.market"), fmt.Sprintf("%.2f×", s.MarketPrice)),
	)

	spark := Sparkline(s.MarketPriceHistory)
	if spark == "" {
		spark = DimStyle.Render(i18n.T("stats.empty_history"))
	} else {
		sparkColor := AccentPurple
		if time.Now().Before(a.statsPulseUntil) {
			sparkColor = BTCGreen
		}
		spark = lipgloss.NewStyle().Foreground(sparkColor).Render(spark)
	}
	lines = append(lines, row(i18n.T("stats.row.spark"), spark))

	lines = append(lines,
		row(i18n.T("stats.row.gpus"), fmt.Sprintf("%d / %d", s.TotalGPUsBought, s.TotalGPUsScrapped)),
		row(i18n.T("stats.row.oc_t1"), formatDuration(s.OCTimeT1Sec)),
		row(i18n.T("stats.row.oc_t2"), formatDuration(s.OCTimeT2Sec)),
		row(i18n.T("stats.row.wages"), BTCStyle.Render(game.FmtBTC(s.TotalWagesPaid))),
		"",
		HeaderStyle.Render(i18n.T("stats.row.events")),
	)

	if len(s.EventsByCategory) == 0 {
		lines = append(lines, DimStyle.Render("  "+i18n.T("stats.empty_events")))
	} else {
		seen := map[string]bool{}
		for _, cat := range statsCategoryOrder {
			seen[cat] = true
			n, ok := s.EventsByCategory[cat]
			if !ok {
				continue
			}
			line := fmt.Sprintf("  %-12s  %d", CategoryStyle(cat).Render(cat), n)
			lines = append(lines, line)
		}
		// Surface any unexpected categories so future-added kinds still show
		// up rather than getting silently swallowed.
		extras := []string{}
		for k := range s.EventsByCategory {
			if !seen[k] {
				extras = append(extras, k)
			}
		}
		sort.Strings(extras)
		for _, k := range extras {
			line := fmt.Sprintf("  %-12s  %d", CategoryStyle(k).Render(k), s.EventsByCategory[k])
			lines = append(lines, line)
		}
	}

	// Mining Pool section — Sprint 2.
	now := time.Now().Unix()
	cur := s.CurrentPool()
	lines = append(lines,
		"",
		HeaderStyle.Render(i18n.T("pool.section")),
		DimStyle.Render(i18n.T("pool.help")),
	)
	lines = append(lines, "  "+i18n.T("pool.current",
		cur.LocalName(), settlementTag(cur.SettlementMode), s.PoolFee()*100, cur.Risk))
	lines = append(lines, DimStyle.Render("  "+i18n.T("pool.shares", s.PoolShares)))
	if s.IsPoolSwitching(now) {
		fromName := s.PoolSwitchFrom
		if def, ok := data.PoolByID(s.PoolSwitchFrom); ok {
			fromName = def.LocalName()
		}
		remain := s.PoolSwitchAt - now
		if remain < 0 {
			remain = 0
		}
		lines = append(lines, DimStyle.Render("  "+i18n.T("pool.switching", fromName, cur.LocalName(), remain)))
	}

	if a.poolPickerActive {
		lines = append(lines, "")
		lines = append(lines, a.renderPoolPicker()...)
	}

	return PanelStyle.Width(fitWidth(80, a.w)).Render(strings.Join(lines, "\n"))
}

// renderPoolPicker draws the inline pool picker rows. Cursor row shows the
// flavor text (same UX as the PSU picker). The player's current pool is
// rendered dim so a no-op switch is visually pre-disabled — the engine
// already errors on it but graying it out makes intent clear.
func (a App) renderPoolPicker() []string {
	out := []string{
		HeaderStyle.Render(i18n.T("pool.picker.title")),
		DimStyle.Render(i18n.T("pool.picker.help")),
	}
	pools := data.Pools()
	for i, def := range pools {
		marker := "  "
		if i == a.poolPickerCursor {
			marker = TitleStyle.Render("▶ ")
		}
		row := i18n.T("pool.picker.row",
			def.LocalName(), def.Fee*100, settlementTag(def.SettlementMode), def.Risk)
		style := lipgloss.NewStyle()
		if def.ID == a.state.PoolID {
			style = DimStyle
		}
		out = append(out, marker+style.Render(row))
		if i == a.poolPickerCursor {
			flavor := lipgloss.NewStyle().Foreground(AccentPurple).Italic(true).Render("    " + def.LocalFlavor())
			out = append(out, flavor)
		}
	}
	return out
}

func (a App) handleStatsKey(key string) (tea.Model, tea.Cmd) {
	pools := data.Pools()

	if a.poolPickerActive {
		switch key {
		case "up", "k":
			if a.poolPickerCursor > 0 {
				a.poolPickerCursor--
			}
			return a, nil
		case "down", "j":
			if a.poolPickerCursor < len(pools)-1 {
				a.poolPickerCursor++
			}
			return a, nil
		case "esc":
			a.poolPickerActive = false
			return a, nil
		case "enter":
			if a.poolPickerCursor >= len(pools) {
				a.poolPickerActive = false
				return a, nil
			}
			def := pools[a.poolPickerCursor]
			if err := a.state.SwitchPool(def.ID, time.Now().Unix()); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.pool_switching", def.LocalName()))
				a.poolPickerActive = false
			}
			return a, nil
		}
		return a, nil
	}

	switch key {
	case "p":
		a.poolPickerActive = true
		// Land the cursor on the player's current pool so the first arrow
		// keystroke moves *away* from the no-op selection rather than past it.
		a.poolPickerCursor = 0
		for i, def := range pools {
			if def.ID == a.state.PoolID {
				a.poolPickerCursor = i
				break
			}
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
