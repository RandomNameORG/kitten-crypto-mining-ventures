# CLAUDE.md — kitten-crypto-mining-ventures

Project-local notes for Claude. Workspace-level rules live at `/Users/jacksonc/i/CLAUDE.md`.

## Ralph-loop mindset: self-verify core changes

**If your change touches `packages/core/game/` — you verify it before reporting done. Do not push "please try this and tell me if it looks right" to the user.** The simulator, `sim_test.go` helpers, and `--debug` mode exist so you can close the feedback loop yourself.

The non-negotiable loop for any `packages/core/game/` edit:

1. **Run the unit tests** — `go test ./packages/core/game/`. Must pass.
2. **Run the simulator for an hour** — `./bin/meowmine-sim --ticks=3600 --seed=1` (build it first if `bin/meowmine-sim` is stale). Read the stderr summary; cross-check BTC / LifetimeEarned / GPU counts against your mental model of what should have changed. If a number looks wrong, that's the bug — investigate, don't ship.
3. **Re-roll the dice** — same command with `--seed=2` and `--seed=3`. Crashes, NaN, or wildly different magnitudes across seeds usually mean an unseeded edge case or a math path that only fires on specific rolls.
4. **If the change is load-bearing, lock it in with a test** — add a case to `packages/core/game/sim_test.go` using `runSim(t, seed, ticks)`. Future sessions (including you) will thank you when they regress it six commits from now. Target the *invariant* you care about (e.g. "billing never drains more than X per hour at this difficulty"), not the exact numbers.

Escalate to the user only when the sim disagrees with the design doc (`docs/GAME_DESIGN.md`) and you can't tell which side is wrong, or when a change crosses a design-level decision the user owns (balance thresholds, new mechanics, difficulty curves). Mechanical "does Tick still work" questions — you answer yourself.

Apply the workspace-level Ralph-loop rules the same way you would elsewhere: search before implementing, no placeholder `// TODO` stubs in shipped code, max three failed approaches before surfacing, scope confined to this project.

## Project shape

- Monorepo layout under `packages/`: `packages/core` (engine), `packages/cli` (Go TUI + dev CLIs), `packages/web` (Go web server + browser frontend). Single root `go.mod`.
- Bubble Tea TUI game. Entry: `packages/cli/cmd/meowmine` (local), `packages/cli/cmd/meowmine-ssh` (remote), `packages/cli/cmd/meowmine-sim` (headless simulator).
- Game engine is pure Go under `packages/core/game/` — no UI dependencies. `State.Tick(now int64)` is the single step function; `now` is virtual unix-seconds.
- TUI lives under `packages/cli/ui/`. The tea loop in `packages/cli/ui/app.go` calls `state.Tick(time.Now().Unix())` once per second.
- Browser frontend lives under `packages/web/frontend/` (Vite + React 18 + TypeScript). Served by `packages/web/cmd/meowmine-web` from `packages/web/frontend/dist/` after `make frontend-build`. Dev workflow: `make run-web` (Go on :8080 serving last build) plus `make run-web-dev` in another shell (Vite on :5173 with `/api`+`/assets` proxied to :8080).
- RNG is the global `math/rand`. Seed via `game.SeedRNG(seed int64)` before touching state if you want reproducibility.

## How to debug & verify game logic

There are three layers. Reach for the lightest one that works.

### 1. Unit tests (`go test ./packages/core/game/`)

Use these for focused assertions on a single system (billing, research, events). Existing pattern: construct `NewState`, manipulate timestamps, call the targeted method or `Tick`, assert on fields. See `economy_test.go`, `research_test.go`, `events_test.go`.

`withTempHome(t)` reroutes HOME so save/legacy writes don't touch your real files — call it at the top of any test that might hit disk.

### 2. Simulator-style tests (`packages/core/game/sim_test.go`)

For regressions that only appear over many ticks (economy balance, modifier churn, GPU wear, billing cadence), use the `runSim(t, seed, ticks)` helper. It mirrors the `packages/cli/cmd/meowmine-sim` inner loop exactly — same fixed epoch, same `SeedRNG` → same `Tick` → same `MaybeFireEvent` sequence.

When you suspect a bug shows up only after minutes of play, add a case here rather than the binary — it runs in CI and keeps the failure reproducible.

### 3. Simulator binary (`packages/cli/cmd/meowmine-sim`)

For exploratory debugging, balance-eyeballing, or "does this new modifier explode after an hour":

```sh
make build-sim                                       # -> bin/meowmine-sim
./bin/meowmine-sim --ticks=3600 --seed=1             # 1 virtual hour, stdout snapshot, stderr summary
./bin/meowmine-sim --ticks=86400 --seed=1 --out=/tmp/day.json   # 24h; summary + full JSON
./bin/meowmine-sim --from=~/.meowmine/save.json --ticks=3600    # advance an existing save
./bin/meowmine-sim --ticks=3600 --seed=1 --snapshot-every=600 --out=/tmp/sim.json   # periodic snapshots
```

The summary on stderr is the fastest sanity signal: check `BTC`, `LifetimeEarned`, GPU counts, `Modifiers active`. If a number looks absurd, diff two snapshots (`diff /tmp/a.json /tmp/b.json`) to pinpoint the diverging field.

**Known non-determinism:** `ShipsAt`, log entry `time`, and a couple of modifier expirations use `time.Now()` directly (not the virtual `now`). Game-mechanical fields are deterministic across same-seed runs; timestamps inside the snapshot can drift by seconds.

### 4. `--debug` flag on the TUI

For reproducing a UI-level bug interactively, or reaching a specific state fast:

```sh
make run-debug                # go run ./packages/cli/cmd/meowmine --debug
./bin/meowmine --debug --debug-seed=42
```

Runtime keys (only when `--debug` is set):

| Key | Action |
| --- | --- |
| `Ctrl+F` | Cycle sim speed: 1× → 4× → 16× → 64× → 1× |
| `Ctrl+D` | Dump full state JSON to `/tmp/meowmine-debug-<unix>.json` |
| `Ctrl+Y` | Cheat: +₿1 |
| `Ctrl+T` | Cheat: +10 TechPoint |
| `Ctrl+B` | Toggle the debug HUD line |

Debug mode is **local only** — the SSH binary never calls `EnableDebug`, so remote sessions can't use cheats or time acceleration.

## Common verification loop

When you change tick-loop behavior:

1. `go test ./packages/core/game/` — catches unit regressions.
2. `./bin/meowmine-sim --ticks=3600 --seed=1 --summary` — does the summary still look reasonable?
3. If behaviour depends on event rolls, try 2–3 seeds (`--seed=1`, `--seed=2`, `--seed=3`) — should see varied outcomes but no crashes / NaN.
4. If the change is UI-facing, `make run-debug` and drive it with `Ctrl+F` to reach the affected state quickly.

## Web frontend dev workflow

`packages/web/` is a *different surface* from the engine. The Ralph loop above is for `packages/core/game/`; web work has its own loop: **HMR + type-check + vitest + manual browser**. Don't run the simulator to verify a CSS change.

### HMR — two terminals, browse :5173/ui/

```sh
# Terminal 1 — backend serves /api and /assets on :8080
make run-web

# Terminal 2 — Vite dev server with HMR on :5173
make run-web-dev
```

Open **http://localhost:5173/ui/** (note the `/ui/` prefix from `vite.config.ts` `base`). Do NOT open `:8080` while developing — that serves the last `make frontend-build` and won't update.

`/api` and `/assets` are proxied through Vite to :8080 (`vite.config.ts` `server.proxy`). The Go server is only authoritative for the API; React is hot-replaced in-place, component state survives most edits.

**When the backend changes** (`packages/web/cmd/meowmine-web/main.go`, `packages/core/game/`), the running `make run-web` binary is stale — restart Terminal 1. Frontend-only edits do not need this.

### Verification loop (web changes)

1. `cd packages/web/frontend && npm run typecheck` — TS errors before runtime.
2. `npm run test` — vitest covers pure helpers in `src/lib/` (sort, slotStats, …). Add a case here for any new pure function; don't reach for component tests yet (no setup).
3. `npm run build` — production bundle. Catches Tailwind class typos that dev mode papers over.
4. Manual browser walkthrough on :5173/ui/ for the affected panels.

If you touched the backend, also rebuild the binary so the next `make run-web` picks it up:

```sh
go build -o bin/meowmine-web ./packages/web/cmd/meowmine-web
```

### File layout (`packages/web/frontend/src/`)

| Path | Holds |
|------|-------|
| `App.tsx` | Tab routing + layout shell + snapshot wiring |
| `hooks/useSnapshot.ts` | 1Hz poll of `/api/snapshot` + dispatcher for `/api/action` |
| `panels/` | One file per tab (Store, GPUs, Defense, Skills, Mercs, Log, Stats, Rooms). Underscore-prefixed files (e.g. `_shipStrip.tsx`) are panel fragments, not standalone tabs |
| `components/` | Reusable visuals: `SlotMeter`, `Hud`, `Tabs`, `EventBanner`, `LogStrip`, `GameStage`, `ActionButton`, plus `tier.ts` (tier visual scale) |
| `lib/` | Pure helpers, vitest-tested: `sort.ts`, `slotStats.ts`, `useNow.ts` |
| `types.ts` | API shape — must stay in sync with `gpuView` / `roomView` / `stateView` etc. in `packages/web/cmd/meowmine-web/main.go` |
| `index.css` | Tailwind v4 layer + a small set of `@keyframes` (only motion that can't be inlined) |

### Style: inline Tailwind for new components

Older components (`.row`, `.metric`, `.tabs`, `.action-btn` in `index.css`) use `@apply`. New work uses inline Tailwind directly in JSX — no class-name layer. Reasons: deletes cleanly, lives next to the markup, no naming bikeshed. Only `@keyframes` and OS-feature `@supports` blocks belong in `index.css`.

### API contract: two places, keep in sync

A new field on a snapshot view requires editing **both**:
- `packages/web/cmd/meowmine-web/main.go` — the `gpuView` / `roomView` / etc. struct + the populating code in `makeSnapshotLocked`
- `packages/web/frontend/src/types.ts` — the corresponding TS interface

Skipping one half = silent runtime breakage that TypeScript can't catch (the JSON just has `undefined`). Recent example: `ship_total_sec` had to be added to both before the ship-progress bar could render correctly.

### Animation: state-change only, transform-only

Three rules from the recent polish pass:

1. **Idle decoration is anti-craft.** Animations earn their place by communicating state change (shipping in flight, broken GPU pulsing, buy in progress). If a card animates while nothing is happening, delete that animation.
2. **Use `transform`, not `left/top/width`.** `transform: translateX(...)` runs on the GPU compositor; `transition: left/top/width` triggers layout every frame. The cat marker on the ship progress bar is the canonical example — `translateX` keeps it on the GPU thread.
3. **One 1Hz clock for all consumers.** Don't put `setInterval` in per-row components. Hoist to a single `useNow()` (`lib/useNow.ts`) and pass the timestamp down — N rows × N timers will tank scroll FPS.

## Don't

- Don't thread `*rand.Rand` through `State` just to make the sim byte-deterministic. The existing global `rand` + `SeedRNG` is intentional; see `events_test.go` for the established pattern.
- Don't write to `~/.meowmine/save.json` from tests. Use `withTempHome(t)`.
- Don't add UI imports to `packages/core/game`. The headless sim depends on that separation.
- Don't reintroduce `@apply` class-name layers for new components in `index.css`. Inline Tailwind, in the JSX, beside the markup. Keyframes are the only exception.
- Don't open `localhost:8080` to verify a frontend change. That's the static build. Use `:5173/ui/`.
- Don't add a field to a snapshot view in `main.go` without also adding it to `types.ts`. TypeScript can't catch the JSON drift.
