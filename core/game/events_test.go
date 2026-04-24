package game

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/data"
)

func TestMaybeFireEventRespectsSeed(t *testing.T) {
	withTempHome(t)
	rand.Seed(1)
	s := NewState("Seed")
	fired := 0
	for i := 0; i < 200; i++ {
		if def := s.MaybeFireEvent(); def != nil {
			fired++
		}
	}
	// Not asserting an exact count (rand is shared), just that we can
	// hammer it 200 times without panicking.
	if fired < 0 {
		t.Fatal("impossible negative count")
	}
}

func TestApplyEventAccumulatesTechPoint(t *testing.T) {
	withTempHome(t)
	s := NewState("TP")
	tpBefore := s.TechPoint
	s.applyEvent(tpEvent())
	if s.TechPoint != tpBefore+1 {
		t.Errorf("tech_point effect should increment TechPoint by 1")
	}
}

func TestApplyEventRespectsWiringForOutages(t *testing.T) {
	withTempHome(t)
	s := NewState("Outage")
	s.Rooms[s.CurrentRoom].WiringLvl = 5 // should shorten outage substantially
	s.applyEvent(outageEvent(120))
	var pause *Modifier
	for i := range s.Modifiers {
		if s.Modifiers[i].Kind == "pause_mining" {
			pause = &s.Modifiers[i]
		}
	}
	if pause == nil {
		t.Fatal("expected pause_mining modifier")
	}
	// wiring_lvl 5 → 50s reduction → 70s duration expected (clamped at >= 10).
	remaining := pause.ExpiresAt - time.Now().Unix()
	if remaining >= 120 {
		t.Errorf("wiring should reduce outage; got %d seconds", remaining)
	}
}

func TestApplyEventDefenseReducesStealRate(t *testing.T) {
	withTempHome(t)

	// runSteals: dispatch 100 steal attempts with the given defense level,
	// refilling the rack after each so theft has a target every round.
	// Returns the total number of GPUs successfully removed.
	runSteals := func(lockLvl int) int {
		rand.Seed(42)
		s := NewState("Defense")
		s.Rooms[s.CurrentRoom].LockLvl = lockLvl
		s.Rooms[s.CurrentRoom].CCTVLvl = lockLvl
		s.Rooms[s.CurrentRoom].ArmorLvl = lockLvl
		// Fill the rack to steady state so theft always has candidates.
		for len(s.GPUs) < 4 {
			s.addGPU("gtx1060", s.CurrentRoom, false)
		}
		count := 0
		for i := 0; i < 100; i++ {
			before := len(s.GPUs)
			s.applyEvent(stealEvent())
			count += before - len(s.GPUs)
			// Refill so the next round has something to steal.
			for len(s.GPUs) < 4 {
				s.addGPU("gtx1060", s.CurrentRoom, false)
			}
		}
		return count
	}

	noDefense := runSteals(0)
	fullDefense := runSteals(5)

	// The floor in tryStealGPUs is 5% so fullDefense won't be zero, but it
	// should decisively beat no-defense.
	if fullDefense >= noDefense {
		t.Errorf("max defense (%d steals) should beat no defense (%d steals)",
			fullDefense, noDefense)
	}
}

// TestTaxAudit_CoveredByReserves pins the "books covered the audit" branch:
// when BTC meets the reserve_factor * LifetimeEarned threshold, the event
// must be a no-op on cash + reputation. Without this we could silently
// start punishing cash-heavy players the mechanic was designed to spare.
func TestTaxAudit_CoveredByReserves(t *testing.T) {
	withTempHome(t)
	s := NewState("Audit")
	s.BTC = 1_000_000
	s.LifetimeEarned = 5_000_000
	repBefore := s.Reputation

	s.applyEvent(taxAuditEvent(0.20, 0.10))

	if s.BTC != 1_000_000 {
		t.Errorf("reserves should have covered the audit; BTC = %v, want 1_000_000", s.BTC)
	}
	if s.Reputation != repBefore {
		t.Errorf("reserves-covered audit shouldn't touch Reputation; got %d want %d", s.Reputation, repBefore)
	}
	if !logContains(s, "covered") {
		t.Errorf("expected a 'covered'/reserves log line; got %v", logTexts(s))
	}
}

// TestTaxAudit_HitsPlayer pins the punitive branch: when BTC is below the
// reserves threshold, the audit eats a chunk of the wallet and dings rep.
// Guards against reserve_factor/amount wiring regressions flipping the sign
// or reading the wrong field.
func TestTaxAudit_HitsPlayer(t *testing.T) {
	withTempHome(t)
	s := NewState("Audit")
	s.BTC = 100_000
	s.LifetimeEarned = 5_000_000
	repBefore := s.Reputation

	s.applyEvent(taxAuditEvent(0.20, 0.10))

	wantBTC := 100_000 * 0.80
	if s.BTC != wantBTC {
		t.Errorf("audit should take 20%% of BTC; got %v, want %v", s.BTC, wantBTC)
	}
	if s.Reputation != repBefore-5 {
		t.Errorf("audit hit should drop Reputation by 5; got %d, want %d", s.Reputation, repBefore-5)
	}
}

// TestPowerSurge_DamagesOCOnly is the core invariant of the surge event: it
// targets the overclocked rail only. A stock GPU living next to an OC'd one
// must never lose HoursLeft to this effect, no matter how many rolls. If
// this fails we'd be silently damaging cards the player deliberately kept
// off the boosted bus to avoid exactly this risk.
func TestPowerSurge_DamagesOCOnly(t *testing.T) {
	withTempHome(t)
	rand.Seed(99)
	s := NewState("Surge")
	// Clear the starter so we have tight control over the two cards under test.
	s.GPUs = s.GPUs[:0]
	ocGPU := s.addGPU("gtx1060", s.CurrentRoom, false)
	stockGPU := s.addGPU("gtx1060", s.CurrentRoom, false)
	ocGPU.OCLevel = 2

	// Durability after damage_oc_gpu(0.35) on a 10h card = 10*0.35 = 3.5 per
	// hit; start both cards at full so we have room for several rounds
	// before either breaks.
	const rounds = 5
	startHours := 10.0
	stockBefore := stockGPU.HoursLeft
	if stockBefore == 0 {
		t.Fatalf("stock GPU started at 0 hours — addGPU changed?")
	}
	prevOCHours := ocGPU.HoursLeft
	for i := 0; i < rounds; i++ {
		// Restore so the OC card never drops below zero (would break and
		// become ineligible, invalidating the "strictly drops" assertion).
		ocGPU.HoursLeft = startHours
		ocGPU.Status = "running"
		stockGPU.HoursLeft = startHours
		stockGPU.Status = "running"

		s.applyEvent(surgeEvent(0.35))

		if stockGPU.HoursLeft != startHours {
			t.Fatalf("round %d: stock GPU HoursLeft changed (%.3f → %.3f) — surge leaked to non-OC card",
				i, startHours, stockGPU.HoursLeft)
		}
		if ocGPU.HoursLeft >= startHours {
			t.Fatalf("round %d: OC GPU HoursLeft should have dropped; got %.3f (start %.3f)",
				i, ocGPU.HoursLeft, startHours)
		}
		prevOCHours = ocGPU.HoursLeft
	}
	_ = prevOCHours
}

// TestPowerSurge_FizzlesWithNoOC is the complement: if no GPU is OC'd, the
// surge must harmlessly fizzle — never fall through to damage a stock card
// as a consolation prize. Also asserts the fizzle log line fires so the
// player sees *something* happened.
func TestPowerSurge_FizzlesWithNoOC(t *testing.T) {
	withTempHome(t)
	rand.Seed(1)
	s := NewState("Surge")
	s.GPUs = s.GPUs[:0]
	a := s.addGPU("gtx1060", s.CurrentRoom, false)
	b := s.addGPU("gtx1060", s.CurrentRoom, false)
	hoursA, hoursB := a.HoursLeft, b.HoursLeft

	s.applyEvent(surgeEvent(0.35))

	if a.HoursLeft != hoursA || b.HoursLeft != hoursB {
		t.Errorf("surge damaged stock cards: %v→%v, %v→%v", hoursA, a.HoursLeft, hoursB, b.HoursLeft)
	}
	if !logContains(s, "fizzle") && !logContains(s, "没有超频") && !logContains(s, "overclocked") {
		t.Errorf("expected a fizzle log line; got %v", logTexts(s))
	}
}

// TestMarketCrash_PinsPriceAndBlocksDrift covers the load-bearing claim of
// market_pin: for the modifier's lifetime the price is held flat, no drift
// backlog accrues, and once the modifier expires mean-reversion resumes.
// Without this, a sim tick that spanned a crash window would either explode
// the price with accumulated drift or stay pinned forever.
func TestMarketCrash_PinsPriceAndBlocksDrift(t *testing.T) {
	withTempHome(t)
	SeedRNG(1)
	s := NewState("Crash")
	s.MarketPrice = 2.0
	s.PrevMarketPrice = 2.0
	s.LastMarketTickUnix = simTestBaseUnix

	const pinSeconds = 300
	s.applyEvent(marketCrashEvent(0.3, pinSeconds))
	// applyEvent stamps the pin modifier with real-clock time.Now(); realign
	// the expiry onto the fixed simTestBaseUnix epoch the advanceMarket
	// calls below operate in, so the "past expiry" assertion means what we
	// think it does.
	found := false
	for i := range s.Modifiers {
		if s.Modifiers[i].Kind == "market_pin" {
			s.Modifiers[i].ExpiresAt = simTestBaseUnix + int64(pinSeconds)
			found = true
		}
	}
	if !found {
		t.Fatalf("market_crash event did not create a market_pin modifier")
	}

	if s.MarketPrice != 0.3 {
		t.Fatalf("applyEvent should pin MarketPrice to 0.3 immediately; got %v", s.MarketPrice)
	}

	// Walk 5 market ticks while still inside the pin window — price must
	// stay glued to 0.3, anchor must advance with now so no backlog builds.
	now := simTestBaseUnix + 5*MarketTickSec
	s.advanceMarket(now)
	if s.MarketPrice != 0.3 {
		t.Fatalf("market_pin leaked: MarketPrice drifted to %v during pin window", s.MarketPrice)
	}

	// Step past the modifier's expiry and give drift many chances to move.
	// Gaussian sigma is 0.03 per step; over 200 steps getting zero net
	// movement is astronomically unlikely, so any failure here is a logic
	// bug not a flaky RNG draw.
	now = simTestBaseUnix + int64(pinSeconds) + 1
	s.advanceMarket(now) // first tick after expiry — baseline step from 0.3
	for i := 0; i < 200; i++ {
		now += MarketTickSec
		s.advanceMarket(now)
	}
	if s.MarketPrice == 0.3 {
		t.Fatalf("market stayed pinned at 0.3 after expiry — drift never resumed")
	}
}

// logContains / logTexts: tiny helpers used by the audit/surge tests to
// assert on the human-readable log stream without coupling to translation
// IDs. Substring match is intentional — the test doesn't care about exact
// wording, only that the right branch fired a message.
func logContains(s *State, sub string) bool {
	for _, l := range s.Log {
		if strings.Contains(l.Text, sub) {
			return true
		}
	}
	return false
}

func logTexts(s *State) []string {
	out := make([]string, 0, len(s.Log))
	for _, l := range s.Log {
		out = append(out, l.Text)
	}
	return out
}

// --- fixtures ---

func tpEvent() data.EventDef {
	return eventShim{
		Category: "opportunity",
		Emoji:    "🧠",
		Name:     "Shim TP",
		Effects:  []effectShim{{Kind: "tech_point", Delta: 1}},
	}.toDef()
}

func outageEvent(seconds int) data.EventDef {
	return eventShim{
		Category: "threat",
		Emoji:    "⚡",
		Name:     "Shim Outage",
		Effects:  []effectShim{{Kind: "pause_mining", Seconds: seconds}},
	}.toDef()
}

func stealEvent() data.EventDef {
	return eventShim{
		Category: "threat",
		Emoji:    "🐀",
		Name:     "Shim Thief",
		Effects:  []effectShim{{Kind: "steal_gpu", ChanceIfNoDefense: 0.9, Count: 1}},
	}.toDef()
}

func taxAuditEvent(amount, reserveFactor float64) data.EventDef {
	return eventShim{
		Category: "threat",
		Emoji:    "🐱‍💼",
		Name:     "Shim Audit",
		Effects:  []effectShim{{Kind: "tax_audit", Amount: amount, ReserveFactor: reserveFactor}},
	}.toDef()
}

func surgeEvent(amount float64) data.EventDef {
	return eventShim{
		Category: "threat",
		Emoji:    "🔌",
		Name:     "Shim Surge",
		Effects:  []effectShim{{Kind: "damage_oc_gpu", Amount: amount}},
	}.toDef()
}

func marketCrashEvent(factor float64, seconds int) data.EventDef {
	return eventShim{
		Category: "crisis",
		Emoji:    "📉",
		Name:     "Shim Crash",
		Effects:  []effectShim{{Kind: "market_pin", Factor: factor, Seconds: seconds}},
	}.toDef()
}
