package game

import (
	"testing"
	"time"
)

// TestTotalTicksIncrements pins the simplest invariant of the new counter —
// after N seconds of virtual sim, TotalTicks should reflect N. Catches regressions
// where Tick stops counting (early-return tweaks) or starts double-counting.
func TestTotalTicksIncrements(t *testing.T) {
	withTempHome(t)
	s := runSim(t, 1, 600)
	if s.TotalTicks != 600 {
		t.Errorf("TotalTicks = %d, want 600", s.TotalTicks)
	}
}

// TestTotalGPUsBoughtScrappedCounters checks that addGPU and SellGPU each
// move the right counter. The starter GPU created in newStateWithLegacy must
// already be reflected in the bought count so the metric tracks "every GPU
// that ever existed", not just post-init purchases.
func TestTotalGPUsBoughtScrappedCounters(t *testing.T) {
	withTempHome(t)
	s := NewState("Counter")
	if s.TotalGPUsBought != 1 {
		t.Errorf("starter GPU should have been counted; TotalGPUsBought = %d, want 1", s.TotalGPUsBought)
	}
	if s.TotalGPUsScrapped != 0 {
		t.Errorf("nothing scrapped yet; got %d, want 0", s.TotalGPUsScrapped)
	}

	s.BTC = 50_000
	if err := s.BuyGPU("gtx1060"); err != nil {
		t.Fatalf("BuyGPU: %v", err)
	}
	if s.TotalGPUsBought != 2 {
		t.Errorf("after one buy, TotalGPUsBought = %d, want 2", s.TotalGPUsBought)
	}

	// Scrap the running starter — SellGPU works on any owned instance.
	var starter *GPU
	for _, g := range s.GPUs {
		if g.Status == "running" {
			starter = g
			break
		}
	}
	if starter == nil {
		t.Fatal("no running GPU found to scrap")
	}
	if err := s.SellGPU(starter.InstanceID); err != nil {
		t.Fatalf("SellGPU: %v", err)
	}
	if s.TotalGPUsScrapped != 1 {
		t.Errorf("after one scrap, TotalGPUsScrapped = %d, want 1", s.TotalGPUsScrapped)
	}
}

// TestEventsByCategoryCount asserts the per-category histogram lights up when
// applyEvent fires. Synthesises a single opportunity-category event (so we
// don't hit the stochastic MaybeFireEvent path).
func TestEventsByCategoryCount(t *testing.T) {
	withTempHome(t)
	s := NewState("Cats")
	before := s.EventsByCategory["opportunity"]
	s.applyEvent(tpEvent())
	if got := s.EventsByCategory["opportunity"]; got != before+1 {
		t.Errorf("opportunity count = %d, want %d", got, before+1)
	}
}

// TestTotalWagesPaid verifies wages-paid bookkeeping when the player actually
// has the BTC to cover wages. The "missed wages" branch deliberately doesn't
// count, but that's a separate path and not what we're pinning here.
func TestTotalWagesPaid(t *testing.T) {
	withTempHome(t)
	s := NewState("Wages")
	s.BTC = 10_000 // plenty to cover any wage
	if err := s.HireMerc("tabby_guard"); err != nil {
		t.Fatalf("HireMerc: %v", err)
	}
	if len(s.Mercs) == 0 {
		t.Fatal("no mercs hired")
	}
	now := time.Now().Unix()
	s.LastWagesUnix = now - 7200 // 2 game-weeks elapsed → wages will fire
	s.payWages(now)
	if s.TotalWagesPaid <= 0 {
		t.Errorf("TotalWagesPaid should be > 0 after a wage cycle; got %v", s.TotalWagesPaid)
	}
}
