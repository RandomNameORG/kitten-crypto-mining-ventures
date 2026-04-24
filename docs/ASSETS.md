# 🎨 Asset Requests — `meowmine` TUI

The game is a TUI, so "assets" mean **ASCII / ANSI art**, not pixel images. This file lists every art slot the game needs with a generator-friendly prompt.

For each entry:
- **file** — where it should live in the repo once generated
- **size** — target character width × height
- **constraints** — terminal-safe (monospace, ANSI-color friendly, ≤ 256 colors)
- **prompt** — paste into an image-to-ASCII or LLM pipeline

All files can stay as `.txt` unless a different extension is noted. If a file doesn't exist yet the UI falls back to a simple text placeholder.

---

## 1. Main-menu logo

- **file**: `assets/ascii/logo.txt`
- **size**: 60×8
- **constraints**: monospace block art, no colors (the game applies lipgloss color afterward)
- **prompt**:
  > ASCII block-letter logo for a game titled "KITTEN CRYPTO MINING". Width 60 columns, height 8 rows, monospace. Blocky, readable from the start menu. Optionally replace the "O" in CRYPTO with a cat head glyph. Plain ASCII only (no Unicode box-drawing).

## 2. Kitten engineer portrait

- **file**: `assets/ascii/kitten_engineer.txt`
- **size**: 32×18
- **constraints**: Unicode allowed (shade blocks `░▒▓█`), plus goggle highlight
- **prompt**:
  > ASCII portrait of a small black cat wearing welding goggles, sitting at a desk beside a GPU and an energy drink. Size 32 columns × 18 rows. Use ░▒▓█ block shading for depth. Cyberpunk vibe. The pupils should be visible through the goggle lenses as tiny circles.

## 3. Room illustrations (one per biome)

Each illustration is 50×10, drawn above the dashboard when the room is active.

| Room id | File | Description |
|---|---|---|
| `alley` | `assets/ascii/room_alley.txt` | A cramped apartment room, radiator, single bare lightbulb, stacks of GPUs on a shelf, small window with neon glow outside. |
| `university` | `assets/ascii/room_university.txt` | A quiet university server room at night, rows of racks, flickering desk lamp, open fire-exit door. |
| `warehouse` | `assets/ascii/room_warehouse.txt` | An abandoned warehouse, roll-up door, broken skylight, puddle on floor, GPU pallets. |
| `basement` | `assets/ascii/room_basement.txt` | A concrete basement with exposed pipes, industrial fan, crack in floor (foreshadow tunnel heist). |
| `seacontainer` | `assets/ascii/room_seacontainer.txt` | A shipping container on a cargo ship deck at dawn, water spray, seagulls, a rope net over the racks. |

- **size**: 50 columns × 10 rows
- **constraints**: monospace, Unicode allowed, no color codes (game colorizes)
- **prompt (template)**:
  > ASCII art of [description]. 50 columns × 10 rows. Monospace. Slight perspective. Use Unicode block/line characters. No text labels inside the art.

## 4. GPU card icons

Tiny 12×4 ASCII icons shown next to each GPU in the rack, color-tinted per tier.

- **files**: `assets/ascii/gpu_{id}.txt` where `{id}` is the GPU id from `internal/data/gpus.json`
- **size**: 12×4
- **prompt (template)**:
  > 12×4 ASCII mini-icon of a PC graphics card labeled "[short name]". Show fan circles (O) and PCB edge (lines). Monospace. Four rows exactly. No color codes.

Needed ids: `scrap`, `gtx1060`, `gtx1060ti`, `rx580`, `gtx1080ti`, `rtx2070s`, `rtx3080`, `rtx4090`, `a100`.

## 5. Event icons / headers

Event popups use emoji now (🐀 🏴‍☠️ 🔥 …). If you want richer illustration, generate a 40×6 banner per crisis event:

| Event id | File | Prompt |
|---|---|---|
| `pirates` | `assets/ascii/event_pirates.txt` | Small ship with skull flag approaching a cargo container. 40×6. |
| `tunnel_heist` | `assets/ascii/event_tunnel_heist.txt` | Cross-section of a floor with a dug tunnel, helmeted figure emerging. 40×6. |
| `power_outage` | `assets/ascii/event_power_outage.txt` | Dark room, one glowing emergency exit sign, GPUs silent. 40×6. |
| `fire` | `assets/ascii/event_fire.txt` | Flames consuming a GPU rack. 40×6. |
| `btc_pump` | `assets/ascii/event_btc_pump.txt` | Rocket-shaped Bitcoin logo with candle chart behind. 40×6. |

## 6. Title-screen idle animation (optional stretch)

- **file**: `assets/ascii/idle_frames.txt`
- **format**: frames separated by `---`, each 40×10
- **prompt**:
  > 6-frame loop of a cat tail swishing over an open laptop. Monospace, 40 columns × 10 rows each. Separate frames with a line containing only `---`.

---

## Pipeline suggestion

1. Generate a PNG per entry (Stable Diffusion / Midjourney / DALL·E).
2. Convert with `jp2a --width=<W> --height=<H> --output=<file>` or an LLM pipeline.
3. Verify width/height in a terminal; trim trailing whitespace.
4. Commit to `assets/ascii/`. The game loads them lazily — missing files just show the text fallback.
