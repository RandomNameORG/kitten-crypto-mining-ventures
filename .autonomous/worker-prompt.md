I received a task from the project owner. Running as `claude -p` (non-interactive).

Project: /Users/jacksonc/i/kitten-crypto-mining-ventures

Phase 5 of Unity→Godot migration. Execute the bash blocks below IN ORDER, then edit one .gd file, then commit. Do NOT deviate. Do NOT read more than you need. Do NOT run long investigations.

## Known facts — trust these, don't re-verify

- macOS FS is case-insensitive, so `assets/` and `Assets/` share the same inode. That's fine. The Unity folder is literally called `Assets/` and the Godot folder we're adding is `assets/` — they coexist in the same directory on disk.
- Source art: `Assets/Resources/Artwork/` (flat, no subdirs). 11 PNGs, 21 aseprite files. Some filenames contain spaces.
- `data/graphiccards.json` references two card image paths: `DemoGraphicCard` and `DemoGraphicCard2`. Neither has a matching PNG in the source tree. Both must be 1×1 placeholders.
- Current `scripts/ui/store_item_slot.gd` loads icons from `res://assets/graphic_cards/<path>.png` — change to `res://assets/art/cards/<path>.png` with a null-safe fallback.
- Model field is `image_source_path` (NOT `icon_name`). Do not rename it.

## Step 1 — run this bash exactly

```bash
cd /Users/jacksonc/i/kitten-crypto-mining-ventures

mkdir -p assets/art/cards assets/sprites/_source assets/tilesets assets/fonts

# Copy PNG/JPG from Unity Artwork to assets/art/ (flat)
shopt -s nullglob
for f in Assets/Resources/Artwork/*.png Assets/Resources/Artwork/*.jpg Assets/Resources/Artwork/*.jpeg; do
  cp -- "$f" assets/art/
done

# Copy aseprite sources
for f in Assets/Resources/Artwork/*.aseprite; do
  cp -- "$f" assets/sprites/_source/
done

# Card placeholders: for each of DemoGraphicCard, DemoGraphicCard2
# No matching PNG exists in source, so write a 1x1 transparent PNG placeholder.
python3 - <<'PY'
import os, struct, zlib, json

def tiny_png(path):
    # 1x1 fully-transparent RGBA PNG, hand-built
    def chunk(tag, data):
        return struct.pack('>I', len(data)) + tag + data + struct.pack('>I', zlib.crc32(tag + data) & 0xffffffff)
    sig = b'\x89PNG\r\n\x1a\n'
    ihdr = struct.pack('>IIBBBBB', 1, 1, 8, 6, 0, 0, 0)  # 1x1, 8-bit RGBA
    raw = b'\x00' + b'\x00\x00\x00\x00'  # filter byte + 1 transparent pixel
    idat = zlib.compress(raw)
    png = sig + chunk(b'IHDR', ihdr) + chunk(b'IDAT', idat) + chunk(b'IEND', b'')
    with open(path, 'wb') as f:
        f.write(png)

with open('data/graphiccards.json') as f:
    data = json.load(f)

for card in data['GraphicCards']:
    name = card['ImageSource']['Path']
    src = f"Assets/Resources/Artwork/{name}.png"
    dst = f"assets/art/cards/{name}.png"
    if os.path.exists(src):
        import shutil; shutil.copy(src, dst)
        print(f"COPIED {name}.png from source")
    else:
        tiny_png(dst)
        print(f"PLACEHOLDER {name}.png (source missing)")
PY

# Report counts
echo "=== COUNTS ==="
echo "art PNG/JPG: $(ls assets/art/*.png assets/art/*.jpg assets/art/*.jpeg 2>/dev/null | wc -l | tr -d ' ')"
echo "aseprite:    $(ls assets/sprites/_source/*.aseprite 2>/dev/null | wc -l | tr -d ' ')"
echo "cards:       $(ls assets/art/cards/*.png 2>/dev/null | wc -l | tr -d ' ')"
```

## Step 2 — write README files

Create `assets/tilesets/README.md` with this exact content:

```markdown
# TileSets

TileSet `.tres` resources must be authored interactively in the Godot editor.

The TileMap + TileSet atlas configuration (tile shapes, collisions, autotile rules, terrain) requires a human to lay out tiles visually — it cannot be generated programmatically as part of the Unity→Godot migration.

Open the Godot editor, create a new TileSet resource here, and import the floor/wall PNGs from `../art/` as atlas sources.
```

Create `assets/fonts/README.md` with this exact content:

```markdown
# Fonts

This project currently uses Godot's default font. No custom font assets are required.

If a custom font is added later, drop `.ttf`/`.otf` files here and reference them via `res://assets/fonts/<name>.ttf`.
```

## Step 3 — edit store_item_slot.gd

Open `scripts/ui/store_item_slot.gd`. In `set_card()`, REPLACE the block:

```
	if card.image_source_path != "":
		var res_path := "res://assets/graphic_cards/%s.png" % card.image_source_path
		if ResourceLoader.exists(res_path):
			_icon.texture = load(res_path)
```

with:

```
	if card.image_source_path != "":
		var res_path := "res://assets/art/cards/%s.png" % card.image_source_path
		if ResourceLoader.exists(res_path):
			var tex := load(res_path)
			if tex != null:
				_icon.texture = tex
```

Change nothing else in the file. Do NOT touch any other script.

## Step 4 — commit

```bash
cd /Users/jacksonc/i/kitten-crypto-mining-ventures
git add -A
git commit -m "sprint 4: Phase 5 art assets + store icon loader"
git log --oneline -1
```

No Claude/AI attribution in the commit message.

## Step 5 — signal done

```bash
python3 -c "import json; json.dump({'status':'done','summary':'Phase 5 art copy complete. Copied N PNGs to assets/art/, M aseprite to assets/sprites/_source/, created placeholders for DemoGraphicCard + DemoGraphicCard2, wrote tileset/font READMEs, re-wired store_item_slot loader.'}, open('.autonomous/comms.json','w'))"
```

Include actual numbers (N, M) from the count output.

## Protocol (only if blocked)

I don't have AskUserQuestion. If you hit something you can't resolve:

Ask: `python3 -c "import json; json.dump({'status':'waiting','questions':[{'question':'...','header':'...','options':[{'label':'A'}],'multiSelect':False}],'rec':'A'}, open('.autonomous/comms.json','w'))"`

Then wait:
```
python3 -c "
import json, time
while True:
    d = json.load(open('.autonomous/comms.json'))
    if d.get('status') == 'answered':
        for a in d.get('answers', []): print(a)
        break
    time.sleep(3)
"
```
