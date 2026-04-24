package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/game"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

// renderDashboard picks a layout based on terminal dims.
//   - w >= 106              → classic side-by-side: full room panel + log
//   - w >= 80               → three-column relayout: key info | log | hints+delivery
//   - w <  80               → single column stack
//
// `compact` (driven by height) strips flavor, blanks, and the standalone
// delivery-lane header so everything still fits on a short terminal.
func (a App) renderDashboard() string {
	roomDef, _ := data.RoomByID(a.state.CurrentRoom)
	compact := a.h < 24
	notif := a.notification()
	switch {
	case a.w >= 106:
		logN := 10
		if compact {
			logN = 6
		}
		left := a.renderRoomPanel(roomDef, 52, compact)
		// Stretch right-hand box to match the room panel's height so the
		// two align bottom-edge to bottom-edge. `lipgloss.Height` returns
		// the OUTER height (includes border), but `Style.Height` sets the
		// CONTENT+padding height — subtract 2 for the top/bottom border
		// via innerFromOuter.
		innerH := innerFromOuter(lipgloss.Height(left))
		var right string
		if notif != nil {
			right = a.renderNotifPanel(notif, 50, innerH)
		} else {
			right = a.renderLogPanel(logN, 50, innerH)
		}
		cols := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
		return lipgloss.NewStyle().Padding(0, 1).Render(cols)
	case a.w >= 80:
		return a.renderDashboardThreeCol(roomDef, compact, notif)
	default:
		panelW := fitWidth(52, a.w)
		logN := 6
		if compact {
			logN = 4
		}
		parts := []string{a.renderRoomPanel(roomDef, panelW, true)}
		if notif != nil {
			parts = append(parts, a.renderNotifPanel(notif, panelW, 0))
		}
		parts = append(parts, a.renderLogPanel(logN, panelW, 0))
		stacked := lipgloss.JoinVertical(lipgloss.Left, parts...)
		return lipgloss.NewStyle().Padding(0, 1).Render(stacked)
	}
}

// renderDashboardThreeCol is the narrow 80–105 layout.
// Left column carries only the must-read stats + rack.
// Middle column is the event log (the most dynamic thing on screen).
// Right column is dashboard hotkeys + the delivery lane, replaced by the
// active notification (event popup or offline summary) when one is live.
func (a App) renderDashboardThreeCol(def data.RoomDef, compact bool, notif *notification) string {
	totalW := a.w - 2
	gap := 1
	leftW, rightW := 26, 24
	centerW := totalW - leftW - rightW - gap*2
	// If centre would collapse, steal width from the sides fairly.
	for centerW < 22 && (leftW > 22 || rightW > 20) {
		if leftW > 22 {
			leftW--
		} else if rightW > 20 {
			rightW--
		}
		centerW = totalW - leftW - rightW - gap*2
	}
	logN := 10
	if compact {
		logN = 6
	}
	left := a.renderKeyInfoPanel(def, leftW, compact)
	targetH := innerFromOuter(lipgloss.Height(left))
	center := a.renderLogPanel(logN, centerW, targetH)
	var right string
	if notif != nil {
		right = a.renderNotifPanel(notif, rightW, targetH)
	} else {
		right = a.renderSidebarPanel(def, rightW, compact, targetH)
	}
	cols := lipgloss.JoinHorizontal(lipgloss.Top, left, " ", center, " ", right)
	return lipgloss.NewStyle().Padding(0, 1).Render(cols)
}

// notification is a small UI-layer struct that flattens the possible sources
// of a "hey, look at this" message (mid-game event popups, welcome-back
// offline catch-up summaries) into a single shape that renderNotifPanel can
// render identically. Priority is fixed at call-sites: event > offline.
type notification struct {
	title  string
	body   string // may contain \n line breaks
	hint   string // small dismiss line under body
	border lipgloss.Color
}

// notification returns the active notification, if any. Event popups take
// priority over the offline-earnings summary because they're mid-game time-
// sensitive and the summary is just a welcome-back recap.
func (a App) notification() *notification {
	if a.showEventPopup != nil {
		e := a.showEventPopup
		return &notification{
			title:  fmt.Sprintf("%s  %s", e.Emoji, e.LocalName()),
			body:   e.LocalText(),
			hint:   i18n.T("event.dismiss"),
			border: KittenPink,
		}
	}
	if a.showOfflineSummary != nil {
		s := a.showOfflineSummary
		body := i18n.T("offline.body",
			formatDuration(s.GapSeconds),
			game.FmtBTC(s.BTCGained))
		if s.Capped {
			body += "\n" + i18n.T("offline.capped")
		}
		return &notification{
			title:  i18n.T("offline.title"),
			body:   body,
			hint:   i18n.T("offline.dismiss"),
			border: MoneyGold,
		}
	}
	return nil
}

// renderNotifPanel draws the active notification as a bordered column.
// It mirrors the log / sidebar columns width-wise so the three-col layout
// stays aligned when it swaps in. minH stretches the box down to match the
// left column's bottom edge.
func (a App) renderNotifPanel(n *notification, width, minH int) string {
	wrapW := width - 4
	if wrapW < 10 {
		wrapW = 10
	}
	lines := []string{TitleStyle.Render(truncate(n.title, wrapW))}
	if n.body != "" {
		lines = append(lines, "")
		for _, para := range strings.Split(n.body, "\n") {
			lines = append(lines, wrap(para, wrapW))
		}
	}
	if n.hint != "" {
		lines = append(lines, "")
		lines = append(lines, DimStyle.Render(n.hint))
	}
	st := PanelStyle.Width(width).BorderForeground(n.border)
	if minH > 0 {
		st = st.Height(minH)
	}
	return st.Render(strings.Join(lines, "\n"))
}

// formatDuration renders a seconds count as "Xh Ym" / "Ym" / "Xs", suitable
// for the welcome-back summary. Input is always non-negative.
func formatDuration(sec int64) string {
	if sec < 0 {
		sec = 0
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	case m > 0:
		return fmt.Sprintf("%dm", m)
	default:
		return fmt.Sprintf("%ds", sec)
	}
}

func (a App) renderRoomPanel(def data.RoomDef, width int, compact bool) string {
	var heat float64
	if rs := a.state.Rooms[a.state.CurrentRoom]; rs != nil {
		heat = rs.Heat
	}

	gpus := a.state.GPUsInRoom(def.ID)
	lines := []string{}

	lines = append(lines, TitleStyle.Render(i18n.T("dash.location", def.LocalName())))
	if !compact {
		lines = append(lines, DimStyle.Render(def.LocalFlavor()))
		lines = append(lines, "")
	}

	var volt float64
	for _, g := range gpus {
		if g.Status != "running" {
			continue
		}
		_, pow, _, _ := a.state.GPUStats(g)
		volt += pow
	}

	roomID := def.ID
	bill := a.state.RoomBillRatePerSec(roomID)
	earn := a.state.RoomEarnRatePerSec(roomID)
	net := earn - bill
	heatDelta, heatTickSec := a.state.RoomHeatDeltaPerTick(roomID)
	heatTickIn := a.state.SecondsUntilNextHeatTick(roomID)
	nextBill := a.state.SecondsUntilNextBill()

	var maxHeat float64 = 90
	if rs := a.state.Rooms[roomID]; rs != nil {
		maxHeat = rs.MaxHeat
	}

	netStyle := lipgloss.NewStyle().Foreground(OppGreen)
	if net < 0 {
		netStyle = lipgloss.NewStyle().Foreground(CrisisRed)
	}

	// Heat: split into three glanceable rows.
	//   row 1: icon + current/max + trend
	//   row 2: a zoned bar (green→orange→red) so the danger zones are
	//          visible even when current fill is low
	//   row 3: a short "what this means right now" impact line, so the
	//          player always sees the effect of the current zone — not
	//          only when they're already in danger.
	_ = heatTickIn // kept for possible future use; new line uses heatTickSec
	heatFrac := 0.0
	if maxHeat > 0 {
		heatFrac = heat / maxHeat
	}
	zoneLabel := "dash.impact.stable"
	zoneStyle := lipgloss.NewStyle().Foreground(OppGreen)
	switch {
	case heatFrac >= 0.95:
		zoneLabel = "dash.impact.critical"
		zoneStyle = lipgloss.NewStyle().Foreground(CrisisRed).Bold(true)
	case heatFrac >= 0.80:
		zoneLabel = "dash.impact.hot"
		zoneStyle = lipgloss.NewStyle().Foreground(ThreatOrange).Bold(true)
	case heatFrac >= 0.60:
		zoneLabel = "dash.impact.warm"
		zoneStyle = lipgloss.NewStyle().Foreground(MoneyGold)
	}
	barW := width - 6 // account for panel padding + a little breathing room
	if barW < 12 {
		barW = 12
	}
	if barW > 30 {
		barW = 30
	}

	// Power: one line + a small "what's it doing to me" hint. If we're net-
	// positive, say "safe"; if bleeding, compute how long cash lasts at the
	// current deficit; if already broke, warn about the 60s blackout.
	powerHint := DimStyle.Render(i18n.T("dash.power.safe"))
	if net < 0 {
		if a.state.BTC <= 0 {
			powerHint = lipgloss.NewStyle().Foreground(CrisisRed).Render(i18n.T("dash.power.broke"))
		} else {
			runway := a.state.BTC / (-net)
			powerHint = lipgloss.NewStyle().Foreground(ThreatOrange).Render(i18n.T("dash.power.deficit", formatDuration(int64(runway))))
		}
	}

	lines = append(lines, fmt.Sprintf("%s   %s",
		VoltStyle.Render(i18n.T("dash.line.power", volt, game.FmtBTC(bill), nextBill)),
		DimStyle.Render(i18n.T("dash.slots_of", len(gpus), def.Slots))))
	lines = append(lines, "  "+powerHint)
	lines = append(lines, HeatStyle.Render(i18n.T("dash.heat.label", heat, maxHeat, heatDelta, heatTickSec)))
	lines = append(lines, "  "+renderHeatBar(heatFrac, barW)+"  "+zoneStyle.Render(i18n.T(zoneLabel)))
	lines = append(lines, netStyle.Render(i18n.T("dash.line.cash2", game.FmtBTC(earn), game.FmtBTCSigned(net))))
	lines = append(lines, renderMarketLine(a.state))
	if !compact {
		lines = append(lines, "")
	}

	var installed, inbound []*game.GPU
	for _, g := range gpus {
		if g.Status == "shipping" {
			inbound = append(inbound, g)
		} else {
			installed = append(installed, g)
		}
	}

	lines = append(lines, HeaderStyle.Render(i18n.T("dash.rack")))
	if len(gpus) == 0 {
		lines = append(lines, DimStyle.Render(i18n.T("dash.empty_hint")))
	}
	for i := 0; i < def.Slots; i++ {
		switch {
		case i < len(installed):
			g := installed[i]
			statusIcon := "●"
			statusColor := OppGreen
			statusText := g.Status
			switch g.Status {
			case "broken":
				statusIcon = "✕"
				statusColor = CrisisRed
				statusText = "broken"
			case "stolen":
				statusIcon = "?"
				statusColor = MutedGrey
				statusText = "stolen"
			}
			indicator := lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon)
			upMark := ""
			if g.UpgradeLevel > 0 {
				upMark = lipgloss.NewStyle().Foreground(AccentPurple).Render(fmt.Sprintf(" +%d", g.UpgradeLevel))
			}
			line := fmt.Sprintf("  %d. %s %s%s  %s", i+1, indicator, gpuDisplayName(a.state, g), upMark, DimStyle.Render(statusText))
			lines = append(lines, line)
		case i < len(installed)+len(inbound):
			lines = append(lines, lipgloss.NewStyle().Foreground(SocialCyan).Render(fmt.Sprintf(i18n.T("dash.slot_reserved"), i+1)))
		default:
			lines = append(lines, DimStyle.Render(fmt.Sprintf(i18n.T("dash.slot_empty"), i+1)))
		}
	}

	if len(inbound) > 0 {
		if !compact {
			lines = append(lines, "")
			lines = append(lines, HeaderStyle.Render(i18n.T("dash.delivery_title")))
		}
		now := time.Now().Unix()
		for _, g := range inbound {
			lines = append(lines, renderDeliveryLine(a.state, g, now, width))
		}
	}
	return PanelStyle.Width(width).Render(strings.Join(lines, "\n"))
}

// renderKeyInfoPanel is the left-column "just the vitals" view used in the
// 80–105 three-column layout. It trades flavor and prose lines for glanceable
// one-liners and a compact rack readout.
func (a App) renderKeyInfoPanel(def data.RoomDef, width int, compact bool) string {
	roomID := def.ID
	var heat float64
	var maxHeat float64 = 90
	if rs := a.state.Rooms[roomID]; rs != nil {
		heat = rs.Heat
		maxHeat = rs.MaxHeat
	}
	gpus := a.state.GPUsInRoom(roomID)

	var volt float64
	for _, g := range gpus {
		if g.Status != "running" {
			continue
		}
		_, pow, _, _ := a.state.GPUStats(g)
		volt += pow
	}
	bill := a.state.RoomBillRatePerSec(roomID)
	earn := a.state.RoomEarnRatePerSec(roomID)
	net := earn - bill
	heatDelta, _ := a.state.RoomHeatDeltaPerTick(roomID)
	nextBill := a.state.SecondsUntilNextBill()

	netStyle := lipgloss.NewStyle().Foreground(OppGreen)
	if net < 0 {
		netStyle = lipgloss.NewStyle().Foreground(CrisisRed)
	}
	heatFrac := 0.0
	if maxHeat > 0 {
		heatFrac = heat / maxHeat
	}
	heatStyle := HeatStyle
	impactKey := "dash.impact.stable"
	impactStyle := lipgloss.NewStyle().Foreground(OppGreen).Faint(true)
	switch {
	case heatFrac >= 0.95:
		heatStyle = lipgloss.NewStyle().Foreground(CrisisRed).Bold(true)
		impactKey = "dash.impact.critical"
		impactStyle = lipgloss.NewStyle().Foreground(CrisisRed).Bold(true)
	case heatFrac >= 0.80:
		heatStyle = lipgloss.NewStyle().Foreground(ThreatOrange).Bold(true)
		impactKey = "dash.impact.hot"
		impactStyle = lipgloss.NewStyle().Foreground(ThreatOrange).Bold(true)
	case heatFrac >= 0.60:
		impactKey = "dash.impact.warm"
		impactStyle = lipgloss.NewStyle().Foreground(MoneyGold)
	}

	innerW := width - 4
	if innerW < 10 {
		innerW = 10
	}
	barW := innerW - 2
	if barW < 8 {
		barW = 8
	}
	if barW > 16 {
		barW = 16
	}
	lines := []string{
		TitleStyle.Render(truncate(i18n.T("dash.location", def.LocalName()), innerW)),
		VoltStyle.Render(fmt.Sprintf("%s %.0fW  −%s/s", IconBolt, volt, game.FmtBTC(bill))),
		netStyle.Render(fmt.Sprintf("%s net %s/s", IconChartUp, game.FmtBTCSigned(net))),
		renderMarketLine(a.state),
		heatStyle.Render(fmt.Sprintf("%s %.0f/%.0f %+.1f/s", IconThermo, heat, maxHeat, heatDelta)),
		renderHeatBar(heatFrac, barW),
		impactStyle.Render(truncate(i18n.T(impactKey), innerW)),
		DimStyle.Render(fmt.Sprintf("%s bill %ds", IconClock, nextBill)),
	}
	if !compact {
		lines = append(lines, "")
	}

	var installed, inbound []*game.GPU
	for _, g := range gpus {
		if g.Status == "shipping" {
			inbound = append(inbound, g)
		} else {
			installed = append(installed, g)
		}
	}
	lines = append(lines, HeaderStyle.Render(fmt.Sprintf("%s %d/%d", i18n.T("dash.rack"), len(gpus), def.Slots)))
	if len(gpus) == 0 {
		lines = append(lines, DimStyle.Render(i18n.T("dash.empty_hint")))
	}
	for i := 0; i < def.Slots; i++ {
		switch {
		case i < len(installed):
			g := installed[i]
			icon := "●"
			col := OppGreen
			switch g.Status {
			case "broken":
				icon, col = "✕", CrisisRed
			case "stolen":
				icon, col = "?", MutedGrey
			}
			indicator := lipgloss.NewStyle().Foreground(col).Render(icon)
			upMark := ""
			up := ""
			if g.UpgradeLevel > 0 {
				up = fmt.Sprintf(" +%d", g.UpgradeLevel)
				upMark = lipgloss.NewStyle().Foreground(AccentPurple).Render(up)
			}
			// "  ● " = 4, up display width = len(up), leave 1 trailing.
			nameBudget := innerW - 4 - len([]rune(up)) - 1
			if nameBudget < 3 {
				nameBudget = 3
			}
			name := truncate(gpuDisplayName(a.state, g), nameBudget)
			lines = append(lines, fmt.Sprintf("  %s %s%s", indicator, name, upMark))
		case i < len(installed)+len(inbound):
			lines = append(lines, lipgloss.NewStyle().Foreground(SocialCyan).Render("  · inbound"))
		default:
			lines = append(lines, DimStyle.Render("  ○ empty"))
		}
	}
	return PanelStyle.Width(width).Render(strings.Join(lines, "\n"))
}

// renderSidebarPanel is the right-column for the 3-col layout: dashboard
// hotkey hints on top (the "keyword" strip), and the shipping kittens at
// the bottom. When the room has no inbound deliveries we just show hints.
func (a App) renderSidebarPanel(def data.RoomDef, width int, compact bool, minH int) string {
	lines := []string{HeaderStyle.Render(i18n.T("dash.sidebar.keys"))}
	hints := []struct{ k, label string }{
		{"space", i18n.T("dash.sidebar.pause")},
		{"V", i18n.T("dash.sidebar.vent")},
		{"p", i18n.T("dash.sidebar.pump")},
		{"S", i18n.T("dash.sidebar.save")},
		{"?", i18n.T("dash.sidebar.help")},
	}
	for _, h := range hints {
		lines = append(lines, KeyHint.Render("["+h.k+"]")+" "+DimStyle.Render(h.label))
	}

	var inbound []*game.GPU
	for _, g := range a.state.GPUsInRoom(def.ID) {
		if g.Status == "shipping" {
			inbound = append(inbound, g)
		}
	}
	if len(inbound) > 0 {
		if !compact {
			lines = append(lines, "")
		}
		lines = append(lines, HeaderStyle.Render(i18n.T("dash.delivery_title")))
		now := time.Now().Unix()
		for _, g := range inbound {
			lines = append(lines, renderDeliveryLineNarrow(a.state, g, now, width))
		}
	}

	st := PanelStyle.Width(width)
	if minH > 0 {
		st = st.Height(minH)
	}
	return st.Render(strings.Join(lines, "\n"))
}

// renderDeliveryLine draws a kitten pacing back and forth on a track, with
// the GPU's display name and ETA. Position is purely decorative (we don't
// store ship-start, so progress can't be derived) — the ETA text carries
// the real progress signal.
func renderDeliveryLine(s *game.State, g *game.GPU, now int64, panelW int) string {
	const (
		sprite = ">^.^<"
		period = 12 // seconds for a one-way traversal
	)
	// Reserve ~24 cols for "  name  ETA Ns" text; the rest goes to the track.
	trackWidth := panelW - 28
	if trackWidth < 10 {
		trackWidth = 10
	}
	if trackWidth > 28 {
		trackWidth = 28
	}
	span := trackWidth - len(sprite)
	// Unique phase offset per GPU so deliveries don't pace in lockstep.
	phase := (now + int64(g.InstanceID)*5) % int64(2*period)
	var pos int
	if phase < int64(period) {
		pos = int(phase) * span / period
	} else {
		pos = span - int(phase-int64(period))*span/period
	}
	if pos < 0 {
		pos = 0
	}
	if pos > span {
		pos = span
	}
	track := strings.Repeat("·", pos) + sprite + strings.Repeat("·", span-pos)
	trackStyled := lipgloss.NewStyle().Foreground(SocialCyan).Render(track)
	name := truncate(gpuDisplayName(s, g), 12)
	eta := g.ShipsAt - now
	if eta < 0 {
		eta = 0
	}
	return fmt.Sprintf(i18n.T("dash.delivery_line"), trackStyled, name, eta)
}

// renderDeliveryLineNarrow is the compact sidebar variant: a short track
// plus inline name/ETA so it fits in ~24 cols.
func renderDeliveryLineNarrow(s *game.State, g *game.GPU, now int64, panelW int) string {
	const (
		sprite = ">^.^<"
		period = 8
	)
	innerW := panelW - 4
	if innerW < 12 {
		innerW = 12
	}
	eta := g.ShipsAt - now
	if eta < 0 {
		eta = 0
	}
	etaStr := fmt.Sprintf(" %ds", eta)
	// Reserve at least 6 cols for the kitten track, use up to 10.
	trackW := 6
	if innerW >= 18 {
		trackW = 8
	}
	if innerW >= 22 {
		trackW = 10
	}
	span := trackW - len(sprite)
	if span < 1 {
		span = 1
	}
	phase := (now + int64(g.InstanceID)*3) % int64(2*period)
	var pos int
	if phase < int64(period) {
		pos = int(phase) * span / period
	} else {
		pos = span - int(phase-int64(period))*span/period
	}
	if pos < 0 {
		pos = 0
	}
	if pos > span {
		pos = span
	}
	track := strings.Repeat("·", pos) + sprite + strings.Repeat("·", span-pos)
	trackStyled := lipgloss.NewStyle().Foreground(SocialCyan).Render(track)
	nameBudget := innerW - trackW - len([]rune(etaStr)) - 1
	if nameBudget < 3 {
		return fmt.Sprintf("%s%s", trackStyled, DimStyle.Render(etaStr))
	}
	name := truncate(gpuDisplayName(s, g), nameBudget)
	return fmt.Sprintf("%s %s%s", trackStyled, name, DimStyle.Render(etaStr))
}

func (a App) renderLogPanel(maxLines, width, minH int) string {
	log := a.state.Log
	lines := []string{TitleStyle.Render(i18n.T("dash.log_title"))}

	start := 0
	if len(log) > maxLines {
		start = len(log) - maxLines
	}
	if start == len(log) {
		lines = append(lines, DimStyle.Render(i18n.T("dash.log_quiet")))
	}
	// "  HH:MM  " prefix = 9 cols, then panel padding/border eats 4. Reserve
	// accordingly so truncation keeps the text column on a single row.
	const stampPrefix = 9
	textW := width - stampPrefix - 4
	if textW < 14 {
		textW = 14
	}
	for i := start; i < len(log); i++ {
		entry := log[i]
		style := CategoryStyle(entry.Category)
		ts := DimStyle.Render(time.Unix(entry.Time, 0).Format("15:04"))
		lines = append(lines, fmt.Sprintf("  %s  %s", ts, style.Render(truncate(entry.Text, textW))))
	}
	st := PanelStyle.Width(width)
	if minH > 0 {
		st = st.Height(minH)
	}
	return st.Render(strings.Join(lines, "\n"))
}

// overlayEvent draws the event popup. The old implementation used
// JoinVertical(content, box), which stacked the popup *below* a full-screen
// dashboard and blew past the bottom of the terminal. We now return only
// the event box — Bubbletea's altScreen clears the frame between renders,
// so no dashboard content bleeds through. Dashboard reappears the moment
// the popup is dismissed.
func (a App) overlayEvent(content string) string {
	if a.showEventPopup == nil {
		return content
	}
	e := a.showEventPopup
	boxW := fitWidth(52, a.w)
	wrapW := boxW - 4
	if wrapW < 20 {
		wrapW = 20
	}
	box := PanelStyle.
		Width(boxW).
		BorderForeground(KittenPink).
		Render(strings.Join([]string{
			TitleStyle.Render(fmt.Sprintf("%s  %s", e.Emoji, e.LocalName())),
			"",
			wrap(e.LocalText(), wrapW),
			"",
			DimStyle.Render(i18n.T("event.dismiss")),
		}, "\n"))
	return lipgloss.NewStyle().Padding(1, 2).Render(box)
}

// overlayOfflineSummary is the non-dashboard fallback for the welcome-back
// notification. Same shape as overlayEvent — returns only the box, altScreen
// does the clearing.
func (a App) overlayOfflineSummary(content string) string {
	if a.showOfflineSummary == nil {
		return content
	}
	s := a.showOfflineSummary
	boxW := fitWidth(52, a.w)
	wrapW := boxW - 4
	if wrapW < 20 {
		wrapW = 20
	}
	body := i18n.T("offline.body",
		formatDuration(s.GapSeconds),
		game.FmtBTC(s.BTCGained))
	if s.Capped {
		body += "\n" + i18n.T("offline.capped")
	}
	box := PanelStyle.
		Width(boxW).
		BorderForeground(MoneyGold).
		Render(strings.Join([]string{
			TitleStyle.Render(i18n.T("offline.title")),
			"",
			wrap(body, wrapW),
			"",
			DimStyle.Render(i18n.T("offline.dismiss")),
		}, "\n"))
	return lipgloss.NewStyle().Padding(1, 2).Render(box)
}

// renderMarketLine formats the BTC market multiplier + trend glyph. Colour
// tracks sign of the trend so a glanceable read matches "green=up / red=down".
func renderMarketLine(s *game.State) string {
	arrow := "·"
	style := DimStyle
	switch s.MarketTrend() {
	case 1:
		arrow = "↑"
		style = lipgloss.NewStyle().Foreground(OppGreen)
	case -1:
		arrow = "↓"
		style = lipgloss.NewStyle().Foreground(CrisisRed)
	}
	return style.Render(i18n.T("dash.market.label", s.MarketPrice, arrow))
}

func truncate(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n-1]) + "…"
}

func wrap(s string, width int) string {
	words := strings.Fields(s)
	var line strings.Builder
	var out []string
	for _, w := range words {
		if line.Len()+len(w)+1 > width {
			out = append(out, line.String())
			line.Reset()
		}
		if line.Len() > 0 {
			line.WriteByte(' ')
		}
		line.WriteString(w)
	}
	if line.Len() > 0 {
		out = append(out, line.String())
	}
	return strings.Join(out, "\n")
}
