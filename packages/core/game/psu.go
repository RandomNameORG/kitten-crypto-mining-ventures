package game

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
)

// psuOverloadChanceCap caps the per-second explosion-roll probability so a
// wildly overloaded rig still has time for the player to react instead of
// vaporising in a single tick. 5%/sec already gives an expected fire window
// of ~20s under sustained heavy overload.
const psuOverloadChanceCap = 0.05

// psuReplacePauseSec is the spec'd downtime when swapping one PSU for
// another in the same room: every GPU there sits idle for this long.
const psuReplacePauseSec int64 = 120

// psuRefundFactor — fraction of original price refunded on RemovePSU.
// Mirrors §4.6 "old PSU sells back at 30% of price."
const psuRefundFactor = 0.30

// psuByInstance returns the PSU instance + its def for a given room+id.
// ok=false if no such instance.
func (s *State) psuByInstance(roomID string, instanceID int) (*PSU, data.PSUDef, bool) {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return nil, data.PSUDef{}, false
	}
	for _, p := range rs.PSUUnits {
		if p.InstanceID == instanceID {
			def, _ := data.PSUByID(p.DefID)
			return p, def, true
		}
	}
	return nil, data.PSUDef{}, false
}

// RoomPSUCapacity sums rated_power over running PSUs in a room.
func (s *State) RoomPSUCapacity(roomID string) float64 {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return 0
	}
	var total float64
	for _, p := range rs.PSUUnits {
		if p.Status != "running" {
			continue
		}
		if def, ok := data.PSUByID(p.DefID); ok {
			total += def.RatedPower
		}
	}
	return total
}

// RoomPSULoad sums effective power_draw over running GPUs in a room. Uses
// GPUStats so OC and upgrade multipliers are honoured.
func (s *State) RoomPSULoad(roomID string) float64 {
	var total float64
	for _, g := range s.GPUs {
		if g.Room != roomID || g.Status != "running" {
			continue
		}
		_, pow, _, _ := s.GPUStats(g)
		total += pow
	}
	return total
}

// RoomPSUEfficiency is the capacity-weighted mean efficiency over running
// PSUs in a room. Returns 1.0 if no running PSUs (defensive — should never
// happen post-migration since unlockRoomInternal seeds psu_builtin).
//
// PSU(next-sprint): RoomPSUEfficiency / RoomPSUHeat ready to multiply in
// once balance retune is scheduled.
func (s *State) RoomPSUEfficiency(roomID string) float64 {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return 1.0
	}
	var weight, weighted float64
	for _, p := range rs.PSUUnits {
		if p.Status != "running" {
			continue
		}
		def, ok := data.PSUByID(p.DefID)
		if !ok {
			continue
		}
		weight += def.RatedPower
		weighted += def.RatedPower * def.Efficiency
	}
	if weight <= 0 {
		return 1.0
	}
	return weighted / weight
}

// RoomPSUHeat sums heat_output over running PSUs in a room.
//
// PSU(next-sprint): RoomPSUEfficiency / RoomPSUHeat ready to multiply in
// once balance retune is scheduled.
func (s *State) RoomPSUHeat(roomID string) float64 {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return 0
	}
	var total float64
	for _, p := range rs.PSUUnits {
		if p.Status != "running" {
			continue
		}
		if def, ok := data.PSUByID(p.DefID); ok {
			total += def.HeatOutput
		}
	}
	return total
}

// RoomPSUOverloadFactor returns load / capacity. ≥1.0 means the room is
// at-or-over rated; how far past 1.0 + the weakest PSU's tolerance feeds
// the explosion-roll probability.
func (s *State) RoomPSUOverloadFactor(roomID string) float64 {
	cap := s.RoomPSUCapacity(roomID)
	if cap <= 0 {
		return 0
	}
	return s.RoomPSULoad(roomID) / cap
}

// roomMinOverloadTolerance returns the lowest overload_tolerance among
// running PSUs in the room — the weakest one blows first. Returns the
// builtin's tolerance (1.0, effectively never) when no other PSU is
// running.
func (s *State) roomMinOverloadTolerance(roomID string) float64 {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return 1.0
	}
	min := -1.0
	for _, p := range rs.PSUUnits {
		if p.Status != "running" {
			continue
		}
		def, ok := data.PSUByID(p.DefID)
		if !ok {
			continue
		}
		if min < 0 || def.OverloadTolerance < min {
			min = def.OverloadTolerance
		}
	}
	if min < 0 {
		return 1.0
	}
	return min
}

// RoomCanFitGPU answers whether buying gpuDefID in roomID would push load
// above capacity. Used by BuyGPU to block oversubscription before any BTC
// is deducted.
func (s *State) RoomCanFitGPU(roomID, gpuDefID string) bool {
	def, ok := data.GPUByID(gpuDefID)
	if !ok {
		return false
	}
	cap := s.RoomPSUCapacity(roomID)
	if cap <= 0 {
		return false
	}
	// Prospective draw uses the catalog base power_draw (the GPU is brand-new,
	// no upgrades or OC yet) but flows through the same PowerDrawMult / etc.
	// the running-stat path uses, so skill modifiers count.
	prospective := def.PowerDraw * s.PowerDrawMult()
	return s.RoomPSULoad(roomID)+prospective <= cap
}

// IsRoomPSUPaused returns true if the room is mid-PSU-replacement and GPUs
// should not earn or generate heat.
func (s *State) IsRoomPSUPaused(roomID string, now int64) bool {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return false
	}
	return rs.PSUResumeAt > now
}

// InstallPSU pays for a new PSU and appends it to the target room.
func (s *State) InstallPSU(roomID, psuDefID string) error {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return fmt.Errorf("room not unlocked: %s", roomID)
	}
	def, ok := data.PSUByID(psuDefID)
	if !ok {
		return fmt.Errorf("no such PSU: %s", psuDefID)
	}
	if def.Price > 0 && s.BTC < float64(def.Price) {
		return fmt.Errorf("need %s, have %s", FmtBTCInt(def.Price), FmtBTC(s.BTC))
	}
	s.BTC -= float64(def.Price)
	if s.NextPSUID < 1 {
		s.NextPSUID = 1
	}
	rs.PSUUnits = append(rs.PSUUnits, &PSU{
		InstanceID:  s.NextPSUID,
		DefID:       def.ID,
		Status:      "running",
		InstalledAt: time.Now().Unix(),
	})
	s.NextPSUID++
	return nil
}

// ReplacePSU swaps one PSU for another. The old unit is removed, the new
// one is paid for and installed, and every GPU in the room pauses for
// psuReplacePauseSec.
func (s *State) ReplacePSU(roomID string, instanceID int, newPSUDefID string) error {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return fmt.Errorf("room not unlocked: %s", roomID)
	}
	idx := -1
	for i, p := range rs.PSUUnits {
		if p.InstanceID == instanceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("no such PSU instance: %d", instanceID)
	}
	newDef, ok := data.PSUByID(newPSUDefID)
	if !ok {
		return fmt.Errorf("no such PSU: %s", newPSUDefID)
	}
	if newDef.Price > 0 && s.BTC < float64(newDef.Price) {
		return fmt.Errorf("need %s, have %s", FmtBTCInt(newDef.Price), FmtBTC(s.BTC))
	}
	s.BTC -= float64(newDef.Price)
	if s.NextPSUID < 1 {
		s.NextPSUID = 1
	}
	// PSUResumeAt anchors on s.LastTickUnix (the sim clock) so headless
	// sim runs and offline catch-up advance through the 120s pause cleanly.
	// Wall-clock InstalledAt is fine — it's only telemetry.
	pauseAnchor := s.LastTickUnix
	if pauseAnchor == 0 {
		pauseAnchor = time.Now().Unix()
	}
	replacement := &PSU{
		InstanceID:  s.NextPSUID,
		DefID:       newDef.ID,
		Status:      "running",
		InstalledAt: time.Now().Unix(),
	}
	s.NextPSUID++
	// Remove old unit, append the replacement.
	rs.PSUUnits = append(rs.PSUUnits[:idx], rs.PSUUnits[idx+1:]...)
	rs.PSUUnits = append(rs.PSUUnits, replacement)
	rs.PSUResumeAt = pauseAnchor + psuReplacePauseSec
	return nil
}

// RemovePSU deletes a PSU and refunds 30% of its original price. Fails if
// removal would drop room capacity below current GPU load.
func (s *State) RemovePSU(roomID string, instanceID int) (int, error) {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return 0, fmt.Errorf("room not unlocked: %s", roomID)
	}
	idx := -1
	for i, p := range rs.PSUUnits {
		if p.InstanceID == instanceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return 0, fmt.Errorf("no such PSU instance: %d", instanceID)
	}
	target := rs.PSUUnits[idx]
	def, _ := data.PSUByID(target.DefID)
	// Capacity guard: would removing this unit drop us below current load?
	postCap := s.RoomPSUCapacity(roomID)
	if target.Status == "running" {
		postCap -= def.RatedPower
	}
	if s.RoomPSULoad(roomID) > postCap {
		return 0, fmt.Errorf("can't remove: GPUs would exceed remaining PSU capacity")
	}
	refund := int(float64(def.Price) * psuRefundFactor)
	s.BTC += float64(refund)
	rs.PSUUnits = append(rs.PSUUnits[:idx], rs.PSUUnits[idx+1:]...)
	return refund, nil
}

// advancePSUOverload rolls a per-second explosion check on every overloaded
// running PSU. Called from Tick. The weakest tolerance (lowest among
// installed running PSUs) gates the roll — that's the unit that blows
// first when load creeps past spec.
func (s *State) advancePSUOverload(now int64, dt float64) {
	if dt <= 0 {
		return
	}
	for roomID, rs := range s.Rooms {
		factor := s.RoomPSUOverloadFactor(roomID)
		tol := s.roomMinOverloadTolerance(roomID)
		// Builtin's tolerance is 1.0 → 1+tol = 2.0, which any realistic
		// load can't exceed. So legacy / fresh-game rooms never roll.
		if factor <= 1.0+tol {
			continue
		}
		// Probability per second scaled by how far past the safe band.
		over := factor - 1.0 - tol
		if tol <= 0 {
			tol = 0.05 // defensive — never divide by 0 even if catalog drifts
		}
		chancePerSec := 0.001 * over / tol
		if chancePerSec > psuOverloadChanceCap {
			chancePerSec = psuOverloadChanceCap
		}
		// One roll per second of dt — keeps offline catch-up consistent
		// with real-time play instead of one big roll for the whole gap.
		// 1 - (1-p)^dt is the equivalent compound probability.
		compound := 1.0 - math.Pow(1.0-chancePerSec, dt)
		if rand.Float64() >= compound {
			continue
		}
		// Pick the weakest running PSU in this room — the one whose
		// tolerance equalled the gating min.
		victim := s.weakestRunningPSU(rs)
		if victim == nil {
			continue
		}
		s.psuExplode(roomID, victim.InstanceID)
	}
}

// weakestRunningPSU returns the running PSU in rs with the lowest overload
// tolerance — the canonical blow-first candidate.
func (s *State) weakestRunningPSU(rs *RoomState) *PSU {
	var pick *PSU
	pickTol := -1.0
	for _, p := range rs.PSUUnits {
		if p.Status != "running" {
			continue
		}
		def, ok := data.PSUByID(p.DefID)
		if !ok {
			continue
		}
		if pickTol < 0 || def.OverloadTolerance < pickTol {
			pickTol = def.OverloadTolerance
			pick = p
		}
	}
	return pick
}

// psuExplode marks one PSU broken and brings down explosion_damage random
// running GPUs in the room.
func (s *State) psuExplode(roomID string, instanceID int) {
	rs, ok := s.Rooms[roomID]
	if !ok {
		return
	}
	var psu *PSU
	for _, p := range rs.PSUUnits {
		if p.InstanceID == instanceID {
			psu = p
			break
		}
	}
	if psu == nil || psu.Status != "running" {
		return
	}
	def, _ := data.PSUByID(psu.DefID)
	psu.Status = "broken"
	// Pull eligible GPU victims from the same room.
	candidates := []*GPU{}
	for _, g := range s.GPUs {
		if g.Room == roomID && g.Status == "running" {
			candidates = append(candidates, g)
		}
	}
	dmg := def.ExplosionDamage
	if dmg > len(candidates) {
		dmg = len(candidates)
	}
	for i := 0; i < dmg; i++ {
		idx := rand.Intn(len(candidates))
		victim := candidates[idx]
		victim.Status = "broken"
		victim.HoursLeft = 0
		candidates = append(candidates[:idx], candidates[idx+1:]...)
	}
	roomName := roomID
	if rdef, ok := data.RoomByID(roomID); ok {
		roomName = rdef.LocalName()
	}
	s.appendLog("crisis", fmt.Sprintf("PSU exploded in %s — %s bricked %d GPU(s)", roomName, def.LocalName(), dmg))
}
