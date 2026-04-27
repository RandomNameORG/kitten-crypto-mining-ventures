package game

import (
	"math"
	"testing"
)

// simTestBaseUnix matches the fixed epoch used by cmd/meowmine-sim, so tests
// and the binary exercise the same pinned-time progression.
const simTestBaseUnix int64 = 1_700_000_000

// runSim mirrors cmd/meowmine-sim's inner loop: seed the RNG, pin every
// timestamp to a fixed epoch, advance N virtual seconds via Tick + event
// rolls. Kept package-private so the test file is the only caller.
func runSim(t *testing.T, seed int64, ticks int) *State {
	t.Helper()
	SeedRNG(seed)
	s := NewState("sim-test")
	s.SetDifficulty("normal")
	s.LastTickUnix = simTestBaseUnix
	s.LastBillUnix = simTestBaseUnix
	s.LastWagesUnix = simTestBaseUnix
	s.LastMarketTickUnix = simTestBaseUnix
	s.StartedUnix = simTestBaseUnix
	for i := 1; i <= ticks; i++ {
		s.Tick(simTestBaseUnix + int64(i))
		_ = s.MaybeFireEvent()
	}
	return s
}

// runSimWithProbe is runSim with a per-tick callback. Used when a test needs
// to assert an invariant holds *every* tick, not just at the end.
func runSimWithProbe(t *testing.T, seed int64, ticks int, probe func(i int, s *State)) *State {
	t.Helper()
	SeedRNG(seed)
	s := NewState("sim-test")
	s.SetDifficulty("normal")
	s.LastTickUnix = simTestBaseUnix
	s.LastBillUnix = simTestBaseUnix
	s.LastWagesUnix = simTestBaseUnix
	s.LastMarketTickUnix = simTestBaseUnix
	s.StartedUnix = simTestBaseUnix
	for i := 1; i <= ticks; i++ {
		s.Tick(simTestBaseUnix + int64(i))
		_ = s.MaybeFireEvent()
		if probe != nil {
			probe(i, s)
		}
	}
	return s
}

// runSimDifficulty is runSimWithProbe but lets the caller pick the difficulty
// tier. Used to compare two runs on the same seed under different knobs.
func runSimDifficulty(t *testing.T, seed int64, ticks int, diffID string, probe func(i int, s *State)) *State {
	t.Helper()
	SeedRNG(seed)
	s := NewState("sim-test")
	s.SetDifficulty(diffID)
	s.LastTickUnix = simTestBaseUnix
	s.LastBillUnix = simTestBaseUnix
	s.LastWagesUnix = simTestBaseUnix
	s.LastMarketTickUnix = simTestBaseUnix
	s.StartedUnix = simTestBaseUnix
	for i := 1; i <= ticks; i++ {
		s.Tick(simTestBaseUnix + int64(i))
		_ = s.MaybeFireEvent()
		if probe != nil {
			probe(i, s)
		}
	}
	return s
}

// TestSimLongRunSanity catches the class of regression the simulator exists
// for: a broken tick path that silently produces NaN earnings, leaves the
// clock stuck, or skips the billing subsystem entirely.
func TestSimLongRunSanity(t *testing.T) {
	withTempHome(t)
	s := runSim(t, 1, 3600) // 1 virtual hour

	if math.IsNaN(s.BTC) || math.IsInf(s.BTC, 0) {
		t.Fatalf("BTC became non-finite: %v", s.BTC)
	}
	if math.IsNaN(s.LifetimeEarned) || s.LifetimeEarned < 0 {
		t.Fatalf("LifetimeEarned invalid: %v", s.LifetimeEarned)
	}
	wantTick := simTestBaseUnix + 3600
	if s.LastTickUnix != wantTick {
		t.Fatalf("LastTickUnix = %d, want %d", s.LastTickUnix, wantTick)
	}
	// Billing fires every 60s. After an hour we expect LastBillUnix to have
	// moved off its starting value — if it hasn't, advanceBilling isn't
	// being reached.
	if s.LastBillUnix <= simTestBaseUnix {
		t.Errorf("billing never advanced in 1h of ticks (LastBillUnix=%d)", s.LastBillUnix)
	}
	// Starter GPU should still exist (may be running or broken, but not gone).
	if len(s.GPUs) == 0 {
		t.Error("GPU list emptied itself during tick loop")
	}
}

// TestSimDeterministicGameState asserts that two fresh runs with the same
// seed produce identical *game* fields. Timestamp fields stamped via
// time.Now() inside appendLog/ShipsAt drift by milliseconds between runs and
// are intentionally excluded — this test is about game logic determinism,
// not wall-clock bookkeeping.
func TestSimDeterministicGameState(t *testing.T) {
	withTempHome(t)
	a := runSim(t, 42, 1800)
	withTempHome(t)
	b := runSim(t, 42, 1800)

	if a.BTC != b.BTC {
		t.Errorf("BTC drift: %v vs %v", a.BTC, b.BTC)
	}
	if a.LifetimeEarned != b.LifetimeEarned {
		t.Errorf("LifetimeEarned drift: %v vs %v", a.LifetimeEarned, b.LifetimeEarned)
	}
	if a.TechPoint != b.TechPoint {
		t.Errorf("TechPoint drift: %d vs %d", a.TechPoint, b.TechPoint)
	}
	if a.Reputation != b.Reputation {
		t.Errorf("Reputation drift: %d vs %d", a.Reputation, b.Reputation)
	}
	if a.Karma != b.Karma {
		t.Errorf("Karma drift: %d vs %d", a.Karma, b.Karma)
	}
	if len(a.GPUs) != len(b.GPUs) {
		t.Errorf("GPU count drift: %d vs %d", len(a.GPUs), len(b.GPUs))
	}
	if len(a.Modifiers) != len(b.Modifiers) {
		t.Errorf("Modifier count drift: %d vs %d", len(a.Modifiers), len(b.Modifiers))
	}
	if len(a.Achievements) != len(b.Achievements) {
		t.Errorf("Achievement count drift: %d vs %d", len(a.Achievements), len(b.Achievements))
	}
}

// TestSimSeedsDiverge proves that seeding actually threads into the game —
// if somebody refactored RNG calls to use a non-seeded source, this catches
// it by observing that two seeds produce at least one different outcome.
func TestSimSeedsDiverge(t *testing.T) {
	withTempHome(t)
	a := runSim(t, 1, 1800)
	withTempHome(t)
	b := runSim(t, 2, 1800)

	same := a.BTC == b.BTC &&
		a.Reputation == b.Reputation &&
		a.TechPoint == b.TechPoint &&
		len(a.Log) == len(b.Log) &&
		len(a.Modifiers) == len(b.Modifiers)
	if same {
		t.Fatal("seed=1 and seed=2 produced identical observable state — RNG is probably not being threaded through")
	}
}

// TestSimOCDrainsDurabilityFaster pins the core promise of overclocking —
// faster earn, faster wear. Two identical-seed runs diverge only in that
// the OC run force-sets every GPU to level 2 at the start of each tick (a
// GPU can arrive from shipping mid-run, so we re-apply to catch it while
// it's still running). After an hour the OC fleet must either be closer
// to the grave (less HoursLeft summed across still-running cards) or have
// actually died more often. Picking an OR lets us stay robust even when
// RNG choreographs a particularly unlucky break.
func TestSimOCDrainsDurabilityFaster(t *testing.T) {
	withTempHome(t)
	baseline := runSim(t, 1, 3600)

	withTempHome(t)
	oc := runSimWithProbe(t, 1, 3600, func(_ int, s *State) {
		for _, g := range s.GPUs {
			if g.Status == "running" {
				g.OCLevel = 2
			}
		}
	})

	sumHours := func(s *State) float64 {
		var h float64
		for _, g := range s.GPUs {
			if g.Status == "running" {
				h += g.HoursLeft
			}
		}
		return h
	}
	countBroken := func(s *State) int {
		n := 0
		for _, g := range s.GPUs {
			if g.Status == "broken" {
				n++
			}
		}
		return n
	}

	baseHours, ocHours := sumHours(baseline), sumHours(oc)
	baseBroken, ocBroken := countBroken(baseline), countBroken(oc)

	hoursDrop := baseHours - ocHours
	moreBroken := ocBroken > baseBroken
	// Meaningful-drop threshold: at level 2 wear is 3× baseline, so over
	// a full virtual hour a healthy fleet should lose at least one extra
	// hour of cumulative durability. If we see neither extra breakage nor
	// ≥1h of extra drain, the OC wearMult path isn't firing.
	if !moreBroken && hoursDrop < 1.0 {
		t.Fatalf("OC did not wear GPUs faster: baseHours=%.2f ocHours=%.2f drop=%.2f baseBroken=%d ocBroken=%d",
			baseHours, ocHours, hoursDrop, baseBroken, ocBroken)
	}
}

// TestSimSyndicateDividendsAccrue pushes the full tick loop through a
// week-plus of virtual time with the player joined to the syndicate. We
// pre-seed LifetimeEarned so the probe can auto-join on tick 1 (otherwise
// reaching the 500K-BTC threshold via the starter GPU alone would stretch
// the test to hundreds of virtual days), then assert that one payout
// window closes cleanly: TotalDividends > 0 and BTC stays finite.
func TestSimSyndicateDividendsAccrue(t *testing.T) {
	withTempHome(t)
	// One week is 604 800 virtual seconds. Run a bit past that so the
	// weekly payout has definitely rolled.
	const ticks = SyndicatePayoutIntervalSec + 200
	joined := false
	s := runSimWithProbe(t, 1, int(ticks), func(i int, s *State) {
		if joined {
			return
		}
		if s.LifetimeEarned < SyndicateJoinThreshold {
			s.LifetimeEarned = SyndicateJoinThreshold
		}
		if err := s.JoinSyndicate(simTestBaseUnix + int64(i)); err == nil {
			joined = true
		}
	})
	if !joined {
		t.Fatal("probe never joined the syndicate")
	}
	if math.IsNaN(s.BTC) || math.IsInf(s.BTC, 0) {
		t.Fatalf("BTC became non-finite: %v", s.BTC)
	}
	if math.IsNaN(s.SyndicateTotalDividends) || math.IsInf(s.SyndicateTotalDividends, 0) {
		t.Fatalf("SyndicateTotalDividends became non-finite: %v", s.SyndicateTotalDividends)
	}
	if s.SyndicateTotalDividends <= 0 {
		t.Fatalf("expected dividends to accrue after one payout window; got %v", s.SyndicateTotalDividends)
	}
}

// TestSimTPScalesWithProgression — sprint-2 invariant: late-game TP income
// must outpace early-game once the new faucets (lifetime milestones,
// achievement bonuses) are wired. Compares a baseline 1h sim to a
// pre-progressed run where LifetimeEarned starts above several milestone
// tiers; the progressed run should end with materially more TP without
// any change to the tick loop's RNG seed.
func TestSimTPScalesWithProgression(t *testing.T) {
	withTempHome(t)
	baseline := runSim(t, 1, 3600)

	// Progressed run: same seed, but the probe pre-loads LifetimeEarned to
	// 50M on tick 1, which should trigger milestone tiers 1 (10K), 2 (100K),
	// 3 (1M) and 4 (10M) — totalling 5+15+30+60 = 110 TP from milestones
	// alone. We pre-load via the probe rather than mutating before the loop
	// because some milestone logic assumes the value moved during play.
	withTempHome(t)
	preloaded := false
	progressed := runSimWithProbe(t, 1, 3600, func(i int, s *State) {
		if !preloaded {
			s.LifetimeEarned = 50_000_000
			preloaded = true
		}
	})

	if baseline.TechPoint >= 50 {
		t.Errorf("baseline starter run produced %d TP — fresh state shouldn't dump hundreds of TP per hour",
			baseline.TechPoint)
	}
	gain := progressed.TechPoint - baseline.TechPoint
	// Expect at least the 110 TP from milestone tiers 1–4. Allow some
	// slack for unrelated grant paths but anchor against the new faucets
	// so a regression that disables them shows up as a near-zero gain.
	if gain < 100 {
		t.Errorf("progressed run gained only %d TP over baseline (baseline=%d, progressed=%d) — milestone faucet not firing?",
			gain, baseline.TechPoint, progressed.TechPoint)
	}
	if progressed.LifetimeMilestonesPaid < 4 {
		t.Errorf("progressed run only crossed %d milestone tiers, expected ≥4 by LE=50M",
			progressed.LifetimeMilestonesPaid)
	}
}

// TestSimMarketPriceInvariants runs a full virtual day through the sim and
// asserts the market price stays finite + clamped every tick, and that it
// actually moves off 1.0 across the run. This catches a drift path that's
// wired into Tick but misbehaves under full-loop interaction (e.g. offline
// catch-up gaps, repeated Tick calls with zero dt, etc.).
func TestSimMarketPriceInvariants(t *testing.T) {
	withTempHome(t)
	var minSeen, maxSeen float64 = 1.0, 1.0
	s := runSimWithProbe(t, 1, 86400, func(_ int, s *State) {
		if math.IsNaN(s.MarketPrice) || math.IsInf(s.MarketPrice, 0) {
			t.Fatalf("MarketPrice non-finite: %v", s.MarketPrice)
		}
		if s.MarketPrice < marketPriceMin || s.MarketPrice > marketPriceMax {
			t.Fatalf("MarketPrice out of bounds: %v", s.MarketPrice)
		}
		if s.MarketPrice < minSeen {
			minSeen = s.MarketPrice
		}
		if s.MarketPrice > maxSeen {
			maxSeen = s.MarketPrice
		}
	})
	if s.MarketPrice == 1.0 {
		t.Errorf("MarketPrice ended at exactly 1.0 after 24h — walk never ran?")
	}
	// A 24h random walk should explore noticeably off the starting point —
	// if min and max are both within ±0.01 of 1.0 something's clamping it.
	if maxSeen-minSeen < 0.02 {
		t.Errorf("MarketPrice range %.4f–%.4f is suspiciously tight — drift may be neutered", minSeen, maxSeen)
	}
}

// TestSimCryptoWinterWiderSwingsAndMoreEvents verifies the two new
// difficulty knobs actually change gameplay: crypto_winter must exhibit a
// wider realized market-price range (MarketVolatilityMult=2.0) and roll more
// successful event fires (EventFreqMult=1.5) than normal over the same 6h at
// the same seed. Same seed across runs isolates the difference to the
// multipliers, not RNG. Also asserts MarketPrice stays finite throughout the
// winter run since the widened clamp band is a new code path.
//
// Per-tick, the probe zeroes EventCooldown *after* MaybeFireEvent has
// already run — so the count of fires is accurate, but the NEXT tick sees a
// clean cooldown map. This is load-bearing: MaybeFireEvent's cooldown
// bookkeeping uses wall-clock time.Now() (see events.go), which in a rapid
// sim means every event goes on effectively-permanent cooldown after one
// fire. Without this reset, total fires saturates at the
// eligible-event count and hides the effect of EventFreqMult entirely.
func TestSimCryptoWinterWiderSwingsAndMoreEvents(t *testing.T) {
	const ticks = 21600 // 6 virtual hours
	const seed = int64(1)

	type runStats struct{ min, max float64; fires, prevEvents int }

	runDiff := func(diff string) *runStats {
		p := &runStats{min: 1.0, max: 1.0}
		withTempHome(t)
		runSimDifficulty(t, seed, ticks, diff, func(_ int, s *State) {
			if math.IsNaN(s.MarketPrice) || math.IsInf(s.MarketPrice, 0) {
				t.Fatalf("%s MarketPrice non-finite: %v", diff, s.MarketPrice)
			}
			if s.MarketPrice < p.min {
				p.min = s.MarketPrice
			}
			if s.MarketPrice > p.max {
				p.max = s.MarketPrice
			}
			cur := 0
			for _, v := range s.EventsByCategory {
				cur += v
			}
			if cur > p.prevEvents {
				p.fires += cur - p.prevEvents
				p.prevEvents = cur
			}
			for k := range s.EventCooldown {
				delete(s.EventCooldown, k)
			}
		})
		return p
	}

	normal := runDiff("normal")
	winter := runDiff("crypto_winter")

	normalRange := normal.max - normal.min
	winterRange := winter.max - winter.min
	if winterRange <= normalRange {
		t.Errorf("crypto_winter market range (%.4f) not wider than normal (%.4f) — MarketVolatilityMult not wired",
			winterRange, normalRange)
	}
	if winter.fires <= normal.fires {
		t.Errorf("crypto_winter fired %d events, normal fired %d — EventFreqMult not wired",
			winter.fires, normal.fires)
	}
}

// TestSimEarningsNotHalvedBySprint4 confirms §8 / §11.2 don't gut mining
// income. Gas fees only fire on SellGPU (the sim never calls it) and
// congestion drift is RNG-free, so seed-N runs should be effectively
// identical to pre-sprint baseline. The threshold below is set well under
// the seed=1 baseline (~2.4 LE at 1h on neutral) so all three seeds clear
// it; a regression that genuinely halved earnings would drop the lowest
// seeds underwater. The byte-for-byte equality check lives in the manual
// verification step (./bin/meowmine-sim --seed=N), not here.
func TestSimEarningsNotHalvedBySprint4(t *testing.T) {
	for _, seed := range []int64{1, 2, 3} {
		withTempHome(t)
		s := runSim(t, seed, 3600)
		if s.LifetimeEarned <= 1.0 {
			t.Errorf("seed=%d: LifetimeEarned=%v collapsed under pre-sprint baseline", seed, s.LifetimeEarned)
		}
		if s.NetworkCongestion < congestionMin-floatEq || s.NetworkCongestion > congestionMax+floatEq {
			t.Errorf("seed=%d: NetworkCongestion=%v out of [%v,%v]",
				seed, s.NetworkCongestion, congestionMin, congestionMax)
		}
	}
}
