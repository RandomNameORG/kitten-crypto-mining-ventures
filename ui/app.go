package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/game"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
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

// splashPhase gates the sim on startup: new saves walk through name → difficulty
// before the dashboard takes over.
type splashPhase int

const (
	splashNone splashPhase = iota
	splashName
	splashDifficulty
)

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
	labBoost1      int // reused as combo index
	labBoost2      int // reserved
	labTier        int // 1..3
	prestigeCursor int

	// Splash overlay phases — name first, then difficulty. Sim is frozen
	// while a splash phase is active so the starter GPU doesn't tick down.
	splashPhase    splashPhase
	nameEntryBuf   string
	diffPickerCur  int

	// Retire confirmation — double-press [R] within a short window.
	retireArmedUntil time.Time

	// Buy rate-limit — prevents held-key auto-repeat from mass-buying.
	lastBuyAt time.Time

	status         string
	statusExpires  time.Time
	showEventPopup *data.EventDef
}

func NewApp(s *game.State) App {
	a := App{
		state:     s,
		view:      viewDashboard,
		labTier:   1,
		labBoost1: 0,
		labBoost2: 1,
	}
	// Name overlay fires first on truly new saves. If the save already has
	// a name but no difficulty (brand-new split flow), jump straight to
	// the difficulty picker.
	switch {
	case s.KittenName == "":
		a.splashPhase = splashName
	case s.Difficulty == "":
		a.splashPhase = splashDifficulty
		a.diffPickerCur = 1 // default cursor on "normal"
	}
	return a
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
		if a.splashPhase == splashNone {
			a.state.Tick(time.Now().Unix())
			if def := a.state.MaybeFireEvent(); def != nil {
				a.showEventPopup = def
			}
		}
		return a, tickCmd()

	case tea.KeyMsg:
		switch a.splashPhase {
		case splashName:
			return a.handleNameEntry(m)
		case splashDifficulty:
			return a.handleDifficultyEntry(m)
		}
		if a.showEventPopup != nil {
			a.showEventPopup = nil
			return a, nil
		}
		return a.handleKey(m)
	}
	return a, nil
}

func (a App) handleNameEntry(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := k.String()
	switch key {
	case "ctrl+c":
		return a, tea.Quit
	case "enter":
		name := strings.TrimSpace(a.nameEntryBuf)
		if name == "" {
			name = i18n.T("welcome.default")
		}
		a.state.KittenName = name
		a.state.AppendLog("info", i18n.T("game.named", name))
		// Advance the splash to the difficulty picker instead of starting
		// the sim — sim stays frozen until the player commits a difficulty.
		a.splashPhase = splashDifficulty
		a.diffPickerCur = 1 // cursor defaults to "normal"
		return a, nil
	case "backspace":
		if r := []rune(a.nameEntryBuf); len(r) > 0 {
			a.nameEntryBuf = string(r[:len(r)-1])
		}
		return a, nil
	default:
		r := k.Runes
		if len(r) == 1 && r[0] >= 0x20 && r[0] != 0x7F && len([]rune(a.nameEntryBuf)) < 20 {
			a.nameEntryBuf += string(r[0])
		}
	}
	return a, nil
}

func (a App) handleDifficultyEntry(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	diffs := data.Difficulties()
	key := k.String()
	switch key {
	case "ctrl+c":
		return a, tea.Quit
	case "up", "k":
		if a.diffPickerCur > 0 {
			a.diffPickerCur--
		}
	case "down", "j":
		if a.diffPickerCur < len(diffs)-1 {
			a.diffPickerCur++
		}
	case "enter":
		picked := diffs[a.diffPickerCur]
		a.state.SetDifficulty(picked.ID)
		a.splashPhase = splashNone
		now := time.Now().Unix()
		a.state.LastTickUnix = now
		a.state.LastBillUnix = now
		a.state.LastWagesUnix = now
		_ = a.saveNow()
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
	case "L":
		next := a.state.CycleLang()
		a = a.withStatus(i18n.T("status.lang", i18n.Label(next)))
		return a, nil
	case "V":
		if err := a.state.EmergencyVent(); err != nil {
			a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
		} else {
			a = a.withStatus(i18n.T("status.vent"))
		}
		return a, nil
	case "S":
		if err := a.saveNow(); err != nil {
			a = a.withStatus(i18n.T("status.save_failed", err))
		} else {
			a = a.withStatus(i18n.T("status.saved"))
		}
		return a, nil
	}

	// View-specific.
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

	// Dashboard-only fallbacks.
	if key == "p" {
		if err := a.state.TriggerPumpDump(); err != nil {
			a = a.withStatus(i18n.T("status.error_prefix") + err.Error())
		} else {
			a = a.withStatus(i18n.T("status.pump_fired"))
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

// saveNow writes to the right destination (local or SSH-keyed override).
func (a App) saveNow() error {
	if a.SavePathOverride != "" {
		return a.state.SaveAs(a.SavePathOverride)
	}
	return a.state.Save()
}

// View renders the full screen.
func (a App) View() string {
	if a.w < 80 || a.h < 22 {
		return i18n.T("warn.terminal_too_small")
	}
	switch a.splashPhase {
	case splashName:
		return a.renderNameEntry()
	case splashDifficulty:
		return a.renderDifficultyPicker()
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
		paused = DimStyle.Render(i18n.T("app.pill_paused"))
	}
	diffBadge := ""
	if d := a.state.Diff(); d.Emoji != "" {
		diffBadge = DimStyle.Render(" " + d.Emoji)
	}
	title := TitleStyle.Render(fmt.Sprintf("%s — %s", i18n.T("app.title"), a.state.KittenName)) + diffBadge

	extras := []string{
		MoneyStyle.Render(fmt.Sprintf("$%.0f", a.state.Money)),
		BTCStyle.Render(fmt.Sprintf("₿%.4f", a.state.BTC)),
		DimStyle.Render(i18n.T("hdr.price", price)),
		DimStyle.Render(i18n.T("hdr.tp", a.state.TechPoint)),
		DimStyle.Render(i18n.T("hdr.rep", a.state.Reputation)),
		DimStyle.Render(i18n.T("hdr.frags", a.state.ResearchFrags)),
		DimStyle.Render(i18n.T("hdr.achievements", len(a.state.Achievements), len(data.Achievements()))),
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
		key, labelKey string
		id            viewID
	}{
		{"1", "nav.dashboard", viewDashboard},
		{"2", "nav.store", viewStore},
		{"3", "nav.gpus", viewGPUs},
		{"4", "nav.rooms", viewRooms},
		{"5", "nav.skills", viewSkills},
		{"6", "nav.log", viewLog},
		{"7", "nav.mercs", viewMercs},
		{"8", "nav.lab", viewLab},
		{"9", "nav.prestige", viewPrestige},
	}
	parts := []string{}
	for _, it := range items {
		label := fmt.Sprintf("[%s]%s", it.key, i18n.T(it.labelKey))
		if it.id == a.view {
			parts = append(parts, TitleStyle.Render(label))
		} else {
			parts = append(parts, DimStyle.Render(label))
		}
	}
	return lipgloss.NewStyle().Padding(0, 1).Render(strings.Join(parts, " "))
}

func (a App) renderNameEntry() string {
	logo := "   /\\_/\\\n  ( o.o )\n   > ^ <"
	prompt := i18n.T("welcome.prompt") + a.nameEntryBuf + "█"
	body := strings.Join([]string{
		TitleStyle.Render(i18n.T("welcome.title")),
		DimStyle.Render(i18n.T("welcome.subtitle")),
		"",
		lipgloss.NewStyle().Foreground(KittenPink).Render(logo),
		"",
		prompt,
		"",
		DimStyle.Render(i18n.T("welcome.keys")),
	}, "\n")
	return lipgloss.NewStyle().Padding(2, 4).Render(body)
}

func (a App) renderDifficultyPicker() string {
	diffs := data.Difficulties()
	lines := []string{
		TitleStyle.Render(i18n.T("splash.diff.title")),
		DimStyle.Render(i18n.T("splash.diff.subtitle", a.state.KittenName)),
		"",
	}
	for i, d := range diffs {
		cursor := "  "
		title := d.Emoji + "  " + d.LocalLabel()
		if i == a.diffPickerCur {
			cursor = TitleStyle.Render("▶ ")
			title = TitleStyle.Render(title)
		} else {
			title = DimStyle.Render(title)
		}
		meta := DimStyle.Render(fmt.Sprintf("(earn ×%.2f · bills ×%.2f · threats ×%.2f · $%.0f start)",
			d.EarnMult, d.BillMult, d.ThreatMult, d.StarterCash))
		lines = append(lines, cursor+title+"   "+meta)
		lines = append(lines, DimStyle.Render("    "+d.LocalDesc()))
		lines = append(lines, "")
	}
	lines = append(lines, DimStyle.Render(i18n.T("splash.diff.help")))
	return lipgloss.NewStyle().Padding(2, 4).Render(strings.Join(lines, "\n"))
}

func (a App) renderFooter() string {
	status := a.status
	if time.Now().After(a.statusExpires) {
		status = ""
	}
	keys := DimStyle.Render(i18n.T("footer.keys"))
	if status != "" {
		return FooterStyle.Render(status + "   ·   " + keys)
	}
	return FooterStyle.Render(keys)
}
