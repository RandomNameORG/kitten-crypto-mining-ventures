package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/game"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

// boostCombos enumerates all 3 valid 2-of-3 boost picks.
var boostCombos = [][]string{
	{"efficiency", "undervolt"},
	{"efficiency", "durability"},
	{"undervolt", "durability"},
}

func (a App) renderLabView() string {
	header := TitleStyle.Render(i18n.T("lab.title"))
	if !a.state.HasUnlock("rd") {
		return strings.Join([]string{
			header, "",
			DimStyle.Render(i18n.T("lab.locked")),
		}, "\n")
	}
	help := DimStyle.Render(i18n.T("lab.help"))

	active := []string{HeaderStyle.Render(i18n.T("lab.active"))}
	if a.state.ActiveResearch == nil {
		active = append(active, DimStyle.Render(i18n.T("lab.active_none")))
	} else {
		ar := a.state.ActiveResearch
		pct := a.state.ResearchProgress()
		bar := progressBar(pct, 30)
		active = append(active, fmt.Sprintf("  tier %d · %s · %s %d%%",
			ar.BlueprintTier, strings.Join(ar.Boosts, "+"), bar, int(pct*100)))
	}

	tiers := game.ResearchTiers()
	var curTier *game.ResearchTierInfo
	for i := range tiers {
		if tiers[i].Tier == a.labTier {
			curTier = &tiers[i]
			break
		}
	}
	combo := boostCombos[a.labBoost1%len(boostCombos)]
	plan := []string{HeaderStyle.Render(i18n.T("lab.plan"))}
	if curTier != nil {
		plan = append(plan, i18n.T("lab.plan_tier", curTier.Tier, curTier.Name))
		plan = append(plan, DimStyle.Render(i18n.T("lab.plan_cost", curTier.Money, curTier.Frags, curTier.Duration/60)))
		plan = append(plan, i18n.T("lab.plan_boosts", combo[0], combo[1]))
		plan = append(plan, DimStyle.Render(i18n.T("lab.plan_hint")))
	}

	bpLines := []string{HeaderStyle.Render(i18n.T("lab.bp_title", len(a.state.Blueprints)))}
	if len(a.state.Blueprints) == 0 {
		bpLines = append(bpLines, DimStyle.Render(i18n.T("lab.bp_empty")))
	}
	for i, bp := range a.state.Blueprints {
		marker := "  "
		if i == a.labCursor {
			marker = TitleStyle.Render("▶ ")
		}
		eff, pow, heat, dur := game.BlueprintStats(bp)
		bpLines = append(bpLines, fmt.Sprintf("%s[%s] tier %d  %s",
			marker, bp.ID, bp.Tier, strings.Join(bp.Boosts, "+")))
		bpLines = append(bpLines, DimStyle.Render(i18n.T("label.bp_line", eff, pow, heat, dur)))
	}

	panel1 := PanelStyle.Width(90).Render(strings.Join(active, "\n"))
	panel2 := PanelStyle.Width(90).Render(strings.Join(plan, "\n"))
	panel3 := PanelStyle.Width(90).Render(strings.Join(bpLines, "\n"))

	return strings.Join([]string{
		header, help, "",
		lipgloss.JoinVertical(lipgloss.Left, panel1, panel2, panel3),
	}, "\n")
}

func (a App) handleLabKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "t":
		a.labTier = (a.labTier % 3) + 1
	case "b":
		a.labBoost1 = (a.labBoost1 + 1) % len(boostCombos)
	case "up", "k":
		if a.labCursor > 0 {
			a.labCursor--
		}
	case "down", "j":
		if a.labCursor < len(a.state.Blueprints)-1 {
			a.labCursor++
		}
	case "r":
		combo := boostCombos[a.labBoost1%len(boostCombos)]
		if err := a.state.StartResearch(a.labTier, combo); err != nil {
			a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
		} else {
			a = a.withStatus(i18n.T("status.research_go"))
		}
	case "p":
		if a.labCursor < len(a.state.Blueprints) {
			bp := a.state.Blueprints[a.labCursor]
			if err := a.state.PrintMEOWCore(bp.ID); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus(i18n.T("status.printed"))
			}
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}

func progressBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	filled := int(float64(width) * pct)
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}
