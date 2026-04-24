---
name: gen-image
description: |
  Generate asset/texture images for this project via OpenRouter's
  openai/gpt-5.4-image-2 model. Use when the user asks to generate, create,
  draft, or regenerate art, sprites, icons, textures, UI assets, or any
  image from a text prompt. Saves files into the project's assets directory.
allowed-tools:
  - Bash
  - Read
  - Write
---

# gen-image: asset image generator

Calls `openai/gpt-5.4-image-2` via OpenRouter using the local
`OPENROUTER_API_KEY`. Saves PNGs into `assets/generated/` by default.

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
  --prompt "pixel-art kitten miner holding a pickaxe, 32x32, transparent bg" \
  -n 4 \
  --size 1024x1024 \
  --output-dir assets/generated/kitten-miner
```

Flags:

| Flag | Default | Notes |
|------|---------|-------|
| `--prompt` | (required) | Text prompt for the image |
| `-n, --num` | `1` | How many images to generate (one API call per image) |
| `-o, --output-dir` | `assets/generated` | Directory to write files into (created if missing) |
| `--name` | slug of prompt | Base filename |
| `--size` | model default | e.g. `512x512`, `1024x1024`, `1536x1024` |
| `--quality` | model default | `low` / `medium` / `high` |
| `--model` | `openai/gpt-5.4-image-2` | Override model id |
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

## When NOT to use

- Editing/compositing existing images (this skill only generates from prompts).
- Vector / SVG output (model returns raster).
- Anything requiring reference images — current script is text-prompt only.
