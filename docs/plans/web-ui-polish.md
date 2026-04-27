<!-- /autoplan restore point: /Users/jacksonc/.gstack/projects/RandomNameORG-kitten-crypto-mining-ventures/feat-2d-assets-autoplan-restore-20260426-170151.md -->
# Web UI Polish — Animations + Page Logic

Branch: `feat-2d-assets` · Author: Jackson · Date: 2026-04-26

## Problem statement

User feedback: "动画太丑了 页面逻辑太丑了 如何优化."

The recent redesign of the `Store` and `GPUs` panels (commits since `f9caffa`) introduced:

1. Animations that look amateur: too many simultaneous shimmers, multi-layer
   `animate-[shipShimmer_…]` / `animate-[shipBeam_…]` / `animate-[shipBob_…]` /
   `animate-[shipSweep_…]` running on the same card; emoji as the primary motion
   element (🐱🐈🐈‍⬛😺😻 cycling, 🐾🐾🐾, 🏠) instead of motion that conveys state;
   per-cell shimmer on `SlotMeter` slots that pulses constantly even when nothing
   is happening. Looks like a Christmas tree, reads as noise, not progress.
2. Page logic that fights the player's mental model:
   - The `Store` tab is the entry point but it shows GPU defs sorted by source order,
     not by what the player can afford or by tier ladder. No filter, no sort.
   - Slot status is a separate `SlotMeter` block, but the buy decision needs slot info
     in the buy CTA itself — currently the button says "机位已满" only after click
     intent forms.
   - The `GPUs` tab puts a duplicate `SlotMeter` AT THE TOP, then a "在途订单" card,
     then the row list. Scrolling to find a specific running GPU is tedious.
   - The `Defense` and `Skills` tabs are separate destinations, breaking the
     "one room, one screen" reading. The player must tab-hop just to upgrade
     a GPU's room cooling.
   - There is no sense of progression hierarchy: tier colors are present but the
     visual weight on a `legendary` is similar to a `trash` because both use the
     same 4px stripe + glow.

## Premises (subject to challenge)

- P1: The TUI is the canonical experience; the web UI is the polished, public-facing
  showcase. Therefore the bar for visual craft is higher in web than in TUI.
- P2: The player's primary loop is "see income → buy GPU → manage heat → repeat,"
  so the UI should optimize for *fast read of state* and *one-click action*, not
  for showing off animation.
- P3: Animations earn their place only when they communicate state change
  (cooldown ending, shipping arriving, heat critical). Idle decoration is anti-craft.
- P4: The current `web/frontend/src/index.css` mixes `@apply` and inline Tailwind.
  The user explicitly preferred inline Tailwind for new components ("保留新的方式").
  Don't reintroduce class-name layers.

## Proposed approach (alternatives)

### Approach A — Restraint pass (3-5 files, ~30 min CC)

Strip animations to one purpose each. Keep:
- `slot-cell.shipping`: a single 1.4s shimmer ONLY while shipping is in flight.
- `slot-cell.broken`: red pulse ONLY while broken.
- `ShipCard` progress bar: smooth `width` transition, no shimmer overlay, no truck emoji marker.
- Hover: scale + border, no rotation.

Remove: `gpuShine`, `shipBeam` rail beam, `shipSweep` card sweep, the cat-face avatar
on the progress bar, the 🐾🐾🐾 trail, the 🏠 destination icon, the "🐱📦 小猫快递服务中"
header decorator. Keep ONE cat motif: a single 🐱 that appears next to "在途订单" and
walks across the progress bar via `transition: left` (no per-frame keyframe spin).

Information architecture untouched. Pure visual triage. Two-day human / 30-min CC.

### Approach B — Restraint + IA refactor (8-10 files, ~90 min CC)

A, plus:
- Move the Store's slot CTA INTO the buy button: button reads
  `购买  ▸  ₿1.20  · 占用 3/8` so slot context lives at the decision point.
  Drop the standalone `<SlotMeter>` from Store; keep it on GPUs (where the player
  is managing slots).
- Sort GPU defs in Store by `tier` ladder first, then by efficiency. Add a single
  affordability filter (`只看买得起的`) as a chip toggle at the top, no dropdown.
- On GPUs tab: collapse "在途订单" into a sticky strip at the top (single line,
  shows count + nearest ETA), expand to full ship cards on click. This avoids the
  big blue panel pushing installed GPUs below the fold.
- Tier visual hierarchy: `legendary` gets a subtle inner border + corner emblem,
  not just a brighter stripe. Make the gap between `common` and `legendary`
  feel like 5x not 1.5x.

This is the "fix the actual reading order" pass.

### Approach C — Full re-skin with motion system (15+ files, ~3-4 hr CC)

A + B, plus:
- Define a motion system: `ease-craft = cubic-bezier(0.32, 0.72, 0, 1)` for state
  changes, `ease-decoration = cubic-bezier(0.4, 0, 0.2, 1)` for hover, both as
  CSS variables. All animations route through one of two timing tokens.
- Replace the radial-gradient tier washes with a single shared `tier-glow.svg`
  applied at low opacity. Tier color tokens move to `@theme` variables
  (`--color-tier-legendary`) so cards can share one component.
- Build a primitive: `<Card tier="…" tone="…">` used by Store, GPUs, ShipCard.
  Single source of card chrome.
- Add reduced-motion respect: `@media (prefers-reduced-motion)` disables all
  shimmer/sweep, leaves transitions only.

This is the "design system" pass. Higher upside, real risk of overshoot.

## Recommended approach (subject to review)

**B (Restraint + IA refactor).** A alone leaves the page-logic complaint
unaddressed. C is overshoot for a pre-launch web client; the design system
upside doesn't pay back unless we expect 3+ more panel types. B fixes both
complaints at the right scope.

## Premise gate decisions (user, 2026-04-26)

1. **Cat charm: one signature cat per panel.** Strip 5/6 of current cat decorations
   wholesale. Keep ONE per panel:
   - **GPUs/ShipCard:** the cat (🐈) walks across the shipping progress bar (this IS
     the signature moment).
   - **Store:** the small `🐈 📦 ×n` shipping badge on cards with in-flight orders.
     Drop everything else.
   - **SlotMeter:** no cat. The colored cells convey state.
   - **Empty states:** drop the 🐾 prefix; plain copy.
   Drop entirely: the `🐱📦` ship-section header decorator + "小猫快递服务中" subtitle,
   `🐾🐾🐾` paw trail, `🏠` destination icon, `🐈‍⬛` empty-slot overflow line, the
   per-card cat avatar inside `ship-card-icon`.
2. **Room-view sketch FIRST.** Before B's IA refactor lands, spend 30 min sketching
   a single "Room" tab that nests Defense + Skills + Mercs under the active room.
   Decision point: if the sketch is obviously better, supersede B's sticky-strip
   work with the Room-view plan; if not, ship B as designed.

These are now binding constraints on Phases 2-3.

## Files affected (Approach B)

- `packages/web/frontend/src/panels/StorePanel.tsx` — sort, filter chip, integrated buy CTA
- `packages/web/frontend/src/panels/GPUsPanel.tsx` — collapsible ship strip
- `packages/web/frontend/src/components/SlotMeter.tsx` — tone-down idle, keep functional
- `packages/web/frontend/src/components/ShipStrip.tsx` *(new)* — collapsed shipping summary
- `packages/web/frontend/src/index.css` — remove gpuShine / shipBeam / shipSweep / pulseGlow on idle elements
- `packages/web/frontend/src/types.ts` — no change

## Out of scope

- Defense/Skills/Mercs panels (separate cleanup)
- Game stage canvas changes (sprite work)
- Audio/haptics
- TUI parity changes (TUI lives elsewhere and has its own bar)
- Hindi/Japanese localization of new copy

## Test plan

- Type-check + production build pass
- Manual browser walkthrough: Store sort, affordability filter, slot integration in CTA, ship strip expand/collapse, no animation when idle
- Reduced motion preview (Chrome devtools) — confirms no animation when user opted out (only if Approach C is selected)

## Success criteria

- A new player can see "what to buy and whether they can afford it" within 1 second of opening the Store.
- Idle pages have at most one animation running (status pill breathe, by design).
- A player on the GPUs tab can see ALL installed GPUs above the fold even when 2-3 shipments are in flight.

---

# /autoplan — Phase 1: CEO Review

## Step 0A — Premise Challenge

| # | Premise | Evidence | Verdict |
|---|---------|----------|---------|
| P1 | Web is the public showcase; bar is higher than TUI | TUI requires SSH; web is the public-facing artifact a casual visitor sees | ACCEPT, reframe goal as "looks pro in 30s" not "more animated than TUI" |
| P2 | Loop = see income → buy → manage heat → repeat | `Hud.tsx` surfaces BTC/Earn/Heat; matches `docs/GAME_DESIGN.md`; `tick.go` reflects same priorities | ACCEPT |
| P3 | Animations only earn their place if they communicate state change | Industry consensus (Material, Apple HIG, Refactoring UI); idle decoration is a maturity signal | ACCEPT |
| P4 | Inline Tailwind for new components, no `@apply` re-introduction | Stated user preference earlier this branch ("保留新的方式") | ACCEPT |

**Hidden premise surfaced:** the per-tab IA (Store / GPUs / Defense / Skills / Mercs) is correct. Worth challenging in a future plan: a single "Room" view that nests Defense + Skills + Mercs under the room context might collapse 4 tabs into 1. Out of scope here. Logged to TODOS.

## Step 0B — Existing Code Leverage

| Sub-problem | Existing code | Notes |
|-------------|---------------|-------|
| Card chrome | `.row.item-row` (index.css) | Old class still used by Mercs/Skills/Defense rows; new gpu-card is inline Tailwind |
| Tier color | `oklch(82% 0.16 155)` etc. inlined 8+ places | No `--color-tier-*` token yet; could promote in C, not B |
| Slot meter | none prior | First attempt — keep |
| Card hover | `.row:hover` `-translate-y-px` | Already mirrored in inline Tailwind |
| Sticky strip | `.event-banner` — single-line attention strip | Closest analog for the new ship strip |
| Filter chip | none | Net new |
| Affordability check | `snapshot.state.btc >= def.price` | Already computed; just need to hoist into a sort/filter |

## Step 0C — Dream State

```
CURRENT (after recent commits)
  ├── 6 simultaneous animations on Store page (shimmer, sweep, beam, bob, shine, pulse)
  ├── Tab-hop to manage one room (Store ↔ GPUs ↔ Defense ↔ Skills)
  ├── Tier hierarchy = 4px stripe + glow (legendary feels ~1.5x of trash)
  └── Slot info disconnected from buy CTA

THIS PLAN (Approach B)
  ├── ≤2 animations per page; only on actual state change
  ├── Store sort by tier ladder + affordability filter chip
  ├── Slot info merged into buy CTA
  ├── Ship strip collapses to single line; expands on click
  └── Legendary vs trash feels 5x via inner border + corner emblem

12-MONTH IDEAL
  ├── One <Card> primitive used by every panel
  ├── Motion tokens (--ease-craft / --ease-decoration) in @theme
  ├── Tier tokens (--color-tier-*) as CSS vars
  ├── prefers-reduced-motion respected throughout
  ├── Single "Room" view nesting Defense + Skills + Mercs (collapses tab count)
  └── Web is link-shareable / marketing surface

DELTA after this plan: closes the visual-noise gap and the buy-CTA gap.
Does NOT yet introduce primitive abstraction or motion tokens.
12-month ideal ~70% reachable in two more polish passes (motion-tokens + room-view).
```

## Step 0C-bis — Implementation Alternatives

| | **A: Restraint pass** | **B: Restraint + IA** ✅ | **C: Full re-skin + tokens** |
|---|---|---|---|
| Files touched | 3 | 5 | 12+ |
| Human effort | ~1 day | ~2-3 days | ~1-2 weeks |
| CC effort | ~30 min | ~90 min | ~3-4 hr |
| Risk | low | low-med | med-high (premature abstraction) |
| Solves "动画太丑" | yes | yes | yes |
| Solves "页面逻辑太丑" | NO | yes | yes |
| Pays back ≥3 panels | no | partially | yes (primitive shared) |

**Recommended: B.** P1 (completeness — addresses both complaints) + P3 (pragmatic — cheaper than C with 80% of the value) + P5 (explicit over clever — avoids primitive abstraction risk before we have 3 callsites).

## Step 0D — Mode-Specific Analysis (SELECTIVE EXPANSION)

Held in scope: 5 files declared above.

**Cherry-picked expansions (in blast radius, < 1d CC):**

| Expansion | Reason | Decision |
|-----------|--------|----------|
| Empty state copy on Store ("还没解锁更多 GPU? 升级技能或解锁房间") | Same `StorePanel.tsx`, ~10 LOC | ✅ Accept |
| Defense slot panel mirroring SlotMeter | Defense doesn't use slots, semantic mismatch | ❌ Reject (P4 DRY) |
| Mercs panel sticky-strip pattern | Out of blast radius | Defer to TODOS |
| Skills panel tier hierarchy | Skills don't have tiers in same sense | ❌ Reject (P4 DRY) |

## Step 0E — Temporal Interrogation

| Time | Player state | What this plan delivers |
|------|--------------|-------------------------|
| HOUR 1 | Opens web client, buys first GPU | Clean Store, no animation noise; sort by tier+affordability shows the realistic options |
| HOUR 6 | 8 rigs running, 2 in flight | Sticky ship strip = one line; installed GPUs above fold |
| HOUR 6+ | Comparing rare vs epic vs legendary | Tier hierarchy reads at a glance; legendary feels 5x not 1.5x |

All three time-points are addressed by Approach B. Approach A breaks at HOUR 6.

## Step 0F — Mode Selection

**Mode: SELECTIVE EXPANSION** — hold the 5-file scope, cherry-pick the empty-state expansion. Defer mercs sticky-strip to TODOS.

## Step 0.5 — Dual Voices (CEO) `[subagent-only]`

**CODEX SAYS (CEO):** unavailable — `gpt-5-codex` and `gpt-5` both rejected by ChatGPT-account auth (`Codex with a ChatGPT account` does not support these models). Continuing with Claude subagent only.

**CLAUDE SUBAGENT (CEO — strategic independence):**

| # | Finding | Severity | Fix |
|---|---------|----------|-----|
| 1 | Plan answers "make it pretty" instead of "why does the player return tomorrow." For an idle/clicker, retention is the load-bearing metric. None of "first-session bounce → session-2 happens" is in scope. | HIGH | Define the one retention metric this UI should move; reframe scope around it OR pair with parallel first-run plan |
| 2 | The web-vs-TUI strategic bet (P1) gets one sentence. If web is canonical, TUI is silently demoted to a dev tool — a real product pivot dressed as styling | HIGH | State explicitly "web is canonical, TUI is novelty" or kill P1 and de-prioritize the plan |
| 3 | Step 0A flags the Room-view as a hidden premise then defers it. The player's "tab-hop to upgrade cooling" complaint IS what Room-view solves. Approach B reorders deck chairs on a 5-tab structure that may not survive | MEDIUM | 30-min Room-view sketch BEFORE committing to B's IA refactor |
| 4 | 6-month regret: removing 🐱🐈🐈‍⬛😺😻/🐾/🏠 wholesale on "Refactoring UI consensus." For a *kitten* mining game, charm is the differentiator vs. Cookie Clicker / Adventure Capitalist. Industry-restraint defaults are wrong for whimsical idle games | MEDIUM | Keep ONE signature cat moment per panel (e.g. shipping cat walks the bar). Strip everything else. Don't sterilize the brand |
| 5 | Approach C dismissed on "premature abstraction." Store + GPUs + ShipCard + Defense + Skills + Mercs = 6 callsites, not 2 | LOW | Acknowledge C is right architecturally, deferred on time budget, not principle |
| 6 | No "screenshot-worthy" moment. Idle-game space wins on shareable mechanics, not polish | LOW | Add one delight moment to scope (e.g. legendary GPU arrival — actually delightful, not just smooth) |
| 7 | Scope is correctly calibrated for what it is — too small for what it should be. 90 min CC visual polish is fine but the user complaint suggests they bounced off the experience entirely | MEDIUM | Pair with 90-min "first 60 seconds of a new player" plan before either ships |

**CEO consensus table (subagent-only mode):**

| Dimension | Subagent | Codex | Consensus |
|-----------|----------|-------|-----------|
| 1. Premises valid? | NO — P1 underjustified | N/A | FLAGGED |
| 2. Right problem to solve? | NO — retention is the real lever | N/A | FLAGGED |
| 3. Scope calibration correct? | locally yes, globally no | N/A | FLAGGED |
| 4. Alternatives sufficiently explored? | NO — Room-view dismissed; C dismissed unfairly | N/A | FLAGGED |
| 5. Competitive/market risks covered? | NO — no shareable moment | N/A | FLAGGED |
| 6. 6-month trajectory sound? | risky — cat charm strip-out may misfire | N/A | FLAGGED |

(N/A = Codex not available. Single-voice findings are flagged regardless per skill protocol.)

## CEO Sections 1-10

### Section 1: Architecture
Component-only change inside `packages/web/frontend/src/`. No new services, no new dependencies, no new build steps. The proposed `<ShipStrip>` is ~80 LOC, lifted from `GPUsPanel`. **No issues.** Examined: import graph, no new external deps, no API surface change.

### Section 2: Error & Rescue Map
The web client already handles `useSnapshot` errors via the toast strip. Sort/filter is pure client-side; affordability filter never breaks data. Empty state on Store (no GPU defs) already reachable. **No issues.** Examined: `useSnapshot.ts` error path, `App.tsx` toast wiring.

### Section 3: Security & Threat Model
No new input surface. Filter chip is local UI state, no server round-trip. **No issues.**

### Section 4: Data Flow & Edge Cases
- Edge: empty `gpu_defs` after sort/filter → render empty state (already handled by `.empty` class)
- Edge: 0 slots room (theoretical — no room has 0 slots in `data.Rooms()`) → `room.slots > 0 ? ... : 0` already guards in `SlotMeter.tsx`
- Edge: shipping count > visible cells → meter overflows; need `Math.min(room.slots, room.slots)` clamp (already correct, but worth a unit assert)

### Section 5: Code Quality
- **DRY violation:** `TIER_STRIPE` / `TIER_CHIP` / `TIER_BG` / `TIER_ART_BG` maps in `StorePanel.tsx` are 4 parallel records keyed by tier. If we keep B's scope, hoist to `web/frontend/src/components/tier.ts` exported helpers. **Auto-decide: P4 DRY** — extract.
- **Naming:** `SHIP_WINDOW_GUESS = 180` magic number; replace with `MAX_SHIP_SECONDS` from a shared constants file or expose from API.
- **Cat-face cycling array** lives in `GPUsPanel.tsx`. If we keep "one cat motif" per subagent finding 4, this stays in one place.

### Section 6: Test Review
- Web frontend has no automated tests today (`packages/web/frontend/` lacks vitest config). The plan adds sort/filter logic — pure functions. **TASTE DECISION:** add a tiny vitest setup for `sortGpuDefs()` and `affordabilityFilter()`? P1 says yes (completeness). P3 says skip (pragmatic — manual browser test catches it). **Auto-decide: skip vitest setup, defer to TODOS.** First test is a high-overhead ceremony for 2 pure functions.
- Manual browser walkthrough is the test plan. The plan has it.
- **Critical gap flagged:** no E2E test exists for the web UI; this plan does not introduce one. Same as today.

### Section 7: Performance
- Removing 4 simultaneous shimmer animations × N cards is a strict perf win. Each `animate-[…]` triggers compositor work; on 12 GPU cards, that's 48 active animations.
- Sort + filter run once per snapshot poll (1s) on a list of ≤25 GPU defs — trivial.
- Sticky strip uses `position: sticky` not JS scroll listeners. **No issues.**

### Section 8: Observability
No logs added or removed. **N/A.**

### Section 9: Deployment & Rollout
- `make frontend-build` produces a fresh dist. Go server reads dist on every request (no embed). Zero-downtime deploy already works. **No issues.**

### Section 10: Long-Term Trajectory
- This plan does not block C (Card primitive + motion tokens). Future C can extract patterns from B's inline Tailwind without rework, since classes-as-strings refactor cleanly into `<Card />`.
- This plan also does not block the Room-view IA refactor — sort/filter logic is portable.
- Risk: if subagent finding 4 is right (cat charm strip-out misfires), we burn one polish cycle reverting and reinstating.

## Required CEO Outputs

### NOT in scope
- Retention mechanics (subagent finding 1) — separate plan
- Web-vs-TUI strategic decision (subagent finding 2) — needs founder call
- Room-view IA refactor (subagent finding 3) — separate plan
- E2E test infra for web frontend
- Card primitive + motion tokens (Approach C)
- Mercs / Skills / Defense panel polish
- Mobile responsive tuning beyond 720px breakpoint already in plan
- Localization

### What already exists (mapped)

| Sub-problem | Existing |
|-------------|----------|
| Snapshot polling | `hooks/useSnapshot.ts` |
| Tab routing | `components/Tabs.tsx`, `App.tsx` `useState<TabId>` |
| Sticky strip pattern | `.event-banner` |
| Card hover | `.row:hover` and inline `hover:-translate-y-px` |
| Tier color palette | inlined `oklch(...)` literals in 8+ places |
| Slot accounting | `SlotMeter.tsx` (just added) |

### Dream state delta
Closes: visual-noise gap, buy-CTA slot context gap, tier hierarchy gap.
Open: primitive abstraction, motion tokens, reduced-motion respect, Room-view IA, retention loop, screenshot-moment.
~70% reachable to ideal in two more polish passes.

### Failure Modes Registry

| # | Mode | Trigger | Severity | Mitigation |
|---|------|---------|----------|------------|
| 1 | Cat charm strip-out backfires (subagent #4) | User dislikes sterilized look post-merge | MEDIUM | Keep ONE signature cat moment (acceptance criterion) |
| 2 | IA refactor obsolete in 1 month if Room-view ships | Future Room-view consolidation | LOW-MED | Sort/filter logic is portable; ship anyway |
| 3 | DRY refactor (tier helpers) introduces a tier-color regression | Mistyped record key | LOW | Manual visual diff; trash/common/rare/epic/legendary all rendered in test |
| 4 | Animation removal kills accessibility-reduced-motion users' (none today) signal | They had no signal anyway | LOW | Acceptable |

### Error & Rescue Registry

(no new error paths introduced)

| Path | Existing rescue | Plan changes |
|------|----------------|--------------|
| Snapshot fetch fail | toast + retry | unchanged |
| Buy fails (slots full / btc < price) | server error → toast | now ALSO disabled at button level |

### Completion Summary (CEO)

- Mode: SELECTIVE EXPANSION
- Premise verdict: 4/4 accepted; 1 hidden surfaced; 1 (Web-vs-TUI) flagged for founder call
- Recommended: Approach B + cherry-picked empty-state copy + ONE retained cat moment per panel (incorporating subagent #4)
- Strategy gate: subagent flags 7 strategic concerns, 2 high-severity. Surfaced at final gate.
- Outside voice: subagent-only mode (Codex unavailable)

---

# /autoplan — Phase 2: Design Review `[subagent-only]`

## Step 0 (Design Scope)

UI scope: yes (28 grep matches). DESIGN.md: not present. Existing patterns mapped (see CEO 0B). Initial design completeness rating: **6/10** (specifies removals well, under-specifies replacements).

## Step 0.5 (Dual Voices — Design)

**CODEX SAYS (design):** unavailable (auth). Tagged `[subagent-only]`.

**CLAUDE SUBAGENT (design — independent review):**

7-dimension scorecard:

| # | Dimension | Score | Rationale |
|---|-----------|-------|-----------|
| 1 | Information hierarchy | 7/10 | Right call to merge slot into buy CTA; but CTA doubling as slot meter is a hierarchy collision (price/action/capacity all on one button) |
| 2 | Missing states | 6/10 | Empty Store gets one line; unspecified: loading skeleton, dispatch-error, partial (shipping > slots), "all affordable" filter = 0 items, reduced-motion |
| 3 | Visual rhythm + density | 6/10 | Three border treatments, three bg tones across SlotMeter / ship strip / GPU list; no density token (`gap-2` vs `gap-3` inconsistent) |
| 4 | Tier hierarchy | 5/10 | Diagnosis correct; fix is a sentence not a system. Adding a 5th token pile without a tier scale won't read as 5x — reads as "more chrome on the orange one" |
| 5 | Motion calibration | 6/10 | Strong on what to cut. Weak on what survives: cat-walk needs duration/easing/non-emoji fallback (Unicode glyph renders different per OS) |
| 6 | Specificity | **4/10** | Plan names files + classes to remove but never specifies the replacement UI: no exact CTA string format, no tier-emblem shape/size, no sticky-strip dimensions, no Room-view sketch artifact format |
| 7 | Brand voice | 8/10 | "One signature cat per panel" is the right call. Risk: Store `🐈 📦 ×n` is redundant with GPUs panel cat — feels like 2 signatures, not 1-per-surface |

**Litmus consensus (subagent-only):**

| Dimension | Subagent | Codex | Consensus |
|-----------|----------|-------|-----------|
| Hierarchy correct? | partially (CTA collision) | N/A | FLAGGED |
| All states specified? | NO | N/A | FLAGGED |
| Tier hierarchy felt as 5x? | NO (system, not chrome) | N/A | FLAGGED |
| Motion earns place? | partially | N/A | FLAGGED |
| Plan specific enough? | NO | N/A | FLAGGED |
| Brand voice survives? | YES (with one calibration) | N/A | CONFIRMED-1 |

## Top design ambiguities (auto-decisions)

| # | Ambiguity | Severity | Auto-decision | Principle |
|---|-----------|----------|---------------|-----------|
| 1 | Room-view sketch deliverable format | CRITICAL | Single annotated HTML/MD mock at `docs/plans/web-ui-room-sketch.md` with 3-bullet "obvious win y/n" checklist. Owner = same session. Decision precedes any IA code change | P1 + P5 |
| 2 | Tier 5x hierarchy needs a system | HIGH | Define tier scale across 3 axes: **frame** (1px / 1.5px / 2px-double), **typography** (12 / 13 / 14px name; gold/orange price for epic+), **motion** (none / none / subtle inner-glow on legendary only). Encode as `tierScale[tier] = {frame, type, motion}` in a shared `tier.ts` | P1 + P4 (DRY) |
| 3 | Buy CTA spec: 6 states | HIGH | Write all 6: `afford+free` "购买 · ₿1.20", `afford+full` "机位已满", `broke+free` "需要 ₿1.20" (greyed), `broke+full` "机位已满 · ₿1.20" (greyed), `in-flight` "下单中…" (pulse), `locked` (room not unlocked) "升级房间". Resolve hierarchy collision: price = META row above button; button = action only | P1 |
| 4 | Cat-walk motion spec | HIGH | `transition: left 700ms cubic-bezier(0.32, 0.72, 0, 1)`. Use 16px Unicode 🐈 with `font-variant-emoji: emoji` to force OS emoji rendering; fall back to a CSS-painted `▸` triangle inside `@media (prefers-reduced-motion)`. No keyframe rotation/bob — purely position | P5 (explicit) |
| 5 | Sticky ship strip collapsed shape | MEDIUM | Single line, 32px tall, layout `[🐈] 在途 N · 最近 ETA Ms · [chevron]`. Click anywhere to expand. At 5+ in-flight: stack as 2 rows; cap at 8, summarize remainder as "+N more" | P1 |
| 6 | Store cat redundancy | LOW | Drop the per-card `🐈 📦 ×n` shipping badge. Cats live ONLY on GPUs panel ship strip + shipping cards. Store cards stay clean | P5 (one signature, one surface) |
| 7 | SlotMeter idle animation | MEDIUM | Remove `slotPulse` entirely. Border-color change only when state changes. Static at all other times | P3 (animation = state change) |
| 8 | Density token | MEDIUM | Adopt project-wide vertical rhythm: panels use `gap-3`, cards use `gap-2`, sub-rows use `gap-1.5`. Codify in plan as constants but inline at use site (don't create a wrapper component yet) | P5 |

## State coverage matrix (mandatory)

| State | Spec |
|-------|------|
| Loading (no snapshot) | Existing skeleton via `App.tsx` "{message}" fallback. Keep |
| Empty (no defs unlocked) | "解锁更多房间或学习技能解锁新显卡" centered, no emoji |
| Error (dispatch fail) | Existing toast strip + button briefly red border 600ms |
| Affordability filter = 0 | "暂时没有买得起的显卡" centered with toggle-off chip |
| Slots full + buy attempt | Button reads "机位已满"; click is no-op |
| Shipping > slots (theoretical) | Cap rendering at `room.slots`; surplus folded into stat counter |
| Reduced motion | All shimmer/sweep/beam off; transitions and color changes only; cat hidden, replaced by static `▸` |
| Broken + full slots | Existing red broken state + slot meter shows red+amber mix |

## Required design outputs

### Tier scale (proposed token shape — keep inline-Tailwind, no class layer)

```ts
export const TIER_SCALE = {
  trash:     { frame: 'border', name: 'text-[12px]', motion: '' },
  common:    { frame: 'border', name: 'text-[13px]', motion: '' },
  rare:      { frame: 'border-2', name: 'text-[13px]', motion: '' },
  epic:      { frame: 'border-2', name: 'text-[14px] text-gold', motion: '' },
  legendary: { frame: 'border-2 [box-shadow:inset_0_0_0_1px_oklch(72%_0.18_40/_0.6)]', name: 'text-[14px] font-bold text-orange', motion: 'animate-[gpuLegendaryGlow_4s_var(--ease-out)_infinite]' },
} as const;
```

Reads as 5x because: (a) frame doubles, (b) typography +2px and color shift, (c) ONLY legendary has a slow inner-glow pulse — the absence of motion on lower tiers is itself the signal.

### Completion Summary (Design)

- Initial completeness: 6/10 → after auto-decisions: 8.5/10
- Critical: Room-view sketch artifact owner + format spec'd
- 8 ambiguities resolved with explicit auto-decisions
- One taste decision (#1 buy-CTA hierarchy collision) flagged at final gate
- Outside voice: subagent-only mode

---

# /autoplan — Phase 3: Eng Review `[subagent-only]`

## Step 0 — Scope Challenge

Read actual code:
- `StorePanel.tsx` — 4 inline `TIER_*` records, magic strings, cat redundancy with GPUs panel
- `GPUsPanel.tsx` — `SHIP_WINDOW_GUESS = 180` constant (line 13), per-card `setInterval` in `ShipCard`, `transition: left` for cat-walk
- `SlotMeter.tsx` — fully presentational, no shareable hook
- `web/cmd/meowmine-web/main.go` — exposes `ships_at` and `ship_eta_sec` but NOT `ship_total_sec`; that's the source of the magic-180 bug
- `core/game/state.go:351` — `g.ShipsAt = time.Now().Unix() + int64(30+rand.Intn(150))`. Window is 30–180s, NOT a fixed 180

The plan as drafted underspecifies the engineering swap. Three real bugs hidden in the existing implementation; without backend support, "smooth progress bar" looks broken on short shipments.

## Step 0.5 — Dual Voices (Eng) `[subagent-only]`

**CODEX SAYS (eng):** unavailable.

**CLAUDE SUBAGENT (eng — independent review):**

| # | Finding | Severity | Auto-decision | Principle |
|---|---------|----------|---------------|-----------|
| E1 | `tier.ts` extraction safe, no cycle. Co-locate `<ShipStrip>` next to `GPUsPanel` instead of `components/` (only one caller) | LOW | ACCEPT — `panels/_shipStrip.tsx` co-located | P5 |
| E2 | 50+ GPU defs × 4 radial-gradient bg = paint thrash on iGPU | MEDIUM | ACCEPT — gate `TIER_BG` heavy gradients to `tier ∈ {rare, epic, legendary}`. Trash/common get a flat `bg-panel/65` | P3 |
| E3 | `ShipCard` runs its own `setInterval`; N ships = N timers re-rendering parent | HIGH | ACCEPT — hoist 1Hz tick to `<ShipStrip>`-level `useNow()` hook; pass `now` down. ShipCard becomes pure | P1 + P5 |
| E4 | Backend exposes `ship_eta_sec` but NOT `ship_total_sec`. Frontend `SHIP_WINDOW_GUESS = 180` is wrong; server window is 30–180s. A 30s shipment paints at 83% on first frame | **HIGH** | ACCEPT — add `ShipTotalSec int64` to `gpuView` and to `core/game/state.go::GPU` (persist on creation). Frontend: `progress = 1 - eta/total`. Drop `SHIP_WINDOW_GUESS` constant | P1 (completeness — fixes a *visible* bug, not just spec rigor) |
| E5 | Web frontend has zero automated tests. Plan defers vitest, but sort/filter are pure functions with trivial vitest setup (~8 lines `vite.config.ts` + `npm i -D vitest`) | MEDIUM | **TASTE DECISION** — surface at final gate (Eng says do now; CEO Section 6 said defer) | conflict |
| E6 | Cat-walk via `transition: left` triggers layout, not composite. Main-thread cost on every frame | **HIGH** | ACCEPT — use `transform: translateX(calc(var(--ship-progress, 0) * (100% - 12px)))` with `will-change: transform`. Same visual, GPU-only path | P5 |
| E7 | Emoji 🐈 + `font-variant-emoji: emoji` is Chrome 132+/Safari 17.4+/FF 141+. Linux without Noto Color Emoji shows tofu | MEDIUM | ACCEPT — wrap fallback in `@supports not (font-variant-emoji: emoji)` AND `@media (prefers-reduced-motion)`. Fallback is `▸` glyph styled with the blue accent | P1 |
| E8 | `<SlotMeter>` is presentational-only; needs a `useSlotStats()` hook split so the buy-CTA can consume slot context without forking rendering | MEDIUM | ACCEPT — split into `lib/slotStats.ts` (pure) + presentational `<SlotMeter>` (consumes hook) + buy-CTA consumes hook directly. Same blast radius (already touching SlotMeter) | P4 (DRY) + P5 |

**Eng consensus (subagent-only):**

| Dimension | Subagent | Codex | Consensus |
|-----------|----------|-------|-----------|
| Architecture sound? | yes-with-fixes | N/A | CONFIRMED-1 |
| Test coverage sufficient? | NO (vitest deferred) | N/A | FLAGGED |
| Performance risks addressed? | NO (magic-180, transform vs left, paint thrash) | N/A | FLAGGED |
| Security threats covered? | yes (no new surface) | N/A | CONFIRMED-1 |
| Error paths handled? | yes (existing toast strip) | N/A | CONFIRMED-1 |
| Deployment risk manageable? | yes | N/A | CONFIRMED-1 |

## Section 1: Architecture (ASCII dependency graph)

```
   panels/StorePanel.tsx ──────┐
                               │
   panels/GPUsPanel.tsx  ──────┼──► components/SlotMeter.tsx ──► lib/slotStats.ts (NEW)
                          │    │                                       ▲
                          │    └────────────────────────────────────────┘
                          │                                            │
                          ├──► panels/_shipStrip.tsx (NEW) ─────────────┤
                          │            │                              │
                          │            └──► lib/useNow.ts (NEW, 1Hz)  │
                          │                                            │
                          └──► components/tier.ts (NEW) ◄───────────────┘
                                       ▲
                                       └── exports TIER_SCALE, TIER_BG, TIER_CHIP

   web/cmd/meowmine-web/main.go ──► gpuView { + ship_total_sec int64 } (NEW field)
   core/game/state.go ──► GPU { + ShipTotalSec int64 } (NEW field, persist on creation)
```

No new external deps. New files: `lib/slotStats.ts`, `lib/useNow.ts`, `components/tier.ts`, `panels/_shipStrip.tsx`. Backend additions: 1 field on `GPU` struct + 1 field on `gpuView`.

## Section 3: Test Diagram

| Codepath | Type | Exists? | Plan |
|----------|------|---------|------|
| `sortGpuDefs(defs, btc)` (new pure fn) | unit | NO | Add `lib/sort.test.ts` — see taste decision |
| `affordableOnly(defs, btc)` filter | unit | NO | Same file |
| `useSlotStats()` hook | unit | NO | Same file (renderless test) |
| `useNow()` hook | unit | NO | Defer (timer mocking is a setup tax) |
| StorePanel render at high def count | manual | YES (browser) | Manual + perf sample at N=50 |
| GPUsPanel ship-strip with N=20 in flight | manual | YES (browser) | Manual + DevTools FPS sample |
| Backend `ship_total_sec` exposure | unit | NO | Add `core/game/state_test.go` case |
| `ShipsAt` randomization stays in 30-180s | unit | partial (sim_test.go) | Augment with explicit min/max assert |

**Test plan artifact:** written to `~/.gstack/projects/RandomNameORG-kitten-crypto-mining-ventures/feat-2d-assets-test-plan-20260426.md` (will write at gate approval).

## Sections 4-9 (concise)

- **Section 4 (Performance):** captured in E2/E3/E6 above. Net: +1 hoisted timer, -19 timers; -4 simultaneous shimmer animations; cat-walk on GPU compositor path. Net win.
- **Section 5 (Code Quality):** DRY violations resolved by `tier.ts` (E1). Magic numbers resolved by `ship_total_sec` (E4). Split for SlotMeter (E8).
- **Section 6 (Test):** see TASTE DECISION at gate.
- **Section 7 (Observability):** no logs added; not needed for UI changes.
- **Section 8 (Deployment):** `make frontend-build` + Go binary rebuild. Ships_at semantic unchanged (only added a sibling field). Existing saves: missing `ShipTotalSec` defaults to 0 → frontend falls back to "已抵达" rendering. Acceptable for old saves; new shipments get the correct value.
- **Section 9 (Long-term):** plan keeps door open for Approach C (motion tokens) and Room-view IA. No regressions to revert.

## Required Eng Outputs

### NOT in scope
- Vitest setup (taste decision at gate)
- Component test infra
- Web E2E tests
- Reduced-motion support beyond the cat fallback (full audit deferred)
- TypeScript strict-mode escalation
- Backend `ship_total_sec` migration of legacy saves (documented graceful fallback only)

### What already exists
- ETA computation seed: `g.ShipsAt = now + 30+rand.Intn(150)` (state.go:351)
- Snapshot poll cadence: 1Hz via `useSnapshot.ts`
- Format helpers: `FmtBTC`, `FmtBTCInt`, `FmtBTCSigned` reusable
- Sticky strip pattern: `.event-banner` precedent

### Failure Modes (eng layer)

| # | Mode | Trigger | Severity | Mitigation |
|---|------|---------|----------|------------|
| F1 | Old save lacks `ShipTotalSec`; ship cards stuck at 0% | Loading pre-fix save | LOW | Treat 0 as "show '已抵达' until next snapshot"; don't divide by zero |
| F2 | Timer hoist regresses ETA mid-tick | useNow updates but child memo stale | LOW | `useNow()` returns `now` from `useSyncExternalStore` so all consumers update same paint |
| F3 | Tier filter regression after `tier.ts` extraction | wrong key | LOW | Visual diff on all 5 tiers in QA |
| F4 | Backend field added without frontend update | partial deploy | LOW (already covered) | Frontend gracefully handles missing field |

### Test plan artifact (preview — will write on approval)

Will create `~/.gstack/projects/RandomNameORG-kitten-crypto-mining-ventures/feat-2d-assets-test-plan-20260426-170000.md` with the test diagram above + acceptance criteria from CEO + design specs.

### Completion Summary (Eng)

- 8 findings auto-decided per principles. 1 taste decision (vitest) → final gate.
- Critical bugs caught and fixed: `transition: left` thrash (E6), magic-180 frontend constant (E4), per-card timer storm (E3).
- Backend touched: 2 fields added (`ShipTotalSec` on `GPU` struct + `gpuView`).
- Outside voice: subagent-only mode.

---

# Cross-phase themes

## Theme 1: "Specificity gap" (CEO ✓ Design ✓ Eng ✓ — high confidence)

The original plan was a sketch, not a spec. All three phases independently flagged underspecification:
- CEO: "no Room-view sketch artifact format"
- Design: 4/10 specificity score, 5 ambiguities
- Eng: magic-180 constant, `transition: left` not `transform`, no `ship_total_sec` field

The auto-decisions above turned the sketch into an executable spec. Future plans should start at this depth, not arrive there via review.

## Theme 2: Tier hierarchy needs a system, not chrome (CEO + Design)

CEO flagged "legendary feels 1.5x not 5x"; Design said "more chrome on the orange one." The `TIER_SCALE` token shape (frame + typography + motion) is the agreed system. Encoded in `components/tier.ts`.

## Theme 3: Cat motif calibration (CEO + Design + User)

User said "保留新的方式 + 一点小猫" → CEO subagent said "ONE per panel" → Design said "one per surface" → user picked "one signature per panel." All three converge: shipping cat walks the bar, everything else cut.

## Theme 4: Single source of truth for slot accounting (Design + Eng)

Design wanted slot info in the buy CTA; Eng said don't fork SlotMeter rendering. Resolved: `useSlotStats()` hook + presentational SlotMeter + buy CTA consumes hook directly.

---

<!-- AUTONOMOUS DECISION LOG -->
## Decision Audit Trail

| # | Phase | Decision | Class | Principle | Rationale |
|---|-------|----------|-------|-----------|-----------|
| 1 | Premise gate | One signature cat per panel | USER | n/a | User answered: kitten brand survives, not sterilized |
| 2 | Premise gate | Sketch Room-view IA before locking B | USER | n/a | User answered: cheap insurance against rework |
| 3 | CEO 5 | Extract tier helpers → `components/tier.ts` | Mech | P4 DRY | 4 parallel records is the textbook DRY trigger |
| 4 | CEO 6 | Defer vitest (initial pass) | Taste | P3 | Surfaced again in eng E5 as taste — see gate |
| 5 | Design 1 | Room-view sketch = single annotated MD/HTML at `docs/plans/web-ui-room-sketch.md` + 3-bullet checklist | Mech | P1 + P5 | Without artifact format, "30-min sketch" is unfalsifiable |
| 6 | Design 2 | Tier scale across 3 axes (frame/type/motion), encoded in `tier.ts` | Mech | P1 + P4 | Chrome alone won't read as 5x; system needs absence-of-motion as a low-tier signal |
| 7 | Design 3 | Buy CTA: 6 states explicit, price=meta-row, button=action only | Mech | P1 | Resolves CTA hierarchy collision flagged by design subagent |
| 8 | Design 4 | Cat-walk = `transform: translateX`, 700ms `cubic-bezier(0.32,0.72,0,1)`, 16px Unicode 🐈 with `font-variant-emoji: emoji`, fallback `▸` | Mech | P5 | Pinned; eng E6 + E7 ratified the exact CSS |
| 9 | Design 5 | Sticky ship strip: 32px tall, [🐈 在途 N · 最近 ETA Ms · ▾]; click anywhere expands; cap at 8, "+N more" | Mech | P1 | Fixes "single line, count + ETA" → testable spec |
| 10 | Design 6 | Drop Store per-card `🐈 📦 ×n`. Cats only on GPUs panel | Mech | P5 (one signature per panel) | One signature per surface; Store cards stay clean |
| 11 | Design 7 | `slotPulse` removed; SlotMeter is static unless state changes | Mech | P3 | Animation only on state change |
| 12 | Design 8 | Density tokens: `gap-3` panels / `gap-2` cards / `gap-1.5` sub-rows. Inline at use site, no wrapper component | Mech | P5 | Avoids premature primitive |
| 13 | Eng E1 | `<ShipStrip>` lives at `panels/_shipStrip.tsx`, not `components/` | Mech | P5 | Promote to components/ when 2nd caller appears |
| 14 | Eng E2 | Gate `TIER_BG` heavy gradients to `tier ∈ {rare,epic,legendary}` | Mech | P3 | 50+ trash/common cards otherwise paint-thrashes iGPU |
| 15 | Eng E3 | Hoist 1Hz tick to `<ShipStrip>` `useNow()`; ShipCard becomes pure | Mech | P1 + P5 | N timers → 1 timer; same paint budget |
| 16 | Eng E4 | Add `ShipTotalSec int64` to backend `GPU` struct + `gpuView` | Mech | P1 | Frontend `SHIP_WINDOW_GUESS=180` is wrong; server window is 30-180s. Visible bug |
| 17 | Eng E5 | Vitest setup for sort/filter/slotStats pure fns | TASTE | conflict | CEO said skip, eng said do. → Gate |
| 18 | Eng E6 | `transform: translateX` not `transition: left` for cat-walk | Mech | P5 | `left` triggers layout; `transform` stays GPU |
| 19 | Eng E7 | Emoji fallback wrapped in `@supports not (font-variant-emoji: emoji)` AND `prefers-reduced-motion` | Mech | P1 | Linux without Noto Color Emoji shows tofu |
| 20 | Eng E8 | Split `<SlotMeter>` into `lib/slotStats.ts` hook + presentational consumer; buy-CTA imports hook | Mech | P4 + P5 | Same blast radius; prevents fork later |



