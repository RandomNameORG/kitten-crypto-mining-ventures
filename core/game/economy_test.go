package game

import (
	"testing"
	"time"
)

func TestTickAccruesBTCAndBills(t *testing.T) {
	withTempHome(t)
	s := NewState("TickFlow")
	for _, g := range s.GPUs {
		g.Status = "running"
	}
	startBTC := s.BTC
	s.LastTickUnix = time.Now().Unix() - 120
	s.LastBillUnix = time.Now().Unix() - 120
	s.Tick(time.Now().Unix())
	if s.BTC <= 0 {
		t.Error("tick bankrupted the player inexplicably")
	}
	if s.LifetimeEarned <= 0 {
		t.Error("LifetimeEarned should accumulate on earnings")
	}
	_ = startBTC
}

func TestModifiersExpireOnTick(t *testing.T) {
	withTempHome(t)
	s := NewState("Expire")
	past := time.Now().Unix() - 10
	s.Modifiers = []Modifier{
		{Kind: "earn_mult", Factor: 1.5, ExpiresAt: past},
		{Kind: "earn_mult", Factor: 1.2, ExpiresAt: time.Now().Unix() + 3600},
	}
	s.LastTickUnix = time.Now().Unix() - 1
	s.Tick(time.Now().Unix())
	if len(s.Modifiers) != 1 {
		t.Fatalf("expected exactly one modifier after prune, got %d", len(s.Modifiers))
	}
	if s.Modifiers[0].Factor != 1.2 {
		t.Error("expected the still-valid earn_mult (factor 1.2) to survive")
	}
}

func TestBlackoutWhenBroke(t *testing.T) {
	withTempHome(t)
	s := NewState("Bankrupt")
	s.BTC = 0
	for i := 0; i < 4; i++ {
		s.addGPU("rtx4090", "alley", false)
	}
	now := time.Now().Unix()
	s.LastBillUnix = now - 70
	s.advanceBilling(now)
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
