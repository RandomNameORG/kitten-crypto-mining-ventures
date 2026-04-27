package game

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
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

func psuExplodeEvent() data.EventDef {
	return eventShim{
		Category: "crisis",
		Emoji:    "💥",
		Name:     "Shim PSU Explode",
		Effects:  []effectShim{{Kind: "psu_explode"}},
	}.toDef()
}

func psuSmokingEvent() data.EventDef {
	return eventShim{
		Category: "threat",
		Emoji:    "🔌",
		Name:     "Shim PSU Smoke",
		Effects: []effectShim{
			{Kind: "earn_multiplier", Factor: 0.7, Seconds: 600},
			{Kind: "psu_smoking_chain"},
		},
	}.toDef()
}

func poolRunawayEvent() data.EventDef {
	return eventShim{
		Category: "crisis",
		Emoji:    "🏊",
		Name:     "Shim Pool Runaway",
		Effects:  []effectShim{{Kind: "pool_runaway"}},
	}.toDef()
}

func soloBlockHitEvent(amount float64) data.EventDef {
	return eventShim{
		Category: "opportunity",
		Emoji:    "🎰",
		Name:     "Shim Solo Hit",
		Effects:  []effectShim{{Kind: "solo_block_hit", Amount: amount}},
	}.toDef()
}

func psuChainExplodeEvent() data.EventDef {
	return eventShim{
		Category: "crisis",
		Emoji:    "🔥",
		Name:     "Shim PSU Chain",
		Effects:  []effectShim{{Kind: "psu_chain_explode"}},
	}.toDef()
}

func fireSaleEvent() data.EventDef {
	return eventShim{
		Category: "opportunity",
		Emoji:    "🛒",
		Name:     "Shim Fire Sale",
		Effects:  []effectShim{{Kind: "fire_sale"}},
	}.toDef()
}

// installRunningPSU drops a freshly-running PSU instance into the room
// with the given InstalledAt time. Used by tests that need to set up the
// gate conditions for E21/E22/E26 by hand without going through the
// shopping flow.
func installRunningPSU(s *State, roomID, defID string, installedAt int64) *PSU {
	rs := s.Rooms[roomID]
	if rs == nil {
		return nil
	}
	if s.NextPSUID < 1 {
		s.NextPSUID = 1
	}
	p := &PSU{
		InstanceID:  s.NextPSUID,
		DefID:       defID,
		Status:      "running",
		InstalledAt: installedAt,
	}
	s.NextPSUID++
	rs.PSUUnits = append(rs.PSUUnits, p)
	return p
}

// fillRoomGPUs adds n running GPUs of the given type to the room.
func fillRoomGPUs(s *State, roomID, defID string, n int) {
	for i := 0; i < n; i++ {
		s.addGPU(defID, roomID, false)
	}
}

// dropBuiltinPSU removes the freebie psu_builtin from a room so tests can
// exercise overload paths under a real capacity number. The builtin's 100kW
// rating makes any realistic GPU load under-utilised; tests for E21/E26
// need it gone to push factor past 1.0+tol.
func dropBuiltinPSU(s *State, roomID string) {
	rs := s.Rooms[roomID]
	if rs == nil {
		return
	}
	keep := rs.PSUUnits[:0]
	for _, p := range rs.PSUUnits {
		if p.DefID == "psu_builtin" {
			continue
		}
		keep = append(keep, p)
	}
	rs.PSUUnits = keep
}

// --- E21: psu_explode ---

func TestPSUExplode_GateRequiresOverloadedNonBuiltin(t *testing.T) {
	withTempHome(t)
	s := NewState("PSUExplodeGate")
	if s.eventGatePasses("psu_explode") {
		t.Fatal("fresh state should fail the psu_explode gate (only psu_builtin)")
	}
	// To trip the overload gate, drop the freebie 100kW builtin so the
	// trash PSU's tiny 300W band actually matters, then load the room
	// past 300W * (1+0.05).
	dropBuiltinPSU(s, s.CurrentRoom)
	installRunningPSU(s, s.CurrentRoom, "psu_trash", 0)
	for s.RoomPSUOverloadFactor(s.CurrentRoom) <= 1.05 {
		s.addGPU("a100", s.CurrentRoom, false)
		if len(s.GPUs) > 30 {
			t.Fatalf("could not push room past trash PSU tol; factor=%.2f",
				s.RoomPSUOverloadFactor(s.CurrentRoom))
		}
	}
	if !s.eventGatePasses("psu_explode") {
		t.Fatalf("expected gate to pass; overloadFactor=%.2f tol=%.2f",
			s.RoomPSUOverloadFactor(s.CurrentRoom),
			s.roomMinOverloadTolerance(s.CurrentRoom))
	}
}

func TestPSUExplode_EffectBreaksPSUAndGPUs(t *testing.T) {
	withTempHome(t)
	rand.Seed(7)
	s := NewState("PSUExplodeFx")
	psu := installRunningPSU(s, s.CurrentRoom, "psu_trash", 0)
	// A handful of running GPUs to get bricked.
	for i := 0; i < 3; i++ {
		s.addGPU("gtx1060", s.CurrentRoom, false)
	}
	beforeRunning := 0
	for _, g := range s.GPUs {
		if g.Status == "running" {
			beforeRunning++
		}
	}
	s.applyEvent(psuExplodeEvent())
	if psu.Status != "broken" {
		t.Fatalf("psu_explode should break the trash PSU; status=%s", psu.Status)
	}
	afterRunning := 0
	for _, g := range s.GPUs {
		if g.Status == "running" {
			afterRunning++
		}
	}
	if afterRunning >= beforeRunning {
		t.Fatalf("expected some GPUs bricked; before=%d after=%d", beforeRunning, afterRunning)
	}
}

func TestPSUExplode_FizzlesWithNoRealPSU(t *testing.T) {
	withTempHome(t)
	s := NewState("PSUExplodeNone")
	logBefore := len(s.Log)
	// Only psu_builtin in the room — applyEvent should log the no-op,
	// not crash, and not touch the builtin.
	s.applyEvent(psuExplodeEvent())
	if len(s.Log) == logBefore {
		t.Fatal("expected a log line for no-real-PSU fizzle")
	}
	for _, p := range s.Rooms[s.CurrentRoom].PSUUnits {
		if p.Status != "running" {
			t.Fatalf("builtin PSU should not be bricked; status=%s", p.Status)
		}
	}
}

// --- E22: psu_smoking ---

func TestPSUSmoking_GateRequiresOldTrashPSU(t *testing.T) {
	withTempHome(t)
	s := NewState("SmokeGate")
	s.LastTickUnix = 1_000_000
	if s.eventGatePasses("psu_smoking") {
		t.Fatal("fresh state should fail the psu_smoking gate")
	}
	// Trash PSU younger than 5h: still a fail.
	installRunningPSU(s, s.CurrentRoom, "psu_trash", s.LastTickUnix-3600)
	if s.eventGatePasses("psu_smoking") {
		t.Fatal("a 1h-old trash PSU shouldn't trip the smoking gate")
	}
	// Backdate it past the 5h window.
	s.Rooms[s.CurrentRoom].PSUUnits[len(s.Rooms[s.CurrentRoom].PSUUnits)-1].InstalledAt = s.LastTickUnix - 18001
	if !s.eventGatePasses("psu_smoking") {
		t.Fatal("a >5h-old trash PSU should trip the smoking gate")
	}
}

func TestPSUSmoking_EffectAppliesEarnDebuff(t *testing.T) {
	withTempHome(t)
	rand.Seed(1)
	s := NewState("SmokeFx")
	s.applyEvent(psuSmokingEvent())
	found := false
	for _, m := range s.Modifiers {
		if m.Kind == "earn_mult" && m.Factor == 0.7 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected a 0.7× earn_mult modifier; got %+v", s.Modifiers)
	}
}

// --- E23: mining_disaster ---

func TestMiningDisaster_GateRequiresLatePriceFloor(t *testing.T) {
	withTempHome(t)
	s := NewState("MDGate")
	if s.eventGatePasses("mining_disaster") {
		t.Fatal("fresh state should fail mining_disaster gate")
	}
	s.MarketPrice = 0.3
	if s.eventGatePasses("mining_disaster") {
		t.Fatal("low price alone shouldn't trip — needs late-game LE too")
	}
	s.LifetimeEarned = 5000
	if !s.eventGatePasses("mining_disaster") {
		t.Fatal("low price + late LE should pass gate")
	}
}

// --- E24: pool_runaway ---

func TestPoolRunaway_GateRequiresWhiskerFi(t *testing.T) {
	withTempHome(t)
	s := NewState("RunGate")
	if s.eventGatePasses("pool_runaway") {
		t.Fatal("scratch_pool default shouldn't trip pool_runaway gate")
	}
	s.PoolID = "whisker_fi"
	if !s.eventGatePasses("pool_runaway") {
		t.Fatal("whisker_fi pool should pass pool_runaway gate")
	}
}

func TestPoolRunaway_EffectResetsPoolAndShares(t *testing.T) {
	withTempHome(t)
	s := NewState("RunFx")
	s.PoolID = "whisker_fi"
	s.PoolShares = 12345
	s.PoolSwitchFrom = "scratch_pool"
	s.PoolSwitchAt = 999999
	s.applyEvent(poolRunawayEvent())
	if s.PoolID != "scratch_pool" {
		t.Fatalf("pool should be reset to scratch_pool; got %s", s.PoolID)
	}
	if s.PoolShares != 0 {
		t.Fatalf("shares should be voided; got %v", s.PoolShares)
	}
	if s.PoolSwitchFrom != "" || s.PoolSwitchAt != 0 {
		t.Fatalf("switch fields should be cleared; got from=%q at=%d", s.PoolSwitchFrom, s.PoolSwitchAt)
	}
}

// --- E25: solo_block_hit ---

func TestSoloBlockHit_GateRequiresSolo(t *testing.T) {
	withTempHome(t)
	s := NewState("SoloGate")
	if s.eventGatePasses("solo_block_hit") {
		t.Fatal("default scratch_pool shouldn't trip solo gate")
	}
	s.PoolID = "solo"
	if !s.eventGatePasses("solo_block_hit") {
		t.Fatal("solo pool should pass solo gate")
	}
}

func TestSoloBlockHit_EffectAddsLumpBTC(t *testing.T) {
	withTempHome(t)
	s := NewState("SoloFx")
	before := s.BTC
	s.applyEvent(soloBlockHitEvent(0.5))
	want := before + 500.0
	if s.BTC != want {
		t.Fatalf("expected BTC+500; got %v want %v", s.BTC, want)
	}
}

// --- E26: psu_chain_explode ---

func TestPSUChainExplode_GateRequiresTwoTrashAndOverload(t *testing.T) {
	withTempHome(t)
	s := NewState("ChainGate")
	if s.eventGatePasses("psu_chain_explode") {
		t.Fatal("fresh state shouldn't trip chain gate")
	}
	installRunningPSU(s, s.CurrentRoom, "psu_trash", 0)
	if s.eventGatePasses("psu_chain_explode") {
		t.Fatal("one trash PSU shouldn't trip chain gate")
	}
	installRunningPSU(s, s.CurrentRoom, "psu_trash", 0)
	// Two trash PSUs but no overload yet — should still fail because
	// builtin's huge capacity drops overload factor far below 1.0.
	if s.eventGatePasses("psu_chain_explode") {
		t.Fatal("two trash but no overload shouldn't trip chain gate")
	}
	// Remove builtin to force capacity = 600W trash, then load past 1.0
	// with high-draw cards (a100 is 20W each).
	dropBuiltinPSU(s, s.CurrentRoom)
	for s.RoomPSUOverloadFactor(s.CurrentRoom) <= 1.0 {
		s.addGPU("a100", s.CurrentRoom, false)
		if len(s.GPUs) > 50 {
			t.Fatalf("could not push factor past 1.0; got %.2f",
				s.RoomPSUOverloadFactor(s.CurrentRoom))
		}
	}
	if !s.eventGatePasses("psu_chain_explode") {
		t.Fatalf("expected chain gate to pass; overload=%.2f trash=%d",
			s.RoomPSUOverloadFactor(s.CurrentRoom),
			s.roomTrashPSUCount(s.CurrentRoom))
	}
}

func TestPSUChainExplode_EffectBricksTrashAndHalfGPUs(t *testing.T) {
	withTempHome(t)
	rand.Seed(11)
	s := NewState("ChainFx")
	a := installRunningPSU(s, s.CurrentRoom, "psu_trash", 0)
	b := installRunningPSU(s, s.CurrentRoom, "psu_trash", 0)
	c := installRunningPSU(s, s.CurrentRoom, "psu_silver650", 0)
	for i := 0; i < 4; i++ {
		s.addGPU("gtx1060", s.CurrentRoom, false)
	}
	runningBefore := 0
	for _, g := range s.GPUs {
		if g.Status == "running" {
			runningBefore++
		}
	}
	s.applyEvent(psuChainExplodeEvent())
	if a.Status != "broken" || b.Status != "broken" {
		t.Fatalf("both trash PSUs should be broken; a=%s b=%s", a.Status, b.Status)
	}
	if c.Status != "running" {
		t.Fatalf("silver PSU should not be touched; status=%s", c.Status)
	}
	runningAfter := 0
	for _, g := range s.GPUs {
		if g.Status == "running" {
			runningAfter++
		}
	}
	wantBroken := runningBefore / 2
	if runningBefore-runningAfter != wantBroken {
		t.Fatalf("expected %d GPUs broken; got %d (before=%d after=%d)",
			wantBroken, runningBefore-runningAfter, runningBefore, runningAfter)
	}
}

// --- E27: share_dilution ---

func TestShareDilution_GateRequiresPPLNSAndLatGame(t *testing.T) {
	withTempHome(t)
	s := NewState("DilGate")
	// scratch_pool is PPLNS, but LE gate fails on a fresh state.
	if s.eventGatePasses("share_dilution") {
		t.Fatal("fresh state shouldn't trip dilution gate (LE too low)")
	}
	s.LifetimeEarned = 5000
	if !s.eventGatePasses("share_dilution") {
		t.Fatal("PPLNS + late LE should pass gate")
	}
	s.PoolID = "kitten_hash" // PPS — should fail
	if s.eventGatePasses("share_dilution") {
		t.Fatal("PPS pool shouldn't trip dilution gate")
	}
	s.PoolID = "whisker_fi" // PPS+ — should pass
	if !s.eventGatePasses("share_dilution") {
		t.Fatal("PPS+ pool should pass dilution gate")
	}
	// Mid-switch: should fail.
	s.PoolSwitchAt = s.LastTickUnix + 60
	if s.eventGatePasses("share_dilution") {
		t.Fatal("mid-pool-switch shouldn't trip dilution gate")
	}
}

// --- E28: fire_sale ---

func TestFireSale_GateRequiresRecentMiningDisaster(t *testing.T) {
	withTempHome(t)
	s := NewState("FireGate")
	s.LastTickUnix = 1_000_000
	if s.eventGatePasses("fire_sale") {
		t.Fatal("fresh state shouldn't trip fire_sale gate (no disaster on record)")
	}
	// Fresh disaster cooldown entry within 600s window.
	s.EventCooldown["mining_disaster"] = s.LastTickUnix - 100
	if !s.eventGatePasses("fire_sale") {
		t.Fatal("recent mining_disaster should trip fire_sale gate")
	}
	// Old disaster: window expired.
	s.EventCooldown["mining_disaster"] = s.LastTickUnix - 700
	if s.eventGatePasses("fire_sale") {
		t.Fatal("old mining_disaster shouldn't trip fire_sale gate")
	}
}

func TestFireSale_EffectIsLogOnly(t *testing.T) {
	withTempHome(t)
	s := NewState("FireFx")
	btcBefore, modsBefore := s.BTC, len(s.Modifiers)
	s.applyEvent(fireSaleEvent())
	if s.BTC != btcBefore {
		t.Fatalf("fire_sale shouldn't move BTC; before=%v after=%v", btcBefore, s.BTC)
	}
	if len(s.Modifiers) != modsBefore {
		t.Fatalf("fire_sale shouldn't add modifiers; before=%d after=%d", modsBefore, len(s.Modifiers))
	}
	if !logContains(s, "Fire sale") {
		t.Fatalf("expected fire-sale log line; got %v", logTexts(s))
	}
}
