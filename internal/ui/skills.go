package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
)

var laneOrder = []string{"engineer", "mogul", "hacker"}

// skillItem is one row in the flattened list used for cursor navigation.
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
	header := TitleStyle.Render("🧠 Skill Tree") +
		"   " + DimStyle.Render(fmt.Sprintf("TP: %d", a.state.TechPoint))
	help := DimStyle.Render("↑/↓ select   [u]/[enter] unlock   [esc]/[1] back")

	laneLabel := map[string]string{
		"engineer": "🔧 Engineer",
		"mogul":    "💰 Mogul",
		"hacker":   "🕶 Hacker",
	}

	col := func(lane string) string {
		lines := []string{TitleStyle.Render(laneLabel[lane])}
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
			label := it.def.Name
			meta := fmt.Sprintf("%s TP", strconv.Itoa(it.def.Cost))
			if owned {
				label = lipgloss.NewStyle().Foreground(OppGreen).Render("✓ " + it.def.Name)
				meta = "owned"
			} else if it.def.Prereq != "" && !a.state.HasSkill(it.def.Prereq) {
				label = DimStyle.Render(label + " (locked)")
			} else if a.state.TechPoint >= it.def.Cost {
				label = MoneyStyle.Render(label)
			}
			lines = append(lines, cursor+label+"  "+gateStyle.Render(meta))
			lines = append(lines, DimStyle.Render("   "+it.def.Desc))
		}
		return PanelStyle.Width(36).Render(strings.Join(lines, "\n"))
	}

	cols := lipgloss.JoinHorizontal(lipgloss.Top,
		col("engineer"), " ", col("mogul"), " ", col("hacker"))

	return strings.Join([]string{header, help, "", cols}, "\n")
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
				a = a.withStatus("❌ " + err.Error())
			} else {
				a = a.withStatus("✓ " + it.def.Name)
			}
		}
	case "esc":
		a.view = viewDashboard
	}
	return a, nil
}
