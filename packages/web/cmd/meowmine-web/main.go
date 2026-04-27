package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/game"
	"github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"
)

type webGame struct {
	mu            sync.Mutex
	state         *game.State
	lastEvent     *eventView
	lastEventRoll int64
	lastEventSeq  int64
}

type snapshot struct {
	State     stateView     `json:"state"`
	Rooms     []roomView    `json:"rooms"`
	GPUs      []gpuView     `json:"gpus"`
	GPUDefs   []gpuDefView  `json:"gpu_defs"`
	Skills    []skillView   `json:"skills"`
	Mercs     []mercView    `json:"mercs"`
	MercDefs  []mercDefView `json:"merc_defs"`
	Log       []logView     `json:"log"`
	LastEvent *eventView    `json:"last_event,omitempty"`
	OK        bool          `json:"ok"`
}

type stateView struct {
	KittenName        string  `json:"kitten_name"`
	BTC               float64 `json:"btc"`
	BTCFmt            string  `json:"btc_fmt"`
	TechPoint         int     `json:"tech_point"`
	ResearchFrags     int     `json:"research_frags"`
	Reputation        int     `json:"reputation"`
	Karma             int     `json:"karma"`
	CurrentRoom       string  `json:"current_room"`
	Paused            bool    `json:"paused"`
	MarketPrice       float64 `json:"market_price"`
	MarketTrend       int     `json:"market_trend"`
	LifetimeEarnedFmt string  `json:"lifetime_earned_fmt"`
	RoomEarnFmt       string  `json:"room_earn_fmt"`
	RoomBillFmt       string  `json:"room_bill_fmt"`
	RoomNetFmt        string  `json:"room_net_fmt"`
	MiningPaused      bool    `json:"mining_paused"`
	SyndicateJoined   bool    `json:"syndicate_joined"`
}

type roomView struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Flavor        string      `json:"flavor"`
	Slots         int         `json:"slots"`
	UnlockCost    int         `json:"unlock_cost"`
	UnlockCostFmt string      `json:"unlock_cost_fmt"`
	Unlocked      bool        `json:"unlocked"`
	Current       bool        `json:"current"`
	GPUCount      int         `json:"gpu_count"`
	Heat          float64     `json:"heat"`
	MaxHeat       float64     `json:"max_heat"`
	HeatPct       float64     `json:"heat_pct"`
	HeatDelta     float64     `json:"heat_delta"`
	HeatTickIn    int         `json:"heat_tick_in"`
	EarnFmt       string      `json:"earn_fmt"`
	BillFmt       string      `json:"bill_fmt"`
	NetFmt        string      `json:"net_fmt"`
	Defense       defenseView `json:"defense"`
	Background    string      `json:"background"`
}

type defenseView struct {
	Lock    int `json:"lock"`
	CCTV    int `json:"cctv"`
	Wiring  int `json:"wiring"`
	Cooling int `json:"cooling"`
	Armor   int `json:"armor"`
}

type gpuView struct {
	InstanceID int     `json:"instance_id"`
	DefID      string  `json:"def_id"`
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	Room       string  `json:"room"`
	Upgrade    int     `json:"upgrade"`
	OCLevel    int     `json:"oc_level"`
	HoursLeft  float64 `json:"hours_left"`
	EarnFmt    string  `json:"earn_fmt"`
	Repairable bool    `json:"repairable"`
	ShipsAt      int64 `json:"ships_at,omitempty"`
	ShipEtaSec   int64 `json:"ship_eta_sec,omitempty"`
	ShipTotalSec int64 `json:"ship_total_sec,omitempty"`
}

type gpuDefView struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Flavor     string  `json:"flavor"`
	Tier       string  `json:"tier"`
	Efficiency float64 `json:"efficiency"`
	PowerDraw  float64 `json:"power_draw"`
	HeatOutput float64 `json:"heat_output"`
	Price      int     `json:"price"`
	PriceFmt   string  `json:"price_fmt"`
	ScrapFmt   string  `json:"scrap_fmt"`
}

type skillView struct {
	ID       string `json:"id"`
	Lane     string `json:"lane"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Cost     int    `json:"cost"`
	Prereq   string `json:"prereq"`
	Unlocked bool   `json:"unlocked"`
}

type mercView struct {
	InstanceID int    `json:"instance_id"`
	DefID      string `json:"def_id"`
	Name       string `json:"name"`
	Loyalty    int    `json:"loyalty"`
	RoomID     string `json:"room_id"`
}

type mercDefView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Flavor      string `json:"flavor"`
	HireCostFmt string `json:"hire_cost_fmt"`
	WageFmt     string `json:"wage_fmt"`
	Specialty   string `json:"specialty"`
}

type logView struct {
	Time     int64  `json:"time"`
	Category string `json:"category"`
	Text     string `json:"text"`
}

type eventView struct {
	Seq      int64  `json:"seq"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Text     string `json:"text"`
}

type actionRequest struct {
	Action     string `json:"action"`
	ID         string `json:"id"`
	Dim        string `json:"dim"`
	InstanceID int    `json:"instance_id"`
}

func main() {
	addr := flag.String("addr", ":8080", "web server address")
	lang := flag.String("lang", "zh", "language: zh or en")
	flag.Parse()

	i18n.SetLang(*lang)
	state := game.NewState("矿业大亨喵")
	state.SetDifficulty("normal")
	wg := &webGame{state: state}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/snapshot", wg.handleSnapshot)
	mux.HandleFunc("/api/action", wg.handleAction)
	mux.Handle("/assets/", http.FileServer(http.Dir(".")))
	mux.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir(frontendDistDir))))
	mux.HandleFunc("/", serveIndex)

	log.Printf("meowmine 2D web running at http://localhost%s", displayAddr(*addr))
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join(frontendDistDir, "index.html"))
}

func (wg *webGame) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	wg.mu.Lock()
	defer wg.mu.Unlock()
	wg.advanceLocked()
	writeJSON(w, wg.makeSnapshotLocked())
}

func (wg *webGame) handleAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req actionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	wg.mu.Lock()
	defer wg.mu.Unlock()

	var err error
	switch req.Action {
	case "buy_gpu":
		err = wg.state.BuyGPU(req.ID)
	case "switch_room":
		err = wg.state.SwitchRoom(req.ID)
	case "unlock_room":
		err = wg.state.UnlockRoom(req.ID)
	case "upgrade_defense":
		err = wg.state.UpgradeDefense(req.Dim)
	case "upgrade_gpu":
		err = wg.state.UpgradeGPU(req.InstanceID)
	case "repair_gpu":
		err = wg.state.RepairGPU(req.InstanceID)
	case "scrap_gpu":
		err = wg.state.SellGPU(req.InstanceID)
	case "cycle_oc":
		err = wg.state.CycleGPUOC(req.InstanceID)
	case "vent":
		err = wg.state.EmergencyVent()
	case "toggle_pause":
		wg.state.TogglePause()
	case "unlock_skill":
		err = wg.state.UnlockSkill(req.ID)
	case "hire_merc":
		err = wg.state.HireMerc(req.ID)
	case "bribe_merc":
		err = wg.state.BribeMerc(req.InstanceID)
	case "fire_merc":
		err = wg.state.FireMerc(req.InstanceID)
	case "reset":
		wg.state = game.NewState("矿业大亨喵")
		wg.state.SetDifficulty("normal")
		wg.lastEvent = nil
		wg.lastEventRoll = 0
		wg.lastEventSeq = 0
	default:
		err = fmt.Errorf("unknown action %q", req.Action)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	wg.advanceLocked()
	writeJSON(w, wg.makeSnapshotLocked())
}

func (wg *webGame) advanceLocked() {
	now := time.Now().Unix()
	wg.state.Tick(now)
	if now > wg.lastEventRoll {
		wg.lastEventRoll = now
		if def := wg.state.MaybeFireEvent(); def != nil {
			wg.lastEventSeq++
			wg.lastEvent = &eventView{
				Seq:      wg.lastEventSeq,
				ID:       def.ID,
				Name:     def.LocalName(),
				Category: def.Category,
				Text:     def.LocalText(),
			}
		}
	}
}

func (wg *webGame) makeSnapshotLocked() snapshot {
	s := wg.state
	currentEarn := s.RoomEarnRatePerSec(s.CurrentRoom)
	currentBill := s.RoomBillRatePerSec(s.CurrentRoom)
	out := snapshot{
		State: stateView{
			KittenName:        s.KittenName,
			BTC:               s.BTC,
			BTCFmt:            game.FmtBTC(s.BTC),
			TechPoint:         s.TechPoint,
			ResearchFrags:     s.ResearchFrags,
			Reputation:        s.Reputation,
			Karma:             s.Karma,
			CurrentRoom:       s.CurrentRoom,
			Paused:            s.Paused,
			MarketPrice:       s.MarketPrice,
			MarketTrend:       s.MarketTrend(),
			LifetimeEarnedFmt: game.FmtBTC(s.LifetimeEarned),
			RoomEarnFmt:       game.FmtBTC(currentEarn) + "/s",
			RoomBillFmt:       game.FmtBTC(currentBill) + "/s",
			RoomNetFmt:        game.FmtBTCSigned(currentEarn-currentBill) + "/s",
			MiningPaused:      s.IsMiningPaused(time.Now().Unix()),
			SyndicateJoined:   s.SyndicateJoined,
		},
		OK: true,
	}

	for _, def := range data.Rooms() {
		rs, unlocked := s.Rooms[def.ID]
		gpuCount := len(s.GPUsInRoom(def.ID))
		room := roomView{
			ID:            def.ID,
			Name:          def.LocalName(),
			Flavor:        def.LocalFlavor(),
			Slots:         def.Slots,
			UnlockCost:    def.UnlockCost,
			UnlockCostFmt: game.FmtBTCInt(def.UnlockCost),
			Unlocked:      unlocked,
			Current:       s.CurrentRoom == def.ID,
			GPUCount:      gpuCount,
			EarnFmt:       game.FmtBTC(s.RoomEarnRatePerSec(def.ID)) + "/s",
			BillFmt:       game.FmtBTC(s.RoomBillRatePerSec(def.ID)) + "/s",
			NetFmt:        game.FmtBTCSigned(s.RoomNetRatePerSec(def.ID)) + "/s",
			Background:    "/assets/2d/backgrounds/rooms/" + def.ID + "/background.png",
		}
		if rs != nil {
			room.Heat = rs.Heat
			room.MaxHeat = rs.MaxHeat
			if rs.MaxHeat > 0 {
				room.HeatPct = rs.Heat / rs.MaxHeat
			}
			room.HeatDelta, _ = s.RoomHeatDeltaPerTick(def.ID)
			room.HeatTickIn = s.SecondsUntilNextHeatTick(def.ID)
			room.Defense = defenseView{
				Lock:    rs.LockLvl,
				CCTV:    rs.CCTVLvl,
				Wiring:  rs.WiringLvl,
				Cooling: rs.CoolingLvl,
				Armor:   rs.ArmorLvl,
			}
		}
		out.Rooms = append(out.Rooms, room)
	}

	nowUnix := time.Now().Unix()
	for _, g := range s.GPUs {
		name := g.DefID
		if def, ok := data.GPUByID(g.DefID); ok {
			name = def.LocalName()
		} else if g.BlueprintID != "" {
			name = "MEOWCore"
		}
		view := gpuView{
			InstanceID: g.InstanceID,
			DefID:      g.DefID,
			Name:       name,
			Status:     g.Status,
			Room:       g.Room,
			Upgrade:    g.UpgradeLevel,
			OCLevel:    g.OCLevel,
			HoursLeft:  g.HoursLeft,
			EarnFmt:    game.FmtBTC(s.GPUEarnRatePerSec(g)) + "/s",
			Repairable: g.Status == "broken",
		}
		if g.Status == "shipping" && g.ShipsAt > 0 {
			view.ShipsAt = g.ShipsAt
			eta := g.ShipsAt - nowUnix
			if eta < 0 {
				eta = 0
			}
			view.ShipEtaSec = eta
			view.ShipTotalSec = g.ShipTotalSec
		}
		out.GPUs = append(out.GPUs, view)
	}

	for _, def := range data.GPUs() {
		out.GPUDefs = append(out.GPUDefs, gpuDefView{
			ID:         def.ID,
			Name:       def.LocalName(),
			Flavor:     def.LocalFlavor(),
			Tier:       def.Tier,
			Efficiency: def.Efficiency,
			PowerDraw:  def.PowerDraw,
			HeatOutput: def.HeatOutput,
			Price:      def.Price,
			PriceFmt:   game.FmtBTCInt(def.Price),
			ScrapFmt:   game.FmtBTCInt(def.ScrapValue),
		})
	}

	for _, def := range data.Skills() {
		out.Skills = append(out.Skills, skillView{
			ID:       def.ID,
			Lane:     def.Lane,
			Name:     def.LocalName(),
			Desc:     def.LocalDesc(),
			Cost:     def.Cost,
			Prereq:   def.Prereq,
			Unlocked: s.HasSkill(def.ID),
		})
	}

	for _, m := range s.Mercs {
		name := m.DefID
		if def, ok := data.MercByID(m.DefID); ok {
			name = def.LocalName()
		}
		out.Mercs = append(out.Mercs, mercView{
			InstanceID: m.InstanceID,
			DefID:      m.DefID,
			Name:       name,
			Loyalty:    m.Loyalty,
			RoomID:     m.RoomID,
		})
	}
	for _, def := range data.Mercs() {
		out.MercDefs = append(out.MercDefs, mercDefView{
			ID:          def.ID,
			Name:        def.LocalName(),
			Flavor:      def.LocalFlavor(),
			HireCostFmt: game.FmtBTCInt(def.HireCost),
			WageFmt:     game.FmtBTCInt(def.WeeklyWage),
			Specialty:   def.Specialty,
		})
	}

	start := len(s.Log) - 9
	if start < 0 {
		start = 0
	}
	for _, e := range s.Log[start:] {
		out.Log = append(out.Log, logView{Time: e.Time, Category: e.Category, Text: e.Text})
	}
	out.LastEvent = wg.lastEvent
	return out
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":    false,
		"error": err.Error(),
	})
}

func displayAddr(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return addr
	}
	if strings.HasPrefix(addr, "0.0.0.0:") {
		return strings.TrimPrefix(addr, "0.0.0.0")
	}
	if _, err := strconv.Atoi(addr); err == nil {
		return ":" + addr
	}
	return addr
}

const frontendDistDir = "packages/web/frontend/dist"

func init() {
	if wd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(wd, frontendDistDir, "index.html")); err != nil {
			log.Printf("warning: %s/index.html missing — run `make frontend-build` first", frontendDistDir)
		}
	}
}
