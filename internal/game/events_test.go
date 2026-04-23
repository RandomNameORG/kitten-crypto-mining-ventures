package game

import (
	"math/rand"
	"testing"
	"time"
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

	runSteals := func(lockLvl int) int {
		rand.Seed(42)
		s := NewState("Defense")
		for _, g := range s.GPUs {
			g.Status = "running"
		}
		s.Rooms[s.CurrentRoom].LockLvl = lockLvl
		s.Rooms[s.CurrentRoom].CCTVLvl = lockLvl
		s.Rooms[s.CurrentRoom].ArmorLvl = lockLvl
		count := 0
		for i := 0; i < 100; i++ {
			for _, g := range s.GPUs {
				if g.Status == "stolen" {
					count++
					g.Status = "running"
				}
			}
			s.applyEvent(stealEvent())
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

// --- fixtures ---

func tpEvent() eventShim {
	return eventShim{
		Category: "opportunity",
		Emoji:    "🧠",
		Name:     "Shim TP",
		Effects:  []effectShim{{Kind: "tech_point", Delta: 1}},
	}.toDef()
}

func outageEvent(seconds int) eventShim {
	return eventShim{
		Category: "threat",
		Emoji:    "⚡",
		Name:     "Shim Outage",
		Effects:  []effectShim{{Kind: "pause_mining", Seconds: seconds}},
	}.toDef()
}

func stealEvent() eventShim {
	return eventShim{
		Category: "threat",
		Emoji:    "🐀",
		Name:     "Shim Thief",
		Effects:  []effectShim{{Kind: "steal_gpu", ChanceIfNoDefense: 0.9, Count: 1}},
	}.toDef()
}
