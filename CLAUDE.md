# CLAUDE.md — kitten-crypto-mining-ventures

Project-local notes for Claude. Workspace-level rules live at `/Users/jacksonc/i/CLAUDE.md`.

## Ralph-loop mindset: self-verify core changes

**If your change touches `core/game/` — you verify it before reporting done. Do not push "please try this and tell me if it looks right" to the user.** The simulator, `sim_test.go` helpers, and `--debug` mode exist so you can close the feedback loop yourself.

The non-negotiable loop for any `core/game/` edit:

1. **Run the unit tests** — `go test ./core/game/`. Must pass.
2. **Run the simulator for an hour** — `./bin/meowmine-sim --ticks=3600 --seed=1` (build it first if `bin/meowmine-sim` is stale). Read the stderr summary; cross-check BTC / LifetimeEarned / GPU counts against your mental model of what should have changed. If a number looks wrong, that's the bug — investigate, don't ship.
3. **Re-roll the dice** — same command with `--seed=2` and `--seed=3`. Crashes, NaN, or wildly different magnitudes across seeds usually mean an unseeded edge case or a math path that only fires on specific rolls.
4. **If the change is load-bearing, lock it in with a test** — add a case to `core/game/sim_test.go` using `runSim(t, seed, ticks)`. Future sessions (including you) will thank you when they regress it six commits from now. Target the *invariant* you care about (e.g. "billing never drains more than X per hour at this difficulty"), not the exact numbers.

Escalate to the user only when the sim disagrees with the design doc (`docs/GAME_DESIGN.md`) and you can't tell which side is wrong, or when a change crosses a design-level decision the user owns (balance thresholds, new mechanics, difficulty curves). Mechanical "does Tick still work" questions — you answer yourself.

Apply the workspace-level Ralph-loop rules the same way you would elsewhere: search before implementing, no placeholder `// TODO` stubs in shipped code, max three failed approaches before surfacing, scope confined to this project.

## Project shape

- Bubble Tea TUI game. Entry: `cmd/meowmine` (local), `cmd/meowmine-ssh` (remote), `cmd/meowmine-sim` (headless simulator).
- Game engine is pure Go under `core/game/` — no UI dependencies. `State.Tick(now int64)` is the single step function; `now` is virtual unix-seconds.
- UI lives under `ui/`. The tea loop in `ui/app.go` calls `state.Tick(time.Now().Unix())` once per second.
- RNG is the global `math/rand`. Seed via `game.SeedRNG(seed int64)` before touching state if you want reproducibility.

## How to debug & verify game logic

There are three layers. Reach for the lightest one that works.

### 1. Unit tests (`go test ./core/game/`)

Use these for focused assertions on a single system (billing, research, events). Existing pattern: construct `NewState`, manipulate timestamps, call the targeted method or `Tick`, assert on fields. See `economy_test.go`, `research_test.go`, `events_test.go`.

`withTempHome(t)` reroutes HOME so save/legacy writes don't touch your real files — call it at the top of any test that might hit disk.

### 2. Simulator-style tests (`core/game/sim_test.go`)

For regressions that only appear over many ticks (economy balance, modifier churn, GPU wear, billing cadence), use the `runSim(t, seed, ticks)` helper. It mirrors the `cmd/meowmine-sim` inner loop exactly — same fixed epoch, same `SeedRNG` → same `Tick` → same `MaybeFireEvent` sequence.

When you suspect a bug shows up only after minutes of play, add a case here rather than the binary — it runs in CI and keeps the failure reproducible.

### 3. Simulator binary (`cmd/meowmine-sim`)

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
make run-debug                # go run ./cmd/meowmine --debug
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

1. `go test ./core/game/` — catches unit regressions.
2. `./bin/meowmine-sim --ticks=3600 --seed=1 --summary` — does the summary still look reasonable?
3. If behaviour depends on event rolls, try 2–3 seeds (`--seed=1`, `--seed=2`, `--seed=3`) — should see varied outcomes but no crashes / NaN.
4. If the change is UI-facing, `make run-debug` and drive it with `Ctrl+F` to reach the affected state quickly.

## Don't

- Don't thread `*rand.Rand` through `State` just to make the sim byte-deterministic. The existing global `rand` + `SeedRNG` is intentional; see `events_test.go` for the established pattern.
- Don't write to `~/.meowmine/save.json` from tests. Use `withTempHome(t)`.
- Don't add UI imports to `core/game`. The headless sim depends on that separation.
