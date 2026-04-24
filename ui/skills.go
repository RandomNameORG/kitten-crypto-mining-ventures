package ui

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

var laneOrder = []string{"engineer", "mogul", "hacker"}

type skillItem struct {
	lane string
	def  data.SkillDef
}

func skillList() []skillItem {
	out := []skillItem{}
	for _, lane := range laneOrder {
		for _, def := range data.SkillsByLane(lane) {
			out = append(out, skillItem{lane: lane, def: def})
		}
	}
	return out
}

func (a App) renderSkillsView() string {
	items := skillList()
	header := TitleStyle.Render(i18n.T("skills.title")) +
		"   " + DimStyle.Render(i18n.T("skills.tp_count", a.state.TechPoint))
	help := DimStyle.Render(i18n.T("skills.help"))

	laneKey := map[string]string{
		"engineer": "skills.lane.engineer",
		"mogul":    "skills.lane.mogul",
		"hacker":   "skills.lane.hacker",
	}

	// 3 cols side-by-side take ~112 chars (36*3 + gaps). Below that, stack
	// them vertically with full terminal width per column.
	var colW int
	col := func(lane string) string {
		lines := []string{TitleStyle.Render(i18n.T(laneKey[lane]))}
		for i, it := range items {
			if it.lane != lane {
				continue
			}
			cursor := "  "
			if i == a.skillsCursor {
				cursor = TitleStyle.Render("▶ ")
			}
			owned := a.state.HasSkill(it.def.ID)
			gateStyle := DimStyle
			label := it.def.LocalName()
			meta := strconv.Itoa(it.def.Cost) + " TP"
			if owned {
				label = lipgloss.NewStyle().Foreground(OppGreen).Render("✓ " + it.def.LocalName())
				meta = i18n.T("skills.owned")
			} else if it.def.Prereq != "" && !a.state.HasSkill(it.def.Prereq) {
				label = DimStyle.Render(label + i18n.T("skills.locked_suffix"))
			} else if a.state.TechPoint >= it.def.Cost {
				label = MoneyStyle.Render(label)
			}
			lines = append(lines, cursor+label+"  "+gateStyle.Render(meta))
			lines = append(lines, DimStyle.Render("   "+it.def.LocalDesc()))
		}
		return PanelStyle.Width(colW).Render(strings.Join(lines, "\n"))
	}

	var body string
	if a.w >= 112 {
		colW = 36
		body = lipgloss.JoinHorizontal(lipgloss.Top,
			col("engineer"), " ", col("mogul"), " ", col("hacker"))
	} else {
		colW = fitWidth(36, a.w)
		body = lipgloss.JoinVertical(lipgloss.Left,
			col("engineer"), col("mogul"), col("hacker"))
	}

	return strings.Join([]string{header, help, "", body}, "\n")
}

func (a App) handleSkillsKey(key string) (tea.Model, tea.Cmd) {
	items := skillList()
	switch key {
	case "up", "k":
		if a.skillsCursor > 0 {
			a.skillsCursor--
		}
	case "down", "j":
		if a.skillsCursor < len(items)-1 {
			a.skillsCursor++
		}
	case "u", "enter":
		if a.skillsCursor < len(items) {
			it := items[a.skillsCursor]
			if err := a.state.UnlockSkill(it.def.ID); err != nil {
				a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
			} else {
				a = a.withStatus("✓ " + it.def.LocalName())
			}
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
