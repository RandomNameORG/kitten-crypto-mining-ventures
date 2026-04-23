package ui

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/lipgloss"
)

// Skill tree is a v0 stub — shows the three lanes and their perks from the
// GDD, but spending TP isn't wired up yet. Lane visibility alone lets players
// plan their build.

type skillNode struct {
	name string
	desc string
	cost int
}

var engineerLane = []skillNode{
	{"Undervolt I", "Reduce GPU power draw by 10%.", 3},
	{"Overclock I", "+10% efficiency, +15% heat.", 4},
	{"PCB Surgery", "Repair success 100%, cost −50%.", 6},
	{"MEOWCore Blueprint", "Unlock custom GPU research.", 12},
}

var mogulLane = []skillNode{
	{"Smart Invoicing", "Electricity bills −15%.", 3},
	{"Tax Optimization", "Scrap/sell value +20%.", 4},
	{"Hedged Wallet", "Halve BTC volatility impact.", 6},
	{"Venture Capital", "Unlock Prestige.", 12},
}

var hackerLane = []skillNode{
	{"Neighbor Leech", "Steal 10% of bill from the grid.", 3},
	{"Pump & Dump", "Trigger a BTC pump (2h cooldown).", 6},
	{"Botnet Whisper", "+1% passive income from elsewhere.", 6},
	{"Chain Ghost", "Police events ignored.", 12},
}

func (a App) renderSkillsView() string {
	lines := []string{
		TitleStyle.Render("🧠 Skill Tree") + "   " + DimStyle.Render("(v0 preview — not yet purchasable)"),
		DimStyle.Render("Your TP: " + strconv.Itoa(a.state.TechPoint)),
		"",
	}

	col := func(title string, nodes []skillNode) string {
		b := []string{TitleStyle.Render(title)}
		for _, n := range nodes {
			b = append(b, " • "+n.name+" "+DimStyle.Render("("+strconv.Itoa(n.cost)+" TP)"))
			b = append(b, DimStyle.Render("   "+n.desc))
		}
		return PanelStyle.Width(34).Render(strings.Join(b, "\n"))
	}

	cols := lipgloss.JoinHorizontal(lipgloss.Top,
		col("🔧 Engineer", engineerLane),
		" ",
		col("💰 Mogul", mogulLane),
		" ",
		col("🕶 Hacker", hackerLane),
	)
	lines = append(lines, cols)
	return strings.Join(lines, "\n")
}

func (a App) handleSkillsKey(key string) (tea.Model, tea.Cmd) {
	if key == "esc" {
		a.view = viewDashboard
	}
	return a, nil
}
