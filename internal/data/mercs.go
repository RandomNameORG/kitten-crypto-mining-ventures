package data

type MercDef struct {
	ID           string
	Name         string
	Flavor       string
	HireCost     int
	WeeklyWage   int     // paid every in-game week (1 week = 60 sim minutes)
	DefenseBonus float64 // added to effective lock/cctv defense vs threats
	Specialty    string  // "guard" | "tech" | "social" | "combat" | "sea"
	LoyaltyBase  int     // starting loyalty (0-100)
}

var mercDefs = []MercDef{
	{ID: "tabby_guard", Name: "Tabby Guard", Flavor: "Watches the window all night. Never blinks.",
		HireCost: 200, WeeklyWage: 90, DefenseBonus: 0.15, Specialty: "guard", LoyaltyBase: 65},
	{ID: "siamese_it", Name: "Siamese IT", Flavor: "Keeps the cable spaghetti under control. Allegedly.",
		HireCost: 350, WeeklyWage: 140, DefenseBonus: 0.05, Specialty: "tech", LoyaltyBase: 55},
	{ID: "ragdoll_pr", Name: "Ragdoll PR Cat", Flavor: "Very good with police, neighbors, and social media.",
		HireCost: 300, WeeklyWage: 120, DefenseBonus: 0.05, Specialty: "social", LoyaltyBase: 70},
	{ID: "persian_ex_mil", Name: "Persian Ex-Military", Flavor: "Retired from a private army. Speaks three languages, bites.",
		HireCost: 600, WeeklyWage: 220, DefenseBonus: 0.30, Specialty: "combat", LoyaltyBase: 50},
	{ID: "sphynx_oracle", Name: "Sphynx Oracle", Flavor: "Unsettlingly lucky. Events just… go well.",
		HireCost: 900, WeeklyWage: 320, DefenseBonus: 0.10, Specialty: "lucky", LoyaltyBase: 40},
	{ID: "pirate_cat", Name: "Retired Pirate Cat", Flavor: "Yarr. Keeps the actual pirates away.",
		HireCost: 1100, WeeklyWage: 380, DefenseBonus: 0.40, Specialty: "sea", LoyaltyBase: 35},
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
