package game

import (
	"fmt"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"
)

// Syndicate — a late-game cooperative mining pool.
//
// Joining diverts SyndicateCutRate of each GPU's per-second earn into a
// contribution accumulator. Once a week (SyndicatePayoutIntervalSec virtual
// seconds) the pool pays the accumulated amount back at SyndicateDividendMult,
// netting a 20% bonus for staying in. Leaving costs SyndicateLeaveFee in cash
// and forfeits any unpaid contribution — a real commitment choice, not a
// free yield switch.

const (
	// SyndicateJoinThreshold is raw LifetimeEarned required before the
	// player can join. Tuned to land roughly mid-late run on normal.
	SyndicateJoinThreshold = 500_000.0

	// SyndicateCutRate is the fraction of per-GPU earn diverted into the
	// contribution pool while Joined. 0.10 = 10%.
	SyndicateCutRate = 0.10

	// SyndicatePayoutIntervalSec is the virtual-seconds cadence of dividend
	// payouts — exactly one week.
	SyndicatePayoutIntervalSec int64 = 7 * 24 * 3600

	// SyndicateDividendMult scales the accumulated contribution at payout
	// time — 1.20 gives a 20% edge vs. keeping the full 100% solo.
	SyndicateDividendMult = 1.20

	// SyndicateLeaveFee is the flat BTC cost to leave the syndicate. The
	// fee is deducted from BTC; any unpaid contribution is forfeited.
	SyndicateLeaveFee = 2500.0

	// SyndicateDividendTPBonus is the flat TP awarded each time the pool
	// pays a non-zero dividend. Weekly cadence keeps it modest (~5 TP per
	// virtual week) while still tying TP income to the late-game co-op.
	SyndicateDividendTPBonus = 5
)

// CanJoinSyndicate reports whether the player meets the threshold. It's
// independent of the current Joined flag — used by the UI to render the
// gate hint even while the player is already in.
func (s *State) CanJoinSyndicate() bool {
	return s.LifetimeEarned >= SyndicateJoinThreshold
}

// JoinSyndicate enrolls the player. Free. Fails below threshold.
func (s *State) JoinSyndicate(now int64) error {
	if s.SyndicateJoined {
		return fmt.Errorf("already in the syndicate")
	}
	if !s.CanJoinSyndicate() {
		s.appendLog("info", i18n.T("log.syndicate.gate_failed",
			FmtBTC(s.LifetimeEarned), FmtBTC(SyndicateJoinThreshold)))
		return fmt.Errorf("need %s lifetime earned; have %s",
			FmtBTC(SyndicateJoinThreshold), FmtBTC(s.LifetimeEarned))
	}
	s.SyndicateJoined = true
	s.SyndicateJoinedAt = now
	s.SyndicateLastPayoutUnix = now
	s.SyndicateContribution = 0
	s.appendLog("opportunity", i18n.T("log.syndicate.joined"))
	return nil
}

// LeaveSyndicate exits the pool for a flat fee. The unpaid accumulator is
// forfeited — if you want its value, wait for the next payout first.
func (s *State) LeaveSyndicate() error {
	if !s.SyndicateJoined {
		return fmt.Errorf("not in the syndicate")
	}
	if s.BTC < SyndicateLeaveFee {
		return fmt.Errorf("need %s to leave, have %s",
			FmtBTC(SyndicateLeaveFee), FmtBTC(s.BTC))
	}
	s.BTC -= SyndicateLeaveFee
	s.SyndicateJoined = false
	s.SyndicateJoinedAt = 0
	s.SyndicateLastPayoutUnix = 0
	s.SyndicateContribution = 0
	s.appendLog("info", i18n.T("log.syndicate.left", FmtBTC(SyndicateLeaveFee)))
	return nil
}

// advanceSyndicate pays out whenever the payout interval has elapsed. It
// pays one dividend per elapsed window, advancing LastPayoutUnix by the
// interval each time so offline catch-up rolls cleanly without drifting
// the cadence. The accumulator is a single bucket, so missed windows
// effectively merge into the next payout — the loop fires exactly once
// per interval regardless.
func (s *State) advanceSyndicate(now int64) {
	if !s.SyndicateJoined {
		return
	}
	if s.SyndicateLastPayoutUnix == 0 {
		s.SyndicateLastPayoutUnix = now
		return
	}
	for now-s.SyndicateLastPayoutUnix >= SyndicatePayoutIntervalSec {
		dividend := s.SyndicateContribution * SyndicateDividendMult
		s.SyndicateContribution = 0
		s.SyndicateLastPayoutUnix += SyndicatePayoutIntervalSec
		if dividend > 0 {
			s.BTC += dividend
			s.LifetimeEarned += dividend
			s.SyndicateTotalDividends += dividend
			s.appendLog("opportunity", i18n.T("log.syndicate.payout", FmtBTC(dividend)))
			s.TechPoint += SyndicateDividendTPBonus
			s.appendLog("opportunity", i18n.T("log.syndicate.tp_bonus",
				SyndicateDividendTPBonus))
		}
	}
}

// SecondsUntilNextSyndicatePayout returns the remaining virtual seconds
// until the next payout fires. Returns 0 when not joined (nothing to show).
func (s *State) SecondsUntilNextSyndicatePayout() int64 {
	if !s.SyndicateJoined || s.SyndicateLastPayoutUnix == 0 {
		return 0
	}
	remaining := SyndicatePayoutIntervalSec - (s.LastTickUnix - s.SyndicateLastPayoutUnix)
	if remaining < 0 {
		return 0
	}
	return remaining
}
