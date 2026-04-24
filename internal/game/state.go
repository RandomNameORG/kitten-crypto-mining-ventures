package game

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/i18n"
)

// GPU is a runtime instance of a graphics card owned by the player.
type GPU struct {
	InstanceID   int     `json:"instance_id"`
	DefID        string  `json:"def_id"`
	Status       string  `json:"status"` // running, broken, shipping, stolen, offline
	UpgradeLevel int     `json:"upgrade_level"`
	HoursLeft    float64 `json:"hours_left"`
	ShipsAt      int64   `json:"ships_at,omitempty"` // unix time when shipping completes
	Room         string  `json:"room"`
	// BlueprintID is set for MEOWCore instances; maps to a Blueprint for stats.
	BlueprintID string `json:"blueprint_id,omitempty"`
}

// RoomState is the runtime instance of a Room owned by the player.
type RoomState struct {
	DefID      string  `json:"def_id"`
	Heat       float64 `json:"heat"`
	MaxHeat    float64 `json:"max_heat"`
	LockLvl    int     `json:"lock_lvl"`     // 0-5, base defense vs theft
	CCTVLvl    int     `json:"cctv_lvl"`     // 0-5, catches thieves + deters merc betrayal
	WiringLvl  int     `json:"wiring_lvl"`   // 0-5, reduces outage + fire chance
	CoolingLvl int     `json:"cooling_lvl"`  // 0-5, extra cooling
	ArmorLvl   int     `json:"armor_lvl"`    // 0-5, defense vs tunnels / pirates
}

// Merc is a runtime mercenary instance.
type Merc struct {
	InstanceID int   `json:"instance_id"`
	DefID      string `json:"def_id"`
	Loyalty    int    `json:"loyalty"` // 0-100
	HiredAt    int64  `json:"hired_at"`
	RoomID     string `json:"room_id"` // which room they guard
}

// Blueprint is a persistent recipe for a custom MEOWCore GPU.
type Blueprint struct {
	ID         string   `json:"id"`
	Tier       int      `json:"tier"`   // 1..3
	Boosts     []string `json:"boosts"` // subset of: efficiency, undervolt, durability, micro, stealth
	CreatedAt  int64    `json:"created_at"`
}

// Research is the player's current active research project.
type Research struct {
	BlueprintTier int      `json:"blueprint_tier"`
	Boosts        []string `json:"boosts"`
	StartedAt     int64    `json:"started_at"`
	DurationSec   int      `json:"duration_sec"`
}

// Modifier is a time-limited multiplier or flag.
type Modifier struct {
	Kind      string  `json:"kind"`   // "btc_mult" | "earn_mult" | "pause_mining"
	Factor    float64 `json:"factor"` // multiplier value; or unused for flags
	ExpiresAt int64   `json:"expires_at"`
}

// LogEntry is a line in the event log.
type LogEntry struct {
	Time     int64  `json:"time"`
	Category string `json:"category"` // info | threat | opportunity | social | crisis
	Text     string `json:"text"`
}

// EventCooldowns tracks when each event was last fired (unix seconds).
type EventCooldowns map[string]int64

// State is the entire save-able game state.
type State struct {
	Version       int                   `json:"version"`
	KittenName    string                `json:"kitten_name"`
	Money         float64               `json:"money"`
	BTC           float64               `json:"btc"`
	BTCPriceSeed  int64                 `json:"btc_price_seed"`
	TechPoint     int                   `json:"tech_point"`
	Reputation    int                   `json:"reputation"`
	Karma         int                   `json:"karma"`
	CurrentRoom   string                `json:"current_room"`
	Rooms         map[string]*RoomState `json:"rooms"`
	GPUs          []*GPU                `json:"gpus"`
	NextGPUID     int                   `json:"next_gpu_id"`
	Modifiers     []Modifier            `json:"modifiers"`
	EventCooldown EventCooldowns        `json:"event_cooldown"`
	LastTickUnix  int64                 `json:"last_tick_unix"`
	LastBillUnix  int64                 `json:"last_bill_unix"`
	LastWagesUnix int64                 `json:"last_wages_unix"`
	Log           []LogEntry            `json:"log"`
	Paused        bool                  `json:"paused"`
	StartedUnix   int64                 `json:"started_unix"`

	// Progression systems.
	UnlockedSkills map[string]bool `json:"unlocked_skills"`
	Mercs          []*Merc         `json:"mercs"`
	NextMercID     int             `json:"next_merc_id"`
	ResearchFrags  int             `json:"research_frags"`
	ActiveResearch *Research       `json:"active_research,omitempty"`
	Blueprints     []*Blueprint    `json:"blueprints"`
	NextBlueprintN int             `json:"next_blueprint_n"`

	// Lifetime + prestige.
	LifetimeEarned float64 `json:"lifetime_earned"`
	// LegacyPoints spent / available this run. True cross-run LP lives in legacy.json.
	LegacyAvailable int `json:"legacy_available"`

	// Lang persists the player's chosen language code ("en" | "zh"). Loaded
	// by LoadFrom into the i18n package at startup.
	Lang string `json:"lang,omitempty"`

	// Difficulty is locked at game start (splash picker) and never changes
	// for the run. Empty string means the splash hasn't been completed yet
	// — the UI will prompt. Loaded saves that pre-date this field are
	// migrated to "normal" by ensureInit.
	Difficulty string `json:"difficulty,omitempty"`
}

// NewState returns a fresh game. An empty kittenName signals that the UI
// should prompt — it's stored as "" on the returned state and the UI's
// name-entry view takes over until the player commits.
func NewState(kittenName string) *State {
	return newStateWithLegacy(kittenName, LoadLegacy())
}

// newStateWithLegacy is the internal constructor that also applies cross-run
// legacy bonuses at new-game time.
func newStateWithLegacy(kittenName string, legacy *LegacyStore) *State {
	now := time.Now().Unix()
	s := &State{
		Version:        1,
		KittenName:     kittenName,
		Money:          150,
		BTC:            0,
		BTCPriceSeed:   rand.Int63(),
		TechPoint:      0,
		Reputation:     0,
		Karma:          0,
		CurrentRoom:    "alley",
		Rooms:          map[string]*RoomState{},
		GPUs:           []*GPU{},
		NextGPUID:      1,
		Modifiers:      []Modifier{},
		EventCooldown:  EventCooldowns{},
		LastTickUnix:   now,
		LastBillUnix:   now,
		LastWagesUnix:  now,
		StartedUnix:    now,
		Log:            []LogEntry{},
		UnlockedSkills: map[string]bool{},
		Mercs:          []*Merc{},
		NextMercID:     1,
		ResearchFrags:  0,
		Blueprints:     []*Blueprint{},
		NextBlueprintN: 1,
		LegacyAvailable: 0,
		Lang:            i18n.Lang(),
	}
	// Unlock every room flagged as default.
	for _, r := range data.Rooms() {
		if r.UnlockedByDefault {
			s.unlockRoomInternal(r)
		}
	}
	// Starter GPU — already on the desk, no shipping wait.
	s.addGPU("gtx1060", "alley", false)
	welcomeName := kittenName
	if welcomeName == "" {
		welcomeName = "friend"
	}
	s.appendLog("info", i18n.T("game.welcome", welcomeName))

	// Apply legacy bonuses at start.
	if legacy != nil {
		if legacy.StarterCash > 0 {
			s.Money += legacy.StarterCash
			s.appendLog("opportunity", fmt.Sprintf("Legacy bonus: +$%.0f starter cash.", legacy.StarterCash))
		}
		if legacy.UnlockedUniversity {
			if def, ok := data.RoomByID("university"); ok {
				s.unlockRoomInternal(def)
				s.appendLog("opportunity", "Legacy bonus: University Server Room pre-unlocked.")
			}
		}
		// Carry over researched blueprints (deep-copied so run-state mutations
		// don't leak back into the legacy bank).
		for _, bp := range legacy.Blueprints {
			dup := *bp
			s.Blueprints = append(s.Blueprints, &dup)
		}
		if len(legacy.Blueprints) > 0 {
			s.appendLog("opportunity", fmt.Sprintf("Legacy bonus: %d blueprints carried over.", len(legacy.Blueprints)))
		}
	}
	return s
}

func (s *State) unlockRoomInternal(r data.RoomDef) {
	if _, ok := s.Rooms[r.ID]; ok {
		return
	}
	s.Rooms[r.ID] = &RoomState{
		DefID:   r.ID,
		Heat:    20,
		MaxHeat: 90,
	}
}

// UnlockRoom unlocks a room if the player can afford it.
func (s *State) UnlockRoom(id string) error {
	if _, ok := s.Rooms[id]; ok {
		return fmt.Errorf("already unlocked")
	}
	def, ok := data.RoomByID(id)
	if !ok {
		return fmt.Errorf("no such room: %s", id)
	}
	if s.Money < float64(def.UnlockCost) {
		return fmt.Errorf("need $%d, have $%.0f", def.UnlockCost, s.Money)
	}
	s.Money -= float64(def.UnlockCost)
	s.unlockRoomInternal(def)
	s.appendLog("info", fmt.Sprintf("Moved into %s.", def.Name))
	return nil
}

func (s *State) SwitchRoom(id string) error {
	if _, ok := s.Rooms[id]; !ok {
		return fmt.Errorf("not unlocked")
	}
	s.CurrentRoom = id
	return nil
}

// addGPU creates a new GPU instance.
func (s *State) addGPU(defID, room string, shipping bool) *GPU {
	def, ok := data.GPUByID(defID)
	if !ok {
		return nil
	}
	g := &GPU{
		InstanceID: s.NextGPUID,
		DefID:      defID,
		Status:     "running",
		HoursLeft:  float64(def.DurabilityHours),
		Room:       room,
	}
	s.NextGPUID++
	if shipping {
		g.Status = "shipping"
		g.ShipsAt = time.Now().Unix() + int64(30+rand.Intn(150))
	}
	s.GPUs = append(s.GPUs, g)
	return g
}

// removeGPU deletes a GPU instance from the list. Used for theft so stolen
// cards don't clutter the dashboard or leak slot accounting.
func (s *State) removeGPU(instanceID int) bool {
	for i, g := range s.GPUs {
		if g.InstanceID == instanceID {
			s.GPUs = append(s.GPUs[:i], s.GPUs[i+1:]...)
			return true
		}
	}
	return false
}

// addMEOWCore creates a GPU instance from a player-researched Blueprint.
func (s *State) addMEOWCore(bp *Blueprint, room string) *GPU {
	g := &GPU{
		InstanceID:  s.NextGPUID,
		DefID:       fmt.Sprintf("meowcore_v%d", bp.Tier),
		Status:      "running",
		HoursLeft:   120, // self-made GPUs are durable
		Room:        room,
		BlueprintID: bp.ID,
	}
	s.NextGPUID++
	s.GPUs = append(s.GPUs, g)
	return g
}

// BuyGPU purchases a GPU by def id, routes it to current room via shipping.
func (s *State) BuyGPU(defID string) error {
	def, ok := data.GPUByID(defID)
	if !ok {
		return fmt.Errorf("no such GPU: %s", defID)
	}
	if s.Money < float64(def.Price) {
		return fmt.Errorf("need $%d, have $%.0f", def.Price, s.Money)
	}
	if !s.RoomHasFreeSlot(s.CurrentRoom) {
		return fmt.Errorf("no free slots in this room")
	}
	s.Money -= float64(def.Price)
	s.addGPU(defID, s.CurrentRoom, true)
	s.appendLog("info", fmt.Sprintf("Ordered %s for $%d. Tracking inbound...", def.Name, def.Price))
	return nil
}

// SellGPU scraps a GPU for its scrap value (boosted by Tax Optimization skill).
func (s *State) SellGPU(instanceID int) error {
	for i, g := range s.GPUs {
		if g.InstanceID == instanceID {
			base := 0
			name := "Unknown"
			if g.BlueprintID != "" {
				// MEOWCore scrap value: mid-tier.
				base = 2000 + (s.blueprintTier(g.BlueprintID)-1)*2000
				name = fmt.Sprintf("MEOWCore v%d", s.blueprintTier(g.BlueprintID))
			} else if def, ok := data.GPUByID(g.DefID); ok {
				base = def.ScrapValue
				name = def.Name
			}
			value := float64(base) * s.ScrapValueMult()
			// Also grant 1-3 research fragments.
			frags := 1 + rand.Intn(3)
			s.Money += value
			s.ResearchFrags += frags
			s.GPUs = append(s.GPUs[:i], s.GPUs[i+1:]...)
			s.appendLog("info", fmt.Sprintf("Scrapped %s for $%.0f + %d research fragments.", name, value, frags))
			return nil
		}
	}
	return fmt.Errorf("no such GPU instance")
}

// RoomHasFreeSlot checks if the given room has capacity for another GPU.
func (s *State) RoomHasFreeSlot(roomID string) bool {
	def, ok := data.RoomByID(roomID)
	if !ok {
		return false
	}
	count := 0
	for _, g := range s.GPUs {
		if g.Room == roomID && g.Status != "stolen" {
			count++
		}
	}
	return count < def.Slots
}

// GPUsInRoom returns GPUs currently placed in the given room.
func (s *State) GPUsInRoom(roomID string) []*GPU {
	out := []*GPU{}
	for _, g := range s.GPUs {
		if g.Room == roomID {
			out = append(out, g)
		}
	}
	return out
}

// MercsInRoom returns mercs currently guarding the given room.
func (s *State) MercsInRoom(roomID string) []*Merc {
	out := []*Merc{}
	for _, m := range s.Mercs {
		if m.RoomID == roomID {
			out = append(out, m)
		}
	}
	return out
}

// BlueprintByID looks up a researched blueprint.
func (s *State) BlueprintByID(id string) *Blueprint {
	for _, bp := range s.Blueprints {
		if bp.ID == id {
			return bp
		}
	}
	return nil
}

func (s *State) blueprintTier(id string) int {
	if bp := s.BlueprintByID(id); bp != nil {
		return bp.Tier
	}
	return 1
}

func (s *State) appendLog(category, text string) {
	s.Log = append(s.Log, LogEntry{
		Time:     time.Now().Unix(),
		Category: category,
		Text:     text,
	})
	// Bound log to the last 200 entries.
	if len(s.Log) > 200 {
		s.Log = s.Log[len(s.Log)-200:]
	}
}

// AppendLog is the external hook for other systems to write log entries.
func (s *State) AppendLog(category, text string) {
	s.appendLog(category, text)
}

// --- Save / Load ---

func saveDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".meowmine"
	}
	return filepath.Join(home, ".meowmine")
}

func SavePath() string {
	dir := saveDir()
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "save.json")
}

// Save writes state to the default save path.
func (s *State) Save() error {
	return s.SaveAs(SavePath())
}

// SaveAs writes state to an arbitrary path. SSH mode uses this to keep
// per-session saves keyed by pubkey.
func (s *State) SaveAs(path string) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(path); dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	return os.WriteFile(path, b, 0o644)
}

// Load reads state from the default save path. Returns nil if no save exists.
func Load() (*State, error) {
	path := SavePath()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return LoadFrom(b)
}

// LoadFrom parses raw save bytes and backfills any missing fields so saves
// written by earlier versions load cleanly into the current schema.
func LoadFrom(b []byte) (*State, error) {
	var s State
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	s.ensureInit()
	// Apply the player's persisted language choice to the i18n singleton.
	if s.Lang != "" {
		i18n.SetLang(s.Lang)
	}
	return &s, nil
}

// CycleLang advances the active language and updates the state's persisted
// field. Returns the new active language code.
func (s *State) CycleLang() string {
	next := i18n.CycleLang()
	s.Lang = next
	s.appendLog("info", i18n.T("game.lang_switched", i18n.Label(next)))
	return next
}

// ensureInit normalises a State so every map/slice is non-nil. Called after
// Unmarshal so older saves work without panics.
func (s *State) ensureInit() {
	if s.Rooms == nil {
		s.Rooms = map[string]*RoomState{}
	}
	if s.GPUs == nil {
		s.GPUs = []*GPU{}
	}
	if s.Modifiers == nil {
		s.Modifiers = []Modifier{}
	}
	if s.EventCooldown == nil {
		s.EventCooldown = EventCooldowns{}
	}
	if s.UnlockedSkills == nil {
		s.UnlockedSkills = map[string]bool{}
	}
	if s.Mercs == nil {
		s.Mercs = []*Merc{}
	}
	if s.Blueprints == nil {
		s.Blueprints = []*Blueprint{}
	}
	if s.Log == nil {
		s.Log = []LogEntry{}
	}
	if s.NextGPUID < 1 {
		s.NextGPUID = 1
	}
	if s.NextMercID < 1 {
		s.NextMercID = 1
	}
	if s.NextBlueprintN < 1 {
		s.NextBlueprintN = 1
	}
	// Ensure every room-state object references a known room. Unknown ids
	// (from removed biomes) silently drop so the game keeps loading.
	for id := range s.Rooms {
		if _, ok := data.RoomByID(id); !ok {
			delete(s.Rooms, id)
		}
	}
	// Migration: drop any lingering `stolen` GPUs from older saves where
	// theft marked-but-didn't-remove. Stolen cards leak into the dashboard
	// slot counter and the GPUs list otherwise.
	alive := s.GPUs[:0]
	for _, g := range s.GPUs {
		if g.Status == "stolen" {
			continue
		}
		alive = append(alive, g)
	}
	s.GPUs = alive
	// Migration: saves from before difficulty existed get "normal" so they
	// don't bounce off the splash picker. Genuinely new saves have both
	// KittenName and Difficulty empty and the UI handles both.
	if s.Difficulty == "" && s.KittenName != "" {
		s.Difficulty = "normal"
	}
}

// Diff returns the active difficulty definition.
func (s *State) Diff() data.DifficultyDef {
	if s.Difficulty == "" {
		return data.DifficultyByID(data.DefaultDifficulty)
	}
	return data.DifficultyByID(s.Difficulty)
}

// DifficultyEarnMult is the earn-rate multiplier for the active difficulty.
func (s *State) DifficultyEarnMult() float64 { return s.Diff().EarnMult }

// DifficultyBillMult is the electricity+rent multiplier for the active difficulty.
func (s *State) DifficultyBillMult() float64 { return s.Diff().BillMult }

// DifficultyThreatMult is the event-fire-probability multiplier.
func (s *State) DifficultyThreatMult() float64 { return s.Diff().ThreatMult }

// SetDifficulty writes the chosen difficulty to state, applies starter cash,
// and logs the choice. Called once from the splash picker.
func (s *State) SetDifficulty(id string) {
	def := data.DifficultyByID(id)
	s.Difficulty = def.ID
	s.Money = def.StarterCash
	s.appendLog("info", i18n.T("game.difficulty_set", def.LocalLabel()))
}
