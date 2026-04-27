package game

import (
	"fmt"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
)

// PoolSwitchSec is the spec'd 10-minute transition window between pools
// (§5.5). During this window mining yields nothing — old pool is
// settling, new pool is registering the worker. The Tycoon-line
// `Pool Hopping` skill (next sprint) shortens it to 3 minutes.
const PoolSwitchSec int64 = 600

// CurrentPool returns the player's current PoolDef. Falls back to
// scratch_pool defensively if PoolID was somehow cleared mid-flight —
// every new game and migrated save lands on scratch_pool, so PoolID
// should never legitimately be empty after ensureInit.
func (s *State) CurrentPool() data.PoolDef {
	if def, ok := data.PoolByID(s.PoolID); ok {
		return def
	}
	if def, ok := data.PoolByID("scratch_pool"); ok {
		return def
	}
	return data.PoolDef{}
}

// IsPoolSwitching reports whether the 10-minute transition window is
// still open. While true, mining is paused exactly the way ReplacePSU's
// PSUResumeAt pauses a single room — see advanceMining.
func (s *State) IsPoolSwitching(now int64) bool {
	return s.PoolSwitchAt > now
}

// PoolFee returns the current pool's fee fraction (0.02 = 2%). Returns
// 0 mid-transition since no earnings flow during the switch window
// anyway — keeps the helper cleanly composable for next sprint's
// payout math.
func (s *State) PoolFee() float64 {
	if s.PoolSwitchAt > 0 {
		return 0
	}
	return s.CurrentPool().Fee
}

// PoolSettlementMode returns the current pool's settlement mode
// ("pps" | "pplns" | "pps_plus" | "solo"). Convenience wrapper over
// CurrentPool().SettlementMode for the call sites that only need the
// mode tag.
func (s *State) PoolSettlementMode() string {
	return s.CurrentPool().SettlementMode
}

// SwitchPool transitions the player from their current pool to newPoolID.
// Validates the target exists, refuses no-op switches and switches while
// already mid-transition, then opens the 10-minute window. Per §5.5,
// leaving a PPLNS pool voids any unsettled shares — this is the place we
// enforce that.
func (s *State) SwitchPool(newPoolID string, now int64) error {
	newDef, ok := data.PoolByID(newPoolID)
	if !ok {
		return fmt.Errorf("no such pool: %s", newPoolID)
	}
	if s.IsPoolSwitching(now) {
		return fmt.Errorf("already switching pools — wait for transition to complete")
	}
	if s.PoolID == newPoolID {
		return fmt.Errorf("already on %s", newDef.LocalName())
	}
	leavingDef := s.CurrentPool()
	// PPLNS shares evaporate when you walk away — that's the structural
	// gimmick that gives PPS its lower-variance appeal. Pool Hopping
	// (mogul T2) softens this to 50% retention via PoolHoppingShareRetention.
	if leavingDef.SettlementMode == "pplns" {
		s.PoolShares *= s.PoolHoppingShareRetention()
	}
	s.PoolSwitchFrom = s.PoolID
	s.PoolID = newPoolID
	switchSec := s.PoolSwitchDurationSec()
	s.PoolSwitchAt = now + switchSec
	s.appendLog("info", fmt.Sprintf("Switching pool: %s → %s (%ds transition)",
		leavingDef.LocalName(), newDef.LocalName(), switchSec))
	return nil
}
