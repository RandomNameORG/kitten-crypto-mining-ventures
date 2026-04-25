package ui

import "github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"

func (a App) renderViewHint() string {
	var key string
	switch a.view {
	case viewDashboard:
		key = "hint.dashboard"
	case viewStore:
		key = "hint.store"
	case viewGPUs:
		key = "hint.gpus"
	case viewRooms:
		key = "hint.rooms"
	case viewSkills:
		key = "hint.skills"
	case viewMercs:
		key = "hint.mercs"
	case viewLab:
		key = "hint.lab"
	case viewPrestige:
		key = "hint.prestige"
	case viewStats:
		key = "hint.stats"
	case viewMastery:
		key = "hint.mastery"
	default:
		return ""
	}
	return DimStyle.Render(i18n.T(key))
}
