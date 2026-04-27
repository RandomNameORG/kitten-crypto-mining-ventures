package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/game"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"
)

// statsCategoryOrder pins the display order of event categories so the panel
// doesn't reshuffle on every render. Matches CategoryStyle's switch arms.
var statsCategoryOrder = []string{"info", "opportunity", "social", "threat", "crisis"}

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

	return PanelStyle.Width(fitWidth(80, a.w)).Render(strings.Join(lines, "\n"))
}
