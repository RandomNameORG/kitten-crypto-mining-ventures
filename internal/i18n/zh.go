package i18n

func init() {
	Register(LangZH, zhStrings)
}

var zhStrings = map[string]string{
	// Nav / global chrome.
	"app.title":       "🐾 喵星挖矿",
	"app.pill_paused": " [已暂停]",
	"nav.dashboard":   "主面板",
	"nav.store":       "商店",
	"nav.gpus":        "显卡",
	"nav.rooms":       "房间",
	"nav.skills":      "技能",
	"nav.log":         "日志",
	"nav.mercs":       "佣兵",
	"nav.lab":         "实验室",
	"nav.prestige":    "转生",

	"hdr.tp":    "TP %d",
	"hdr.rep":   "声望 %+d",
	"hdr.frags": "碎片 %d",
	"hdr.price": "$%.0f/BTC",

	"footer.keys": "[空格]暂停  [s]存档  [L]语言  [?]帮助  [q]退出",

	// Welcome / splash.
	"welcome.title":    "🐾 喵星挖矿：加密猫的修罗场",
	"welcome.subtitle": "一款尊重你注意力的增量器",
	"welcome.prompt":   "  给你的小猫工程师起个名字：",
	"welcome.keys":     "  [Enter]开始   [Ctrl+C]退出",
	"welcome.default":  "小橘",

	// Dashboard.
	"dash.location":   "📍 %s",
	"dash.meters":     "⚡ 耗电 %.0f V/s   🌡  %.0f°C   槽位 %d/%d",
	"dash.rack":       "显卡架",
	"dash.empty_hint": "  （空的——按 [2] 去商店）",
	"dash.slot_empty": "  %d. （空）",
	"dash.log_title":  "📜 事件日志",
	"dash.log_quiet":  "  （目前风平浪静）",

	// Status feedback.
	"status.saved":        "💾 已存档",
	"status.save_failed":  "存档失败：%v",
	"status.lang":         "🌐 语言：%s",
	"status.order":        "📦 已下单 %s",
	"status.upgrade":      "⚙️  升级尝试中",
	"status.repaired":     "🔧 已修好",
	"status.sold":         "💵 已出售",
	"status.unlocked":     "🔓 %s 已解锁",
	"status.now_in":       "📍 当前在 %s",
	"status.defense_up":   "🛡 %s 已升级",
	"status.hired":        "🐾 雇佣了 %s",
	"status.dismissed":    "已解雇",
	"status.bribed":       "🎁 忠诚度提升",
	"status.research_go":  "🔬 研究开始",
	"status.printed":      "🛠 已打印 MEOWCore",
	"status.perk_bought":  "🎁 特权已购买",
	"status.retired":      "🐾 你退休了。+%d LP 入账。新周目开始。",
	"status.retire_arm":   "⚠ 5 秒内再按一次 [R] 确认退休",
	"status.retire_deny":  "❌ 还不够资格退休",
	"status.pump_fired":   "📈 拉盘已启动",
	"status.error_prefix": "❌ ",

	// Store.
	"store.title": "🛒 商店  ·  快递：约 30–180 秒",
	"store.help":  "↑/↓ 选择   [b] 购买   [esc]/[1] 返回",

	// GPUs view.
	"gpus.title": "🖥  我的显卡",
	"gpus.help":  "↑/↓ 选择   [u] 升级   [r] 维修   [s] 拆解   [esc]/[1] 返回",
	"gpus.empty": "  （还没有显卡，去商店看看）",

	// Rooms view.
	"rooms.title":      "🏠 房间",
	"rooms.help":       "↑/↓ 选房间   [u] 解锁   [enter] 切换   [l/c/w/o/a] 升级当前房间的防御   [esc]/[1] 返回",
	"rooms.here":       "● 此处",
	"rooms.unlocked":   "已解锁",
	"rooms.to_unlock":  "解锁 $%d",
	"rooms.stats":      "  散热 %.1f · 电费 ×%.2f · 基础威胁 %.2f",
	"rooms.defense":    "🛡  防御 —— 当前房间（%s）",
	"rooms.dim.lock":    "门锁",
	"rooms.dim.cctv":    "监控",
	"rooms.dim.wiring":  "电路",
	"rooms.dim.cooling": "散热",
	"rooms.dim.armor":   "护甲",

	// Skills.
	"skills.title":         "🧠 技能树",
	"skills.tp_count":      "TP：%d",
	"skills.help":          "↑/↓ 选择   [u]/[enter] 解锁   [esc]/[1] 返回",
	"skills.lane.engineer": "🔧 工程师",
	"skills.lane.mogul":    "💰 大亨",
	"skills.lane.hacker":   "🕶 黑客",
	"skills.owned":         "已有",
	"skills.locked_suffix": "（未解锁前置）",

	// Log.
	"log.title": "📜 完整事件日志",
	"log.help":  "[esc]/[1] 返回",
	"log.empty": "  （空）",

	// Help.
	"help.title":       "🐾 帮助",
	"help.views":       "视图",
	"help.view.1":      "[1]  主面板 —— 显卡架 + 实时事件日志",
	"help.view.2":      "[2]  商店 —— 买新显卡（有快递延迟）",
	"help.view.3":      "[3]  我的显卡 —— 升级·维修·拆解",
	"help.view.4":      "[4]  房间 —— 解锁·切换·升级防御",
	"help.view.5":      "[5]  技能 —— 花 TechPoint",
	"help.view.6":      "[6]  日志 —— 完整历史",
	"help.view.7":      "[7]  佣兵 —— 雇佣·解雇·贿赂",
	"help.view.8":      "[8]  实验室 —— 研究自制 MEOWCore 显卡",
	"help.view.9":      "[9]  转生 —— 退休 + 购买 Legacy 特权",
	"help.global":      "全局",
	"help.g.space":     "[空格]   暂停 / 继续",
	"help.g.save":      "[s]      存档（仅主面板——其他视图里 s 有其他含义）",
	"help.g.pump":      "[p]      拉盘技能（主面板，需解锁）",
	"help.g.lang":      "[L]      循环切换语言",
	"help.g.quit":      "[q]      退出（自动存档）",
	"help.defense":     "房间防御（在房间视图里）",
	"help.defense_row": "[l] 锁 · [c] 监控 · [w] 电路 · [o] 散热 · [a] 护甲",
	"help.tip.idle":    "提示：这是增量器——放心开着 tmux 挂后台。",
	"help.tip.offline": "离线进度在重启时追赶（上限 8 小时）。",

	// Mercs.
	"mercs.title":      "🐾 佣兵",
	"mercs.help":       "[tab] 切换标签   ↑/↓ 选择   [h] 雇佣   [f] 解雇   [b] 贿赂（忠诚 +15，$200）   [esc]/[1] 返回",
	"mercs.yours":      "你的佣兵",
	"mercs.empty":      "  （没有——切到雇佣标签）",
	"mercs.hire":       "雇佣",
	"mercs.owned_line": "房间 %s  周薪 $%d  忠诚 %d",
	"mercs.hire_line":  "雇佣 $%d",
	"mercs.defbonus":   "防御 +%.0f%%",
	"mercs.loyalty":    "忠诚 %d",

	// Lab.
	"lab.title":       "🔬 实验室 —— 自研 MEOWCore",
	"lab.locked":      "实验室未解锁。先去工程师技能里解锁「MEOWCore Blueprint」。",
	"lab.help":        "[t] 切档位   [b] 切加成组合   [r] 开始研究   [↑/↓] 选蓝图   [p] 打印   [esc]/[1] 返回",
	"lab.active":      "当前研究",
	"lab.active_none": "  （无）",
	"lab.plan":        "下一次研究计划",
	"lab.plan_tier":   "  档位 %d —— %s",
	"lab.plan_cost":   "  消耗：$%d + %d 碎片  ·  用时：%d 分钟",
	"lab.plan_boosts": "  加成：%s + %s",
	"lab.plan_hint":   "  （按 [r] 开始研究）",
	"lab.bp_title":    "蓝图（%d 个）—— 按 [p] 打印选中的",
	"lab.bp_empty":    "  （还没研究出来）",

	// Prestige.
	"prestige.title":        "🎓 转生 —— 退休 + 重开",
	"prestige.locked":       "转生未解锁。先去大亨技能里解锁「Venture Capital」。",
	"prestige.help":         "[↑/↓] 选特权   [p] 购买   [R] 退休（按两次确认）   [esc]/[1] 返回",
	"prestige.status":       "状态",
	"prestige.lifetime":     "  累计收入：$%.0f / $%.0f",
	"prestige.eligible_yes": "可以转生",
	"prestige.eligible_no":  "暂时还不行",
	"prestige.eligible_row": "  退休资格：%s",
	"prestige.reward":       "  退休奖励：%d LP",
	"prestige.bank":         "  LP 存款：总计 %d · 已花 %d · 可用 %d",
	"prestige.perks":        "Legacy 特权",
	"prestige.perk_owned":   "已拥有 / 满级",

	// Small labels used across multiple views.
	"label.eff":     "效率 %.4f ₿/秒",
	"label.bp_line": "    效率 %.4f ₿/秒 · %.0fV · %.0f°C · %.0fh 耐久",

	// Event popup prompts.
	"event.dismiss": "[按任意键关闭]",

	// Game log messages.
	"game.welcome":       "欢迎，%s。第一张显卡已上线。",
	"game.named":         "小猫命名：%s。",
	"game.paused":        "⏸  已暂停。",
	"game.resumed":       "▶️  继续运行。",
	"game.lang_switched": "🌐 语言切换为 %s。",

	"warn.terminal_too_small": "终端太小了，请至少调到 80x22。",
}
