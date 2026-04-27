# Web API audit — `meowmine-web` vs engine vs design v2.1

Snapshot date: 2026-04-26.
Surfaces audited: `packages/web/cmd/meowmine-web/main.go` (`/api/snapshot`,
`/api/action`), `packages/core/game/state.go` and siblings, `docs/GAME_DESIGN.md`
(v2.1).

The audit is split into three sections:

1. Things the **design doc** wants but the **engine** doesn't have yet — out of
   scope for this sprint, listed so we don't accidentally try to expose them.
2. Things the **engine** already exposes on `*State` but the **web snapshot**
   doesn't surface — the work this sprint actually does.
3. Things the **engine** already exposes as `*State` methods but `/api/action`
   doesn't dispatch — the action handlers added this sprint.

Sections 2 & 3 drive the implementation in `main.go` + `types.ts`. Section 1 is
backlog.

---

## 1. v2.1 design requires that the engine doesn't have yet

**Engine-level work, deferred. Do NOT implement in this sprint.** Listed only
so that any web-facing field for these systems has to wait until the engine
side lands. Where the design lives in `GAME_DESIGN.md` is noted in
parentheses.

- **PSU system** — `PSUDef`, per-room `psu_units`, rated power constraint,
  efficiency (75–95%), `heat_output`, overload tolerance, replace flow,
  E21/E22/E26 events. (§4)
- **Mining pool system** — `pool_id` on the run, the catalog (Solo,
  ScratchPool, KittenHash, DarkClaw, WhiskerFi), per-pool fee + settlement
  mode (PPS / PPLNS / PPS+), 10-minute switch transition, share accumulator,
  pool-runaway/dilution/Solo-block events (E24, E25, E27). (§5)
- **Stale Share rate** — per-room network-quality field, applied as
  `effective_eff = card.eff × (1 - stale_rate)`. (§7.2)
- **Gas Fee** — per-cashout fee `base_gas × (1 + network_congestion)` plus
  flat per-cashout floor; `Gas Optimizer` skill. (§11.2)
- **BTC-linked GPU resale price** — `base_resale_ratio` + `btc_sensitivity`
  per `GPUDef`, dynamic resale formula, scrap/buy price both tied to
  `current_btc / 30000`. (§8.1–§8.3)
- **Net Worth panel** — `cash + btc_held × btc_price + Σ(card.resale) +
  Σ(psu.resale × 0.7) − debt`. Drives the future Prestige threshold. (§8.4)
- **Newton cooling temperature equilibrium** — `equilibrium = ambient +
  max(0, total_heat - cooling) / dissipation`, `temp += (eq - temp) ×
  approach_speed`. Today the engine uses a simpler heat-tick model in
  `RoomState.Heat` / `RoomHeatDeltaPerTick`. Fields on `RoomDef` for
  `ambient`, `dissipation`, and `approach_speed` don't exist yet. (§3)
- **New events E21–E28** — PSU explode, smoking PSU, mining-disaster sell-off,
  pool runaway, Solo block hit, PSU chain explode, share dilution, miner
  fire-sale. None are in `data.Events()`. (§10.1)
- **New skill nodes** — `Wiring Optimization`, `Pool Hopping`,
  `Asset Hedging`, `Network Optimization`, `Pool Infiltration`. Not in
  `data.Skills()`. (§9)

Until the above ship in `packages/core/game/`, the snapshot must not invent
zero-valued placeholders for them — the audit explicitly excludes that.

---

## 2. Engine already has it, snapshot doesn't expose it

Each row: the engine field/method, where to find it, and what we're surfacing
on the snapshot. Field naming follows the existing snake_case JSON style
already used by `stateView` / `gpuView`.

### 2.1 Top-level on `stateView`

| Field on `*State`                              | Source                                                | Snapshot key (proposed)              |
| ---------------------------------------------- | ----------------------------------------------------- | ------------------------------------ |
| `LifetimeEarned` (raw float)                   | `state.go:132`                                         | `state.lifetime_earned`              |
| `LegacyAvailable` (int)                        | `state.go:134`                                         | `state.legacy_available`             |
| `Difficulty` (string)                          | `state.go:163`, `Diff()` `state.go:692`                | `state.difficulty`                   |
| `Lang` (string)                                | `state.go:157`                                         | `state.lang`                         |
| `PrevMarketPrice` (float64)                    | `state.go:128`                                         | `state.prev_market_price`            |

`LifetimeEarnedFmt`, `MarketPrice`, `MarketTrend()`, `SyndicateJoined`,
`Paused`, `MiningPaused` are already exposed; the new fields above sit
next to them.

### 2.2 Syndicate (extend existing `state.syndicate_joined`)

| Engine                                                                                    | Snapshot key                       |
| ----------------------------------------------------------------------------------------- | ---------------------------------- |
| `CanJoinSyndicate()` — `syndicate.go:48`                                                   | `state.syndicate_can_join`         |
| `SyndicateContribution` — `state.go:152`                                                   | `state.syndicate_contribution`     |
| `SyndicateTotalDividends` — `state.go:153`                                                 | `state.syndicate_total_dividends`  |
| `SecondsUntilNextSyndicatePayout()` — `syndicate.go:122`                                   | `state.syndicate_next_payout_sec`  |

### 2.3 GPU (extend existing `gpuView`)

| Engine field on `*GPU`                | Snapshot key on each gpu          |
| ------------------------------------- | --------------------------------- |
| `BlueprintID` — `state.go:26`         | `blueprint_id` (`omitempty`)      |

`BlueprintID` is non-empty only for printed MEOWCore instances; the existing
naming logic at `main.go:357–361` already special-cases this — exposing the
ID lets the future UI link a card back to its blueprint.

### 2.4 Modifiers — new top-level `modifiers: []modifierView`

`Modifier` shape lives at `state.go:76`. `Kind` is currently `"earn_mult"` or
`"pause_mining"`. Snapshot view per modifier:

| Field             | Source                                                  |
| ----------------- | ------------------------------------------------------- |
| `kind`            | `Modifier.Kind`                                          |
| `factor`          | `Modifier.Factor`                                       |
| `expires_at`      | `Modifier.ExpiresAt`                                    |
| `seconds_left`    | `expires_at - time.Now().Unix()`, clamped ≥ 0           |

Active list is bounded by `pruneModifiers` at `economy.go:39` — already
called inside `Tick`, so the snapshot just iterates `s.Modifiers`.

### 2.5 Active research — new top-level `active_research: researchView | null`

Engine: `Research` struct at `state.go:68`, populated by `StartResearch`
(`research.go:33`), finalised by `advanceResearch` (`research.go:89`).
Progress helper `ResearchProgress()` at `research.go:77`.

| Field           | Source                                                                  |
| --------------- | ----------------------------------------------------------------------- |
| `tier`          | `ActiveResearch.BlueprintTier`                                           |
| `boosts`        | `ActiveResearch.Boosts`                                                 |
| `started_at`    | `ActiveResearch.StartedAt`                                              |
| `duration_sec`  | `ActiveResearch.DurationSec`                                            |
| `progress`      | `ResearchProgress()` — already returns 0..1                             |
| `seconds_left`  | `started_at + duration_sec - time.Now().Unix()`, clamped ≥ 0            |

### 2.6 Research-tier catalog — new top-level `research_tiers: []researchTierView`

Engine: `ResearchTiers()` at `research.go:30` returning
`[]ResearchTierInfo` (`research.go:15`).

Per tier: `tier`, `name`, `duration_sec`, `frags`, `money`, `min_lvl`. The UI
needs this to render the picker that drives `start_research`.

### 2.7 Blueprints — new top-level `blueprints: []blueprintView`

Engine: `[]*Blueprint` at `state.go:120`, struct at `state.go:60`.

| Field           | Source                                                                                         |
| --------------- | ---------------------------------------------------------------------------------------------- |
| `id`            | `Blueprint.ID`                                                                                  |
| `tier`          | `Blueprint.Tier`                                                                                |
| `boosts`        | `Blueprint.Boosts`                                                                              |
| `created_at`    | `Blueprint.CreatedAt`                                                                           |
| `can_print`     | derived — `BTC ≥ tierCost*0.3 && ResearchFrags ≥ tierFrags/5 && RoomHasFreeSlot(CurrentRoom)`  |

`can_print` mirrors the gate inside `PrintMEOWCore` (`research.go:127`); the
snapshot doesn't mutate, just reports affordability. Tier cost and fragments
are looked up by walking `ResearchTiers()` for the matching `Tier`.

### 2.8 Achievements — new top-level `achievements: []string`

Engine: `state.go:167` (`s.Achievements`), checked by `CheckAchievements()`
(`achievements.go:44`). Just the IDs; the catalog of names/descriptions lives
in `data.Achievements()` (separate sprint to expose definitions).

### 2.9 Mastery — new top-level `mastery_levels: map[string]int`

Engine: `state.go:172` (`s.MasteryLevels`). Reading via `MasteryLevel(id)` is
optional — the raw map is already in canonical form.

### 2.10 Stats counters — new top-level `stats: statsView`

Group all the lifetime counters that today live unexposed on `*State`:

| Engine field                                                  | Snapshot key in `stats`                         |
| ------------------------------------------------------------- | ----------------------------------------------- |
| `TotalTicks` — `state.go:178`                                  | `total_ticks`                                   |
| `TotalGPUsBought` — `state.go:179`                             | `total_gpus_bought`                             |
| `TotalGPUsScrapped` — `state.go:180`                           | `total_gpus_scrapped`                           |
| `OCTimeT1Sec` — `state.go:181`                                 | `oc_time_t1_sec`                                |
| `OCTimeT2Sec` — `state.go:182`                                 | `oc_time_t2_sec`                                |
| `TotalWagesPaid` — `state.go:184`                              | `total_wages_paid` + `total_wages_paid_fmt`    |
| `MarketCrashCount` — `state.go:186`                            | `market_crash_count`                            |
| `LifetimeEarned` (raw + fmt)                                   | `lifetime_earned` + `lifetime_earned_fmt`       |
| `EventsByCategory` — `state.go:183`                            | `events_by_category`                            |
| `MarketPriceHistory` — `state.go:185` (cap `MarketHistoryCap`) | `market_price_history`                          |

### 2.12 Sprint 2 — design catalogs and small enrichments

Added in Sprint 2 so the frontend can render Achievement / Mastery / Legacy
panels without hardcoding the catalog. All additive — existing keys (`achievements`,
`mastery_levels`) are unchanged.

- **`achievement_defs: []achievementDefView`** — full catalog from
  `data.Achievements()`. Per entry: `id`, `emoji`, `name` (`def.LocalName()`),
  `desc` (`def.LocalDesc()`), `tp_reward`, `earned` (`s.HasAchievement(id)`).
- **`mastery_tracks: []masteryTrackView`** — full catalog from
  `data.MasteryTracks()` joined with the player's per-track level. Per entry:
  `id`, `emoji`, `name`, `desc`, `effect`, `per_level`, `level`
  (`s.MasteryLevel(t.ID)`), `max_level`, `next_cost` (`t.CostFor(level)` —
  `-1` sentinel when maxed, kept as-is per engine contract), `maxed`.
- **`legacy_perks: []legacyPerkView`** — `game.LegacyPerks()` joined with the
  loaded `*LegacyStore`. Per entry: `id`, `name`, `desc`, `cost`, `available`
  (`p.Available(legacy)`), `owned` (derived as `!available`). Note: perks like
  `efficiency_5pct` cap at 0.50, so `available=false` can mean either
  "owned-once" or "maxed-out" — the UI disambiguates by reading the legacy
  summary fields.
- **`legacy: legacyView`** — flat summary of `*LegacyStore` (loaded once per
  snapshot via `game.LoadLegacy()`): `total_earned`, `total_earned_fmt`,
  `total_lp`, `spent_lp`, `lp_available` (`legacy.LPAvailable()`),
  `starter_cash`, `efficiency_boost`, `unlocked_university`, `carried_tp`.

Two small enrichments to existing views:

- **`blueprintView.print_btc_cost`** (int) and **`blueprintView.print_frag_cost`**
  (int) — same `info.Money * 3 / 10` and `info.Frags / 5` already used to
  compute `can_print`, exposed so the UI can render a "Print (₿X / Yfrags)"
  button label without recomputing.
- **`stateView.pump_dump_unlocked`** (bool) and **`stateView.pump_dump_cooldown_sec`**
  (int64) — `unlocked = s.HasUnlock("pump_dump_action")`; cooldown mirrors
  the formula in `TriggerPumpDump` (`tick.go:486`): base 1800s, halved to
  900s when `pump_dump_ii` is unlocked, minus `now - s.EventCooldown["pump_dump"]`,
  clamped to 0. Both fields are zero when not unlocked — that's the "hidden"
  signal for the UI.

### 2.11 What's already exposed (kept for clarity, no change)

`state.kitten_name`, `state.btc(_fmt)`, `state.tech_point`,
`state.research_frags`, `state.reputation`, `state.karma`,
`state.current_room`, `state.paused`, `state.market_price`,
`state.market_trend`, `state.lifetime_earned_fmt`, `state.room_*_fmt`,
`state.mining_paused`, `state.syndicate_joined`, the entire `rooms`,
`gpus`, `gpu_defs`, `skills`, `mercs`, `merc_defs`, `log`, `last_event`.
Those keep their current shape — this sprint is additive.

---

## 3. Engine has it, `/api/action` doesn't dispatch it

| Action key           | Engine call                                                       | Body fields read                                              |
| -------------------- | ----------------------------------------------------------------- | ------------------------------------------------------------- |
| `join_syndicate`     | `state.JoinSyndicate(time.Now().Unix())` — `syndicate.go:53`      | none                                                          |
| `leave_syndicate`    | `state.LeaveSyndicate()` — `syndicate.go:73`                      | none                                                          |
| `start_research`     | `state.StartResearch(tier, boosts)` — `research.go:33`            | `tier int` (default 1), `boosts []string` (must be 2 entries) |
| `print_meowcore`     | `state.PrintMEOWCore(id)` — `research.go:127`                     | reuse existing `id` field                                     |
| `retire`             | `state.Retire()` — `prestige.go:73` — replaces `wg.state` on success | none                                                       |
| `set_difficulty`     | `state.SetDifficulty(id)` — `state.go:732` (no return)            | reuse existing `id` field                                     |
| `cycle_lang`         | `state.CycleLang()` — `state.go:589`                              | none                                                          |

Notes for the implementer:

- `Retire` returns `(*State, int, error)`. On success swap `wg.state` and
  reset `wg.lastEvent` / `wg.lastEventRoll` / `wg.lastEventSeq`, mirroring
  the existing `reset` case in the action switch.
- `actionRequest` already has `id` (used by `buy_gpu`, `unlock_skill`, etc.)
  — `print_meowcore` and `set_difficulty` reuse it. Add optional
  `tier int` and `boosts []string` for `start_research`; they're zero values
  for every other action so JSON omitempty isn't required server-side.
- `CycleLang` shifts the i18n singleton; subsequent `LocalName()` /
  `LocalFlavor()` calls in the same snapshot will already render in the new
  language — no extra plumbing.

Future actions still missing from this list (engine has them, but they're
already wired today, so not part of this sprint): `buy_gpu`, `switch_room`,
`unlock_room`, `upgrade_defense`, `upgrade_gpu`, `repair_gpu`, `scrap_gpu`,
`cycle_oc`, `vent`, `toggle_pause`, `unlock_skill`, `hire_merc`,
`bribe_merc`, `fire_merc`, `reset`. These are all already in the switch.

### Sprint 2 wired

The following actions, listed as out-of-scope for Sprint 1, were wired in
Sprint 2:

| Action key             | Engine call                                       | Body fields read           |
| ---------------------- | ------------------------------------------------- | -------------------------- |
| `level_up_mastery`     | `state.LevelUpMastery(id)` — `mastery.go:13`      | reuse `id`                 |
| `convert_frags_to_btc` | `state.ConvertFragsToBTC(frags)` — `mastery.go:89`| **new** `frags int`        |
| `trigger_pump_dump`    | `state.TriggerPumpDump()` — `tick.go:486`         | none                       |
| `repair_all_broken`    | `state.RepairAllBroken()` — `events.go:362`       | none (returns no error)    |
| `buy_legacy_perk`      | `game.BuyLegacyPerk(id)` — `prestige.go:102`      | reuse `id` (package-level) |

Notes for future maintainers:
- `RepairAllBroken` returns `(int, int)` with no error — the dispatch just
  calls and discards. Surfacing the (repaired, skipped) counts is a future
  enrichment if the UI needs it.
- `BuyLegacyPerk` is a package-level func that mutates the on-disk
  `LegacyStore`, not `wg.state`. The next snapshot reflects the new
  `legacy.*` fields automatically — no extra plumbing.
- `actionRequest` gained one new field this sprint: `frags int` (omitempty).
  Reuses the existing `id` for the other four actions.
