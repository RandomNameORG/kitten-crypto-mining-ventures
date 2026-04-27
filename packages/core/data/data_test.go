package data

import (
	"testing"
)

// TestEmbeddedCatalogsLoad verifies every embedded JSON parses and has the
// minimum required content — a cheap smoke test that fires at `go test`.
func TestEmbeddedCatalogsLoad(t *testing.T) {
	if len(GPUs()) < 5 {
		t.Errorf("expected at least 5 GPUs, got %d", len(GPUs()))
	}
	if len(Rooms()) < 3 {
		t.Errorf("expected at least 3 rooms, got %d", len(Rooms()))
	}
	if len(Events()) < 10 {
		t.Errorf("expected at least 10 events, got %d", len(Events()))
	}
}

func TestGPUIDsUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, g := range GPUs() {
		if seen[g.ID] {
			t.Errorf("duplicate GPU id: %s", g.ID)
		}
		seen[g.ID] = true
		if g.Name == "" {
			t.Errorf("GPU %s has empty name", g.ID)
		}
		if g.Efficiency <= 0 {
			t.Errorf("GPU %s has non-positive efficiency", g.ID)
		}
		if g.Price <= 0 {
			t.Errorf("GPU %s has non-positive price", g.ID)
		}
	}
}

func TestRoomsValid(t *testing.T) {
	hasDefault := false
	for _, r := range Rooms() {
		if r.ID == "" {
			t.Errorf("room has empty id")
		}
		if r.Slots < 1 {
			t.Errorf("room %s has no slots", r.ID)
		}
		if len(r.ThreatPool) == 0 {
			t.Errorf("room %s has empty threat pool", r.ID)
		}
		if r.UnlockedByDefault {
			hasDefault = true
		}
	}
	if !hasDefault {
		t.Error("no room is unlocked by default — new games would have nowhere to start")
	}
}

func TestEventsValid(t *testing.T) {
	validCats := map[string]bool{
		"threat": true, "opportunity": true, "social": true, "crisis": true,
	}
	validKinds := map[string]bool{
		"steal_gpu": true, "pause_mining": true, "rep_change": true,
		"tech_point": true, "gift_gpu": true,
		"earn_multiplier": true, "damage_gpu": true, "burn_room_chance": true,
		"eviction_warning": true, "money_loss": true,
		"tax_audit": true, "damage_oc_gpu": true, "market_pin": true,
	}
	for _, e := range Events() {
		if !validCats[e.Category] {
			t.Errorf("event %s has invalid category %q", e.ID, e.Category)
		}
		if e.Weight <= 0 {
			t.Errorf("event %s has non-positive weight", e.ID)
		}
		for _, eff := range e.Effects {
			if !validKinds[eff.Kind] {
				t.Errorf("event %s uses unknown effect kind %q", e.ID, eff.Kind)
			}
		}
	}
}

func TestSkillsPrereqsResolve(t *testing.T) {
	ids := map[string]bool{}
	for _, s := range Skills() {
		ids[s.ID] = true
	}
	for _, s := range Skills() {
		if s.Prereq != "" && !ids[s.Prereq] {
			t.Errorf("skill %s has unknown prereq %q", s.ID, s.Prereq)
		}
		if s.Cost < 1 {
			t.Errorf("skill %s has non-positive cost", s.ID)
		}
	}
}

func TestMercsValid(t *testing.T) {
	for _, m := range Mercs() {
		if m.HireCost <= 0 || m.WeeklyWage <= 0 {
			t.Errorf("merc %s has non-positive price/wage", m.ID)
		}
		if m.LoyaltyBase < 0 || m.LoyaltyBase > 100 {
			t.Errorf("merc %s loyalty_base out of range: %d", m.ID, m.LoyaltyBase)
		}
	}
}

func TestLookupHelpers(t *testing.T) {
	if _, ok := GPUByID("gtx1060"); !ok {
		t.Error("GPUByID(gtx1060) should resolve")
	}
	if _, ok := GPUByID("nonexistent"); ok {
		t.Error("GPUByID(nonexistent) should fail")
	}
	if _, ok := RoomByID("alley"); !ok {
		t.Error("RoomByID(alley) should resolve")
	}
	if _, ok := EventByID("petty_thief"); !ok {
		t.Error("EventByID(petty_thief) should resolve")
	}
	if _, ok := SkillByID("undervolt_i"); !ok {
		t.Error("SkillByID(undervolt_i) should resolve")
	}
	if _, ok := MercByID("tabby_guard"); !ok {
		t.Error("MercByID(tabby_guard) should resolve")
	}
	if _, ok := PSUByID("psu_builtin"); !ok {
		t.Error("PSUByID(psu_builtin) should resolve")
	}
}

func TestPSUsValid(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range PSUs() {
		if p.ID == "" {
			t.Error("PSU has empty id")
		}
		if seen[p.ID] {
			t.Errorf("duplicate PSU id: %s", p.ID)
		}
		seen[p.ID] = true
		if p.RatedPower <= 0 {
			t.Errorf("PSU %s rated_power must be positive, got %v", p.ID, p.RatedPower)
		}
		if p.Efficiency < 0.5 || p.Efficiency > 1.0 {
			t.Errorf("PSU %s efficiency %v out of [0.5, 1.0]", p.ID, p.Efficiency)
		}
	}
	// Spec'd catalog (§4.3) plus the built-in passthrough used for migration.
	required := []string{
		"psu_builtin",
		"psu_trash",
		"psu_bronze500",
		"psu_silver650",
		"psu_gold850",
		"psu_gold1200",
		"psu_platinum1600",
		"psu_meowcore",
	}
	for _, id := range required {
		if !seen[id] {
			t.Errorf("required PSU %q missing from catalog", id)
		}
	}
}
