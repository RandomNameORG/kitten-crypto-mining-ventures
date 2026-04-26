// meowmine-sim is a headless simulator for the kitten-crypto-mining-ventures
// game. It drives game.State.Tick against virtual time (no UI, no sleeps),
// which lets us run 1h / 1d / 1w of game time in seconds and dump snapshots
// for inspection or regression-diffing.
//
// Examples:
//
//	meowmine-sim --ticks=3600 --seed=1 --out=/tmp/sim.json
//	meowmine-sim --from=~/.meowmine/save.json --ticks=86400 --out=-
//	meowmine-sim --ticks=3600 --seed=1 --snapshot-every=600 --out=/tmp/sim
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/game"
)

// fixedBaseUnix is the virtual-time origin for fresh sims. Choosing a fixed
// epoch (rather than time.Now) keeps runs reproducible across invocations.
const fixedBaseUnix int64 = 1_700_000_000

func main() {
	ticks := flag.Int("ticks", 3600, "number of 1-second virtual ticks to advance")
	seed := flag.Int64("seed", 1, "RNG seed (deterministic across runs with same seed)")
	from := flag.String("from", "", "optional save file to start from; empty = fresh NewState")
	kittenName := flag.String("name", "sim-kitten", "kitten name for a fresh sim (ignored if --from set)")
	difficulty := flag.String("difficulty", "normal", "difficulty for a fresh sim (ignored if --from set)")
	out := flag.String("out", "-", "final snapshot destination; \"-\" for stdout")
	snapshotEvery := flag.Int("snapshot-every", 0, "if >0, also write a snapshot every N ticks to <out>.<tick>.json")
	summary := flag.Bool("summary", true, "print a human summary to stderr on completion")
	flag.Parse()

	if *ticks < 0 {
		fmt.Fprintln(os.Stderr, "--ticks must be >= 0")
		os.Exit(2)
	}
	if *snapshotEvery > 0 && *out == "-" {
		fmt.Fprintln(os.Stderr, "--snapshot-every requires --out to be a file path, not '-'")
		os.Exit(2)
	}

	game.SeedRNG(*seed)

	state, err := loadOrNew(*from, *kittenName, *difficulty)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(1)
	}

	startBTC := state.BTC
	startLifetime := state.LifetimeEarned
	startTech := state.TechPoint
	startLogLen := len(state.Log)
	baseUnix := state.LastTickUnix

	for i := 1; i <= *ticks; i++ {
		state.Tick(baseUnix + int64(i))
		// Drive event rolls too — in the real UI this happens right after Tick.
		_ = state.MaybeFireEvent()
		if *snapshotEvery > 0 && i%*snapshotEvery == 0 {
			path := snapshotPath(*out, i)
			if err := writeJSON(path, state); err != nil {
				fmt.Fprintf(os.Stderr, "snapshot %s: %v\n", path, err)
				os.Exit(1)
			}
		}
	}

	if err := writeJSON(*out, state); err != nil {
		fmt.Fprintf(os.Stderr, "write final: %v\n", err)
		os.Exit(1)
	}

	if *summary {
		printSummary(os.Stderr, state, *ticks, *seed, startBTC, startLifetime, startTech, startLogLen)
	}
}

func loadOrNew(from, kittenName, difficulty string) (*game.State, error) {
	if from == "" {
		s := game.NewState(kittenName)
		s.SetDifficulty(difficulty)
		// Pin every timestamp to a fixed epoch so two fresh runs with the same
		// seed produce identical state (save for log wall-clock timestamps,
		// which are stamped via time.Now inside appendLog).
		s.LastTickUnix = fixedBaseUnix
		s.LastBillUnix = fixedBaseUnix
		s.LastWagesUnix = fixedBaseUnix
		s.LastMarketTickUnix = fixedBaseUnix
		s.StartedUnix = fixedBaseUnix
		return s, nil
	}
	path := expandHome(from)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return game.LoadFrom(b)
}

func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

func snapshotPath(out string, tick int) string {
	// <out>.<tick>.json — strip a trailing .json on <out> if present so the
	// tick number isn't sandwiched awkwardly between two extensions.
	base := strings.TrimSuffix(out, ".json")
	return fmt.Sprintf("%s.%d.json", base, tick)
}

func writeJSON(dst string, state *game.State) error {
	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if dst == "-" {
		_, err = os.Stdout.Write(append(b, '\n'))
		return err
	}
	if dir := filepath.Dir(dst); dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	return os.WriteFile(dst, b, 0o644)
}

func printSummary(w *os.File, s *game.State, ticks int, seed int64, startBTC, startLifetime float64, startTech, startLogLen int) {
	gpuRunning, gpuShipping, gpuBroken := 0, 0, 0
	for _, g := range s.GPUs {
		switch g.Status {
		case "shipping":
			gpuShipping++
		case "broken", "stolen", "offline":
			gpuBroken++
		default:
			gpuRunning++
		}
	}
	fmt.Fprintf(w, "── sim summary ──────────────────────────────\n")
	fmt.Fprintf(w, " ticks:            %d (seed=%d)\n", ticks, seed)
	fmt.Fprintf(w, " virtual time:     %ds -> %ds\n", s.LastTickUnix-int64(ticks), s.LastTickUnix)
	fmt.Fprintf(w, " BTC:              %.4f  (Δ %+.4f)\n", s.BTC, s.BTC-startBTC)
	fmt.Fprintf(w, " LifetimeEarned:   %.4f  (Δ %+.4f)\n", s.LifetimeEarned, s.LifetimeEarned-startLifetime)
	fmt.Fprintf(w, " MarketPrice:      %.4f×\n", s.MarketPrice)
	fmt.Fprintf(w, " TechPoint:        %d  (Δ %+d)\n", s.TechPoint, s.TechPoint-startTech)
	fmt.Fprintf(w, " Reputation:       %d\n", s.Reputation)
	fmt.Fprintf(w, " GPUs:             %d running, %d shipping, %d broken\n", gpuRunning, gpuShipping, gpuBroken)
	fmt.Fprintf(w, " Mercs:            %d\n", len(s.Mercs))
	fmt.Fprintf(w, " Blueprints:       %d\n", len(s.Blueprints))
	fmt.Fprintf(w, " Modifiers active: %d\n", len(s.Modifiers))
	fmt.Fprintf(w, " Log entries:      +%d (total %d)\n", len(s.Log)-startLogLen, len(s.Log))
	fmt.Fprintf(w, "─────────────────────────────────────────────\n")
}
