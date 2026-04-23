package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/game"
)

// boostCombos enumerates all 3 valid 2-of-3 boost picks.
var boostCombos = [][]string{
	{"efficiency", "undervolt"},
	{"efficiency", "durability"},
	{"undervolt", "durability"},
}

func (a App) renderLabView() string {
	header := TitleStyle.Render("🔬 Lab — Custom MEOWCore Research")
	if !a.state.HasUnlock("rd") {
		return strings.Join([]string{
			header, "",
			DimStyle.Render("R&D is locked. Unlock 'MEOWCore Blueprint' in the Engineer skill lane first."),
		}, "\n")
	}
	help := DimStyle.Render("[t] cycle tier   [b] cycle boost combo   [r] start research   [↑/↓] select blueprint   [p] print   [esc]/[1] back")

	// Active research bar.
	active := []string{HeaderStyle.Render("Active research")}
	if a.state.ActiveResearch == nil {
		active = append(active, DimStyle.Render("  (none)"))
	} else {
		ar := a.state.ActiveResearch
		pct := a.state.ResearchProgress()
		bar := progressBar(pct, 30)
		active = append(active, fmt.Sprintf("  tier %d · %s · %s %d%%",
			ar.BlueprintTier, strings.Join(ar.Boosts, "+"), bar, int(pct*100)))
	}

	// Plan next research.
	tiers := game.ResearchTiers()
	var curTier *game.ResearchTierInfo
	for i := range tiers {
		if tiers[i].Tier == a.labTier {
			curTier = &tiers[i]
			break
		}
	}
	combo := boostCombos[a.labBoostCombo()%len(boostCombos)]
	plan := []string{HeaderStyle.Render("Plan next research")}
	if curTier != nil {
		plan = append(plan, fmt.Sprintf("  Tier %d — %s", curTier.Tier, curTier.Name))
		plan = append(plan, DimStyle.Render(fmt.Sprintf("  costs: $%d + %d frags  ·  duration: %dm",
			curTier.Money, curTier.Frags, curTier.Duration/60)))
		plan = append(plan, fmt.Sprintf("  boosts: %s + %s", combo[0], combo[1]))
		plan = append(plan, DimStyle.Render("  (press [r] to start)"))
	}

	// Blueprints.
	bpLines := []string{HeaderStyle.Render(fmt.Sprintf("Blueprints (%d) — [p] to print selected", len(a.state.Blueprints)))}
	if len(a.state.Blueprints) == 0 {
		bpLines = append(bpLines, DimStyle.Render("  (none researched yet)"))
	}
	for i, bp := range a.state.Blueprints {
		marker := "  "
		if i == a.labCursor {
			marker = TitleStyle.Render("▶ ")
		}
		eff, pow, heat, dur := game.BlueprintStats(bp)
		bpLines = append(bpLines, fmt.Sprintf("%s[%s] tier %d  %s",
			marker, bp.ID, bp.Tier, strings.Join(bp.Boosts, "+")))
		bpLines = append(bpLines, DimStyle.Render(fmt.Sprintf("    eff %.4f ₿/s · %.0fV · %.0f°C · %.0fh durability",
			eff, pow, heat, dur)))
	}

	panel1 := PanelStyle.Width(90).Render(strings.Join(active, "\n"))
	panel2 := PanelStyle.Width(90).Render(strings.Join(plan, "\n"))
	panel3 := PanelStyle.Width(90).Render(strings.Join(bpLines, "\n"))

	return strings.Join([]string{
		header, help, "",
		lipgloss.JoinVertical(lipgloss.Left, panel1, panel2, panel3),
	}, "\n")
}

// labBoostCombo is stored in labBoost1 (we reuse the field as the combo index).
func (a App) labBoostCombo() int {
	return a.labBoost1
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
			a = a.withStatus("❌ " + err.Error())
		} else {
			a = a.withStatus("🔬 research started")
		}
	case "p":
		if a.labCursor < len(a.state.Blueprints) {
			bp := a.state.Blueprints[a.labCursor]
			if err := a.state.PrintMEOWCore(bp.ID); err != nil {
				a = a.withStatus("❌ " + err.Error())
			} else {
				a = a.withStatus("🛠 printed MEOWCore")
			}
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}

// progressBar renders a simple ascii progress bar.
func progressBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	filled := int(float64(width) * pct)
	s := "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
	return s
}
