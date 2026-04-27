// update_splash.go renders the "new version available" pre-game splash
// and handles its key bindings. It's layered BEFORE the name / difficulty
// splash phases so returning players see it first on launch.
//
// The panel is populated from a ui.UpdateAvailableMsg delivered by the
// main binary's startup goroutine (see cmd/meowmine/main.go). If no
// message arrives before the player reaches the splash loop, this phase
// is skipped entirely — offline / slow-network users see zero friction.
package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/update"
)

// UpdateAvailableMsg is dispatched by the startup update check once the
// HTTP response has been successfully parsed and the version comparison
// confirms the release is newer than the running binary.
type UpdateAvailableMsg struct {
	CurrentVersion string
	LatestVersion  string
	HTMLURL        string
	// Body is pre-stripped / truncated plaintext ready to render.
	Body string
	// DismissFilePath is where to write "skip this version" state.
	DismissFilePath string
}

// updateSplashOptions maps cursor index → semantic key. Kept in one place
// so arrow-key + enter-commit stays consistent with the hotkey handlers.
const (
	updateOptYes  = 0
	updateOptNo   = 1
	updateOptSkip = 2
)

var updateOptCount = 3

// handleUpdateSplash dispatches key events for the update-available
// splash phase. Mirrors the arrow-key style of the difficulty splash so
// players don't have to relearn navigation.
func (a App) handleUpdateSplash(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := k.String()
	switch key {
	case "ctrl+c":
		return a, tea.Quit
	case "up", "k":
		if a.updateCursor > 0 {
			a.updateCursor--
		}
		return a, nil
	case "down", "j":
		if a.updateCursor < updateOptCount-1 {
			a.updateCursor++
		}
		return a, nil
	case "enter":
		return a.applyUpdateChoice(a.updateCursor)
	case "y", "Y":
		return a.applyUpdateChoice(updateOptYes)
	case "o", "O":
		// Explicit "open release notes" shortcut — same action as Yes
		// but advertised separately because players often just want
		// to skim the changelog without committing to an upgrade.
		return a.applyUpdateChoice(updateOptYes)
	case "n", "N", "esc":
		return a.applyUpdateChoice(updateOptNo)
	case "s", "S":
		return a.applyUpdateChoice(updateOptSkip)
	}
	return a, nil
}

// applyUpdateChoice executes the player's selection and advances past
// the update splash. All three paths leave the app in the same post-
// splash state; the difference is only what gets persisted / launched.
func (a App) applyUpdateChoice(choice int) (tea.Model, tea.Cmd) {
	info := a.updateInfo
	switch choice {
	case updateOptYes:
		// Best-effort: fire and forget. If opening fails we still
		// advance — the URL was visible in the panel, the player can
		// copy it. We set a transient status so the next splash /
		// dashboard shows a breadcrumb.
		_ = openBrowser(info.HTMLURL)
		a = a.withStatus(i18n.T("update.opening", info.HTMLURL))
	case updateOptSkip:
		// Persist the tag; ignore errors — a write failure means the
		// prompt will return next launch, which is annoying but not
		// data loss. Silent log entry helps debugging.
		list, _ := update.LoadDismissed(info.DismissFilePath)
		list = update.AppendDismissed(list, info.LatestVersion)
		if err := update.SaveDismissed(info.DismissFilePath, list); err != nil {
			a.state.AppendLog("info", fmt.Sprintf("update: failed to persist skip: %v", err))
		}
	case updateOptNo:
		// No-op — session-only dismiss.
	}
	a.updateActive = false
	// Advance to whichever splash phase the name / difficulty flow
	// would normally have started on.
	switch {
	case a.state.KittenName == "":
		a.splashPhase = splashName
	case a.state.Difficulty == "":
		a.splashPhase = splashDifficulty
		a.diffPickerCur = 1
	default:
		a.splashPhase = splashNone
	}
	return a, nil
}

// renderUpdateSplash draws the panel. Visual style matches the
// difficulty splash so it feels native — same TitleStyle, PanelStyle,
// bullet indicator.
func (a App) renderUpdateSplash() string {
	info := a.updateInfo
	title := TitleStyle.Render(fmt.Sprintf("%s — %s → %s",
		i18n.T("update.title"),
		info.CurrentVersion,
		info.LatestVersion,
	))

	fromTo := DimStyle.Render(i18n.T("update.from_to", info.CurrentVersion, info.LatestVersion))

	body := strings.TrimSpace(info.Body)
	if body == "" {
		body = i18n.T("update.no_notes")
	}
	changelog := strings.Join([]string{
		DimStyle.Render(i18n.T("update.changelog")),
		body,
	}, "\n")

	opts := []string{
		i18n.T("update.opt.yes"),
		i18n.T("update.opt.no"),
		i18n.T("update.opt.skip"),
	}
	rendered := make([]string, 0, len(opts))
	for i, label := range opts {
		cursor := "  "
		line := label
		if i == a.updateCursor {
			cursor = TitleStyle.Render("▶ ")
			line = TitleStyle.Render(line)
		} else {
			line = DimStyle.Render(line)
		}
		rendered = append(rendered, cursor+line)
	}
	optBlock := strings.Join(rendered, "\n")

	urlLine := DimStyle.Render(info.HTMLURL)
	help := DimStyle.Render(i18n.T("update.help"))

	panel := PanelStyle.Render(strings.Join([]string{
		title,
		fromTo,
		"",
		changelog,
		"",
		optBlock,
		"",
		urlLine,
	}, "\n"))

	return lipgloss.NewStyle().Padding(2, 4).Render(panel + "\n" + help)
}

// openBrowser asks the host OS to open `url`. This is a best-effort
// cross-platform launcher — we explicitly avoid pulling in a dependency
// (github.com/pkg/browser) because std-lib is enough for the three
// platforms we ship binaries for.
func openBrowser(url string) error {
	if strings.TrimSpace(url) == "" {
		return fmt.Errorf("update: empty URL")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		// linux, *bsd, etc. — xdg-open is the conventional launcher.
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
