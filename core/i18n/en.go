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
	"dash.line.volt":  "⚡ %.0fV draw  ·  bill −%s/s  ·  next bill %ds",
	"dash.line.heat":  "🌡 %.0f°C / %.0f max  ·  %+.1f°C every %ds  ·  next in %ds",
	"dash.line.cash":  "📈 earn +%s/s  ·  net %s/s",
	"dash.slots_of":   "slots %d/%d",
	"dash.heat.warning":  "\uf071 HOT — efficiency ½ · wear 3×",
	"dash.heat.critical": "\uf06d CRITICAL — wear 8× · GPU failure imminent",
	"dash.impact.stable":   "stable · eff 100% · wear 1×",
	"dash.impact.warm":     "warm · eff 100% · wear 1×",
	"dash.impact.hot":      "HOT · eff ½ · wear 3×",
	"dash.impact.critical": "CRITICAL · wear 8× · failure imminent",
	"dash.power.safe":      "safe",
	"dash.power.deficit":   "losing ₿/s — balance lasts %s",
	"dash.power.broke":     "empty wallet → 60s blackout",
	"dash.market.label":    "📊 %.2f× %s",
	"dash.line.power":      "\uf0e7 %.0fV  −%s/s  (next bill %ds)",
	"dash.line.cash2":      "\uf201 +%s/s earn   net %s/s",
	"dash.heat.label":      "\uf2c7 Heat  %.0f/%.0f°C  %+.1f/%ds",
	"dash.rack":           "GPU Rack",
	"dash.empty_hint":     "  (empty — press [2] to go to the store)",
	"dash.slot_empty":     "  %d. (empty)",
	"dash.slot_reserved":  "  %d. (reserved — inbound)",
	"dash.delivery_title": "📦 Delivery Lane",
	"dash.delivery_line":  "  %s  %s  ETA %ds",
	"dash.log_title":      "📜 Event Log",
	"dash.log_quiet":      "  (quiet so far)",
	"dash.sidebar.keys":   "Keys",
	"dash.sidebar.pause":  "pause",
	"dash.sidebar.vent":   "vent",
	"dash.sidebar.pump":   "pump&dump",
	"dash.sidebar.save":   "save",
	"dash.sidebar.help":   "help",

	// Status feedback.
	"status.saved":        "💾 saved",
	"status.save_failed":  "save failed: %v",
	"status.lang":         "\uf0ac language: %s",
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
	"rooms.to_unlock":  "%s to unlock",
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
	"help.g.vent":     "[V]       emergency vent — reset room heat · %s · 30s pause · 2m cooldown",
	"help.defense":    "Room defense (from rooms view)",
	"help.defense_row": "[l] lock · [c] CCTV · [w] wiring · [o] cooling · [a] armor",
	"help.tip.idle":    "Tip: it's an incremental game — feel free to leave it running in tmux.",
	"help.tip.offline": "Offline progress catches up on relaunch (capped at 8h).",

	"help.mechanics":     "Game mechanics",
	"help.mech.heat":     "\uf2c7 Heat — GPUs produce it; rooms have a max ceiling.",
	"help.mech.heat.z1":  "    0–80%   stable · no penalty",
	"help.mech.heat.z2":  "   80–95%   HOT · mining efficiency halved, wear ×3",
	"help.mech.heat.z3":  "  95–100%   CRITICAL · wear ×8, GPUs will break",
	"help.mech.heat.act": "  → buy cooling in [4] rooms, or hit [V] for emergency vent.",
	"help.mech.power":    "\uf0e7 Power — each GPU draws volts; bill settles every 60s.",
	"help.mech.power.2":  "  Bill = voltage × room multiplier × difficulty. Rent adds a fixed ₿/hour.",
	"help.mech.power.3":  "  If balance hits ₿0 mid-bill: 60-second blackout — no earnings.",
	"help.mech.power.act": "  → watch \"net\" on the dashboard. Negative = bleeding.",
	"help.mech.earn":     "\uf201 Earnings — GPUs mint ₿ per tick, added to your balance.",
	"help.mech.earn.2":   "  MEOWCore blueprints (from [8] lab) unlock custom stat trade-offs.",

	// Mercs.
	"mercs.title":     "🐾 Mercenaries",
	"mercs.help":      "[tab] switch tab   ↑/↓ select   [h] hire   [f] fire   [b] bribe (+15 loyalty, %s)   [esc]/[1] back",
	"mercs.yours":     "Your Mercs",
	"mercs.empty":     "  (none — switch to Hire tab)",
	"mercs.hire":      "Hire",
	"mercs.owned_line": "room %s  wage %s/wk  loyalty %d",
	"mercs.hire_line":  "hire %s",
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
	"lab.plan_cost":   "  costs: %s + %d frags  ·  duration: %dm",
	"lab.plan_boosts": "  boosts: %s + %s",
	"lab.plan_hint":   "  (press [r] to start)",
	"lab.bp_title":    "Blueprints (%d) — [p] to print selected",
	"lab.bp_empty":    "  (none researched yet)",

	// Prestige.
	"prestige.title":   "🎓 Prestige — Retire & Restart",
	"prestige.locked":  "Prestige is locked. Unlock 'Venture Capital' in the Mogul skill lane.",
	"prestige.help":    "[↑/↓] select perk   [p] buy perk   [R] RETIRE (press twice to confirm)   [esc]/[1] back",
	"prestige.status":  "Status",
	"prestige.lifetime": "  lifetime earned: %s / %s",
	"prestige.eligible_yes": "ELIGIBLE",
	"prestige.eligible_no":  "not eligible",
	"prestige.eligible_row": "  retirement status: %s",
	"prestige.reward": "  retire reward: %d LP",
	"prestige.bank":   "  bank balance: %d LP total · %d spent · %d available",
	"prestige.perks":  "Legacy Perks",
	"prestige.perk_owned": "owned / maxed",

	// Small labels used across multiple views.
	"label.eff":       "%s/s",
	"label.bp_line":   "    eff %s/s · %.0fV · %.0f°C · %.0fh durability",

	// Event popup prompts.
	"event.dismiss": "[press any key to dismiss]",

	// Offline catch-up welcome-back notification.
	"offline.title":   "😽 Welcome back!",
	"offline.body":    "You were away for %s.\nEarned %s while the rack kept humming.",
	"offline.capped":  "(capped at 8h — longer gaps stop accruing.)",
	"offline.dismiss": "[press any key to dismiss]",

	// Game log messages (rendered from game layer).
	"game.welcome":        "Welcome, %s. Your first GPU hums to life.",
	"game.named":          "Named kitten: %s.",
	"game.paused":         "\uf04c  Paused.",
	"game.resumed":        "\uf04b  Resumed.",
	"game.lang_switched":  "\uf0ac Language set to %s.",
	"game.difficulty_set": "Difficulty locked: %s.",
	"game.achievement":    "\uf091 Achievement unlocked — %s",

	// Log strings emitted from the game layer. Glyphs are Nerd Font
	// codepoints — see ui/icons.go for the semantic mapping.
	"log.research.started":      "\uf0c3 Research started: %s (%ds).",
	"log.research.breakthrough": "\uf06b Research had a breakthrough — extra boost included.",
	"log.research.complete":     "\uf00c Research complete: %s blueprint [%s] ready.",
	"log.research.printed":      "\uf0ad Printed a %s (%s).",

	"log.skill.learned": "\uf0eb Learned: %s.",

	"log.merc.hired":              "\uf1b0 Hired %s for %s.",
	"log.merc.dismissed":          "Dismissed %s. The other mercs noticed.",
	"log.merc.bribed":             "\uf06b Bribed %s — loyalty now %d.",
	"log.merc.wages":              "\uf0b1 Paid %s in mercenary wages.",
	"log.merc.betray.unlock":      "\uf1b0 %s unlocked the door on the way out.",
	"log.merc.betray.sabotage":    "\uf1e2 %s sabotaged something on the way out.",
	"log.merc.betray.sold_story":  "\uf0a1 %s sold your story to a rival kitten. Rep −10.",
	"log.merc.betray.stole_gpu":   "\uf140 %s made off with your best GPU.",
	"log.merc.betray.pirate_crew": "\uf091 %s invited their old crew back. Brace yourself.",
	"log.merc.betray.generic":     "%s betrayed you and left with some gear.",

	"log.event.chain_ghost":    "%s A threat was averted silently by Chain Ghost.",
	"log.event.tp_gained":      "\uf0eb +%d TechPoint.",
	"log.event.gift.installed": "\uf06b Free %s installed.",
	"log.event.gift.sold":      "…but no room. Sold %s.",
	"log.event.fire.averted":   "\uf132 Armor held. Nothing caught fire.",
	"log.event.fire.warning":   "You've been warned. One more incident and the room is gone.",
	"log.event.fire.money":     "\uf155 Lost %s (%.0f%% of balance).",
	"log.event.fire.destroyed": "\uf06d Fire! %d GPUs destroyed in %s.",
	"log.event.thief.empty":    "Thief found nothing worth taking. Huh.",
	"log.event.thief.defended": "\uf132 Defense held. Nothing stolen.",
	"log.event.thief.took_gpu": "\uf21b They took your %s. Gone.",
	"log.event.thief.took_bp":  "\uf21b They took one of your MEOWCores. Devastating.",
	"log.event.gpu.broken":     "\uf1e2 A GPU took too much damage — broken.",
	"log.event.gpu.damaged":    "\uf071 A GPU is damaged.",
	"log.event.repair.free":    "\uf0ad PCB surgery — free repair.",
	"log.event.repair.paid":    "\uf0ad Repaired for %s.",

	"log.gpu.arrived":         "\uf1b2 %s arrived and is online.",
	"log.gpu.failed":          "\uf1e2 %s failed. It needs repair or scrapping.",
	"log.gpu.upgrade.success": "\uf013 GPU upgraded to level %d.",
	"log.gpu.upgrade.bricked": "\uf06d Upgrade failed — GPU is bricked.",
	"log.gpu.ordered":         "Ordered %s for %s. Tracking inbound…",
	"log.gpu.scrapped":        "Scrapped %s for %s + %d research fragments.",

	"log.bills.settled":  "\uf155 Bills settled: %s electricity, %s rent.",
	"log.bills.blackout": "\uf1e6 Couldn't pay the bill. Blackout for 60s.",

	"log.room.vent":        "\uf2dc Emergency vent — heat reset, 30s power cycle, -%s.",
	"log.room.moved":       "Moved into %s.",
	"log.defense.upgraded": "\uf132 %s upgraded to level %d.",

	"defense.lock":    "Lock",
	"defense.cctv":    "CCTV",
	"defense.wiring":  "Wiring",
	"defense.cooling": "Cooling",
	"defense.armor":   "Armor",

	"log.pump.fired": "\uf201 Pump & Dump — BTC price ×1.5 for 5 minutes.",

	"log.prestige.retired":  "\uf1b0 You retired rich. +%d LegacyPoints banked.",
	"log.legacy.cash":       "Legacy bonus: +%s starter balance.",
	"log.legacy.room":       "Legacy bonus: University Server Room pre-unlocked.",
	"log.legacy.blueprints": "Legacy bonus: %d blueprints carried over.",
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
