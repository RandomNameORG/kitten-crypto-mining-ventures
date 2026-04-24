package data

import "github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"

type AchievementDef struct {
	ID     string
	Emoji  string
	Name   string
	NameZH string
	Desc   string
	DescZH string
}

func (a AchievementDef) LocalName() string { return i18n.Pick(a.Name, a.NameZH) }
func (a AchievementDef) LocalDesc() string { return i18n.Pick(a.Desc, a.DescZH) }

// Ten simple milestones. Checked at end of tick.
var achievements = []AchievementDef{
	{ID: "first_drop", Emoji: "💧",
		Name: "First Drop", NameZH: "第一滴",
		Desc: "Earned your first ₿1.", DescZH: "赚到第一枚币。"},
	{ID: "first_ten_k", Emoji: "💵",
		Name: "Ten Thousand", NameZH: "第一万",
		Desc: "Banked ₿10,000 lifetime.", DescZH: "累计收入 ₿10,000。"},
	{ID: "first_million", Emoji: "💰",
		Name: "First Million", NameZH: "百万喵产",
		Desc: "₿1,000,000 lifetime earnings.", DescZH: "累计收入 ₿1,000,000。"},
	{ID: "first_blueprint", Emoji: "🔬",
		Name: "Silicon Alchemist", NameZH: "硅片炼金师",
		Desc: "Researched your first MEOWCore blueprint.", DescZH: "研究出第一张 MEOWCore 蓝图。"},
	{ID: "first_retire", Emoji: "🎓",
		Name: "Cat-tharsis", NameZH: "喵式退休",
		Desc: "Retired at least once.", DescZH: "至少退休过一次。"},
	{ID: "full_stack", Emoji: "🖥",
		Name: "Full Stack", NameZH: "堆满机架",
		Desc: "Filled every slot in a room.", DescZH: "把一个房间的槽位填满。"},
	{ID: "merc_employer", Emoji: "🐾",
		Name: "Pawsitive Employer", NameZH: "爪上老板",
		Desc: "Hired your first mercenary.", DescZH: "雇佣第一名佣兵。"},
	{ID: "all_rooms", Emoji: "🏠",
		Name: "Real Estate Mogul", NameZH: "地产大亨",
		Desc: "Unlocked every room.", DescZH: "解锁所有房间。"},
	{ID: "polyglot", Emoji: "🌐",
		Name: "Polyglot Purr", NameZH: "多语言喵",
		Desc: "Switched the game language.", DescZH: "切换过游戏语言。"},
	{ID: "hot_cat", Emoji: "🔥",
		Name: "Playing with Fire", NameZH: "玩火",
		Desc: "Pushed a room into the critical heat zone.", DescZH: "让房间进入热量临界区。"},
}

func Achievements() []AchievementDef { return achievements }

func AchievementByID(id string) (AchievementDef, bool) {
	for _, a := range achievements {
		if a.ID == id {
			return a, true
		}
	}
	return AchievementDef{}, false
}
