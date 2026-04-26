package game

// MiningScale converts raw GPU efficiency (the pre-unification BTC/sec rate)
// into the balance-scale BTC units shown in the UI. Tuned so a GTX 1060
// (eff 0.0012) at scale 300 yields ₿0.36/sec — a 1060 pays back in ~5 min,
// an A100 (₿60k) in ~80 min of dedicated uptime. Numbers preserved from
// the earlier USD-priced economy so difficulty curves stay intact.
const MiningScale = 300.0

// ElectricPerVoltMin is the per-V per-minute price of electricity. Rooms
// multiply this by their electric_cost_mult. Tuned so bills are ~5-10% of
// earnings on the base room — meaningful, but not death-spiral-inducing
// unless you run an oversubscribed rack.
const ElectricPerVoltMin = 0.25

// earnMultiplier returns the aggregate production multiplier from active
// time-limited modifiers (events, skill actions).
func (s *State) earnMultiplier(now int64) float64 {
	mult := 1.0
	for _, m := range s.Modifiers {
		if m.Kind == "earn_mult" && m.ExpiresAt > now {
			mult *= m.Factor
		}
	}
	return mult
}

// IsMiningPaused returns true if a pause_mining modifier is active.
func (s *State) IsMiningPaused(now int64) bool {
	for _, m := range s.Modifiers {
		if m.Kind == "pause_mining" && m.ExpiresAt > now {
			return true
		}
	}
	return false
}

// pruneModifiers removes expired modifiers.
func (s *State) pruneModifiers(now int64) {
	alive := s.Modifiers[:0]
	for _, m := range s.Modifiers {
		if m.ExpiresAt > now {
			alive = append(alive, m)
		}
	}
	s.Modifiers = alive
}
