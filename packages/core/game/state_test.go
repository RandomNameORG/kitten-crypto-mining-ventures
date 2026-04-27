package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// withTempHome reroutes HOME to a t.TempDir() so save/legacy writes don't
// touch the developer's real files.
func withTempHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	// Windows uses USERPROFILE.
	t.Setenv("USERPROFILE", dir)
}

func TestNewStateInitialized(t *testing.T) {
	withTempHome(t)
	s := NewState("Test")
	if s.KittenName != "Test" {
		t.Errorf("kitten name not applied")
	}
	if s.BTC <= 0 {
		t.Error("new game should start with positive cash")
	}
	if len(s.GPUs) == 0 {
		t.Error("new game should include a starter GPU")
	}
	if s.Rooms["alley"] == nil {
		t.Error("new game should unlock alley by default")
	}
	if s.UnlockedSkills == nil {
		t.Error("UnlockedSkills must be initialized")
	}
}

func TestStateSaveLoadRoundtrip(t *testing.T) {
	withTempHome(t)
	s := NewState("Roundtrip")
	s.BTC = 12345
	s.ResearchFrags = 7
	s.UnlockedSkills["undervolt_i"] = true
	if err := s.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.BTC != 12345 || loaded.ResearchFrags != 7 {
		t.Errorf("numeric state did not roundtrip: %+v", loaded)
	}
	if !loaded.UnlockedSkills["undervolt_i"] {
		t.Error("unlocked skill was not preserved")
	}
}

func TestLoadFromBackfillsOldSaves(t *testing.T) {
	// Simulate an old save that has no UnlockedSkills / Mercs / Blueprints.
	legacy := `{
      "version": 1,
      "kitten_name": "OldCat",
      "money": 500,
      "current_room": "alley",
      "rooms": {"alley": {"def_id":"alley","heat":20,"max_heat":90}},
      "gpus": [],
      "next_gpu_id": 1,
      "started_unix": 1000
    }`
	s, err := LoadFrom([]byte(legacy))
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if s.UnlockedSkills == nil || s.Mercs == nil || s.Blueprints == nil {
		t.Error("ensureInit should have backfilled new maps/slices")
	}
	if s.NextMercID < 1 || s.NextBlueprintN < 1 {
		t.Error("ensureInit should have set next-id defaults to 1")
	}
}

func TestBuyAndSellGPUFlow(t *testing.T) {
	withTempHome(t)
	s := NewState("Flow")
	s.BTC = 10000
	if err := s.BuyGPU("gtx1060"); err != nil {
		t.Fatalf("buy: %v", err)
	}
	// Shipping: should add 1 GPU in shipping state.
	var shipped *GPU
	for _, g := range s.GPUs {
		if g.DefID == "gtx1060" && g.Status == "shipping" {
			shipped = g
		}
	}
	if shipped == nil {
		t.Fatal("new GPU should be in shipping state after BuyGPU")
	}
	// Window invariant: BuyGPU stamps both ShipsAt and ShipTotalSec from the
	// same 30..180s draw, so the UI can render a correct progress bar without
	// guessing the server-side window. Drift here = visible regression.
	if shipped.ShipTotalSec < 30 || shipped.ShipTotalSec > 180 {
		t.Errorf("ShipTotalSec out of [30,180]: %d", shipped.ShipTotalSec)
	}
	if shipped.ShipsAt == 0 {
		t.Error("ShipsAt should be set on shipping GPU")
	}
	// Sell the starter GPU (which is running).
	var starter *GPU
	for _, g := range s.GPUs {
		if g.Status == "running" {
			starter = g
			break
		}
	}
	if starter == nil {
		t.Fatal("expected a running starter GPU")
	}
	before := s.BTC
	if err := s.SellGPU(starter.InstanceID); err != nil {
		t.Fatalf("sell: %v", err)
	}
	if s.BTC <= before {
		t.Error("selling should add money")
	}
	if s.ResearchFrags <= 0 {
		t.Error("scrapping should yield research fragments")
	}
}

func TestBuyGPUInsufficientFunds(t *testing.T) {
	withTempHome(t)
	s := NewState("Broke")
	s.BTC = 1
	if err := s.BuyGPU("rtx4090"); err == nil {
		t.Error("buying a $18k card with $1 should fail")
	}
}

func TestUnlockRoomChargesMoney(t *testing.T) {
	withTempHome(t)
	s := NewState("Unlocker")
	s.BTC = 5000
	before := s.BTC
	if err := s.UnlockRoom("warehouse"); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	if s.Rooms["warehouse"] == nil {
		t.Error("warehouse should be unlocked")
	}
	if s.BTC >= before {
		t.Error("unlocking should deduct money")
	}
}

func TestSavePathInsideHome(t *testing.T) {
	withTempHome(t)
	home, _ := os.UserHomeDir()
	path := SavePath()
	if rel, err := filepath.Rel(home, path); err != nil || rel == "" {
		t.Errorf("save path should be inside HOME, got %s (home=%s)", path, home)
	}
}

func TestLoadFromRejectsGarbage(t *testing.T) {
	if _, err := LoadFrom([]byte("not json")); err == nil {
		t.Error("expected error on garbage input")
	}
}

func TestLegacyStoreRoundtrip(t *testing.T) {
	withTempHome(t)
	l := LoadLegacy()
	l.TotalLP = 42
	l.StarterCash = 500
	if err := l.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	l2 := LoadLegacy()
	if l2.TotalLP != 42 || l2.StarterCash != 500 {
		t.Errorf("legacy did not roundtrip: %+v", l2)
	}
}

// Ensure state marshals to stable JSON (not required by game, but catches
// field tag regressions cheaply).
func TestStateJSONTagsPresent(t *testing.T) {
	withTempHome(t)
	s := NewState("JSON")
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, wanted := range []string{"kitten_name", "btc", "current_room", "unlocked_skills", "mercs", "blueprints"} {
		if !containsField(b, wanted) {
			t.Errorf("marshaled state missing %q: %s", wanted, b)
		}
	}
}

func containsField(b []byte, field string) bool {
	needle := []byte("\"" + field + "\"")
	for i := 0; i+len(needle) <= len(b); i++ {
		match := true
		for j := range needle {
			if b[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
