# Kitten Crypto Mining Ventures 🐾

> A cat, a pile of GPUs, and a dream — in your terminal.

An incremental/tycoon game where you play a cat running a crypto mining operation. Buy GPUs (they ship in 30-180s), rack them in your current biome, balance electricity bills against earnings, survive petty thieves and the occasional Caribbean pirate, research your own custom silicon, and retire rich to a kittenly paradise. Built as a TUI with **Go + [Bubbletea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss)**.

The idea is you can leave it open in a `tmux` pane and come back to richer cats.

## 🎮 Running it

Requires **Go 1.22+**. On first clone run `go mod tidy` to fetch dependencies.

```bash
git clone https://github.com/RandomNameORG/kitten-crypto-mining-ventures
cd kitten-crypto-mining-ventures
go mod tidy
go run ./cmd/meowmine
```

Build static binaries:

```bash
go build -o meowmine     ./cmd/meowmine
go build -o meowmine-ssh ./cmd/meowmine-ssh
```

Saves live at `~/.meowmine/save.json`. Offline progress catches up (capped at 8h) when you relaunch. Start a new game with `./meowmine -new`.

## 🛰 SSH server mode

Run the game as an SSH service (terminal.shop-style) with [`charmbracelet/wish`](https://github.com/charmbracelet/wish):

```bash
./meowmine-ssh                # listens on 0.0.0.0:23234 by default
./meowmine-ssh -port 2022
```

Then anyone can play with:

```bash
ssh -p 23234 your.host
```

Each client gets their own save keyed by SHA-256 of their SSH public key, stored at `~/.meowmine/ssh_saves/<hash>.json` on the server. Anonymous (no-pubkey) connections work but don't persist meaningfully across disconnects.

## 🎹 Keys

| Key | Action |
|---|---|
| `1`-`9` | Switch view (dashboard / store / gpus / rooms / skills / log / mercs / lab / prestige) |
| `↑` `↓` / `k` `j` | Move cursor |
| `enter` / `b` | Confirm / buy / hire |
| `u` | Upgrade GPU · Unlock room · Unlock skill |
| `r` | Repair GPU · Start research (lab) |
| `s` | Scrap GPU · Save (dashboard) |
| `f` | Fire merc |
| `b` | Bribe merc (+15 loyalty for $200) · cycle lab boost combo |
| `t` | Cycle lab tier |
| `p` | Print MEOWCore (lab) · Pump & Dump (dashboard, if unlocked) · Buy legacy perk (prestige) |
| `l c w o a` | Upgrade defense: Lock / CCTV / Wiring / cOoling / Armor (from rooms view) |
| `R` | **Retire** (prestige view, only when eligible) |
| `space` | Pause / resume |
| `?` | Help |
| `q` / `ctrl+c` | Save and quit |

## 📖 Design docs

- [`docs/GAME_DESIGN.md`](docs/GAME_DESIGN.md) — full design doc (18 sections, mechanics, catalogs, numerical baselines, roadmap).
- [`docs/ASSETS.md`](docs/ASSETS.md) — art slots + AI-ready prompts. Drop generated ASCII art into `assets/ascii/` and the game will pick it up.

## 🗂 Layout

```
cmd/
  meowmine/              local TUI entry point
  meowmine-ssh/          Wish-based SSH server
internal/
  data/                  GPU · room · event · skill · merc catalogs
  game/                  state · economy · tick · events · skills · mercs · research · prestige · save/load
  ui/                    Bubbletea views — dashboard, store, gpus, rooms, skills, log, mercs, lab, prestige
assets/ascii/            ASCII art placeholders (see docs/ASSETS.md)
docs/                    design + asset docs
```

## 🛠 Status

Playable, feature-complete v0 for the core loop plus all the post-launch systems from the GDD:

- ✅ Tick loop — BTC earnings, volt draw, heat, overheating debuff
- ✅ BTC price oscillator (long+medium+short sine waves) + auto cash-out + lifetime earnings tracking
- ✅ 9 off-the-shelf GPU tiers with shipping delay, upgrade, scrap, repair
- ✅ 5 rooms/biomes with distinct cost × threat × cooling profiles
- ✅ Event engine — 15 scripted events on ~5-10 min cadence, per-room threat pools, defense-modified outcomes
- ✅ Spendable skill tree — 12 skills across Engineer / Mogul / Hacker lanes; effects ripple through tick
- ✅ Per-room defense upgrades — lock · CCTV · wiring · cooling · armor
- ✅ Mercenary system — hire · fire · bribe · loyalty drift · specialty-driven betrayal crises
- ✅ Custom MEOWCore R&D — blueprint-tier research (v1/v2/Purrfect), pick-2-of-3 boost axis, print from blueprints
- ✅ Prestige — LifetimeEarned threshold, LegacyPoints, cross-run legacy perks, carry-over blueprints
- ✅ SSH multiplayer mode (per-connection save keyed by pubkey)
- ✅ Save / offline catch-up (8h cap)

## License

MIT — see [LICENSE](LICENSE).
