package data

import "github.com/RandomNameORG/kitten-crypto-mining-ventures/packages/core/i18n"

type AchievementDef struct {
	ID     string
	Emoji  string
	Name   string
	NameZH string
	Desc   string
	DescZH string
	// TPReward is a one-shot TechPoint bounty granted the first time the
	// achievement unlocks. Zero means no bonus (preserves the silent unlock
	// behaviour for older saves and any future cosmetic-only milestone).
	TPReward int
}

func (a AchievementDef) LocalName() string { return i18n.Pick(a.Name, a.NameZH) }
func (a AchievementDef) LocalDesc() string { return i18n.Pick(a.Desc, a.DescZH) }

// Ten simple milestones. Checked at end of tick.
//
// TPReward tiers: trivial=1, easy=3, medium=5, hard=10. Total ≈87 TP across
// the catalog — small enough that achievements alone don't trivialise the
// skill tree, big enough that they meaningfully feed the mastery ladder
// over a multi-prestige run.
var achievements = []AchievementDef{
	{ID: "first_drop", Emoji: "💧",
		Name: "First Drop", NameZH: "第一滴",
		Desc: "Earned your first ₿1.", DescZH: "赚到第一枚币。",
		TPReward: 1},
	{ID: "first_ten_k", Emoji: "💵",
		Name: "Ten Thousand", NameZH: "第一万",
		Desc: "Banked ₿10,000 lifetime.", DescZH: "累计收入 ₿10,000。",
		TPReward: 3},
	{ID: "first_million", Emoji: "💰",
		Name: "First Million", NameZH: "百万喵产",
		Desc: "₿1,000,000 lifetime earnings.", DescZH: "累计收入 ₿1,000,000。",
		TPReward: 5},
	{ID: "first_blueprint", Emoji: "🔬",
		Name: "Silicon Alchemist", NameZH: "硅片炼金师",
		Desc: "Researched your first MEOWCore blueprint.", DescZH: "研究出第一张 MEOWCore 蓝图。",
		TPReward: 5},
	{ID: "first_retire", Emoji: "🎓",
		Name: "Cat-tharsis", NameZH: "喵式退休",
		Desc: "Retired at least once.", DescZH: "至少退休过一次。",
		TPReward: 5},
	{ID: "full_stack", Emoji: "🖥",
		Name: "Full Stack", NameZH: "堆满机架",
		Desc: "Filled every slot in a room.", DescZH: "把一个房间的槽位填满。",
		TPReward: 3},
	{ID: "merc_employer", Emoji: "🐾",
		Name: "Pawsitive Employer", NameZH: "爪上老板",
		Desc: "Hired your first mercenary.", DescZH: "雇佣第一名佣兵。",
		TPReward: 3},
	{ID: "all_rooms", Emoji: "🏠",
		Name: "Real Estate Mogul", NameZH: "地产大亨",
		Desc: "Unlocked every room.", DescZH: "解锁所有房间。",
		TPReward: 5},
	{ID: "polyglot", Emoji: "🌐",
		Name: "Polyglot Purr", NameZH: "多语言喵",
		Desc: "Switched the game language.", DescZH: "切换过游戏语言。",
		TPReward: 1},
	{ID: "hot_cat", Emoji: "🔥",
		Name: "Playing with Fire", NameZH: "玩火",
		Desc: "Pushed a room into the critical heat zone.", DescZH: "让房间进入热量临界区。",
		TPReward: 5},
	{ID: "market_timing", Emoji: "📉",
		Name: "Buy the Dip", NameZH: "抄底喵",
		Desc: "Bought a GPU while the market price was below 0.7×.", DescZH: "在市场价格低于 0.7× 时购买 GPU。",
		TPReward: 3},
	{ID: "oc_mastery", Emoji: "🎛",
		Name: "Overclock Master", NameZH: "超频大师",
		Desc: "Accumulated a full hour of overclocked mining.", DescZH: "累计超频运行一小时。",
		TPReward: 5},
	{ID: "tax_survivor", Emoji: "🧾",
		Name: "Books in Order", NameZH: "账目清白",
		Desc: "Survived a tax audit thanks to clean reserves.", DescZH: "靠充足储备扛过一次税务稽查。",
		TPReward: 5},
	{ID: "overdrive", Emoji: "🚀",
		Name: "Overdrive", NameZH: "全员超频",
		Desc: "Had every installed GPU running at max overclock.", DescZH: "所有在机 GPU 同时以最高档超频运行。",
		TPReward: 10},
	{ID: "peak_sell", Emoji: "📈",
		Name: "Sell the Peak", NameZH: "顶部出货",
		Desc: "Sold a GPU while the market price was above 1.5×.", DescZH: "在市场价格高于 1.5× 时出售 GPU。",
		TPReward: 3},
	{ID: "event_veteran", Emoji: "🎟",
		Name: "Event Veteran", NameZH: "事件老手",
		Desc: "Lived through 50 random events.", DescZH: "经历过 50 次随机事件。",
		TPReward: 5},
	{ID: "marathon", Emoji: "⏱",
		Name: "Marathon Miner", NameZH: "马拉松矿工",
		Desc: "Ran the mine for 100,000 virtual seconds.", DescZH: "累计运转矿场 100,000 虚拟秒。",
		TPReward: 10},
	{ID: "crisis_manager", Emoji: "🆘",
		Name: "Crisis Manager", NameZH: "危机处理",
		Desc: "Weathered three market crashes on a single save.", DescZH: "单档内挺过三次市场崩盘。",
		TPReward: 10},
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
