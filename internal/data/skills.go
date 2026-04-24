package data

import "github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/i18n"

// SkillEffect describes what a skill changes when unlocked.
type SkillEffect struct {
	Kind    string  // power_mult, bill_mult, scrap_mult, btc_damp, repair_free, merc_loyalty, unlock
	Value   float64 // multiplier or delta, depending on Kind
	Unlocks string  // feature gate (e.g. "rd", "prestige", "pump_dump_action")
}

type SkillDef struct {
	ID     string
	Lane   string // "engineer" | "mogul" | "hacker"
	Name   string
	NameZH string
	Desc   string
	DescZH string
	Cost   int
	Prereq string // another SkillDef.ID, or "" for top-of-lane
	Effect SkillEffect
}

func (s SkillDef) LocalName() string { return i18n.Pick(s.Name, s.NameZH) }
func (s SkillDef) LocalDesc() string { return i18n.Pick(s.Desc, s.DescZH) }

var skillDefs = []SkillDef{
	// Engineer.
	{ID: "undervolt_i", Lane: "engineer", Cost: 3, Effect: SkillEffect{Kind: "power_mult", Value: 0.90},
		Name:   "Undervolt I",
		NameZH: "降压 I",
		Desc:   "Reduce GPU power draw by 10%.",
		DescZH: "显卡耗电 −10%。"},
	{ID: "undervolt_ii", Lane: "engineer", Cost: 4, Prereq: "undervolt_i", Effect: SkillEffect{Kind: "power_mult", Value: 0.90},
		Name:   "Undervolt II",
		NameZH: "降压 II",
		Desc:   "Reduce GPU power draw by another 10%.",
		DescZH: "显卡耗电再 −10%。"},
	{ID: "overclock_i", Lane: "engineer", Cost: 4, Effect: SkillEffect{Kind: "overclock", Value: 0.10},
		Name:   "Overclock I",
		NameZH: "超频 I",
		Desc:   "+10% efficiency, +15% heat output.",
		DescZH: "效率 +10%，产热 +15%。"},
	{ID: "pcb_surgery", Lane: "engineer", Cost: 6, Prereq: "overclock_i", Effect: SkillEffect{Kind: "repair_free"},
		Name:   "PCB Surgery",
		NameZH: "PCB 外科手术",
		Desc:   "Repairs are free.",
		DescZH: "维修不花钱。"},
	{ID: "rd_unlock", Lane: "engineer", Cost: 12, Prereq: "undervolt_ii", Effect: SkillEffect{Kind: "unlock", Unlocks: "rd"},
		Name:   "MEOWCore Blueprint",
		NameZH: "MEOWCore 蓝图",
		Desc:   "Unlock custom GPU research.",
		DescZH: "解锁自研显卡研究。"},

	// Mogul.
	{ID: "smart_invoicing", Lane: "mogul", Cost: 3, Effect: SkillEffect{Kind: "bill_mult", Value: 0.85},
		Name:   "Smart Invoicing",
		NameZH: "智能报税",
		Desc:   "Electricity bills −15%.",
		DescZH: "电费账单 −15%。"},
	{ID: "tax_opt", Lane: "mogul", Cost: 4, Effect: SkillEffect{Kind: "scrap_mult", Value: 1.20},
		Name:   "Tax Optimization",
		NameZH: "税务优化",
		Desc:   "Scrap / sell value +20%.",
		DescZH: "拆解/出售价值 +20%。"},
	{ID: "hedged_wallet", Lane: "mogul", Cost: 6, Prereq: "smart_invoicing", Effect: SkillEffect{Kind: "btc_damp", Value: 0.50},
		Name:   "Hedged Wallet",
		NameZH: "对冲钱包",
		Desc:   "BTC volatility halved.",
		DescZH: "BTC 价格波动减半。"},
	{ID: "venture_cap", Lane: "mogul", Cost: 12, Prereq: "hedged_wallet", Effect: SkillEffect{Kind: "unlock", Unlocks: "prestige"},
		Name:   "Venture Capital",
		NameZH: "风险投资",
		Desc:   "Unlock Prestige retirement.",
		DescZH: "解锁「退休转生」。"},

	// Hacker.
	{ID: "neighbor_leech", Lane: "hacker", Cost: 3, Effect: SkillEffect{Kind: "bill_mult", Value: 0.90},
		Name:   "Neighbor Leech",
		NameZH: "偷邻居电",
		Desc:   "Electricity bills −10% (stacks with Invoicing).",
		DescZH: "电费账单 −10%（与智能报税叠加）。"},
	{ID: "pump_dump", Lane: "hacker", Cost: 6, Effect: SkillEffect{Kind: "unlock", Unlocks: "pump_dump_action"},
		Name:   "Pump & Dump",
		NameZH: "拉盘砸盘",
		Desc:   "Manually trigger a BTC pump (cooldown).",
		DescZH: "手动触发 BTC 拉盘（有冷却）。"},
	{ID: "chain_ghost", Lane: "hacker", Cost: 12, Prereq: "pump_dump", Effect: SkillEffect{Kind: "merc_loyalty", Value: 15},
		Name:   "Chain Ghost",
		NameZH: "链上幽灵",
		Desc:   "Merc loyalty floor +15.",
		DescZH: "佣兵忠诚下限 +15。"},
}

func Skills() []SkillDef { return skillDefs }

func SkillByID(id string) (SkillDef, bool) {
	for _, s := range skillDefs {
		if s.ID == id {
			return s, true
		}
	}
	return SkillDef{}, false
}

func SkillsByLane(lane string) []SkillDef {
	out := []SkillDef{}
	for _, s := range skillDefs {
		if s.Lane == lane {
			out = append(out, s)
		}
	}
	return out
}
