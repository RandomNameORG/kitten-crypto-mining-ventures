package data

import "github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"

// MasteryTrack is one infinite-ladder TP sink. Unlike the bounded skill
// tree, mastery levels can keep climbing — costs ramp linearly, effect
// stacks multiplicatively. The point is to absorb late-game TP overflow
// and give players a "numbers go up" goal that survives every prestige.
type MasteryTrack struct {
	ID       string
	Name     string
	NameZH   string
	Desc     string
	DescZH   string
	Emoji    string
	Effect   string  // "mining" | "power" | "cooling" | "frags"
	PerLevel float64 // multiplicative factor applied per level
	BaseCost int     // first level's TP cost; total cost = base + level*step
	StepCost int
	MaxLevel int
}

func (t MasteryTrack) LocalName() string { return i18n.Pick(t.Name, t.NameZH) }
func (t MasteryTrack) LocalDesc() string { return i18n.Pick(t.Desc, t.DescZH) }

// CostFor returns the TP cost to advance FROM `currentLevel` to currentLevel+1.
// Cost ramps linearly so early gains are cheap and 50→51 is a real commitment.
func (t MasteryTrack) CostFor(currentLevel int) int {
	if currentLevel >= t.MaxLevel {
		return -1
	}
	return t.BaseCost + currentLevel*t.StepCost
}

var masteryTracks = []MasteryTrack{
	{
		ID: "mining", Emoji: "⛏",
		Name: "Mining Mastery", NameZH: "挖矿精通",
		Desc:   "+1% earn rate per level, multiplicative.",
		DescZH: "每级 +1% 产出（乘法叠加）。",
		Effect: "mining", PerLevel: 0.01, BaseCost: 3, StepCost: 2, MaxLevel: 50,
	},
	{
		ID: "power", Emoji: "⚡",
		Name: "Power Engineering", NameZH: "电力工程",
		Desc:   "−1% electricity bills per level, multiplicative.",
		DescZH: "每级 −1% 电费（乘法叠加）。",
		Effect: "power", PerLevel: -0.01, BaseCost: 3, StepCost: 2, MaxLevel: 50,
	},
	{
		ID: "cooling", Emoji: "❄",
		Name: "Thermal Mastery", NameZH: "散热精通",
		Desc:   "+2% room cooling per level, multiplicative.",
		DescZH: "每级 +2% 房间散热（乘法叠加）。",
		Effect: "cooling", PerLevel: 0.02, BaseCost: 4, StepCost: 2, MaxLevel: 50,
	},
	{
		ID: "frags", Emoji: "🔬",
		Name: "Fragment Mastery", NameZH: "碎片精通",
		Desc:   "+2% research fragments from scrapping per level.",
		DescZH: "每级拆解 +2% 碎片产出。",
		Effect: "frags", PerLevel: 0.02, BaseCost: 4, StepCost: 2, MaxLevel: 50,
	},
	{
		ID: "scrap", Emoji: "♻",
		Name: "Salvage Mastery", NameZH: "拆解精通",
		Desc:   "+1.5% scrap value per level.",
		DescZH: "每级 +1.5% 拆解售价。",
		Effect: "scrap", PerLevel: 0.015, BaseCost: 3, StepCost: 2, MaxLevel: 50,
	},
}

// MasteryTracks returns the catalog (read-only — mutating the slice is UB).
func MasteryTracks() []MasteryTrack { return masteryTracks }

// MasteryByID returns the track or false if unknown.
func MasteryByID(id string) (MasteryTrack, bool) {
	for _, t := range masteryTracks {
		if t.ID == id {
			return t, true
		}
	}
	return MasteryTrack{}, false
}
