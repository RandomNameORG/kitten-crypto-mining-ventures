package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/game"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/ui"
)

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
		name := promptKittenName()
		state = game.NewState(name)
	} else {
		// Offline progress: catch up the simulation from the last tick to now,
		// capped at 8 hours. Use a fake large-step Tick call; state.Tick already
		// handles arbitrary dt. We just bill in chunks so a long offline
		// doesn't bankrupt the player in a single blow.
		now := time.Now().Unix()
		gap := now - state.LastTickUnix
		if gap > 8*3600 {
			gap = 8 * 3600
			state.LastTickUnix = now - gap
			state.LastBillUnix = now - gap
			state.AppendLog("info", "Offline > 8h — capped progress at 8 hours.")
		}
		if gap > 60 {
			state.AppendLog("info", fmt.Sprintf("Caught up on %d offline minutes.", gap/60))
		}
		state.Tick(now)
	}

	p := tea.NewProgram(ui.NewApp(state), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ui error: %v\n", err)
		os.Exit(1)
	}
	// Final save on clean exit.
	_ = state.Save()
}

func promptKittenName() string {
	fmt.Print("Name your kitten (press enter for Whiskers): ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name != "" {
			return name
		}
	}
	return "Whiskers"
}
