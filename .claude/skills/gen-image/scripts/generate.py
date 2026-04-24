#!/usr/bin/env python3
"""Generate asset images via OpenRouter (openai/gpt-5.4-image-2).

Reads OPENROUTER_API_KEY from env. Saves images to --output-dir.
"""

from __future__ import annotations

import argparse
import base64
import json
import os
import re
import sys
import time
import urllib.request
import urllib.error
from pathlib import Path

API_URL = "https://openrouter.ai/api/v1/chat/completions"
DEFAULT_MODEL = "openai/gpt-5.4-image-2"


def slugify(text: str, max_len: int = 40) -> str:
    s = re.sub(r"[^a-zA-Z0-9]+", "-", text.strip().lower()).strip("-")
    return (s[:max_len] or "image").rstrip("-")


def build_payload(model: str, prompt: str, size: str | None, quality: str | None) -> dict:
    payload = {
        "model": model,
        "messages": [{"role": "user", "content": prompt}],
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
    p.add_argument("-n", "--num", type=int, default=1, help="Number of images (separate API calls)")
    p.add_argument("-o", "--output-dir", default="assets/generated", help="Output directory")
    p.add_argument("--name", default=None, help="Base filename (defaults to slug of prompt)")
    p.add_argument("--size", default=None, help="Image size, e.g. 1024x1024")
    p.add_argument("--quality", default=None, help="Image quality: low | medium | high")
    p.add_argument("--model", default=DEFAULT_MODEL, help="OpenRouter model id")
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

    payload = build_payload(args.model, args.prompt, args.size, args.quality)
    if args.dry_run:
        print(json.dumps(payload, indent=2))
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
