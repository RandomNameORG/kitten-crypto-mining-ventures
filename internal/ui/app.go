package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/game"
)

type viewID int

const (
	viewDashboard viewID = iota
	viewStore
	viewGPUs
	viewRooms
	viewSkills
	viewLog
	viewMercs
	viewLab
	viewPrestige
	viewHelp
)

// tickMsg is emitted every second for sim + UI updates.
type tickMsg time.Time

type App struct {
	state *game.State
	view  viewID
	w, h  int

	// SavePathOverride, when non-empty, makes all save actions write to this
	// path instead of the default (~/.meowmine/save.json). Used by the SSH
	// server to keep per-connection saves.
	SavePathOverride string

	storeCursor    int
	gpusCursor     int
	roomsCursor    int
	skillsCursor   int
	mercsCursor    int // index into hireable list when hiring; else 0
	mercsOwnedCur  int // cursor in owned list
	mercsTab       int // 0 = owned, 1 = hireable
	labCursor      int
	labBoost1      int // index in ResearchBoosts
	labBoost2      int
	labTier        int // 1..3
	prestigeCursor int
	defenseCursor  int // 0..4 for lock/cctv/wiring/cooling/armor

	status         string
	statusExpires  time.Time
	showEventPopup *data.EventDef
}

func NewApp(s *game.State) App {
	return App{
		state:    s,
		view:     viewDashboard,
		labTier:  1,
		labBoost1: 0,
		labBoost2: 1,
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(tickCmd(), tea.SetWindowTitle("Kitten Crypto Mining Ventures"))
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = m.Width, m.Height
		return a, nil

	case tickMsg:
		a.state.Tick(time.Now().Unix())
		if def := a.state.MaybeFireEvent(); def != nil {
			a.showEventPopup = def
		}
		return a, tickCmd()

	case tea.KeyMsg:
		if a.showEventPopup != nil {
			a.showEventPopup = nil
			return a, nil
		}
		return a.handleKey(m)
	}
	return a, nil
}

func (a App) handleKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := k.String()
	// Universal keys.
	switch key {
	case "ctrl+c", "q":
		_ = a.saveNow()
		return a, tea.Quit
	case "1":
		a.view = viewDashboard
		return a, nil
	case "2":
		a.view = viewStore
		return a, nil
	case "3":
		a.view = viewGPUs
		return a, nil
	case "4":
		a.view = viewRooms
		return a, nil
	case "5":
		a.view = viewSkills
		return a, nil
	case "6":
		a.view = viewLog
		return a, nil
	case "7":
		a.view = viewMercs
		return a, nil
	case "8":
		a.view = viewLab
		return a, nil
	case "9":
		a.view = viewPrestige
		return a, nil
	case "?":
		a.view = viewHelp
		return a, nil
	case " ":
		a.state.TogglePause()
		return a, nil
	}

	// Per-view delegates (so overlapping keys resolve sensibly).
	switch a.view {
	case viewStore:
		return a.handleStoreKey(key)
	case viewGPUs:
		return a.handleGPUsKey(key)
	case viewRooms:
		return a.handleRoomsKey(key)
	case viewSkills:
		return a.handleSkillsKey(key)
	case viewMercs:
		return a.handleMercsKey(key)
	case viewLab:
		return a.handleLabKey(key)
	case viewPrestige:
		return a.handlePrestigeKey(key)
	}

	// Dashboard fallback: 's' = save.
	if key == "s" {
		if err := a.saveNow(); err != nil {
			a = a.withStatus(fmt.Sprintf("save failed: %v", err))
		} else {
			a = a.withStatus("💾 saved")
		}
	}
	// 'p' on dashboard = Pump & Dump (if unlocked).
	if key == "p" {
		if err := a.state.TriggerPumpDump(); err != nil {
			a = a.withStatus("❌ " + err.Error())
		} else {
			a = a.withStatus("📈 Pump & Dump fired")
		}
	}
	return a, nil
}

// withStatus returns a copy of `a` with a transient status banner set.
func (a App) withStatus(text string) App {
	a.status = text
	a.statusExpires = time.Now().Add(3 * time.Second)
	return a
}

// saveNow writes the current state to the right destination (local save path
// or an SSH-keyed override).
func (a App) saveNow() error {
	if a.SavePathOverride != "" {
		return a.state.SaveAs(a.SavePathOverride)
	}
	return a.state.Save()
}

// View renders the full screen.
func (a App) View() string {
	if a.w < 80 || a.h < 22 {
		return "Please widen your terminal to at least 80x22."
	}

	header := a.renderHeader()
	nav := a.renderNav()

	var body string
	switch a.view {
	case viewDashboard:
		body = a.renderDashboard()
	case viewStore:
		body = a.renderStore()
	case viewGPUs:
		body = a.renderGPUsView()
	case viewRooms:
		body = a.renderRoomsView()
	case viewSkills:
		body = a.renderSkillsView()
	case viewLog:
		body = a.renderLogView()
	case viewMercs:
		body = a.renderMercsView()
	case viewLab:
		body = a.renderLabView()
	case viewPrestige:
		body = a.renderPrestigeView()
	case viewHelp:
		body = a.renderHelpView()
	}

	footer := a.renderFooter()

	content := lipgloss.JoinVertical(lipgloss.Left, header, nav, body, footer)

	if a.showEventPopup != nil {
		return a.overlayEvent(content)
	}
	return content
}

func (a App) renderHeader() string {
	price := a.state.CurrentBTCPrice()
	paused := ""
	if a.state.Paused {
		paused = DimStyle.Render(" [PAUSED]")
	}
	title := TitleStyle.Render(fmt.Sprintf("🐾 Kitten Crypto Mining — %s", a.state.KittenName))

	extras := []string{
		MoneyStyle.Render(fmt.Sprintf("$%.0f", a.state.Money)),
		BTCStyle.Render(fmt.Sprintf("₿%.4f", a.state.BTC)),
		DimStyle.Render(fmt.Sprintf("$%.0f/BTC", price)),
		DimStyle.Render(fmt.Sprintf("TP %d", a.state.TechPoint)),
		DimStyle.Render(fmt.Sprintf("Rep %+d", a.state.Reputation)),
		DimStyle.Render(fmt.Sprintf("frags %d", a.state.ResearchFrags)),
	}
	if a.state.ActiveResearch != nil {
		pct := int(a.state.ResearchProgress() * 100)
		extras = append(extras, lipgloss.NewStyle().Foreground(AccentPurple).Render(fmt.Sprintf("🔬 %d%%", pct)))
	}
	stats := strings.Join(extras, "  ")
	line := title + paused + "  " + stats
	return HeaderStyle.Render(line)
}

func (a App) renderNav() string {
	items := []struct {
		key, label string
		id         viewID
	}{
		{"1", "dashboard", viewDashboard},
		{"2", "store", viewStore},
		{"3", "gpus", viewGPUs},
		{"4", "rooms", viewRooms},
		{"5", "skills", viewSkills},
		{"6", "log", viewLog},
		{"7", "mercs", viewMercs},
		{"8", "lab", viewLab},
		{"9", "prestige", viewPrestige},
	}
	parts := []string{}
	for _, it := range items {
		label := fmt.Sprintf("[%s]%s", it.key, it.label)
		if it.id == a.view {
			parts = append(parts, TitleStyle.Render(label))
		} else {
			parts = append(parts, DimStyle.Render(label))
		}
	}
	return lipgloss.NewStyle().Padding(0, 1).Render(strings.Join(parts, " "))
}

func (a App) renderFooter() string {
	status := a.status
	if time.Now().After(a.statusExpires) {
		status = ""
	}
	keys := DimStyle.Render("[space] pause  [s] save  [?] help  [q] quit")
	if status != "" {
		return FooterStyle.Render(status + "   ·   " + keys)
	}
	return FooterStyle.Render(keys)
}
