# Kitten Crypto Mining Ventures 🐾

> A cat, a pile of GPUs, and a dream — in your terminal.

An incremental/tycoon game where you play a cat running a crypto mining operation. Buy GPUs (they ship in 30-180s), rack them in your current biome, balance electricity bills against earnings, survive petty thieves and the occasional Caribbean pirate, and grow from a cramped apartment to a cargo-ship container. Built as a TUI with **Go + [Bubbletea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss)**.

The idea is you can leave it open in a `tmux` pane and come back to richer cats.

## 🎮 Running the game

Requires **Go 1.22+**.

```bash
git clone https://github.com/RandomNameORG/kitten-crypto-mining-ventures
cd kitten-crypto-mining-ventures
go run ./cmd/meowmine
```

Or build a static binary:

```bash
go build -o meowmine ./cmd/meowmine
./meowmine
```

Saves live at `~/.meowmine/save.json`. Offline progress catches up (capped at 8h) when you relaunch. Start a new game with `./meowmine -new`.

## 🎹 Keys

| Key | Action |
|---|---|
| `1`-`6` | Switch view (dashboard / store / GPUs / rooms / skills / log) |
| `↑` `↓` / `k` `j` | Move cursor |
| `enter` / `b` | Confirm / buy |
| `u` | Upgrade (GPUs view) · Unlock (rooms view) |
| `r` | Repair GPU |
| `s` | Scrap GPU · or save (anywhere else) |
| `space` | Pause / resume |
| `?` | Help |
| `q` / `ctrl+c` | Save and quit |

## 📖 Design docs

- [`docs/GAME_DESIGN.md`](docs/GAME_DESIGN.md) — full design doc (18 sections, mechanics, GPU/room/event catalogs, numerical baselines, roadmap).
- [`docs/ASSETS.md`](docs/ASSETS.md) — art slots + AI-ready prompts for each. Drop generated ASCII art into `assets/ascii/` and the game will pick it up.

## 🗂 Layout

```
cmd/meowmine/           main entry point
internal/
  data/                 GPU, room, event catalogs (embedded JSON)
  game/                 state · economy · tick · events · save/load
  ui/                   Bubbletea views (dashboard, store, rooms, skills, log)
assets/ascii/           placeholder ASCII art (see docs/ASSETS.md)
docs/                   design + asset docs
```

## 🛠 Status

v0 — playable vertical slice:

- ✅ Core tick loop (earn BTC, consume Volt, accumulate Heat)
- ✅ BTC price oscillator + auto-trickle cash-out
- ✅ 9 GPU tiers with shipping delay, upgrade, scrap, repair
- ✅ 5 rooms/biomes with per-room electricity cost + threat pool
- ✅ Event scheduler with cooldowns and per-room weights (~5-10 min cadence)
- ✅ Pause / save / offline catch-up
- 🔜 Spendable skill tree (preview only right now)
- 🔜 Mercenaries, defense upgrades, custom MEOWCore R&D
- 🔜 Prestige / LegacyPoints
- 🔜 SSH server mode (`ssh play.meowmine.sh`) via [`charmbracelet/wish`](https://github.com/charmbracelet/wish)

## License

MIT — see [LICENSE](LICENSE).
