package game

import (
	"math"
	"testing"
)

// TestSyndicateGateEnforced — JoinSyndicate should refuse below the
// threshold and accept at or above it. The gate is the only protection
// against an early-game player tanking their BTC on the 10% cut before
// the pool ever pays out.
func TestSyndicateGateEnforced(t *testing.T) {
	withTempHome(t)
	s := NewState("Gate")
	now := simTestBaseUnix

	s.LifetimeEarned = SyndicateJoinThreshold - 1
	if err := s.JoinSyndicate(now); err == nil {
		t.Fatalf("expected gate to refuse below threshold")
	}
	if s.SyndicateJoined {
		t.Fatalf("Joined flag should not flip on a refused join")
	}

	s.LifetimeEarned = SyndicateJoinThreshold
	if err := s.JoinSyndicate(now); err != nil {
		t.Fatalf("expected join at threshold to succeed; got %v", err)
	}
	if !s.SyndicateJoined {
		t.Fatalf("Joined flag should flip on a successful join")
	}
	if s.SyndicateJoinedAt != now || s.SyndicateLastPayoutUnix != now {
		t.Errorf("join should stamp JoinedAt and LastPayoutUnix to now")
	}
}

// TestSyndicatePayoutMath — seed a ripe accumulator, call advanceSyndicate,
// and assert the dividend lands with the expected 1.20× bonus, the
// accumulator zeroes, and the total-dividend ledger updates.
func TestSyndicatePayoutMath(t *testing.T) {
	withTempHome(t)
	s := NewState("Payout")
	now := simTestBaseUnix + SyndicatePayoutIntervalSec + 1
	s.SyndicateJoined = true
	s.SyndicateContribution = 1000
	s.SyndicateLastPayoutUnix = now - (SyndicatePayoutIntervalSec + 1)

	btcBefore := s.BTC
	s.advanceSyndicate(now)

	wantDividend := 1000 * SyndicateDividendMult // 1200
	got := s.BTC - btcBefore
	if math.Abs(got-wantDividend) > 1e-9 {
		t.Errorf("BTC gain = %v, want %v", got, wantDividend)
	}
	if s.SyndicateContribution != 0 {
		t.Errorf("contribution should reset to 0, got %v", s.SyndicateContribution)
	}
	if math.Abs(s.SyndicateTotalDividends-wantDividend) > 1e-9 {
		t.Errorf("TotalDividends = %v, want %v", s.SyndicateTotalDividends, wantDividend)
	}
	// LastPayoutUnix should advance by exactly the interval, not snap to now —
	// otherwise offline catch-up would drift the weekly cadence.
	wantLast := (now - (SyndicatePayoutIntervalSec + 1)) + SyndicatePayoutIntervalSec
	if s.SyndicateLastPayoutUnix != wantLast {
		t.Errorf("LastPayoutUnix = %d, want %d", s.SyndicateLastPayoutUnix, wantLast)
	}
}

// TestSyndicateJoinEarnLeave — full lifecycle against a live tick loop. Joining
// applies the 10% haircut to BTC, contribution accumulates proportionally,
// and Leave pays the flat fee while forfeiting the pending contribution.
func TestSyndicateJoinEarnLeave(t *testing.T) {
	withTempHome(t)
	SeedRNG(1)
	s := NewState("Lifecycle")
	s.SetDifficulty("normal")

	base := simTestBaseUnix
	s.LastTickUnix = base
	s.LastBillUnix = base
	s.LastWagesUnix = base
	s.LastMarketTickUnix = base
	// Make sure we're eligible — no actual earning needed to pass the gate.
	s.LifetimeEarned = SyndicateJoinThreshold
	// Seed enough cash to survive the leave fee later.
	s.BTC = SyndicateLeaveFee * 2

	if err := s.JoinSyndicate(base); err != nil {
		t.Fatalf("join: %v", err)
	}

	// Run 10 ticks of mining so the starter GPU accumulates some earn.
	for i := 1; i <= 10; i++ {
		s.Tick(base + int64(i))
	}
	if s.SyndicateContribution <= 0 {
		t.Fatalf("contribution should accumulate while joined; got %v", s.SyndicateContribution)
	}

	// Now leave. Pending contribution should zero out; BTC drops by the fee.
	btcBefore := s.BTC
	pendingForfeited := s.SyndicateContribution
	if err := s.LeaveSyndicate(); err != nil {
		t.Fatalf("leave: %v", err)
	}
	if s.SyndicateJoined {
		t.Errorf("Joined flag should be false after leave")
	}
	if s.SyndicateContribution != 0 {
		t.Errorf("leaving must forfeit contribution; got %v", s.SyndicateContribution)
	}
	drop := btcBefore - s.BTC
	if math.Abs(drop-SyndicateLeaveFee) > 1e-9 {
		t.Errorf("leave fee drop = %v, want %v", drop, SyndicateLeaveFee)
	}
	// Sanity: the forfeited pool wasn't silently re-added to BTC.
	_ = pendingForfeited
}

// TestSyndicateMultiWeekCatchup — a single advanceSyndicate call pays a
// dividend per elapsed interval, but since the accumulator is a single
// bucket, the first payout empties it and subsequent payouts in the same
// call are zero-value. The key invariant is that LastPayoutUnix advances
// cleanly by N * interval so the cadence doesn't drift after a long
// offline gap.
func TestSyndicateMultiWeekCatchup(t *testing.T) {
	withTempHome(t)
	s := NewState("Catchup")
	weeks := int64(3)
	s.SyndicateJoined = true
	s.SyndicateContribution = 500
	anchor := simTestBaseUnix
	s.SyndicateLastPayoutUnix = anchor
	now := anchor + weeks*SyndicatePayoutIntervalSec

	btcBefore := s.BTC
	s.advanceSyndicate(now)

	// First payout clears 500 * 1.20 = 600. Subsequent payouts this call
	// see an empty bucket and add nothing.
	wantGain := 500 * SyndicateDividendMult
	got := s.BTC - btcBefore
	if math.Abs(got-wantGain) > 1e-9 {
		t.Errorf("BTC gain = %v, want %v (one-bucket payout only)", got, wantGain)
	}
	// LastPayoutUnix must advance by exactly weeks*interval so the cadence
	// anchor stays aligned — not snap to `now`.
	if s.SyndicateLastPayoutUnix != anchor+weeks*SyndicatePayoutIntervalSec {
		t.Errorf("LastPayoutUnix = %d, want %d (clean multi-week advance)",
			s.SyndicateLastPayoutUnix, anchor+weeks*SyndicatePayoutIntervalSec)
	}
	if s.SyndicateContribution != 0 {
		t.Errorf("contribution should be empty after catch-up; got %v", s.SyndicateContribution)
	}
}

// TestSyndicateUnjoinedAdvanceNoop — advanceSyndicate on a not-joined state
// must be inert. Guards against a refactor that accidentally credits
// dividends to someone who never joined.
func TestSyndicateUnjoinedAdvanceNoop(t *testing.T) {
	withTempHome(t)
	s := NewState("NotJoined")
	s.SyndicateContribution = 9999 // leftover from a prior run, e.g. after leave
	btcBefore := s.BTC
	s.advanceSyndicate(simTestBaseUnix + 10*SyndicatePayoutIntervalSec)
	if s.BTC != btcBefore {
		t.Errorf("unjoined state should not receive dividends")
	}
}
