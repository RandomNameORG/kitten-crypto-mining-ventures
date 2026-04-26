---
name: gen-image
description: |
  Generate asset/texture images for this project via OpenRouter's
  openai/gpt-5.4-image-2 model. Use when the user asks to generate, create,
  draft, or regenerate art, sprites, icons, textures, UI assets, or any
  image from a text prompt, optionally with one or more reference images.
  Saves files into the project's assets directory.
allowed-tools:
  - Bash
  - Read
  - Write
---

# gen-image: asset image generator

Calls `openai/gpt-5.4-image-2` via OpenRouter using the local
`OPENROUTER_API_KEY`. Saves PNGs into `assets/generated/` by default.
Supports text-only generation and image-referenced generation.

## Asset modes

Use `--mode` for project assets unless the user explicitly wants a raw,
unconstrained generation. `--mod` is accepted as an alias. Mode names and
numbers are both accepted.

### Mode 1: character

For characters, NPCs, monsters, and the cat protagonist.

- Style: 2D pixel art
- Background: `#FF00FF`
- Frame size: `64x64`
- Default frames: `4`
- Default layout: `1x4`
- Default canvas: `256x64`

If the prompt action is walk / walking / move / moving / 移动 / 行走:

- Generate all four directions in one sheet: `down`, `left`, `right`, `up`
- Each direction has 4 frames
- Layout: `4x4`
- Canvas: `256x256`

### Mode 2: map

For scenes, rooms, biomes, and room backgrounds.

- Resolution: `640x360`
- Style: 2D top-down / 3/4 pixel art
- Full scene image, no magenta background
- Single image, no frame slicing
- No UI, no text
- No characters by default unless the prompt explicitly asks

### Mode 3: item_gpu

For items, GPUs, devices, machines, and props.

- Style: 2D pixel art
- Background: `#FF00FF`
- Frame size: `64x64`
- Frames: `4`
- Layout: `1x4`
- Canvas: `256x64`

### Mode 4: ui

For UI assets. Select with `--ui-subtype`.

| Subtype | Size | Frames | Background |
|---------|------|--------|------------|
| `icon_small` | `16x16` | `1` | `#FF00FF` |
| `icon` | `32x32` | `1` | `#FF00FF` |
| `icon_large` | `64x64` | `1` | `#FF00FF` |
| `button` | `160x48` | `1` | transparent-ready / `#FF00FF` |
| `panel` | `320x180` | `1` | transparent-ready / `#FF00FF` |
| `card` | `180x240` | `1` | transparent-ready / `#FF00FF` |
| `popup` | `360x200` | `1` | transparent-ready / `#FF00FF` |

### Mode 5: fx

For visual effects.

- Style: 2D pixel art
- Background: `#FF00FF`
- Default frame size: `64x64`
- Default frames: `4`
- Default layout: `1x4`
- Default canvas: `256x64`
- Large FX: `--fx-size large`, frame size `96x96`, canvas `384x96`

## Preflight

```bash
test -n "$OPENROUTER_API_KEY" && echo OK || echo "MISSING OPENROUTER_API_KEY"
command -v python3 >/dev/null && echo PY_OK || echo "MISSING python3"
```

If `OPENROUTER_API_KEY` is missing, stop and ask the user to export it
(`export OPENROUTER_API_KEY=sk-or-...`) before retrying. Do not hard-code
the key into any file.

## Usage

Invoke the script from the project root:

```bash
python3 .claude/skills/gen-image/scripts/generate.py \
  --mode character \
  --prompt "kitten miner holding a pickaxe idle animation" \
  -n 4 \
  --output-dir assets/generated/kitten-miner
```

With reference images:

```bash
python3 .claude/skills/gen-image/scripts/generate.py \
  --mode character \
  --prompt "redraw this kitten miner as a walking sprite" \
  --reference-image assets/2d/spritesheet/characters/player/idle/down-1.png \
  --reference-image assets/2d/spritesheet/characters/player/idle/right-1.png \
  -n 2 \
  --output-dir assets/generated/kitten-miner
```

UI example:

```bash
python3 .claude/skills/gen-image/scripts/generate.py \
  --mode ui \
  --ui-subtype button \
  --prompt "green upgrade button with beveled pixel edges" \
  --output-dir assets/generated/ui
```

Flags:

| Flag | Default | Notes |
|------|---------|-------|
| `--prompt` | (required) | Text prompt for the image |
| `--mode`, `--mod` | none | `1`/`character`, `2`/`map`, `3`/`item_gpu`, `4`/`ui`, `5`/`fx`; omit only for raw prompts |
| `--ui-subtype`, `--subtype` | `icon` | UI subtype when `--mode ui` is used |
| `--fx-size` | `normal` | Use `large` for `96x96 x 4 = 384x96` FX sheets |
| `-n, --num` | `1` | How many images to generate (one API call per image) |
| `-o, --output-dir` | `assets/generated` | Directory to write files into (created if missing) |
| `--name` | slug of prompt | Base filename |
| `--size` | mode target size | Override the image request size |
| `--quality` | model default | `low` / `medium` / `high` |
| `--model` | `openai/gpt-5.4-image-2` | Override model id |
| `-r, --reference-image` | none | Reference image path, URL, or data URL. Repeat for multiple images |
| `--dry-run` | off | Print request payload without calling API |

Files are named `<base>-<timestamp>-<NN>.<ext>`. If the model returns
multiple images in one response, they are suffixed `-1`, `-2`, ...

## Choosing `n`

- Exploring a new concept: start with `-n 4` to compare variants.
- Refining a locked design: `-n 1` or `-n 2` per iteration.
- Batch generating matching assets (e.g. a set of icons): issue one
  call per asset with distinct prompts, not a single `-n 10` blob.

Each image is a separate API call, so cost scales linearly with `n`.

## Output conventions

- Project-root-relative output paths only (stay inside `assets/`).
- Prefer project-root-relative reference image paths when using local assets.
- Do not commit generated images unless the user asks.
- If the response contains no image (content policy, rate limit, etc.),
  the raw JSON is saved next to the expected filename for debugging.

## After generating

Read the file back with the Read tool so the user can see it:

```
Read({file_path: "/absolute/path/to/assets/generated/...png"})
```

Show all generated files this way before reporting `DONE`.

## When to use this skill

- "generate a sprite for X"
- "draft some concept art for the mine background"
- "make 4 icon variants for the upgrade button"
- "regenerate the title screen art"
- "use this image as reference and make it match the game style"

## When NOT to use

- Precise editing/compositing existing images where pixel-accurate preservation is required.
- Vector / SVG output (model returns raster).
- Reference-image workflows that require masking, inpainting, or control over exact regions.
