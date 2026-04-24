package game

import (
	"fmt"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

// HasAchievement reports whether the given id has been earned.
func (s *State) HasAchievement(id string) bool {
	for _, a := range s.Achievements {
		if a == id {
			return true
		}
	}
	return false
}

// grantAchievement idempotently unlocks an achievement and logs it.
func (s *State) grantAchievement(id string) {
	if s.HasAchievement(id) {
		return
	}
	def, ok := data.AchievementByID(id)
	if !ok {
		return
	}
	s.Achievements = append(s.Achievements, id)
	s.appendLog("opportunity", i18n.T("game.achievement",
		fmt.Sprintf("%s %s — %s", def.Emoji, def.LocalName(), def.LocalDesc())))
}

// CheckAchievements evaluates every achievement and grants any that have
// flipped true. Called once per tick — O(10) checks, cheap.
func (s *State) CheckAchievements() {
	if s.LifetimeEarned > 0 {
		s.grantAchievement("first_drop")
	}
	if s.LifetimeEarned >= 10_000 {
		s.grantAchievement("first_ten_k")
	}
	if s.LifetimeEarned >= 1_000_000 {
		s.grantAchievement("first_million")
	}
	if len(s.Blueprints) > 0 {
		s.grantAchievement("first_blueprint")
	}
	if len(s.Mercs) > 0 {
		s.grantAchievement("merc_employer")
	}
	// "full_stack": any room that's full (non-stolen count == room slots).
	for roomID, _ := range s.Rooms {
		def, ok := data.RoomByID(roomID)
		if !ok {
			continue
		}
		if len(s.GPUsInRoom(roomID)) >= def.Slots {
			s.grantAchievement("full_stack")
			break
		}
	}
	// "all_rooms": every room in the data catalog is unlocked.
	if len(s.Rooms) >= len(data.Rooms()) {
		s.grantAchievement("all_rooms")
	}
	// "hot_cat": any owned room is in the critical heat band.
	for _, rs := range s.Rooms {
		if rs.MaxHeat > 0 && rs.Heat >= 0.95*rs.MaxHeat {
			s.grantAchievement("hot_cat")
			break
		}
	}
	// "oc_mastery": an hour of accumulated overclocked wall-time.
	if s.OCTimeT1Sec+s.OCTimeT2Sec >= 3600 {
		s.grantAchievement("oc_mastery")
	}
	// "overdrive": every installed (non-shipping/non-stolen) GPU is pegged
	// at OCLevel == 2. Requires at least one GPU so an empty rack can't
	// trivially satisfy the universal quantifier.
	if len(s.GPUs) > 0 {
		allMax := true
		counted := 0
		for _, g := range s.GPUs {
			if g.Status == "shipping" || g.Status == "stolen" {
				continue
			}
			counted++
			if g.OCLevel != 2 {
				allMax = false
				break
			}
		}
		if allMax && counted > 0 {
			s.grantAchievement("overdrive")
		}
	}
	// "event_veteran": 50 events total across all categories.
	eventTotal := 0
	for _, n := range s.EventsByCategory {
		eventTotal += n
	}
	if eventTotal >= 50 {
		s.grantAchievement("event_veteran")
	}
	// "marathon": virtual-time endurance milestone.
	if s.TotalTicks >= 100_000 {
		s.grantAchievement("marathon")
	}
	// "crisis_manager": three market crashes on this save.
	if s.MarketCrashCount >= 3 {
		s.grantAchievement("crisis_manager")
	}
}
