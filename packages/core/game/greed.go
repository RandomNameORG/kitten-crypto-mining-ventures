package game

// GreedScore returns a 0..1 multiplier that rises with how much cash the
// player has accumulated recently. It feeds into the event fire rate:
// success attracts thieves. The formula is intentionally gentle — at
// $10k recent earnings the modifier is ~+5%, at $100k ~+20%, capped at
// +40%.
//
// We don't track a rolling window explicitly — that'd mean a timeseries
// in State. Instead we use a proxy: liquid cash on hand compared to the
// player's lifetime total. Rich-right-now > rich-lifetime.
func (s *State) GreedScore() float64 {
	if s.LifetimeEarned < 1 {
		return 0
	}
	// Fraction of lifetime earnings that's sitting liquid right now.
	liquid := s.BTC / s.LifetimeEarned
	if liquid > 1 {
		liquid = 1
	}
	// Log-ish ramp so early game stays calm.
	score := liquid * 0.4
	if score > 0.4 {
		score = 0.4
	}
	return score
}
