package game

import "testing"

// TestRetireCarriesTP — Retire banks 25% of unspent TP (floor) into the
// legacy store, capped at PrestigeTPCarryCap. The fresh state then loads
// that carry into TechPoint exactly once and zeroes the bank so a re-load
// can't re-credit it.
func TestRetireCarriesTP(t *testing.T) {
	// Retire grants "first_retire" inline, which itself credits a small
	// TPReward on the *outgoing* state before the carry is computed. We
	// pre-mark it owned so each scenario sees a clean pre-carry TechPoint
	// equal to the value we set, with no implicit +TP from the achievement
	// chain leaking into the snapshot.
	preRetire := func(name string, tp int) *State {
		s := NewState(name)
		s.UnlockedSkills["venture_cap"] = true
		s.LifetimeEarned = PrestigeThreshold + 1
		s.Achievements = append(s.Achievements, "first_retire")
		s.TechPoint = tp
		return s
	}

	t.Run("100 TP carries 25", func(t *testing.T) {
		withTempHome(t)
		s := preRetire("Carry100", 100)

		fresh, _, err := s.Retire()
		if err != nil {
			t.Fatalf("Retire: %v", err)
		}
		if fresh.TechPoint != 25 {
			t.Errorf("fresh TechPoint = %d, want 25 (25%% of 100)", fresh.TechPoint)
		}

		// Bank should now be drained — a second NewState read off the same
		// legacy file must NOT re-credit the carry.
		again := NewState("Carry100Reload")
		if again.TechPoint != 0 {
			t.Errorf("re-loaded fresh state inherited %d TP from a drained carry; bank not zeroed",
				again.TechPoint)
		}
	})

	t.Run("1000 TP caps at 200", func(t *testing.T) {
		withTempHome(t)
		s := preRetire("CarryCap", 1000)

		fresh, _, err := s.Retire()
		if err != nil {
			t.Fatalf("Retire: %v", err)
		}
		if fresh.TechPoint != PrestigeTPCarryCap {
			t.Errorf("fresh TechPoint = %d, want %d (cap)", fresh.TechPoint, PrestigeTPCarryCap)
		}
	})

	t.Run("zero TP carries nothing", func(t *testing.T) {
		withTempHome(t)
		s := preRetire("CarryZero", 0)

		fresh, _, err := s.Retire()
		if err != nil {
			t.Fatalf("Retire: %v", err)
		}
		if fresh.TechPoint != 0 {
			t.Errorf("fresh TechPoint = %d, want 0", fresh.TechPoint)
		}
	})
}
