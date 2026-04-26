package data

import "github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"

// DifficultyDef describes one of the fixed difficulty tiers. Picked once per
// save at game start via the splash screen; locks the multipliers for the
// entire run (player must `-new` to change).
type DifficultyDef struct {
	ID          string  // "easy" | "normal" | "hard" | "crypto_winter"
	Emoji       string
	LabelEN     string
	LabelZH     string
	DescEN      string
	DescZH      string
	EarnMult    float64 // applied to BTC produced per tick
	BillMult    float64 // applied to electricity + rent
	ThreatMult  float64 // applied to MaybeFireEvent per-tick probability
	StarterCash float64 // replaces the hardcoded $150 opener
	// MarketVolatilityMult scales the Gaussian drift kick in advanceMarket and
	// widens the price clamp band symmetrically around 1.0. 1.0 is the
	// default; higher values produce wilder market swings.
	MarketVolatilityMult float64
	// EventFreqMult multiplies MaybeFireEvent's per-tick fire probability.
	// 1.0 is the default; kept separate from ThreatMult so a tier can tune
	// event cadence and severity independently.
	EventFreqMult float64
}

func (d DifficultyDef) LocalLabel() string { return i18n.Pick(d.LabelEN, d.LabelZH) }
func (d DifficultyDef) LocalDesc() string  { return i18n.Pick(d.DescEN, d.DescZH) }

var difficulties = []DifficultyDef{
	{
		ID: "easy", Emoji: "🐾",
		LabelEN: "Kitten Kindergarten", LabelZH: "小猫幼儿园",
		DescEN: "Relaxed pacing. Earnings flow fast, bills are gentle, events are rare.",
		DescZH: "轻松节奏。收益快、电费温柔、事件稀疏。",
		EarnMult: 1.5, BillMult: 0.75, ThreatMult: 0.5, StarterCash: 300,
		MarketVolatilityMult: 1.0, EventFreqMult: 1.0,
	},
	{
		ID: "normal", Emoji: "🐈",
		LabelEN: "Alley Cat", LabelZH: "流浪猫",
		DescEN: "The tuned defaults. Balanced for most players.",
		DescZH: "调好的默认值。平衡向,推荐大多数玩家。",
		EarnMult: 1.0, BillMult: 1.0, ThreatMult: 1.0, StarterCash: 150,
		MarketVolatilityMult: 1.0, EventFreqMult: 1.0,
	},
	{
		ID: "hard", Emoji: "😾",
		LabelEN: "Feral", LabelZH: "野猫硬核",
		DescEN: "Tight margins. Pirates find you faster. For players who like to sweat the numbers.",
		DescZH: "利润紧绷。海盗找你更勤。适合喜欢算账的玩家。",
		EarnMult: 0.75, BillMult: 1.25, ThreatMult: 1.5, StarterCash: 100,
		MarketVolatilityMult: 1.0, EventFreqMult: 1.0,
	},
	{
		ID: "crypto_winter", Emoji: "🧊",
		LabelEN: "Crypto Winter", LabelZH: "加密寒冬",
		DescEN: "Frozen markets, empty wallets, constant crisis. For veterans who want the charts to scream.",
		DescZH: "市场冰封,钱包空空,危机不断。适合想看行情疯狂跳动的老玩家。",
		EarnMult: 0.6, BillMult: 1.4, ThreatMult: 1.75, StarterCash: 100,
		MarketVolatilityMult: 2.0, EventFreqMult: 1.5,
	},
}

const DefaultDifficulty = "normal"

func Difficulties() []DifficultyDef { return difficulties }

func DifficultyByID(id string) DifficultyDef {
	for _, d := range difficulties {
		if d.ID == id {
			return d
		}
	}
	// Fall back to normal so missing/unknown IDs never crash.
	for _, d := range difficulties {
		if d.ID == DefaultDifficulty {
			return d
		}
	}
	return difficulties[0]
}
