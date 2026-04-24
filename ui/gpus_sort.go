package ui

import (
	"sort"

	"github.com/charmbracelet/lipgloss"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/game"
)

// gpuSortMode cycles through the ways the GPUs view can order its rows.
// Stored on App (UI-only) and cycled via the `b` keypress. The game State
// never sees this — sort is a display concern, not a save-worthy one.
type gpuSortMode int

const (
	gpuSortDefault gpuSortMode = iota // original InstanceID / insertion order
	gpuSortEarnDesc
	gpuSortEffDesc
	gpuSortDurAsc
	gpuSortModeCount
)

// cycleGPUSortMode returns the next mode in the fixed cycle.
func cycleGPUSortMode(m gpuSortMode) gpuSortMode {
	return (m + 1) % gpuSortModeCount
}

// rankTier buckets a running GPU's earn-per-second into quartile colour bands.
// Non-running GPUs always land in rankNone — they render in the dim palette
// with no coloured earn cell.
type rankTier int

const (
	rankNone rankTier = iota
	rankTop           // top 25% → OppGreen
	rankMid           // middle 50% → MoneyGold
	rankLow           // bottom 25% → CrisisRed
)

// rankColour maps a rankTier to the colour used for the earn cell. Callers
// should skip styling entirely when the tier is rankNone so the dim dash
// stays visible.
func rankColour(t rankTier) lipgloss.Color {
	switch t {
	case rankTop:
		return OppGreen
	case rankMid:
		return MoneyGold
	case rankLow:
		return CrisisRed
	}
	return MutedGrey
}

// gpuMetrics is the precomputed per-GPU snapshot the renderer and the sort
// comparator share. Calculated once by prepareGPUView so the sort doesn't
// re-hit time.Now() on every compare.
type gpuMetrics struct {
	earn    float64 // BTC/s (0 when not running)
	power   float64 // watts (0 when not running)
	eff     float64 // earn / power (0 when power == 0)
	hours   float64 // durability remaining
	running bool
}

// prepareGPUView caches stats per GPU, returns a sorted copy of `gpus`
// according to `mode`, and assigns rank tiers computed over the
// *running-only* earn distribution. The input slice is never mutated — all
// sorting happens on a fresh copy.
//
// Rank assignment rules:
//   - 0 running → no tiers
//   - 1 running → rankTop
//   - 2 running → top=Top, bottom=Low
//   - 3 running → Top, Mid, Low
//   - 4+ running → top max(1, n/4) → Top, bottom max(1, n/4) → Low, rest → Mid
func prepareGPUView(s *game.State, gpus []*game.GPU, mode gpuSortMode) (sorted []*game.GPU, metrics map[int]gpuMetrics, ranks map[int]rankTier) {
	metrics = make(map[int]gpuMetrics, len(gpus))
	for _, g := range gpus {
		m := gpuMetrics{hours: g.HoursLeft, running: g.Status == "running"}
		if m.running {
			_, pow, _, _ := s.GPUStats(g)
			m.power = pow
			m.earn = s.GPUEarnRatePerSec(g)
			if pow > 0 {
				m.eff = m.earn / pow
			}
		}
		metrics[g.InstanceID] = m
	}

	sorted = make([]*game.GPU, len(gpus))
	copy(sorted, gpus)
	switch mode {
	case gpuSortEarnDesc:
		sort.SliceStable(sorted, func(i, j int) bool {
			return metrics[sorted[i].InstanceID].earn > metrics[sorted[j].InstanceID].earn
		})
	case gpuSortEffDesc:
		sort.SliceStable(sorted, func(i, j int) bool {
			return metrics[sorted[i].InstanceID].eff > metrics[sorted[j].InstanceID].eff
		})
	case gpuSortDurAsc:
		sort.SliceStable(sorted, func(i, j int) bool {
			return metrics[sorted[i].InstanceID].hours < metrics[sorted[j].InstanceID].hours
		})
	}

	ranks = assignRanks(gpus, metrics)
	return sorted, metrics, ranks
}

// assignRanks sorts the running subset by earn desc and buckets it into
// quartiles (with small-n fallbacks so 1/2/3-GPU setups still render
// meaningful colours). Returned map is keyed by InstanceID; non-running
// GPUs are absent → renderer treats them as rankNone.
func assignRanks(gpus []*game.GPU, metrics map[int]gpuMetrics) map[int]rankTier {
	ranks := make(map[int]rankTier, len(gpus))
	running := make([]*game.GPU, 0, len(gpus))
	for _, g := range gpus {
		if metrics[g.InstanceID].running {
			running = append(running, g)
		}
	}
	sort.SliceStable(running, func(i, j int) bool {
		return metrics[running[i].InstanceID].earn > metrics[running[j].InstanceID].earn
	})

	n := len(running)
	switch {
	case n == 0:
		// nothing to rank
	case n == 1:
		ranks[running[0].InstanceID] = rankTop
	case n == 2:
		ranks[running[0].InstanceID] = rankTop
		ranks[running[1].InstanceID] = rankLow
	case n == 3:
		ranks[running[0].InstanceID] = rankTop
		ranks[running[1].InstanceID] = rankMid
		ranks[running[2].InstanceID] = rankLow
	default:
		nTop := n / 4
		if nTop < 1 {
			nTop = 1
		}
		nBot := n / 4
		if nBot < 1 {
			nBot = 1
		}
		for i, g := range running {
			switch {
			case i < nTop:
				ranks[g.InstanceID] = rankTop
			case i >= n-nBot:
				ranks[g.InstanceID] = rankLow
			default:
				ranks[g.InstanceID] = rankMid
			}
		}
	}
	return ranks
}

// indexOfGPU returns the position of the GPU with `id` in `gpus`, or -1 if
// absent. Used to keep the cursor anchored to the same GPU across sort
// changes and list mutations.
func indexOfGPU(gpus []*game.GPU, id int) int {
	for i, g := range gpus {
		if g.InstanceID == id {
			return i
		}
	}
	return -1
}
