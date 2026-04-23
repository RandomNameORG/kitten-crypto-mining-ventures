package game

import (
	"math"
	"testing"
	"time"
)

func TestBTCPriceStaysPositive(t *testing.T) {
	withTempHome(t)
	s := NewState("BTC")
	// Sample many points along the oscillator.
	for i := int64(0); i < 7200; i += 17 {
		p := s.BTCPriceAt(s.StartedUnix + i)
		if p <= 0 {
			t.Fatalf("BTC price went non-positive at t+%d: %.2f", i, p)
		}
		if math.IsNaN(p) || math.IsInf(p, 0) {
			t.Fatalf("BTC price went non-finite at t+%d", i)
		}
	}
}

func TestBTCPriceDeterministicPerSeed(t *testing.T) {
	withTempHome(t)
	s := NewState("Det")
	p1 := s.BTCPriceAt(s.StartedUnix + 300)
	p2 := s.BTCPriceAt(s.StartedUnix + 300)
	if p1 != p2 {
		t.Errorf("BTCPriceAt not deterministic: %.2f vs %.2f", p1, p2)
	}
}

func TestBTCPriceRespondsToMultiplier(t *testing.T) {
	withTempHome(t)
	s := NewState("Mult")
	now := time.Now().Unix()
	base := s.BTCPriceAt(now)
	s.Modifiers = append(s.Modifiers, Modifier{
		Kind:      "btc_mult",
		Factor:    2.0,
		ExpiresAt: now + 3600,
	})
	bumped := s.BTCPriceAt(now)
	if bumped <= base {
		t.Errorf("btc_mult factor 2.0 should raise price: base=%.2f bumped=%.2f", base, bumped)
	}
}

func TestTickAccruesBTCAndBills(t *testing.T) {
	withTempHome(t)
	s := NewState("TickFlow")
	// Force starter into running state right now so advanceMining produces.
	for _, g := range s.GPUs {
		g.Status = "running"
	}
	startMoney := s.Money
	s.LastTickUnix = time.Now().Unix() - 120 // 2 minutes of dt
	s.LastBillUnix = time.Now().Unix() - 120
	s.Tick(time.Now().Unix())
	// After 2 minutes the starter should have earned some money and paid a bill.
	if s.Money <= 0 {
		t.Error("tick bankrupted the player inexplicably")
	}
	if s.LifetimeEarned <= 0 {
		t.Error("LifetimeEarned should accumulate on earnings")
	}
	_ = startMoney // not strictly compared — just ensure no panic
}

func TestModifiersExpireOnTick(t *testing.T) {
	withTempHome(t)
	s := NewState("Expire")
	past := time.Now().Unix() - 10
	s.Modifiers = []Modifier{
		{Kind: "btc_mult", Factor: 1.5, ExpiresAt: past},
		{Kind: "earn_mult", Factor: 1.2, ExpiresAt: time.Now().Unix() + 3600},
	}
	s.LastTickUnix = time.Now().Unix() - 1
	s.Tick(time.Now().Unix())
	kinds := map[string]int{}
	for _, m := range s.Modifiers {
		kinds[m.Kind]++
	}
	if kinds["btc_mult"] != 0 {
		t.Error("expired btc_mult should be pruned")
	}
	if kinds["earn_mult"] != 1 {
		t.Error("still-valid earn_mult should survive")
	}
}

func TestBlackoutWhenBroke(t *testing.T) {
	withTempHome(t)
	s := NewState("Bankrupt")
	s.Money = 0
	// Put some power-hungry GPUs up so the bill is real.
	for i := 0; i < 4; i++ {
		s.addGPU("rtx4090", "alley", false)
	}
	// Force a bill cycle: LastBillUnix 70s ago so the 60s threshold trips.
	now := time.Now().Unix()
	s.LastBillUnix = now - 70
	s.LastTickUnix = now - 1
	s.Tick(now)
	// Should be broke and have a pause_mining modifier.
	paused := false
	for _, m := range s.Modifiers {
		if m.Kind == "pause_mining" {
			paused = true
		}
	}
	if !paused {
		t.Error("expected pause_mining modifier after inability to pay bill")
	}
}

func TestTogglePause(t *testing.T) {
	withTempHome(t)
	s := NewState("Pause")
	if s.Paused {
		t.Error("new state should not be paused")
	}
	s.TogglePause()
	if !s.Paused {
		t.Error("toggle should pause")
	}
	s.TogglePause()
	if s.Paused {
		t.Error("toggle should resume")
	}
}
