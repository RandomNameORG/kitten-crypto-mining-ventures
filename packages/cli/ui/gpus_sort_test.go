package ui

import (
	"testing"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/game"
)

// testState returns a minimally-populated *game.State that's safe for
// GPUEarnRatePerSec / GPUStats calls — normal difficulty (EarnMult=1.0),
// neutral market price, one cold room, no modifiers or skills. HOME is
// rerouted to a tempdir so LoadLegacy (called transitively via
// EfficiencyMult) can't pick up real legacy bonuses.
func testState(t *testing.T) *game.State {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	return &game.State{
		Difficulty:  "normal",
		MarketPrice: 1.0,
		Rooms: map[string]*game.RoomState{
			"alley": {DefID: "alley", Heat: 20, MaxHeat: 100},
		},
		UnlockedSkills: map[string]bool{},
	}
}

// runningGPU returns a catalog-backed running GPU instance with the given ID,
// def, and durability. Room is fixed to "alley" — the one the testState
// already spins up.
func runningGPU(id int, defID string, hours float64) *game.GPU {
	return &game.GPU{
		InstanceID: id,
		DefID:      defID,
		Status:     "running",
		Room:       "alley",
		HoursLeft:  hours,
	}
}

// TestSortByEarnDesc: three running GPUs with distinct efficiencies (and
// therefore distinct earn rates at MarketPrice=1) must come out in
// strictly descending earn order.
func TestSortByEarnDesc(t *testing.T) {
	s := testState(t)
	// earn ordering (under MarketPrice=1, EarnMult=1): a100 > rtx4090 > scrap
	gpus := []*game.GPU{
		runningGPU(1, "scrap", 5),
		runningGPU(2, "a100", 50),
		runningGPU(3, "rtx4090", 20),
	}
	s.GPUs = gpus

	sorted, metrics, _ := prepareGPUView(s, gpus, gpuSortEarnDesc)
	if len(sorted) != 3 {
		t.Fatalf("expected 3 GPUs, got %d", len(sorted))
	}
	for i := 0; i < len(sorted)-1; i++ {
		a := metrics[sorted[i].InstanceID].earn
		b := metrics[sorted[i+1].InstanceID].earn
		if a < b {
			t.Errorf("earn desc broken at index %d: %.6f < %.6f", i, a, b)
		}
	}
	// First row should be the a100 — highest efficiency in the set.
	if sorted[0].DefID != "a100" {
		t.Errorf("expected a100 first under earn desc, got %s", sorted[0].DefID)
	}
}

// TestSortByEfficiencyDesc: rtx4090 has a higher earn/watt ratio than a100
// even though a100 earns more in absolute terms, so the eff-desc sort must
// reorder them relative to earn-desc.
func TestSortByEfficiencyDesc(t *testing.T) {
	s := testState(t)
	gpus := []*game.GPU{
		runningGPU(1, "a100", 10),     // earn=0.04*300=12, pow=20 → eff=0.0020
		runningGPU(2, "rtx4090", 10),  // earn=0.025*300=7.5, pow=12 → eff≈0.00208
		runningGPU(3, "scrap", 10),    // earn=0.0005*300=0.15, pow=1 → eff=0.00050
	}
	s.GPUs = gpus

	sorted, metrics, _ := prepareGPUView(s, gpus, gpuSortEffDesc)
	if sorted[0].DefID != "rtx4090" {
		t.Errorf("expected rtx4090 (highest eff ratio) first, got %s", sorted[0].DefID)
	}
	for i := 0; i < len(sorted)-1; i++ {
		a := metrics[sorted[i].InstanceID].eff
		b := metrics[sorted[i+1].InstanceID].eff
		if a < b {
			t.Errorf("eff desc broken at index %d: %.6f < %.6f", i, a, b)
		}
	}
}

// TestSortByDurabilityAsc: the "what's about to die" view must surface the
// GPU with the lowest HoursLeft first, regardless of earn.
func TestSortByDurabilityAsc(t *testing.T) {
	s := testState(t)
	gpus := []*game.GPU{
		runningGPU(1, "gtx1060", 10.0),
		runningGPU(2, "rtx4090", 2.5), // lowest — should come first
		runningGPU(3, "a100", 7.3),
	}
	s.GPUs = gpus

	sorted, _, _ := prepareGPUView(s, gpus, gpuSortDurAsc)
	if sorted[0].InstanceID != 2 {
		t.Errorf("expected instance 2 (2.5h) first, got %d", sorted[0].InstanceID)
	}
	for i := 0; i < len(sorted)-1; i++ {
		if sorted[i].HoursLeft > sorted[i+1].HoursLeft {
			t.Errorf("dur asc broken at index %d: %.2f > %.2f",
				i, sorted[i].HoursLeft, sorted[i+1].HoursLeft)
		}
	}
}

// TestSortDefaultPreservesOrder: the default mode returns the GPUs in the
// same order as the input (by InstanceID, which mirrors insertion order).
func TestSortDefaultPreservesOrder(t *testing.T) {
	s := testState(t)
	// Intentionally scramble the natural earn order so a non-default sort
	// would visibly rearrange them.
	gpus := []*game.GPU{
		runningGPU(1, "a100", 10),
		runningGPU(2, "scrap", 10),
		runningGPU(3, "rtx4090", 10),
	}
	s.GPUs = gpus

	sorted, _, _ := prepareGPUView(s, gpus, gpuSortDefault)
	for i, g := range sorted {
		if g.InstanceID != gpus[i].InstanceID {
			t.Errorf("default sort mutated order at index %d: got #%d, want #%d",
				i, g.InstanceID, gpus[i].InstanceID)
		}
	}
	// Input slice must not be mutated either.
	if gpus[0].InstanceID != 1 || gpus[1].InstanceID != 2 || gpus[2].InstanceID != 3 {
		t.Error("prepareGPUView mutated the input slice; it must work on a copy")
	}
}

// TestRankQuartilesForFourRunning: 4 running GPUs should split into
// 1 top / 2 mid / 1 low so quartile colour bands are visible.
func TestRankQuartilesForFourRunning(t *testing.T) {
	s := testState(t)
	// Distinct efficiencies → distinct earn rates.
	gpus := []*game.GPU{
		runningGPU(1, "a100", 10),     // highest earn
		runningGPU(2, "rtx4090", 10),
		runningGPU(3, "gtx1060", 10),
		runningGPU(4, "scrap", 10),    // lowest earn
	}
	s.GPUs = gpus

	_, _, ranks := prepareGPUView(s, gpus, gpuSortDefault)
	if ranks[1] != rankTop {
		t.Errorf("instance 1 (a100) should be rankTop, got %v", ranks[1])
	}
	if ranks[4] != rankLow {
		t.Errorf("instance 4 (scrap) should be rankLow, got %v", ranks[4])
	}
	// Middle two should both be rankMid.
	if ranks[2] != rankMid {
		t.Errorf("instance 2 (rtx4090) should be rankMid, got %v", ranks[2])
	}
	if ranks[3] != rankMid {
		t.Errorf("instance 3 (gtx1060) should be rankMid, got %v", ranks[3])
	}
}

// TestRankFallbacks: with <4 running GPUs the quartile math has to degrade
// gracefully (see assignRanks). Non-running GPUs must never receive a
// tier — the renderer relies on absence to pick the dim style.
func TestRankFallbacks(t *testing.T) {
	s := testState(t)

	// n=1 → single GPU is rankTop.
	s.GPUs = []*game.GPU{runningGPU(1, "rtx4090", 10)}
	_, _, r1 := prepareGPUView(s, s.GPUs, gpuSortDefault)
	if r1[1] != rankTop {
		t.Errorf("n=1: expected rankTop, got %v", r1[1])
	}

	// n=2 → top + low, no middle.
	s.GPUs = []*game.GPU{
		runningGPU(1, "scrap", 10),
		runningGPU(2, "rtx4090", 10),
	}
	_, _, r2 := prepareGPUView(s, s.GPUs, gpuSortDefault)
	if r2[2] != rankTop {
		t.Errorf("n=2: expected rtx4090=rankTop, got %v", r2[2])
	}
	if r2[1] != rankLow {
		t.Errorf("n=2: expected scrap=rankLow, got %v", r2[1])
	}

	// n=3 → top / mid / low.
	s.GPUs = []*game.GPU{
		runningGPU(1, "scrap", 10),
		runningGPU(2, "rtx4090", 10),
		runningGPU(3, "gtx1060", 10),
	}
	_, _, r3 := prepareGPUView(s, s.GPUs, gpuSortDefault)
	if r3[2] != rankTop {
		t.Errorf("n=3: expected rtx4090=rankTop, got %v", r3[2])
	}
	if r3[3] != rankMid {
		t.Errorf("n=3: expected gtx1060=rankMid, got %v", r3[3])
	}
	if r3[1] != rankLow {
		t.Errorf("n=3: expected scrap=rankLow, got %v", r3[1])
	}

	// Non-running GPUs must not be ranked.
	s.GPUs = []*game.GPU{
		runningGPU(1, "rtx4090", 10),
		{InstanceID: 2, DefID: "gtx1060", Status: "broken", Room: "alley", HoursLeft: 0},
	}
	_, _, r4 := prepareGPUView(s, s.GPUs, gpuSortDefault)
	if _, ok := r4[2]; ok {
		t.Errorf("broken GPU should not receive a rank tier, got %v", r4[2])
	}
}

// TestEfficiencyMath: the per-GPU eff metric returned by the helper must
// match a manual earn/power recomputation. Guards against accidental
// unit bugs if either primitive changes.
func TestEfficiencyMath(t *testing.T) {
	s := testState(t)
	gpus := []*game.GPU{runningGPU(1, "rtx4090", 10)}
	s.GPUs = gpus

	_, metrics, _ := prepareGPUView(s, gpus, gpuSortDefault)
	m := metrics[1]
	want := m.earn / m.power
	// Non-zero sanity: if power is zero the test tells us nothing.
	if m.power == 0 {
		t.Fatal("rtx4090 power should not be zero in this state")
	}
	if m.eff != want {
		t.Errorf("eff mismatch: got %.9f, want %.9f (earn=%.6f, power=%.3f)",
			m.eff, want, m.earn, m.power)
	}
	// And independently sanity-check against GPUEarnRatePerSec, so a
	// future refactor that silently severs the link fails loudly here.
	if got := s.GPUEarnRatePerSec(gpus[0]); got != m.earn {
		t.Errorf("cached earn %.6f disagrees with GPUEarnRatePerSec %.6f", m.earn, got)
	}
}

// TestCursorPreservationViaIndexOfGPU locks in the "sort keeps the cursor
// on the same GPU" invariant the key handler relies on. Index lookup
// must track the GPU even after the display order flips.
func TestCursorPreservationViaIndexOfGPU(t *testing.T) {
	s := testState(t)
	gpus := []*game.GPU{
		runningGPU(1, "scrap", 10),    // lowest earn
		runningGPU(2, "rtx4090", 10),
		runningGPU(3, "a100", 10),     // highest earn
	}
	s.GPUs = gpus

	// Cursor starts on the middle row (index 1 → instance 2).
	anchorID := gpus[1].InstanceID

	// After earn-desc sort, instance 2 should land at index 1 again
	// (a100 > rtx4090 > scrap); but instance 3 lands at index 0 — so
	// naive index-based preservation would silently move the cursor
	// to the wrong GPU. Use indexOfGPU to recover the right position.
	sorted, _, _ := prepareGPUView(s, gpus, gpuSortEarnDesc)
	idx := indexOfGPU(sorted, anchorID)
	if idx < 0 {
		t.Fatalf("indexOfGPU failed to find instance %d after sort", anchorID)
	}
	if sorted[idx].InstanceID != anchorID {
		t.Errorf("indexOfGPU returned wrong index: sorted[%d] = #%d, want #%d",
			idx, sorted[idx].InstanceID, anchorID)
	}

	// A missing ID must return -1 so callers can fall back cleanly.
	if got := indexOfGPU(sorted, 999); got != -1 {
		t.Errorf("indexOfGPU(missing) = %d, want -1", got)
	}
}
