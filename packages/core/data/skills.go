package data

import "github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"

// SkillEffect describes what a skill changes when unlocked.
type SkillEffect struct {
	Kind    string  // power_mult, bill_mult, scrap_mult, earn_damp, repair_free, merc_loyalty, unlock
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
	{ID: "undervolt_iii", Lane: "engineer", Cost: 6, Prereq: "undervolt_ii", Effect: SkillEffect{Kind: "power_mult", Value: 0.90},
		Name:   "Undervolt III",
		NameZH: "降压 III",
		Desc:   "Reduce GPU power draw by another 10% (cumulative ≈−27%).",
		DescZH: "显卡耗电再 −10%（累计约 −27%）。"},
	{ID: "overclock_i", Lane: "engineer", Cost: 4, Effect: SkillEffect{Kind: "overclock", Value: 0.10},
		Name:   "Overclock I",
		NameZH: "超频 I",
		Desc:   "+10% efficiency, +15% heat output.",
		DescZH: "效率 +10%，产热 +15%。"},
	{ID: "overclock_ii", Lane: "engineer", Cost: 6, Prereq: "overclock_i", Effect: SkillEffect{Kind: "overclock", Value: 0.10},
		Name:   "Overclock II",
		NameZH: "超频 II",
		Desc:   "Stacks another +10% efficiency / +15% heat.",
		DescZH: "再叠加 +10% 效率 / +15% 产热。"},
	{ID: "overclock_iii", Lane: "engineer", Cost: 8, Prereq: "overclock_ii", Effect: SkillEffect{Kind: "overclock", Value: 0.10},
		Name:   "Overclock III",
		NameZH: "超频 III",
		Desc:   "+10% more efficiency, +15% heat. Risk-reward sharpens.",
		DescZH: "再 +10% 效率 / +15% 产热。收益与风险并升。"},
	{ID: "pcb_surgery", Lane: "engineer", Cost: 6, Prereq: "overclock_i", Effect: SkillEffect{Kind: "repair_discount", Value: 0.5},
		Name:   "PCB Surgery I",
		NameZH: "PCB 外科手术 I",
		Desc:   "Repair costs cut in half.",
		DescZH: "维修费用半价。"},
	{ID: "pcb_surgery_ii", Lane: "engineer", Cost: 8, Prereq: "pcb_surgery", Effect: SkillEffect{Kind: "repair_discount", Value: 0.6},
		Name:   "PCB Surgery II",
		NameZH: "PCB 外科手术 II",
		Desc:   "Stacks another 40% off (≈30% of original cost).",
		DescZH: "再 −40%（约为原价 30%）。"},
	{ID: "auto_repair", Lane: "engineer", Cost: 8, Prereq: "pcb_surgery", Effect: SkillEffect{Kind: "auto_repair"},
		Name:   "Auto-Repair Loop I",
		NameZH: "自动维修 I",
		Desc:   "Auto-fixes one broken GPU every 60s.",
		DescZH: "每 60 秒自动修好一张坏掉的显卡。"},
	{ID: "auto_repair_ii", Lane: "engineer", Cost: 6, Prereq: "auto_repair", Effect: SkillEffect{Kind: "auto_repair_fast"},
		Name:   "Auto-Repair Loop II",
		NameZH: "自动维修 II",
		Desc:   "Auto-repair fires every 30s instead of 60s.",
		DescZH: "自动维修间隔从 60 秒缩短到 30 秒。"},
	{ID: "auto_repair_iii", Lane: "engineer", Cost: 8, Prereq: "auto_repair_ii", Effect: SkillEffect{Kind: "auto_repair_burst"},
		Name:   "Auto-Repair Loop III",
		NameZH: "自动维修 III",
		Desc:   "Auto-repair fixes ALL broken GPUs each cycle, not just one.",
		DescZH: "每次自动维修一次性修好所有坏卡，不再一次只修一张。"},
	{ID: "rd_unlock", Lane: "engineer", Cost: 12, Prereq: "undervolt_ii", Effect: SkillEffect{Kind: "unlock", Unlocks: "rd"},
		Name:   "MEOWCore Blueprint",
		NameZH: "MEOWCore 蓝图",
		Desc:   "Unlock custom GPU research.",
		DescZH: "解锁自研显卡研究。"},

	// Mogul.
	{ID: "smart_invoicing", Lane: "mogul", Cost: 3, Effect: SkillEffect{Kind: "bill_mult", Value: 0.85},
		Name:   "Smart Invoicing I",
		NameZH: "智能报税 I",
		Desc:   "Electricity bills −15%.",
		DescZH: "电费账单 −15%。"},
	{ID: "smart_invoicing_ii", Lane: "mogul", Cost: 5, Prereq: "smart_invoicing", Effect: SkillEffect{Kind: "bill_mult", Value: 0.85},
		Name:   "Smart Invoicing II",
		NameZH: "智能报税 II",
		Desc:   "Bills −15% more (cumulative ≈−28%).",
		DescZH: "电费再 −15%（累计约 −28%）。"},
	{ID: "smart_invoicing_iii", Lane: "mogul", Cost: 7, Prereq: "smart_invoicing_ii", Effect: SkillEffect{Kind: "bill_mult", Value: 0.85},
		Name:   "Smart Invoicing III",
		NameZH: "智能报税 III",
		Desc:   "Bills −15% more (cumulative ≈−39%).",
		DescZH: "电费再 −15%（累计约 −39%）。"},
	{ID: "tax_opt", Lane: "mogul", Cost: 4, Effect: SkillEffect{Kind: "scrap_mult", Value: 1.20},
		Name:   "Tax Optimization I",
		NameZH: "税务优化 I",
		Desc:   "Scrap / sell value +20%.",
		DescZH: "拆解/出售价值 +20%。"},
	{ID: "tax_opt_ii", Lane: "mogul", Cost: 5, Prereq: "tax_opt", Effect: SkillEffect{Kind: "scrap_mult", Value: 1.20},
		Name:   "Tax Optimization II",
		NameZH: "税务优化 II",
		Desc:   "Scrap +20% more (cumulative +44%).",
		DescZH: "拆解再 +20%（累计 +44%）。"},
	{ID: "tax_opt_iii", Lane: "mogul", Cost: 7, Prereq: "tax_opt_ii", Effect: SkillEffect{Kind: "scrap_mult", Value: 1.20},
		Name:   "Tax Optimization III",
		NameZH: "税务优化 III",
		Desc:   "Scrap +20% more (cumulative +73%).",
		DescZH: "拆解再 +20%（累计 +73%）。"},
	{ID: "hedged_wallet", Lane: "mogul", Cost: 6, Prereq: "smart_invoicing", Effect: SkillEffect{Kind: "earn_damp", Value: 0.50},
		Name:   "Hedged Wallet I",
		NameZH: "对冲钱包 I",
		Desc:   "Event earn-rate swings halved (both ways).",
		DescZH: "事件带来的产出波动减半（上下皆减）。"},
	{ID: "hedged_wallet_ii", Lane: "mogul", Cost: 8, Prereq: "hedged_wallet", Effect: SkillEffect{Kind: "earn_damp", Value: 0.50},
		Name:   "Hedged Wallet II",
		NameZH: "对冲钱包 II",
		Desc:   "Event earn-rate swings halved again (cumulative ÷4).",
		DescZH: "再减半（累计 ÷4）。"},
	{ID: "venture_cap", Lane: "mogul", Cost: 12, Prereq: "hedged_wallet", Effect: SkillEffect{Kind: "unlock", Unlocks: "prestige"},
		Name:   "Venture Capital",
		NameZH: "风险投资",
		Desc:   "Unlock Prestige retirement.",
		DescZH: "解锁「退休转生」。"},

	// Hacker.
	{ID: "neighbor_leech", Lane: "hacker", Cost: 3, Effect: SkillEffect{Kind: "bill_mult", Value: 0.90},
		Name:   "Neighbor Leech I",
		NameZH: "偷邻居电 I",
		Desc:   "Electricity bills −10% (stacks with Invoicing).",
		DescZH: "电费账单 −10%（与智能报税叠加）。"},
	{ID: "neighbor_leech_ii", Lane: "hacker", Cost: 5, Prereq: "neighbor_leech", Effect: SkillEffect{Kind: "bill_mult", Value: 0.90},
		Name:   "Neighbor Leech II",
		NameZH: "偷邻居电 II",
		Desc:   "Bills another −10% (stacks).",
		DescZH: "电费再 −10%（叠加）。"},
	{ID: "neighbor_leech_iii", Lane: "hacker", Cost: 7, Prereq: "neighbor_leech_ii", Effect: SkillEffect{Kind: "bill_mult", Value: 0.90},
		Name:   "Neighbor Leech III",
		NameZH: "偷邻居电 III",
		Desc:   "Bills another −10% (stacks).",
		DescZH: "电费再 −10%（叠加）。"},
	{ID: "pump_dump", Lane: "hacker", Cost: 6, Effect: SkillEffect{Kind: "unlock", Unlocks: "pump_dump_action"},
		Name:   "Pump & Dump I",
		NameZH: "拉盘砸盘 I",
		Desc:   "Manually trigger a ×1.5 mining boost (5 min, 30 min cooldown).",
		DescZH: "手动触发 ×1.5 产出加成（5 分钟，冷却 30 分钟）。"},
	{ID: "pump_dump_ii", Lane: "hacker", Cost: 8, Prereq: "pump_dump", Effect: SkillEffect{Kind: "pump_dump_cd", Value: 0.5},
		Name:   "Pump & Dump II",
		NameZH: "拉盘砸盘 II",
		Desc:   "Pump cooldown halved (15 min instead of 30).",
		DescZH: "拉盘冷却减半（15 分钟）。"},
	{ID: "chain_ghost", Lane: "hacker", Cost: 10, Prereq: "pump_dump", Effect: SkillEffect{Kind: "merc_loyalty", Value: 15},
		Name:   "Chain Ghost I",
		NameZH: "链上幽灵 I",
		Desc:   "Merc loyalty floor +15.",
		DescZH: "佣兵忠诚下限 +15。"},
	{ID: "chain_ghost_ii", Lane: "hacker", Cost: 8, Prereq: "chain_ghost", Effect: SkillEffect{Kind: "merc_loyalty", Value: 15},
		Name:   "Chain Ghost II",
		NameZH: "链上幽灵 II",
		Desc:   "Merc loyalty floor +15 more (total +30).",
		DescZH: "佣兵忠诚下限再 +15（共 +30）。"},
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
