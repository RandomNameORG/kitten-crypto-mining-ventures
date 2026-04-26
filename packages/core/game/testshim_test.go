package game

import "github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/data"

// eventShim / effectShim are test-only convenience wrappers that let tests
// construct synthetic events without touching the JSON catalogs.
type effectShim struct {
	Kind              string
	Seconds           int
	Delta             int
	Factor            float64
	Tier              string
	Amount            float64
	Chance            float64
	ChanceIfNoDefense float64
	Count             int
	ReserveFactor     float64
}

type eventShim struct {
	Category string
	Emoji    string
	Name     string
	Effects  []effectShim
}

func (e eventShim) toDef() data.EventDef {
	out := data.EventDef{
		ID:       "shim",
		Name:     e.Name,
		Category: e.Category,
		Emoji:    e.Emoji,
		Text:     "(shim)",
		Weight:   1,
	}
	for _, eff := range e.Effects {
		out.Effects = append(out.Effects, data.EventEffect{
			Kind:              eff.Kind,
			Seconds:           eff.Seconds,
			Delta:             eff.Delta,
			Factor:            eff.Factor,
			Tier:              eff.Tier,
			Amount:            eff.Amount,
			Chance:            eff.Chance,
			ChanceIfNoDefense: eff.ChanceIfNoDefense,
			Count:             eff.Count,
			ReserveFactor:     eff.ReserveFactor,
		})
	}
	return out
}
