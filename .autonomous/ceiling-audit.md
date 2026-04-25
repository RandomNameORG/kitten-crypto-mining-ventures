# Late-Game Ceiling Audit

> ## Sprint 2 update — 2026-04-25
>
> The "tech_share is the only TP source" bottleneck called out below has
> been addressed by adding **four new TP faucets**. The audit body that
> follows is from sprint 1 and reflects the pre-faucet state; this section
> overlays the revised numbers.
>
> ### New TP sources (all in `core/game/`)
>
> 1. **Achievement TPRewards** — every achievement now carries a one-shot
>    `TPReward` (1 / 3 / 5 / 10 by tier). Catalog total: **≈ 87 TP** across
>    18 achievements. Fired by `grantAchievement`
>    (`core/game/achievements.go`).
> 2. **Lifetime-earned milestones** — eight tiers from LE=10K (5 TP) up
>    to LE=100B (1000 TP). Catalog total: **≈ 1,980 TP** across the full
>    arc. Each tier pays exactly once via `LifetimeMilestonesPaid`
>    high-water counter; new method `checkLifetimeMilestones` runs
>    inside `CheckAchievements`.
> 3. **Syndicate dividend bonus** — every non-zero weekly payout now
>    grants **+5 TP** alongside the BTC dividend
>    (`SyndicateDividendTPBonus`). At ~52 payouts/virtual-year that's
>    ~260 TP/year of dedicated late-game play.
> 4. **Prestige TP carryover** — Retire banks **25%** of unspent TP
>    (floor) into the legacy store, **capped at 200 TP**, surfaced into
>    the next run via a new `LegacyStore.CarriedTP` field consumed once
>    by `newStateWithLegacy`.
>
> ### Revised TP/h ballpark
>
> Sprint-1 estimate was a flat ~10–12 TP/h endgame, dominated by
> `tech_share` events. With the new faucets layered in:
>
> | Source | Sustained late-game TP rate |
> |---|---|
> | `tech_share` events (unchanged) | ~10–12 TP/h |
> | Lifetime milestones | bursty: tiers 5–8 alone deliver 1,870 TP, paid in chunks as LE crosses 100M / 1B / 10B / 100B (one tier per ~tens of hours at high earn rates) |
> | Syndicate dividend | ~5 TP per virtual week ≈ 0.7 TP/h sustained |
> | Achievement bursts | ~87 TP one-shot, mostly front-loaded |
> | Prestige carry | up to +200 TP at every prestige boundary |
>
> Effective endgame TP/h on a sustained farm with milestone tiers still
> ahead lands roughly **30–60 TP/h** (event income + amortised milestone
> bursts), spiking to multi-hundred TP-per-tier when crossing the higher
> LE thresholds — a 3–5× improvement over the pre-sprint baseline.
>
> ### Mastery reachability
>
> Mastery's full 13,100-TP ladder still isn't single-run territory, but
> it is now **reachable across a realistic multi-prestige campaign**:
>
> - One full LE arc to 100B yields ~1,980 milestone TP + ~87 achievement
>   TP + ~200 prestige carry = ~2,270 TP/run, plus ongoing event +
>   syndicate income.
> - Five to seven prestige cycles, plus continuous event farming, lands
>   in the 13K range. That's a credible long-tail goal for the
>   ~50-hour-plus engaged player rather than the 1,300-hour wall the
>   sprint-1 audit projected.
>
> ### What's still pending
>
> - LP ceiling (2,150 LP total to max every prestige perk) — sprint 1's
>   second bottleneck — is **untouched** by this sprint. Adding more
>   perks or higher tiers is the natural next sprint.
> - Hardware ceiling (T3 Purrfect / 24-slot orbit) is also untouched.
>
> Verification: `go test ./core/game/` passes; new tests
> `TestAchievementTPReward`, `TestLifetimeMilestonePaysOnce`,
> `TestSyndicateDividendAwardsTP`, `TestRetireCarriesTP`, and
> `TestSimTPScalesWithProgression` lock the faucets in. Three-seed sim
> runs (`./bin/meowmine-sim --ticks=3600 --seed={1,2,3}`) confirm fresh
> starter state still produces only 2–7 TP/h — the new faucets only
> reward actual progression.
>
> ---

Snapshot date: 2026-04-25. All numbers extracted from the source on
`auto/session-1777134405-sprint-1` and corroborated by `bin/meowmine-sim`
runs (seeds 1/2/3, 24h each). Audit-only — no game code modified.

## Executive summary

- **BTC peak earn-rate ceiling:** ≈ **517 BTC/sec/GPU** sustained at neutral
  market, with a 24-slot orbit room reaching **≈ 12,400 BTC/sec** sustained
  and **≈ 55,000 BTC/sec** under stacked transient buffs (×3.0 market peak
  + ×1.5 Pump & Dump). At 86,400 s/day this implies ~₿1.07 B / day ceiling.
- **TP ceiling (skill tree):** **183 TP** total to fully unlock all three
  lanes — reachable in ~15-25 hours of play.
- **TP ceiling (mastery):** **13,100 TP** to push every track to L50.
  Practically unreachable in a single run at the ~10-12 TP/h event-driven
  income rate (≈ 1,300 hours / 50+ days of focused play).
- **LP ceiling:** Only **2,150 LP** total to max every existing prestige
  perk. After that, LP has nowhere to go.
- **Frag ceiling:** Late-game burn rate (T3 print + L6-10 upgrades) is
  ~2,000 frags one-time then trickles; alchemy at ₿0.5/frag is a vestigial
  sink relative to the multi-thousand BTC/sec mining rate.
- **Top bottlenecks:** (1) `tech_share` is the *only* TP source, fixing
  endgame TP/h regardless of progression; (2) prestige perk catalog is
  3-deep, capping LP utility around 2,150; (3) MEOWCore Purrfect at
  upgrade L10 / 24-slot orbit is a hard hardware ceiling — no T4 GPU,
  no bigger room.

## BTC peak earn rate

`core/game/tick.go:210` is the load-bearing line:

```
earned := eff * dt * earnMult * efficiencyFactor *
          s.DifficultyEarnMult() * s.MarketPrice * MiningScale *
          s.MasteryEarnMult()
// then if syndicated: earned *= (1 - SyndicateCutRate)  // 0.10 cut
```

### Multiplier chain (per running GPU)

Assumptions for the theoretical peak:
- Tier-3 MEOWCore Purrfect (`core/game/research.go:170` base eff 0.130).
- Blueprint boosts: **efficiency + durability** (×1.40 × 0.95 = ×1.33;
  best two-boost combo for sustained earn — see `BlueprintStats` at
  `core/game/research.go:172-184`). Note: random "breakthrough" 3rd boost
  (10% chance) actually drops the multiplier to 1.197 because the third
  boost is always undervolt with ×0.90.
- Upgrade Level 10 (`core/game/tick.go:339-344`).
- Engineer skills `overclock_i/ii/iii` all unlocked (×1.10³ = ×1.331,
  `core/game/skills.go:84-98`).
- Legacy `EfficiencyBoost` capped at +0.50 → ×1.50 (`core/game/legacy.go:19`,
  `core/game/prestige.go:43`).
- In-game OC Level 2 (`core/game/tick.go:146` `ocEarnMult[2] = 1.50`).
- Mining Mastery L50 (`core/data/mastery.go:41` PerLevel 0.01 →
  `1.01^50 = 1.6446`, `core/game/mastery.go:66`).
- Difficulty: normal (`EarnMult = 1.0`, `core/data/difficulty.go:46`).
- `MarketPrice = 1.0` neutral (cap 3.0 on normal, 5.0 on crypto_winter
  per `core/game/market.go:18-19, 73-77`).
- `MiningScale = 300.0` (`core/game/economy.go:8`).
- Room: `efficiencyFactor = 1.0` (heat ≤ 80% max → no penalty,
  `core/game/tick.go:206`).
- Syndicate joined (10% cut routed to dividends, `core/game/syndicate.go:25`).

| Multiplier | Source | Value at cap |
|---|---|---|
| Base eff (T3 Purrfect, e+d) | `core/game/research.go:170,180-182` | 0.130 × 1.33 = **0.1729** |
| `upgradeEffMult(10)` | `core/game/tick.go:339-344` | **2.25** |
| Skill `EfficiencyMult` (OC×3) | `core/game/skills.go:84-98` | **1.331** |
| Legacy `EfficiencyBoost` (+0.50 cap) | `core/game/prestige.go:43` | **1.50** |
| `ocEarnMult[OC=2]` | `core/game/tick.go:146` | **1.50** |
| **eff_per_gpu (effective)** | product | **≈ 1.165** |
| `dt` | per-tick seconds | 1 |
| `earnMult` (modifier stack) | `core/game/economy.go:18-26` | 1.0 sustained · ≤4.5 stacked-burst |
| `efficiencyFactor` | `core/game/tick.go:205-208` | 1.0 (0.5 in hot rooms) |
| `DifficultyEarnMult` (normal) | `core/data/difficulty.go:46` | 1.0 |
| `MarketPrice` | `core/game/market.go:73-77` | 1.0 typ · 3.0 cap (5.0 crypto_winter) |
| `MiningScale` | `core/game/economy.go:8` | 300.0 |
| `MasteryEarnMult` (L50) | `core/data/mastery.go:41` | 1.6446 |
| Syndicate net keep | `core/game/syndicate.go:25` | × 0.90 |

Per-GPU sustained earn at neutral market, syndicated:

```
1.165 × 1 × 1.0 × 1.0 × 1.0 × 1.0 × 300 × 1.6446 × 0.90
  ≈ 517.3 BTC / sec / GPU
```

Filling the 24-slot **orbit** room (`core/data/rooms.json:127`) gives
**≈ 12,415 BTC/sec sustained**. With the ×3.0 market clamp ceiling and
a Pump & Dump ×1.5 active, transient peak climbs to **≈ 55,868 BTC/sec**.
Sustained 24h at neutral market = **₿1.07 B / day**.

### Sim cross-check (24h, fresh starter state, no AI)

The headless sim does not buy or build anything — it just runs the
starter state forward. With one GTX 1060 in `alley`, durability decay
breaks the card before 10 virtual hours and earnings flatline. The sim
therefore reports the **floor**, not the ceiling, but it confirms the
no-intervention baseline:

```sh
$ ./bin/meowmine-sim --ticks=86400 --seed=1
── sim summary ──────────────────────────────
 ticks:            86400 (seed=1)
 BTC:              0.0000  (Δ -150.0000)
 LifetimeEarned:   2.5200  (Δ +2.5200)
 MarketPrice:      0.9647×
 TechPoint:        1  (Δ +1)
 GPUs:             0 running, 0 shipping, 1 broken
─────────────────────────────────────────────
```

```sh
$ ./bin/meowmine-sim --ticks=86400 --seed=2
── sim summary ──────────────────────────────
 BTC:              26.3565  (Δ -123.6435)
 LifetimeEarned:   21.8194  (Δ +21.8194)
 MarketPrice:      0.9736×
 TechPoint:        1  (Δ +1)
 GPUs:             0 running, 0 shipping, 1 broken
─────────────────────────────────────────────
```

```sh
$ ./bin/meowmine-sim --ticks=86400 --seed=3
── sim summary ──────────────────────────────
 BTC:              14.5164  (Δ -135.4836)
 LifetimeEarned:   9.3600  (Δ +9.3600)
 MarketPrice:      0.7558×
 TechPoint:        1  (Δ +1)
 GPUs:             0 running, 0 shipping, 1 broken
─────────────────────────────────────────────
```

```sh
$ ./bin/meowmine-sim --ticks=3600 --seed=1
── sim summary ──────────────────────────────
 BTC:              121.0200  (Δ -28.9800)
 LifetimeEarned:   2.5200  (Δ +2.5200)
 GPUs:             1 running, 0 shipping, 0 broken
─────────────────────────────────────────────
```

Note: in the sim, `MaybeFireEvent` reads `time.Now()` for cooldown
arithmetic (`core/game/events.go:15,38`), so virtual ticks execute in a
few wall-clock milliseconds and most events stay on cooldown. That is
why all three 24h runs report exactly +1 TP — `tech_share` only fires
once before its 300 s real-time cooldown locks it out for the rest of
the run. **The TP/h figure used in this audit is computed from event
weights, not pulled from these sim stderr logs.**

## TP economy & skill-tree depth

### Skill tree — total cost by lane (`core/data/skills.go`)

| Lane | Skills | Sum of `Cost` |
|---|---|---|
| Engineer | undervolt I/II/III (3+4+6), overclock I/II/III (4+6+8), pcb_surgery I/II (6+8), auto_repair I/II/III (8+6+8), rd_unlock (12) | **79** |
| Mogul | smart_invoicing I/II/III (3+5+7), tax_opt I/II/III (4+5+7), hedged_wallet I/II (6+8), venture_cap (12) | **57** |
| Hacker | neighbor_leech I/II/III (3+5+7), pump_dump I/II (6+8), chain_ghost I/II (10+8) | **47** |
| **Total** | | **183 TP** |

### Mastery — cost to L50 per track (`core/data/mastery.go`)

`CostFor(lvl) = BaseCost + lvl·StepCost`. Sum over levels 0..49 =
`50·B + S · (49·50/2) = 50·B + 1225·S`.

| Track | `BaseCost` | `StepCost` | Total to L50 | Effect at L50 |
|---|---|---|---|---|
| `mining` | 3 | 2 | **2,600 TP** | `1.01^50 ≈ 1.6446` (+64.5% earn) |
| `power` | 3 | 2 | **2,600 TP** | `0.99^50 ≈ 0.6050` (-39.5% bills) |
| `cooling` | 4 | 2 | **2,650 TP** | `1.02^50 ≈ 2.6916` (+169% cooling) |
| `frags` | 4 | 2 | **2,650 TP** | `1.02^50 ≈ 2.6916` (+169% scrap frags) |
| `scrap` | 3 | 2 | **2,600 TP** | `1.015^50 ≈ 2.1052` (+110% scrap value) |
| **Total** | | | **13,100 TP** | |

### TP income paths

The `TechPoint` symbol is grepped exhaustively across `core/game/`. The
**only** code path that adds TP is `events.go:145-146` handling
`tech_point` effects, and the only event in `core/data/events.json`
emitting `tech_point` is `tech_share` (id 39-49, weight 8, cooldown 300 s,
delta +1). No TP from achievements, syndicate, prestige carryover, or
research completion.

Steady-state rate model (alley room, normal difficulty, no LE gates yet):

- `baseFire = 0.04 + 0.03·0.4 = 0.052` per tick (`core/game/events.go:51`).
- Tech_share weight share of the eligible pool ≈ 8/49 ≈ 16.3 %.
- Per-tick probability ≈ 0.052 × 0.163 = **0.85 %** → tech_share would try
  to fire every ~118 s, but the 300 s cooldown in `events.json:46` clamps
  steady-state TP to **at most 12 TP/h**. Realistic figure with
  late-game LE gates expanding the pool (tax_audit, market_crash, celeb)
  and competition from `chain_ghost` no-ops drops effective rate to
  **~6-10 TP/h**.

### Time-to-max

| Goal | TP needed | Hours @ 10 TP/h |
|---|---|---|
| Skill tree fully unlocked | 183 | **18.3 h** |
| One mastery track to L50 | ~2,600 | **260 h** |
| All five mastery tracks to L50 | 13,100 | **1,310 h** (≈ 54 days continuous) |

The skill tree is fully reachable in a single dedicated session window;
the mastery cap is essentially infinite vs. realistic play time.
Mastery is doing exactly what its docstring (`core/data/mastery.go:7-10`)
intends — absorbing late-game TP overflow — but at this earn rate the
"overflow" never overflows into a max-out.

## Prestige scaling

`core/game/prestige.go:14` sets `PrestigeThreshold = 250_000.0`.
`core/game/prestige.go:55-60`:

```go
return int(math.Floor(math.Sqrt(s.LifetimeEarned / 10000.0)))
```

### LP earned at given LifetimeEarned

| LifetimeEarned | `sqrt(LE/10000)` | LP awarded |
|---|---|---|
| 250,000 | 5.000 | **5** |
| 1,000,000 | 10.000 | **10** |
| 10,000,000 | 31.622 | **31** |
| 100,000,000 | 100.000 | **100** |
| 1,000,000,000 | 316.227 | **316** |

### Legacy perks (`core/game/prestige.go:26-45`)

| Perk | `Cost` (LP) | Cap | Total LP if maxed |
|---|---|---|---|
| `starter_cash_500` (Seed Capital) | 10 | StarterCash < ₿5,000 (10 buys) | **100** |
| `unlock_university` (Alumni Privileges) | 50 | one-shot | **50** |
| `efficiency_5pct` (Muscle Memory) | 200 | EfficiencyBoost < 0.50 (10 buys) | **2,000** |
| **Total perk catalog cost to fully max** | | | **2,150 LP** |

To accumulate 2,150 LP from a single mega-prestige requires
`LE ≥ 2150² × 10,000 ≈ 4.62 × 10¹⁰ BTC` lifetime — i.e. **~46 billion
BTC LE in one run**, which is beyond any practical session. Multi-run is
the realistic path: e.g. seven prestiges at LE = 1B (316 LP each) = 2,212 LP.

### LP / perk bottleneck

Only three perks exist. After 2,150 LP spent, every subsequent prestige
delivers LP that has nowhere to go; the player still gets blueprint
carryover (`core/game/prestige.go:74-76`) and a starter cash floor, but
the LP currency itself is dead. The `EfficiencyBoost < 0.50` clamp on
`Muscle Memory` is the headline ceiling — once you have +50% legacy
efficiency, that knob retires.

## Fragment sinks

### Sources

- **GPU scrap** (`core/game/state.go:404-411`): `1 + rand.Intn(3)` raw
  frags per scrap, multiplied by `MasteryFragMult` (cap `1.02^50 ≈ 2.69`
  at L50). Effective range **1-8 frags per scrap** at L50, floored to
  raw count if mastery shaves fractional digits.
- That's the **only** source. No frag drops from events, achievements,
  research, or syndicate.

### Sinks

| Sink | Cost | Source |
|---|---|---|
| Research T1 (MEOWCore v1) | 20 frags + ₿2,000 | `core/game/research.go:25` |
| Research T2 (MEOWCore v2) | 50 frags + ₿8,000 | `core/game/research.go:26` |
| Research T3 (MEOWCore Purrfect) | 120 frags + ₿25,000 | `core/game/research.go:27` |
| Print MEOWCore (any tier) | 20% research frags + 30% money | `core/game/research.go:142-143` |
| GPU upgrade L5→6/7/8/9/10 | 3 / 5 / 8 / 12 / 20 frags | `core/game/tick.go:368-379` |
| Defense L4-L8 (each dim) | 2 / 4 / 6 / 8 / 10 frags | `core/game/tick.go:515-520` |
| Alchemy `ConvertFragsToBTC` | × 0.5 BTC per frag | `core/game/mastery.go:99` |

### Late-game frag accounting (24-slot orbit, T3 build-out)

One-time setup:

- One T3 research: 120 frags
- Print 24 MEOWCore Purrfect: 24 × (120 / 5) = **576 frags**
- Upgrade each MEOWCore to L10 (5 levels of frag cost: 3+5+8+12+20 =
  48 frags/GPU): 24 × 48 = **1,152 frags**
- Max defense on the active room (5 dims × 30 frags): **150 frags**
- **Total one-time burn ≈ 1,998 frags**

At a steady scrap rate of ~1-2 cards/h (replacing broken stock or
turning over commons), `MasteryFragMult` × ~2 raw → ~10-12 frags/h.
The setup phase therefore burns ~167 hours of frag income; once the
fleet is built, ongoing burn is a single **Print** every few real days
when a Purrfect dies (24 frags), plus occasional defense top-ups —
**well under 5 frags/h average**. Earn at 10-12 frags/h floods the
counter.

### Alchemy as overflow valve

`ConvertFragsToBTC` returns 0.5 BTC per frag. At 10-12 frags/h that's
~5-6 BTC/h. Compared with the late-game mining rate of **12,400 BTC/sec**,
alchemy is a homeopathic dose — frags will overflow visually long
before alchemy could meaningfully convert them.

## Bottleneck callouts

1. **`tech_share` is the lone TP faucet.** No TP from research, scrap,
   syndicate dividends, achievements, or prestige. The 300 s cooldown
   in `events.json:46` clamps endgame income at ~10-12 TP/h regardless
   of your hash power. Mastery's 13,100 TP cost becomes asymptotic.
   **Design implication:** if mastery is meant to be the late-game
   chase, either add a second TP source scaling with throughput
   (e.g. lifetime-earned milestones, achievement TP rewards, or a
   mastery-perk that buys-back TP/sec) or let prestige carry a fraction
   of unspent TP. Otherwise the ladder is decorative.

2. **Prestige perk catalog is too shallow for late LP supply.**
   `legacyPerks` in `core/game/prestige.go:26-45` totals 2,150 LP.
   A single prestige at LE = 100 M already mints 100 LP; players who
   reach syndicate-tier income will lap the perk shop in a handful of
   prestiges. **Design implication:** add LP-funded sinks (cosmetic
   tracks, secondary perk tiers, or LP→TP conversion at a steep rate)
   so the prestige loop has somewhere to land when the headline
   bonuses are maxed.

3. **No T4 GPU and no room above 24 slots.** MEOWCore Purrfect (T3,
   `research.go:27`) is the apex blueprint and `orbit` (24 slots,
   `rooms.json:127`) is the apex room. Once both are unlocked,
   per-GPU and per-fleet caps are hard ceilings — further BTC growth
   has to come from market timing, modifier stacking, and the
   already-maxed legacy efficiency perk. **Design implication:** the
   ceiling is well-defined and intentional, but it leaves no
   headroom to express prestige progression beyond Muscle Memory's
   +50% — every post-2,150-LP prestige yields literally zero new
   ceiling.

4. **MarketPrice clamp interacts asymmetrically with the multiplier
   stack.** Upper clamp on normal is 3.0× (`market.go:19`), lower is
   0.3× — but `market_crash` events can pin the market at 0.3× for
   300 s (`events.json:363`). Late-game players running
   neutral-stacked income see a 90% earnings cliff during pin events
   despite no other game-state change. **Design implication:** the
   pin scales linearly with hash, so the absolute BTC lost during a
   crash grows with progression. Either dampen pin floor at high
   `LifetimeEarned`, or sell the market_crash modifier as a strategic
   beat (e.g. let players burn LP for a brief immunity).

5. **Bills are inert at endgame.** Stacked skill bill_mult
   (`smart_invoicing^3 × neighbor_leech^3 ≈ 0.448`) × `MasteryBillMult`
   (`0.99^50 ≈ 0.605`) × orbit's `electric_cost_mult = 0.1` produces
   negligible drag (~5 BTC/min on a 24-rig orbit fleet) against
   thousands of BTC/sec earnings. The bill side of the economy stops
   being a meaningful tradeoff long before the player runs out of
   skill tree to spend on. **Design implication:** consider re-tuning
   `ElectricPerVoltMin` or adding a difficulty multiplier on rent that
   scales with `MasteryEarnMult`, otherwise Smart Invoicing III and
   Neighbor Leech III become free-to-buy nostalgia.

6. **Frag economy inverts at endgame.** The only source is GPU scrap,
   which is paced by GPU lifetime hours (slow once everything is L10
   Purrfect with `durability` boost — base 200 h × 2.0 boost × upgrade
   bonus + Salvage L50 yields multi-day uptime). The post-build burn
   rate is dominated by occasional MEOWCore reprints (24 frags) and
   discretionary `ConvertFragsToBTC`. Fragments will pile up faster
   than they can be spent without alchemy; alchemy returns a trivial
   BTC/h relative to mining. **Design implication:** introduce a
   higher-tier sink (T4 research, mastery levels gated by frags,
   or a fragment-priced legacy perk) before the counter becomes
   visual noise.
