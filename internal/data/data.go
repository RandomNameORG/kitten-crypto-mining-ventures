package data

import (
	_ "embed"
	"encoding/json"
)

//go:embed gpus.json
var gpusJSON []byte

//go:embed rooms.json
var roomsJSON []byte

//go:embed events.json
var eventsJSON []byte

type GPUDef struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Flavor          string  `json:"flavor"`
	Tier            string  `json:"tier"`
	Efficiency      float64 `json:"efficiency"`
	PowerDraw       float64 `json:"power_draw"`
	HeatOutput      float64 `json:"heat_output"`
	DurabilityHours int     `json:"durability_hours"`
	Price           int     `json:"price"`
	ScrapValue      int     `json:"scrap_value"`
	Special         string  `json:"special,omitempty"`
}

type RoomDef struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Flavor            string   `json:"flavor"`
	Slots             int      `json:"slots"`
	BaseCooling       float64  `json:"base_cooling"`
	ElectricCostMult  float64  `json:"electric_cost_mult"`
	RentPerHour       int      `json:"rent_per_hour"`
	ThreatBase        float64  `json:"threat_base"`
	ThreatPool        []string `json:"threat_pool"`
	UnlockCost        int      `json:"unlock_cost"`
	UnlockedByDefault bool     `json:"unlocked_by_default"`
}

type EventEffect struct {
	Kind    string  `json:"kind"`
	Chance  float64 `json:"chance,omitempty"`
	ChanceIfNoDefense float64 `json:"chance_if_no_defense,omitempty"`
	Count   int     `json:"count,omitempty"`
	Seconds int     `json:"seconds,omitempty"`
	Delta   int     `json:"delta,omitempty"`
	Factor  float64 `json:"factor,omitempty"`
	Tier    string  `json:"tier,omitempty"`
	Amount  float64 `json:"amount,omitempty"`
}

type EventDef struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Category    string        `json:"category"`
	Emoji       string        `json:"emoji"`
	Text        string        `json:"text"`
	Weight      int           `json:"weight"`
	CooldownSec int           `json:"cooldown_sec"`
	Effects     []EventEffect `json:"effects"`
}

var (
	gpus   []GPUDef
	rooms  []RoomDef
	events []EventDef
)

func init() {
	if err := json.Unmarshal(gpusJSON, &gpus); err != nil {
		panic("bad gpus.json: " + err.Error())
	}
	if err := json.Unmarshal(roomsJSON, &rooms); err != nil {
		panic("bad rooms.json: " + err.Error())
	}
	if err := json.Unmarshal(eventsJSON, &events); err != nil {
		panic("bad events.json: " + err.Error())
	}
}

func GPUs() []GPUDef   { return gpus }
func Rooms() []RoomDef { return rooms }
func Events() []EventDef { return events }

func GPUByID(id string) (GPUDef, bool) {
	for _, g := range gpus {
		if g.ID == id {
			return g, true
		}
	}
	return GPUDef{}, false
}

func RoomByID(id string) (RoomDef, bool) {
	for _, r := range rooms {
		if r.ID == id {
			return r, true
		}
	}
	return RoomDef{}, false
}

func EventByID(id string) (EventDef, bool) {
	for _, e := range events {
		if e.ID == id {
			return e, true
		}
	}
	return EventDef{}, false
}
