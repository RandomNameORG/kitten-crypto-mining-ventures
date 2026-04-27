package game

// Pool-risk modifiers (additive) folded into the per-room baseline stale
// rate. Catalog-style block so future tuning lands in one place rather
// than as scattered magic numbers in advanceMining. Values picked tame
// (≤2pp) so even an extreme-risk pool still leaves the conservative
// starter path inside the sprint balance budget.
const (
	staleRiskLow     = 0.000
	staleRiskMedium  = 0.005
	staleRiskHigh    = 0.010
	staleRiskExtreme = 0.020
)

// staleRateMax caps the combined room+pool stale rate. Stops a future
// hot-room + extreme-pool + event multiplier from black-holing earnings.
const staleRateMax = 0.5

// poolRiskStaleModifier returns the additive stale-rate bump for a pool
// risk tier. Unknown tiers map to 0 — defensive against catalog drift.
func poolRiskStaleModifier(risk string) float64 {
	switch risk {
	case "low":
		return staleRiskLow
	case "medium":
		return staleRiskMedium
	case "high":
		return staleRiskHigh
	case "extreme":
		return staleRiskExtreme
	}
	return 0
}

// EffectiveStaleRate returns the room's runtime stale-rate after folding
// in the current pool's risk modifier. Mid pool-switch we return the
// room baseline only — mining is paused during the transition (see
// IsPoolSwitching) so the modifier wouldn't actually flow through to
// earnings. Same convention as PoolFee returning 0 mid-switch.
func (s *State) EffectiveStaleRate(roomID string) float64 {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return 0
	}
	base := rs.StaleRate
	if s.PoolSwitchAt > 0 {
		return clampStale(base)
	}
	return clampStale(base + poolRiskStaleModifier(s.CurrentPool().Risk))
}

func clampStale(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > staleRateMax {
		return staleRateMax
	}
	return v
}
