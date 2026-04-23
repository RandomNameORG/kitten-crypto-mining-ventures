package game

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/data"
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
}

// RoomState is the runtime instance of a Room owned by the player.
type RoomState struct {
	DefID    string  `json:"def_id"`
	Heat     float64 `json:"heat"`
	MaxHeat  float64 `json:"max_heat"`
	LockLvl  int     `json:"lock_lvl"`
	CCTVLvl  int     `json:"cctv_lvl"`
	WiringLv int     `json:"wiring_lvl"`
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
	Version       int                 `json:"version"`
	KittenName    string              `json:"kitten_name"`
	Money         float64             `json:"money"`
	BTC           float64             `json:"btc"`
	BTCPriceSeed  int64               `json:"btc_price_seed"`
	TechPoint     int                 `json:"tech_point"`
	Reputation    int                 `json:"reputation"`
	Karma         int                 `json:"karma"`
	CurrentRoom   string              `json:"current_room"`
	Rooms         map[string]*RoomState `json:"rooms"`
	GPUs          []*GPU              `json:"gpus"`
	NextGPUID     int                 `json:"next_gpu_id"`
	Modifiers     []Modifier          `json:"modifiers"`
	EventCooldown EventCooldowns      `json:"event_cooldown"`
	LastTickUnix  int64               `json:"last_tick_unix"`
	LastBillUnix  int64               `json:"last_bill_unix"`
	Log           []LogEntry          `json:"log"`
	Paused        bool                `json:"paused"`
	StartedUnix   int64               `json:"started_unix"`
}

// NewState returns a fresh game.
func NewState(kittenName string) *State {
	if kittenName == "" {
		kittenName = "Whiskers"
	}
	now := time.Now().Unix()
	s := &State{
		Version:       1,
		KittenName:    kittenName,
		Money:         150,
		BTC:           0,
		BTCPriceSeed:  rand.Int63(),
		TechPoint:     0,
		Reputation:    0,
		Karma:         0,
		CurrentRoom:   "alley",
		Rooms:         map[string]*RoomState{},
		GPUs:          []*GPU{},
		NextGPUID:     1,
		Modifiers:     []Modifier{},
		EventCooldown: EventCooldowns{},
		LastTickUnix:  now,
		LastBillUnix:  now,
		StartedUnix:   now,
		Log:           []LogEntry{},
	}
	// Unlock every room flagged as default.
	for _, r := range data.Rooms() {
		if r.UnlockedByDefault {
			s.unlockRoomInternal(r)
		}
	}
	// Starter GPU.
	s.addGPU("gtx1060", "alley", true)
	s.appendLog("info", fmt.Sprintf("Welcome, %s. Your first GPU hums to life.", kittenName))
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

// addGPU creates a new GPU instance. If shipping is true it enters shipping
// state with a random 30-180s delivery window.
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

// BuyGPU purchases a GPU by def id, routes it to current room via shipping.
func (s *State) BuyGPU(defID string) error {
	def, ok := data.GPUByID(defID)
	if !ok {
		return fmt.Errorf("no such GPU: %s", defID)
	}
	if s.Money < float64(def.Price) {
		return fmt.Errorf("need $%d, have $%.0f", def.Price, s.Money)
	}
	// Check room slot availability (shipping GPUs count toward slots).
	if !s.RoomHasFreeSlot(s.CurrentRoom) {
		return fmt.Errorf("no free slots in this room")
	}
	s.Money -= float64(def.Price)
	s.addGPU(defID, s.CurrentRoom, true)
	s.appendLog("info", fmt.Sprintf("Ordered %s for $%d. Tracking inbound...", def.Name, def.Price))
	return nil
}

// SellGPU scraps a GPU for its scrap value.
func (s *State) SellGPU(instanceID int) error {
	for i, g := range s.GPUs {
		if g.InstanceID == instanceID {
			def, ok := data.GPUByID(g.DefID)
			if !ok {
				return fmt.Errorf("ghost GPU")
			}
			s.Money += float64(def.ScrapValue)
			s.GPUs = append(s.GPUs[:i], s.GPUs[i+1:]...)
			s.appendLog("info", fmt.Sprintf("Sold %s for $%d.", def.Name, def.ScrapValue))
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

// AppendLog is the external hook for the event system to write entries.
func (s *State) AppendLog(category, text string) {
	s.appendLog(category, text)
}

// --- Save / Load ---

func SavePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".meowmine.save.json"
	}
	dir := filepath.Join(home, ".meowmine")
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "save.json")
}

// Save writes state to disk.
func (s *State) Save() error {
	path := SavePath()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Load reads state from disk, or returns nil if no save exists.
func Load() (*State, error) {
	path := SavePath()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var s State
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
