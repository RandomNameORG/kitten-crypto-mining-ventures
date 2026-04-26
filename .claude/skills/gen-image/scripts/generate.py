#!/usr/bin/env python3
"""Generate asset images via OpenRouter (openai/gpt-5.4-image-2).

Reads OPENROUTER_API_KEY from env. Saves images to --output-dir.
"""

from __future__ import annotations

import argparse
import base64
import json
import mimetypes
import os
import re
import sys
import time
import urllib.request
import urllib.error
from pathlib import Path

API_URL = "https://openrouter.ai/api/v1/chat/completions"
DEFAULT_MODEL = "openai/gpt-5.4-image-2"
MODE_ALIASES = {
    "1": "character",
    "char": "character",
    "character": "character",
    "npc": "character",
    "monster": "character",
    "cat": "character",
    "2": "map",
    "map": "map",
    "scene": "map",
    "room": "map",
    "biome": "map",
    "background": "map",
    "3": "item_gpu",
    "item": "item_gpu",
    "item_gpu": "item_gpu",
    "gpu": "item_gpu",
    "device": "item_gpu",
    "machine": "item_gpu",
    "prop": "item_gpu",
    "props": "item_gpu",
    "4": "ui",
    "ui": "ui",
    "5": "fx",
    "fx": "fx",
    "effect": "fx",
    "effects": "fx",
}
UI_SUBTYPES = {
    "icon_small": {"size": "16x16", "frames": "1", "background": "#FF00FF"},
    "icon": {"size": "32x32", "frames": "1", "background": "#FF00FF"},
    "icon_large": {"size": "64x64", "frames": "1", "background": "#FF00FF"},
    "button": {"size": "160x48", "frames": "1", "background": "transparent-ready / #FF00FF"},
    "panel": {"size": "320x180", "frames": "1", "background": "transparent-ready / #FF00FF"},
    "card": {"size": "180x240", "frames": "1", "background": "transparent-ready / #FF00FF"},
    "popup": {"size": "360x200", "frames": "1", "background": "transparent-ready / #FF00FF"},
}


def slugify(text: str, max_len: int = 40) -> str:
    s = re.sub(r"[^a-zA-Z0-9]+", "-", text.strip().lower()).strip("-")
    return (s[:max_len] or "image").rstrip("-")


def normalize_mode(mode: str | None) -> str | None:
    if mode is None:
        return None
    key = mode.strip().lower().replace("-", "_")
    if key not in MODE_ALIASES:
        valid = ", ".join(sorted(set(MODE_ALIASES.values())))
        raise SystemExit(f"Unknown mode: {mode}. Valid modes: {valid}")
    return MODE_ALIASES[key]


def prompt_mentions_walk(prompt: str) -> bool:
    return bool(re.search(r"\b(walk|walking|move|moving)\b|移动|行走", prompt, re.IGNORECASE))


def build_mode_prompt(user_prompt: str, mode: str | None, ui_subtype: str, fx_size: str) -> tuple[str, str | None]:
    if mode is None:
        return user_prompt, None

    if mode == "character":
        if prompt_mentions_walk(user_prompt):
            target_size = "256x256"
            rules = [
                "Mode: character walk sheet.",
                "Use for a character, NPC, monster, or cat protagonist.",
                "Style: 2D pixel art.",
                "Background: solid #FF00FF chroma key on every empty pixel.",
                "Generate all four directions in one sheet: down, left, right, up.",
                "Each direction has exactly 4 frames.",
                "Layout: 4 columns x 4 rows.",
                "Rows, top to bottom: down, left, right, up.",
                "Frame size: 64x64 pixels.",
                "Final canvas size: 256x256 pixels.",
            ]
        else:
            target_size = "256x64"
            rules = [
                "Mode: character sprite sheet.",
                "Use for a character, NPC, monster, or cat protagonist.",
                "Style: 2D pixel art.",
                "Background: solid #FF00FF chroma key on every empty pixel.",
                "Default action frames: exactly 4 frames.",
                "Layout: 1 row x 4 columns.",
                "Frame size: 64x64 pixels.",
                "Final canvas size: 256x64 pixels.",
            ]
    elif mode == "map":
        target_size = "640x360"
        rules = [
            "Mode: map.",
            "Use for a scene, room, biome, or room background.",
            "Style: 2D top-down / 3/4 pixel art.",
            "Resolution: 640x360 pixels.",
            "Full scene image; do not use a magenta background.",
            "Single image only; no frame slicing or sprite sheet.",
            "No UI and no text.",
            "Do not include characters unless the prompt explicitly asks for them.",
        ]
    elif mode == "item_gpu":
        target_size = "256x64"
        rules = [
            "Mode: item_gpu.",
            "Use for an item, GPU, device, machine, or prop.",
            "Style: 2D pixel art.",
            "Background: solid #FF00FF chroma key on every empty pixel.",
            "Action: exactly 4 frames.",
            "Layout: 1 row x 4 columns.",
            "Frame size: 64x64 pixels.",
            "Final canvas size: 256x64 pixels.",
        ]
    elif mode == "ui":
        if ui_subtype not in UI_SUBTYPES:
            valid = ", ".join(UI_SUBTYPES)
            raise SystemExit(f"Unknown UI subtype: {ui_subtype}. Valid UI subtypes: {valid}")
        spec = UI_SUBTYPES[ui_subtype]
        target_size = spec["size"]
        rules = [
            f"Mode: ui, subtype: {ui_subtype}.",
            "Use for a game UI asset.",
            "Style: 2D pixel art UI.",
            f"Canvas size: {spec['size']} pixels.",
            f"Frames: {spec['frames']}.",
            f"Background: {spec['background']}.",
            "No extra labels or text unless the prompt explicitly asks for readable text.",
        ]
    elif mode == "fx":
        large = fx_size == "large"
        target_size = "384x96" if large else "256x64"
        frame_size = "96x96" if large else "64x64"
        rules = [
            "Mode: fx.",
            "Use for a visual effect sprite sheet.",
            "Style: 2D pixel art.",
            "Background: solid #FF00FF chroma key on every empty pixel.",
            f"Frame size: {frame_size} pixels.",
            "Frames: exactly 4.",
            "Layout: 1 row x 4 columns.",
            f"Final canvas size: {target_size} pixels.",
        ]
    else:
        raise AssertionError(f"Unhandled mode: {mode}")

    common_rules = [
        "Use crisp pixel edges with no antialiasing blur.",
        "Keep the asset centered within each frame.",
        "Do not add UI, labels, explanatory text, watermarks, borders, or mockup framing.",
    ]
    prompt = "\n".join(
        [
            "Create a game asset from this request:",
            user_prompt,
            "",
            "Mandatory output rules:",
            *[f"- {rule}" for rule in rules + common_rules],
        ]
    )
    return prompt, target_size


def is_url(value: str) -> bool:
    return value.startswith(("http://", "https://", "data:"))


def image_file_to_data_url(path: Path) -> str:
    if not path.is_file():
        raise SystemExit(f"Reference image not found: {path}")
    mime, _ = mimetypes.guess_type(path.name)
    if not mime or not mime.startswith("image/"):
        raise SystemExit(f"Reference image must be an image file: {path}")
    data = base64.b64encode(path.read_bytes()).decode("ascii")
    return f"data:{mime};base64,{data}"


def build_message_content(prompt: str, reference_images: list[str] | None) -> str | list[dict]:
    if not reference_images:
        return prompt

    content: list[dict] = [{"type": "text", "text": prompt}]
    for ref in reference_images:
        url = ref if is_url(ref) else image_file_to_data_url(Path(ref))
        content.append({"type": "image_url", "image_url": {"url": url}})
    return content


def build_payload(
    model: str,
    prompt: str,
    size: str | None,
    quality: str | None,
    reference_images: list[str] | None = None,
) -> dict:
    payload = {
        "model": model,
        "messages": [{"role": "user", "content": build_message_content(prompt, reference_images)}],
        "modalities": ["image", "text"],
    }
    if size or quality:
        img_opts = {}
        if size:
            img_opts["size"] = size
        if quality:
            img_opts["quality"] = quality
        payload["image"] = img_opts
    return payload


def redact_payload_for_print(payload: dict) -> dict:
    redacted = json.loads(json.dumps(payload))
    for message in redacted.get("messages", []):
        content = message.get("content")
        if not isinstance(content, list):
            continue
        for part in content:
            image_url = part.get("image_url") if isinstance(part, dict) else None
            if not isinstance(image_url, dict):
                continue
            url = image_url.get("url")
            if isinstance(url, str) and url.startswith("data:"):
                header, _, _ = url.partition(",")
                image_url["url"] = f"{header},<base64 redacted>"
    return redacted


def call_api(payload: dict, api_key: str, timeout: int = 180) -> dict:
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        API_URL,
        data=data,
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
            "HTTP-Referer": "https://github.com/local/kitten-crypto-mining-ventures",
            "X-Title": "kitten-crypto-mining-ventures asset gen",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        body = e.read().decode("utf-8", errors="replace")
        raise SystemExit(f"HTTP {e.code}: {body}") from e


def extract_images(resp: dict) -> list[tuple[str, bytes]]:
    """Extract (extension, bytes) for each image in the response.

    Handles multiple OpenRouter response shapes:
    - message.images[].image_url.url (data URL or http URL)
    - message.content as list with {"type": "image_url", "image_url": ...}
    - message.content as list with {"type": "output_image", "image": "<b64>"}
    """
    out: list[tuple[str, bytes]] = []
    choices = resp.get("choices") or []
    for choice in choices:
        msg = choice.get("message") or {}
        for item in msg.get("images") or []:
            url = (item.get("image_url") or {}).get("url") if isinstance(item.get("image_url"), dict) else item.get("image_url")
            if url:
                out.append(decode_image(url))
        content = msg.get("content")
        if isinstance(content, list):
            for part in content:
                if not isinstance(part, dict):
                    continue
                t = part.get("type")
                if t in ("image_url", "image"):
                    url_obj = part.get("image_url") or part.get("image") or {}
                    url = url_obj.get("url") if isinstance(url_obj, dict) else url_obj
                    if url:
                        out.append(decode_image(url))
                elif t in ("output_image", "image_base64"):
                    b64 = part.get("image") or part.get("data") or part.get("b64_json")
                    if b64:
                        out.append(("png", base64.b64decode(b64)))
    return out


def decode_image(url: str) -> tuple[str, bytes]:
    if url.startswith("data:"):
        header, _, b64 = url.partition(",")
        ext = "png"
        m = re.match(r"data:image/([a-zA-Z0-9+]+)", header)
        if m:
            ext = m.group(1).replace("jpeg", "jpg")
        return ext, base64.b64decode(b64)
    # remote URL — fetch it
    with urllib.request.urlopen(url, timeout=60) as r:
        ext = "png"
        ctype = r.headers.get("content-type", "")
        m = re.search(r"image/([a-zA-Z0-9+]+)", ctype)
        if m:
            ext = m.group(1).replace("jpeg", "jpg")
        return ext, r.read()


def main() -> int:
    p = argparse.ArgumentParser(description="Generate asset images via OpenRouter.")
    p.add_argument("--prompt", required=True, help="Image prompt")
    p.add_argument(
        "--mode",
        "--mod",
        dest="mode",
        default=None,
        help="Asset mode: 1/character, 2/map, 3/item_gpu, 4/ui, 5/fx. Omit for raw prompt.",
    )
    p.add_argument(
        "--ui-subtype",
        "--subtype",
        dest="ui_subtype",
        default="icon",
        choices=sorted(UI_SUBTYPES),
        help="UI subtype when --mode ui is used.",
    )
    p.add_argument(
        "--fx-size",
        default="normal",
        choices=("normal", "large"),
        help="FX sheet size when --mode fx is used.",
    )
    p.add_argument("-n", "--num", type=int, default=1, help="Number of images (separate API calls)")
    p.add_argument("-o", "--output-dir", default="assets/generated", help="Output directory")
    p.add_argument("--name", default=None, help="Base filename (defaults to slug of prompt)")
    p.add_argument("--size", default=None, help="Image request size. Defaults to the mode target size when --mode is set.")
    p.add_argument("--quality", default=None, help="Image quality: low | medium | high")
    p.add_argument("--model", default=DEFAULT_MODEL, help="OpenRouter model id")
    p.add_argument(
        "-r",
        "--reference-image",
        action="append",
        default=[],
        help="Reference image path, URL, or data URL. Repeat for multiple references.",
    )
    p.add_argument("--dry-run", action="store_true", help="Print payload and exit")
    args = p.parse_args()

    api_key = os.environ.get("OPENROUTER_API_KEY")
    if not api_key and not args.dry_run:
        print("ERROR: OPENROUTER_API_KEY is not set in the environment.", file=sys.stderr)
        return 2

    out_dir = Path(args.output_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    base_name = args.name or slugify(args.prompt)
    ts = time.strftime("%Y%m%d-%H%M%S")

    mode = normalize_mode(args.mode)
    prompt, mode_size = build_mode_prompt(args.prompt, mode, args.ui_subtype, args.fx_size)
    request_size = args.size or mode_size
    payload = build_payload(args.model, prompt, request_size, args.quality, args.reference_image)
    if args.dry_run:
        print(json.dumps(redact_payload_for_print(payload), indent=2))
        return 0

    saved: list[str] = []
    for i in range(1, args.num + 1):
        print(f"[{i}/{args.num}] requesting...", file=sys.stderr)
        resp = call_api(payload, api_key)
        images = extract_images(resp)
        if not images:
            debug_path = out_dir / f"{base_name}-{ts}-{i:02d}.response.json"
            debug_path.write_text(json.dumps(resp, indent=2))
            print(f"  no image in response — raw saved to {debug_path}", file=sys.stderr)
            continue
        for j, (ext, data) in enumerate(images, start=1):
            suffix = f"-{j}" if len(images) > 1 else ""
            path = out_dir / f"{base_name}-{ts}-{i:02d}{suffix}.{ext}"
            path.write_bytes(data)
            saved.append(str(path))
            print(f"  saved {path} ({len(data)} bytes)", file=sys.stderr)

    print("\n".join(saved))
    return 0 if saved else 1


if __name__ == "__main__":
    raise SystemExit(main())
