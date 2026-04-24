package data

import "github.com/RandomNameORG/kitten-crypto-mining-ventures/internal/i18n"

type MercDef struct {
	ID           string
	Name         string
	NameZH       string
	Flavor       string
	FlavorZH     string
	HireCost     int
	WeeklyWage   int     // paid every in-game week (1 week = 60 sim minutes)
	DefenseBonus float64 // added to effective lock/cctv defense vs threats
	Specialty    string  // "guard" | "tech" | "social" | "combat" | "sea"
	LoyaltyBase  int     // starting loyalty (0-100)
}

func (m MercDef) LocalName() string   { return i18n.Pick(m.Name, m.NameZH) }
func (m MercDef) LocalFlavor() string { return i18n.Pick(m.Flavor, m.FlavorZH) }

var mercDefs = []MercDef{
	{ID: "tabby_guard", HireCost: 200, WeeklyWage: 90, DefenseBonus: 0.15, Specialty: "guard", LoyaltyBase: 65,
		Name:     "Tabby Guard",
		NameZH:   "虎斑看门猫",
		Flavor:   "Watches the window all night. Never blinks.",
		FlavorZH: "整晚盯着窗户。眼都不眨。"},
	{ID: "siamese_it", HireCost: 350, WeeklyWage: 140, DefenseBonus: 0.05, Specialty: "tech", LoyaltyBase: 55,
		Name:     "Siamese IT",
		NameZH:   "暹罗 IT 猫",
		Flavor:   "Keeps the cable spaghetti under control. Allegedly.",
		FlavorZH: "把线缆乱麻整理清爽——据说。"},
	{ID: "ragdoll_pr", HireCost: 300, WeeklyWage: 120, DefenseBonus: 0.05, Specialty: "social", LoyaltyBase: 70,
		Name:     "Ragdoll PR Cat",
		NameZH:   "布偶公关猫",
		Flavor:   "Very good with police, neighbors, and social media.",
		FlavorZH: "搞定警察、邻居和社交媒体都很在行。"},
	{ID: "persian_ex_mil", HireCost: 600, WeeklyWage: 220, DefenseBonus: 0.30, Specialty: "combat", LoyaltyBase: 50,
		Name:     "Persian Ex-Military",
		NameZH:   "波斯退伍兵",
		Flavor:   "Retired from a private army. Speaks three languages, bites.",
		FlavorZH: "退役自私人武装。会三门外语，也会咬人。"},
	{ID: "sphynx_oracle", HireCost: 900, WeeklyWage: 320, DefenseBonus: 0.10, Specialty: "lucky", LoyaltyBase: 40,
		Name:     "Sphynx Oracle",
		NameZH:   "无毛神谕猫",
		Flavor:   "Unsettlingly lucky. Events just… go well.",
		FlavorZH: "玄学的运气好。事件总是「刚好」顺利。"},
	{ID: "pirate_cat", HireCost: 1100, WeeklyWage: 380, DefenseBonus: 0.40, Specialty: "sea", LoyaltyBase: 35,
		Name:     "Retired Pirate Cat",
		NameZH:   "退役海盗猫",
		Flavor:   "Yarr. Keeps the actual pirates away.",
		FlavorZH: "YARR。负责把真正的海盗劝退。"},
}

func Mercs() []MercDef { return mercDefs }

func MercByID(id string) (MercDef, bool) {
	for _, m := range mercDefs {
		if m.ID == id {
			return m, true
		}
	}
	return MercDef{}, false
}
