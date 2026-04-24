package i18n

func init() {
	Register(LangEN, enStrings)
}

var enStrings = map[string]string{
	// Nav / global chrome.
	"app.title":       "🐾 Kitten Crypto Mining",
	"app.pill_paused": " [PAUSED]",
	"nav.dashboard":   "dashboard",
	"nav.store":       "store",
	"nav.gpus":        "gpus",
	"nav.rooms":       "rooms",
	"nav.skills":      "skills",
	"nav.log":         "log",
	"nav.mercs":       "mercs",
	"nav.lab":         "lab",
	"nav.prestige":    "prestige",

	"hdr.tp":    "TP %d",
	"hdr.rep":   "Rep %+d",
	"hdr.frags": "frags %d",

	"footer.keys": "[space] pause  [S] save  [L] lang  [?] help  [q] quit",

	// Welcome / splash.
	"welcome.title":    "🐾 Kitten Crypto Mining Ventures",
	"welcome.subtitle": "an incremental game that respects your attention",
	"welcome.prompt":   "  Name your kitten engineer: ",
	"welcome.keys":     "  [enter] start   [ctrl+c] quit",
	"welcome.default":  "Whiskers",

	// Difficulty splash.
	"splash.diff.title":    "Choose your difficulty",
	"splash.diff.subtitle": "Good luck, %s. This choice is locked for the run.",
	"splash.diff.help":     "[↑/↓] select   [enter] commit (permanent)   [ctrl+c] quit",

	// Dashboard.
	"dash.location":   "📍 %s",
	"dash.line.volt":  "⚡ %.0fV draw  ·  bill −₿%.3f/s  ·  next bill %ds",
	"dash.line.heat":  "🌡 %.0f°C / %.0f max  ·  %+.1f°C every %ds  ·  next in %ds",
	"dash.line.cash":  "📈 earn +₿%.3f/s  ·  net %+.3f/s",
	"dash.slots_of":   "slots %d/%d",
	"dash.heat.warning":  "⚠ HOT — efficiency ½ · wear 3×",
	"dash.heat.critical": "🔥 CRITICAL — wear 8× · GPU failure imminent",
	"dash.rack":           "GPU Rack",
	"dash.empty_hint":     "  (empty — press [2] to go to the store)",
	"dash.slot_empty":     "  %d. (empty)",
	"dash.slot_reserved":  "  %d. (reserved — inbound)",
	"dash.delivery_title": "📦 Delivery Lane",
	"dash.delivery_line":  "  %s  %s  ETA %ds",
	"dash.log_title":      "📜 Event Log",
	"dash.log_quiet":      "  (quiet so far)",

	// Status feedback.
	"status.saved":        "💾 saved",
	"status.save_failed":  "save failed: %v",
	"status.lang":         "🌐 language: %s",
	"status.order":        "📦 Ordered %s",
	"status.upgrade":      "⚙️  upgrade attempted",
	"status.repaired":     "🔧 repaired",
	"status.sold":         "💵 sold",
	"status.unlocked":     "🔓 %s unlocked",
	"status.now_in":       "📍 now in %s",
	"status.defense_up":   "🛡 %s upgraded",
	"status.hired":        "🐾 hired %s",
	"status.dismissed":    "dismissed",
	"status.bribed":       "🎁 loyalty boosted",
	"status.research_go":  "🔬 research started",
	"status.printed":      "🛠 printed MEOWCore",
	"status.perk_bought":  "🎁 perk purchased",
	"status.retired":      "🐾 retired. +%d LP banked. New run begins.",
	"status.retire_arm":   "⚠ press [R] again within 5s to confirm retirement",
	"status.retire_deny":  "❌ not eligible to retire yet",
	"status.pump_fired":   "📈 Pump & Dump fired",
	"status.vent":         "🧊 Emergency vent fired",
	"status.error_prefix": "❌ ",

	// Store.
	"store.title": "🛒 Store  ·  Shipping: ~30–180s",
	"store.help":  "↑/↓ select   [b] buy   [esc]/[1] back",

	// GPUs view.
	"gpus.title": "🖥  Your GPUs",
	"gpus.help":  "↑/↓ select   [u] upgrade   [r] repair   [s] scrap   [esc]/[1] back",
	"gpus.empty": "  (no GPUs yet — visit the store)",

	// Rooms view.
	"rooms.title":      "🏠 Rooms",
	"rooms.help":       "↑/↓ room   [u] unlock   [enter] switch   [l/c/w/o/a] upgrade defense on current room   [esc]/[1] back",
	"rooms.here":       "● here",
	"rooms.unlocked":   "unlocked",
	"rooms.to_unlock":  "₿%d to unlock",
	"rooms.stats":      "  cooling %.1f · elec ×%.2f · threat base %.2f",
	"rooms.defense":    "🛡  Defense — current room (%s)",
	"rooms.dim.lock":    "Lock",
	"rooms.dim.cctv":    "CCTV",
	"rooms.dim.wiring":  "Wiring",
	"rooms.dim.cooling": "Cooling",
	"rooms.dim.armor":   "Armor",

	// Skills.
	"skills.title":        "🧠 Skill Tree",
	"skills.tp_count":     "TP: %d",
	"skills.help":         "↑/↓ select   [u]/[enter] unlock   [esc]/[1] back",
	"skills.lane.engineer": "🔧 Engineer",
	"skills.lane.mogul":    "💰 Mogul",
	"skills.lane.hacker":   "🕶 Hacker",
	"skills.owned":         "owned",
	"skills.locked_suffix": " (locked)",

	// Log.
	"log.title": "📜 Full Event Log",
	"log.help":  "[esc]/[1] back",
	"log.empty": "  (empty)",

	// Help.
	"help.title":      "🐾 Help",
	"help.views":      "Views",
	"help.view.1":     "[1]  dashboard — GPU rack + live event log",
	"help.view.2":     "[2]  store — buy new GPUs (shipping delay)",
	"help.view.3":     "[3]  your GPUs — upgrade · repair · scrap",
	"help.view.4":     "[4]  rooms — unlock · switch · defense upgrades",
	"help.view.5":     "[5]  skills — spend TechPoints",
	"help.view.6":     "[6]  log — full history",
	"help.view.7":     "[7]  mercs — hire · fire · bribe",
	"help.view.8":     "[8]  lab — research custom MEOWCore GPUs",
	"help.view.9":     "[9]  prestige — retire & buy legacy perks",
	"help.global":     "Global",
	"help.g.space":    "[space]  pause / resume",
	"help.g.save":     "[S]       save (any view)",
	"help.g.pump":     "[p]       Pump & Dump ability (dashboard, if unlocked)",
	"help.g.lang":     "[L]       cycle language",
	"help.g.quit":     "[q]       quit (auto-saves)",
	"help.g.vent":     "[V]       emergency vent — reset room heat · -₿100 · 30s pause · 2m cooldown",
	"help.defense":    "Room defense (from rooms view)",
	"help.defense_row": "[l] lock · [c] CCTV · [w] wiring · [o] cooling · [a] armor",
	"help.tip.idle":    "Tip: it's an incremental game — feel free to leave it running in tmux.",
	"help.tip.offline": "Offline progress catches up on relaunch (capped at 8h).",

	// Mercs.
	"mercs.title":     "🐾 Mercenaries",
	"mercs.help":      "[tab] switch tab   ↑/↓ select   [h] hire   [f] fire   [b] bribe (+15 loyalty, ₿200)   [esc]/[1] back",
	"mercs.yours":     "Your Mercs",
	"mercs.empty":     "  (none — switch to Hire tab)",
	"mercs.hire":      "Hire",
	"mercs.owned_line": "room %s  wage ₿%d/wk  loyalty %d",
	"mercs.hire_line":  "hire ₿%d",
	"mercs.defbonus":   "def +%.0f%%",
	"mercs.loyalty":    "loyalty %d",

	// Lab.
	"lab.title":       "🔬 Lab — Custom MEOWCore Research",
	"lab.locked":      "R&D is locked. Unlock 'MEOWCore Blueprint' in the Engineer skill lane first.",
	"lab.help":        "[t] cycle tier   [b] cycle boost combo   [r] start research   [↑/↓] select blueprint   [p] print   [esc]/[1] back",
	"lab.active":      "Active research",
	"lab.active_none": "  (none)",
	"lab.plan":        "Plan next research",
	"lab.plan_tier":   "  Tier %d — %s",
	"lab.plan_cost":   "  costs: ₿%d + %d frags  ·  duration: %dm",
	"lab.plan_boosts": "  boosts: %s + %s",
	"lab.plan_hint":   "  (press [r] to start)",
	"lab.bp_title":    "Blueprints (%d) — [p] to print selected",
	"lab.bp_empty":    "  (none researched yet)",

	// Prestige.
	"prestige.title":   "🎓 Prestige — Retire & Restart",
	"prestige.locked":  "Prestige is locked. Unlock 'Venture Capital' in the Mogul skill lane.",
	"prestige.help":    "[↑/↓] select perk   [p] buy perk   [R] RETIRE (press twice to confirm)   [esc]/[1] back",
	"prestige.status":  "Status",
	"prestige.lifetime": "  lifetime earned: ₿%.0f / ₿%.0f",
	"prestige.eligible_yes": "ELIGIBLE",
	"prestige.eligible_no":  "not eligible",
	"prestige.eligible_row": "  retirement status: %s",
	"prestige.reward": "  retire reward: %d LP",
	"prestige.bank":   "  bank balance: %d LP total · %d spent · %d available",
	"prestige.perks":  "Legacy Perks",
	"prestige.perk_owned": "owned / maxed",

	// Small labels used across multiple views.
	"label.eff":       "eff %.4f ₿/s",
	"label.bp_line":   "    eff %.4f ₿/s · %.0fV · %.0f°C · %.0fh durability",

	// Event popup prompts.
	"event.dismiss": "[press any key to dismiss]",

	// Game log messages (rendered from game layer).
	"game.welcome":       "Welcome, %s. Your first GPU hums to life.",
	"game.named":         "Named kitten: %s.",
	"game.paused":        "⏸  Paused.",
	"game.resumed":       "▶️  Resumed.",
	"game.lang_switched": "🌐 Language set to %s.",
	"game.difficulty_set": "Difficulty locked: %s.",
	"game.achievement":    "🏆 Achievement unlocked — %s",
	"hdr.achievements":    "🏆 %d/%d",

	// Minimum terminal size warning.
	"warn.terminal_too_small": "Please widen your terminal to at least 80x22.",

	// Update-available splash — shown on startup when GitHub reports a
	// newer release than the running binary. See core/update and
	// ui/update_splash.go.
	"update.title":     "🆕 Update available",
	"update.from_to":   "Current: %s   →   Latest: %s",
	"update.changelog": "What's new:",
	"update.no_notes":  "(no changelog)",
	"update.opt.yes":   "Open release page",
	"update.opt.no":    "Remind me next time",
	"update.opt.skip":  "Skip this version",
	"update.help":      "[↑/↓] select  [enter] confirm  [y] open  [n] later  [s] skip  [o] release notes  [ctrl+c] quit",
	"update.opening":   "Opening %s ...",
}
