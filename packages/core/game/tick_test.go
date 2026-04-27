package game

import (
	"math"
	"testing"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
)

// expectedEquilibrium recomputes the §3.2 target temperature for roomID
// off the live State so a test assertion can read what advanceMining sees
// instead of guessing constants.
func expectedEquilibrium(s *State, roomID string) float64 {
	def, _ := data.RoomByID(roomID)
	rs := s.Rooms[roomID]
	var totalHeat float64
	for _, g := range s.GPUs {
		if g.Room != roomID || g.Status != "running" {
			continue
		}
		_, _, hOut, _ := s.GPUStats(g)
		totalHeat += hOut
	}
	totalHeat += s.RoomPSUHeat(roomID)
	cooling := def.BaseCooling * (1.0 + 0.25*float64(rs.CoolingLvl)) * s.MasteryCoolingMult()
	netLoad := totalHeat - cooling
	if netLoad < 0 {
		netLoad = 0
	}
	return def.Ambient + netLoad/def.Dissipation
}

// TestNewtonianHeatSingleStep pins the per-tick math to the exact §3.2
// formula: heat moves by (equilibrium - heat) * approach_speed * dt, not
// the entire gap. Uses alley (Ambient 30, Dissipation 0.15, ApproachSpeed
// 0.03, BaseCooling 0.5) plus two RTX 4090s to push equilibrium well above
// ambient so the gap is unambiguous.
func TestNewtonianHeatSingleStep(t *testing.T) {
	withTempHome(t)
	s := NewState("test")
	s.SetDifficulty("normal")
	now := simTestBaseUnix
	s.LastTickUnix = now
	s.LastBillUnix = now
	s.LastWagesUnix = now
	s.LastMarketTickUnix = now
	s.StartedUnix = now

	// Force load high enough to make equilibrium well above ambient.
	s.addGPU("rtx4090", "alley", false)
	s.addGPU("rtx4090", "alley", false)

	rs := s.Rooms["alley"]
	// Pin a starting temperature distinct from both ambient and the
	// equilibrium target so the assertion catches an over- or under-shoot.
	rs.Heat = 50.0

	def, _ := data.RoomByID("alley")
	startHeat := rs.Heat
	eq := expectedEquilibrium(s, "alley")
	want := startHeat + (eq-startHeat)*def.ApproachSpeed*1.0

	// One second of advance — call advanceMining directly so billing /
	// events can't interfere with the heat assertion.
	s.LastTickUnix = now
	s.advanceMining(now+1, 1.0)

	if math.Abs(rs.Heat-want) > 1e-9 {
		t.Fatalf("after one tick: got %.6f, want %.6f (eq=%.4f, start=%.4f)",
			rs.Heat, want, eq, startHeat)
	}
	// Sanity: we moved a fraction of the gap, not the whole thing.
	if math.Abs(rs.Heat-eq) < math.Abs(eq-startHeat)*0.5 {
		t.Fatalf("single tick closed too much of the gap: heat=%.4f, start=%.4f, eq=%.4f",
			rs.Heat, startHeat, eq)
	}
}

// TestNewtonianHeatLongRunConvergence pumps 200s of constant load through
// advanceMining and asserts the temperature lands within 1°C of the
// analytical equilibrium. With approach_speed 0.03, 200 ticks closes
// 1 - exp(-0.03*200) ≈ 99.75% of any starting gap — comfortably inside 1°C
// for the alley + 2x RTX 4090 setup (equilibrium ≈ 76°C, gap ≈ 26°C from
// the 50°C start, so 0.25% of 26°C ≈ 0.07°C residual).
func TestNewtonianHeatLongRunConvergence(t *testing.T) {
	withTempHome(t)
	s := NewState("test")
	s.SetDifficulty("normal")
	now := simTestBaseUnix
	s.LastTickUnix = now
	s.LastBillUnix = now
	s.LastWagesUnix = now
	s.LastMarketTickUnix = now
	s.StartedUnix = now

	s.addGPU("rtx4090", "alley", false)
	s.addGPU("rtx4090", "alley", false)

	rs := s.Rooms["alley"]
	rs.Heat = 50.0
	// Give the GPUs absurd durability so a hot-room wear cliff during the
	// run can't flip a card to broken and silently change the load.
	for _, g := range s.GPUs {
		g.HoursLeft = 1_000_000
	}

	for i := 1; i <= 200; i++ {
		s.LastTickUnix = now + int64(i-1)
		s.advanceMining(now+int64(i), 1.0)
	}

	eq := expectedEquilibrium(s, "alley")
	// MaxHeat clamp on alley is 80°C; if equilibrium overshoots that,
	// convergence will land at the clamp. Cap target accordingly so the
	// assertion stays valid in either regime.
	if eq > rs.MaxHeat {
		eq = rs.MaxHeat
	}
	if math.Abs(rs.Heat-eq) > 1.0 {
		t.Fatalf("after 200 ticks: heat=%.4f, equilibrium=%.4f (gap=%.4f)",
			rs.Heat, eq, math.Abs(rs.Heat-eq))
	}
}

// TestNewtonianHeatAmbientFloor: a room with no load (cooling >= total
// heat) should sit at exactly ambient — no drift below it. Catches a
// regression where the ambient floor was removed in favor of a hard 20°C
// limit (which would warm arctic from -10 to 20 spuriously).
func TestNewtonianHeatAmbientFloor(t *testing.T) {
	withTempHome(t)
	s := NewState("test")
	s.SetDifficulty("normal")
	now := simTestBaseUnix
	s.LastTickUnix = now
	s.LastBillUnix = now
	s.LastWagesUnix = now
	s.LastMarketTickUnix = now
	s.StartedUnix = now

	// Unlock arctic (Ambient -10°C, Dissipation 0.50). Default-unlocked
	// list does not include it, so seed via the internal helper.
	def, _ := data.RoomByID("arctic")
	s.unlockRoomInternal(def)
	rs := s.Rooms["arctic"]
	if rs == nil {
		t.Fatal("arctic not unlocked")
	}
	// No GPUs in arctic — cooling >> load. Force initial heat above
	// ambient so we can watch it decay back.
	rs.Heat = 5.0

	for i := 1; i <= 600; i++ {
		s.LastTickUnix = now + int64(i-1)
		s.advanceMining(now+int64(i), 1.0)
	}

	// Equilibrium with zero net load is ambient itself.
	if math.Abs(rs.Heat-def.Ambient) > 0.1 {
		t.Fatalf("arctic heat after 600 ticks: %.4f, want %.4f (ambient)",
			rs.Heat, def.Ambient)
	}
}
