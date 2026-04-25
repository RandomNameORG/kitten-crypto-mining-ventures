package game

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// LegacyStore persists across prestiges. It lives next to the save file at
// ~/.meowmine/legacy.json.
type LegacyStore struct {
	// LP totals.
	TotalEarned float64 `json:"total_earned"`
	TotalLP     int     `json:"total_lp"`
	SpentLP     int     `json:"spent_lp"`

	// Purchased perks.
	StarterCash        float64 `json:"starter_cash"`
	EfficiencyBoost    float64 `json:"efficiency_boost"` // added fraction (0.05 = +5%)
	UnlockedUniversity bool    `json:"unlocked_university"`

	// Carried-over blueprints.
	Blueprints []*Blueprint `json:"blueprints"`

	// CarriedTP banks the slice of unspent TP retained at Retire. Consumed
	// by newStateWithLegacy: read into the fresh state's TechPoint, then
	// zeroed and persisted so a single retire can't re-credit on a later
	// load. Capped at PrestigeTPCarryCap when produced.
	CarriedTP int `json:"carried_tp,omitempty"`
}

func legacyPath() string {
	dir := saveDir()
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "legacy.json")
}

// LoadLegacy returns the legacy store (empty if missing).
func LoadLegacy() *LegacyStore {
	b, err := os.ReadFile(legacyPath())
	if err != nil {
		return &LegacyStore{}
	}
	var l LegacyStore
	if err := json.Unmarshal(b, &l); err != nil {
		return &LegacyStore{}
	}
	return &l
}

// Save persists the legacy store.
func (l *LegacyStore) Save() error {
	b, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(legacyPath(), b, 0o644)
}

// LPAvailable returns how many LP can still be spent.
func (l *LegacyStore) LPAvailable() int { return l.TotalLP - l.SpentLP }
