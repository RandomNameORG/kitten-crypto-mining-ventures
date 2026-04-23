package data

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
	Desc   string
	Cost   int
	Prereq string // another SkillDef.ID, or "" for top-of-lane
	Effect SkillEffect
}

var skillDefs = []SkillDef{
	// Engineer.
	{ID: "undervolt_i", Lane: "engineer", Name: "Undervolt I", Desc: "Reduce GPU power draw by 10%.",
		Cost: 3, Effect: SkillEffect{Kind: "power_mult", Value: 0.90}},
	{ID: "undervolt_ii", Lane: "engineer", Name: "Undervolt II", Desc: "Reduce GPU power draw by another 10%.",
		Cost: 4, Prereq: "undervolt_i", Effect: SkillEffect{Kind: "power_mult", Value: 0.90}},
	{ID: "overclock_i", Lane: "engineer", Name: "Overclock I", Desc: "+10% efficiency, +15% heat output.",
		Cost: 4, Effect: SkillEffect{Kind: "overclock", Value: 0.10}},
	{ID: "pcb_surgery", Lane: "engineer", Name: "PCB Surgery", Desc: "Repairs are free.",
		Cost: 6, Prereq: "overclock_i", Effect: SkillEffect{Kind: "repair_free"}},
	{ID: "rd_unlock", Lane: "engineer", Name: "MEOWCore Blueprint", Desc: "Unlock custom GPU research.",
		Cost: 12, Prereq: "undervolt_ii", Effect: SkillEffect{Kind: "unlock", Unlocks: "rd"}},

	// Mogul.
	{ID: "smart_invoicing", Lane: "mogul", Name: "Smart Invoicing", Desc: "Electricity bills −15%.",
		Cost: 3, Effect: SkillEffect{Kind: "bill_mult", Value: 0.85}},
	{ID: "tax_opt", Lane: "mogul", Name: "Tax Optimization", Desc: "Scrap / sell value +20%.",
		Cost: 4, Effect: SkillEffect{Kind: "scrap_mult", Value: 1.20}},
	{ID: "hedged_wallet", Lane: "mogul", Name: "Hedged Wallet", Desc: "BTC volatility halved.",
		Cost: 6, Prereq: "smart_invoicing", Effect: SkillEffect{Kind: "btc_damp", Value: 0.50}},
	{ID: "venture_cap", Lane: "mogul", Name: "Venture Capital", Desc: "Unlock Prestige retirement.",
		Cost: 12, Prereq: "hedged_wallet", Effect: SkillEffect{Kind: "unlock", Unlocks: "prestige"}},

	// Hacker.
	{ID: "neighbor_leech", Lane: "hacker", Name: "Neighbor Leech", Desc: "Electricity bills −10% (stacks with Invoicing).",
		Cost: 3, Effect: SkillEffect{Kind: "bill_mult", Value: 0.90}},
	{ID: "pump_dump", Lane: "hacker", Name: "Pump & Dump", Desc: "Manually trigger a BTC pump (cooldown).",
		Cost: 6, Effect: SkillEffect{Kind: "unlock", Unlocks: "pump_dump_action"}},
	{ID: "chain_ghost", Lane: "hacker", Name: "Chain Ghost", Desc: "Merc loyalty floor +15.",
		Cost: 12, Prereq: "pump_dump", Effect: SkillEffect{Kind: "merc_loyalty", Value: 15}},
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
