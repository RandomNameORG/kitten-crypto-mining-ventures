package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/game"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/update"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/ui"
)

// Version is stamped at build time via ldflags on release. Dev builds
// (go run / go build with no ldflags) see "dev" and skip the update
// check entirely — we don't want to nag developers on every iteration.
var Version = "dev"

func main() {
	newGame := flag.Bool("new", false, "start a new game, discarding any save")
	flag.Parse()

	var state *game.State
	if !*newGame {
		loaded, err := game.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "save corrupt: %v\n", err)
		}
		if loaded != nil {
			state = loaded
		}
	}

	if state == nil {
		// Empty name makes the UI open a splash/name-entry overlay on start.
		state = game.NewState("")
	} else {
		// Catch the sim up to wall-clock, capped at 8 hours. The helper
		// leaves an OfflineSummary on `state` for the UI to surface as a
		// notification on first render.
		state.RunOfflineCatchup(time.Now().Unix())
	}

	p := tea.NewProgram(ui.NewApp(state), tea.WithAltScreen())

	// Fire the update check in a goroutine BEFORE the tea Program runs.
	// Once it resolves we deliver the result via p.Send so the App
	// transitions into the splashUpdate phase. Anything that fails
	// (offline, HTTP error, same version, dismissed tag) is silent.
	go runStartupUpdateCheck(p)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ui error: %v\n", err)
		os.Exit(1)
	}
	// Final save on clean exit.
	_ = state.Save()
}

// runStartupUpdateCheck performs the GitHub releases lookup and, if a
// newer stable release is available and hasn't been "skip"-dismissed by
// the player, dispatches an ui.UpdateAvailableMsg into the tea program.
//
// Every failure mode here is swallowed — silence is the feature. A
// player offline at startup must see zero difference from a player who
// happens to be on the latest version.
func runStartupUpdateCheck(p *tea.Program) {
	// Dev builds (Version == "dev", empty, or anything that doesn't
	// parse as semver) skip entirely — the update check is only for
	// published binaries that know their own tag.
	if Version == "" || Version == "dev" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	info, err := update.Check(ctx)
	if err != nil || info == nil {
		return
	}
	if !update.IsNewer(Version, info.TagName) {
		return
	}
	dismissPath := update.DefaultDismissFile()
	dismissed, _ := update.LoadDismissed(dismissPath)
	if update.IsDismissed(dismissed, info.TagName) {
		return
	}

	body := update.TruncateLines(update.StripMarkdown(info.Body), 12)
	p.Send(ui.UpdateAvailableMsg{
		CurrentVersion:  Version,
		LatestVersion:   info.TagName,
		HTMLURL:         info.HTMLURL,
		Body:            body,
		DismissFilePath: dismissPath,
	})
}
