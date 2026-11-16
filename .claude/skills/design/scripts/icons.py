"""Render the Freebies: LA Sports Deals app icon (Trophy Burger over LA Triple victory rays) into Expo assets.

The icon is a 32x32 pixel grid in NES palette colors, scaled by whole numbers
(nearest neighbor). Pure standard library; no PIL.

Usage:
  python3 -I icons.py assets <apps/mobile/assets dir>   # writes all icon files
  python3 -I icons.py preview <out.png>                 # 32x preview of the icon only
  python3 -I icons.py logoart <logoArt.ts>              # the in-app logo bitmap (same art)
"""
import math
import struct
import sys
import zlib

N = 32
K, W = "#000000", "#FCFCFC"
GOLD = "#F8B800"  # NES gold: the trophy and cheese
# Victory rays, "LA Triple": the teams' official Pantone colors, cycling (owner decision,
# 2026-11-15). Lakers purple PMS 268 C, Lakers gold PMS 123 C, Dodger blue PMS 294 C.
LAKERS_PURPLE, LAKERS_GOLD, DODGER_BLUE = "#552583", "#FDB927", "#005A9C"
RAYS = (LAKERS_PURPLE, LAKERS_GOLD, DODGER_BLUE)
RAY_COUNT, RAY_CX, RAY_CY = 18, 16, 17

# Sprite keys -> NES colors. Trophy golds; burger bun, lettuce, cheese, patty.
COLORS = {
    "K": K, "W": W, "Y": GOLD, "y": "#F8D878", "d": "#AC7C00",
    "b": "#E45C10", "h": "#FCA044", "g": "#00A800", "c": GOLD, "p": "#881400",
}
BURGER = [
    "....KKKKKKKK....",
    "..KKbbhbbbhbKK..",
    ".KbhhbbbhbbbbbK.",
    ".KbhbbbbbbbhbbK.",
    "KbbbbbbbbbbbbbbK",
    "KKKKKKKKKKKKKKKK",
    "KggggggggggggggK",
    "KccccccccccccccK",
    "KppppppppppppppK",
    "KppppppppppppppK",
    "KKKKKKKKKKKKKKKK",
    "KbbbbbbbbbbbbbbK",
    ".KbbbbbbbbbbbbK.",
    "..KKKKKKKKKKKK..",
]
TROPHY = [
    "..KKKKKKKKKKKKKK..",
    "KKKYYYYYYYYYYYYKKK",
    "KYKYyyYYYYYYYYdKYK",
    "KYKYyYYYYYYYYYdKYK",
    ".KKYYYYYYYYYYYdKK.",
    "...KYYYYYYYYYYK...",
    "....KYYYYYYYYK....",
    ".....KKYYYYKK.....",
    ".......KYYK.......",
    ".......KYdK.......",
    ".....KKYYYYKK.....",
    "....KYYYYYYddK....",
    "....KKKKKKKKKK....",
]
SPARK = [".W.", "WWW", ".W."]


def stamp(layer, rows, x0, y0):
    for y, row in enumerate(rows):
        for x, ch in enumerate(row):
            if ch != "." and 0 <= y0 + y < N and 0 <= x0 + x < N:
                layer[y0 + y][x0 + x] = ch


def subject():
    """Trophy Burger: a burger in a winner's cup, two sparkles. Keys, None = empty."""
    layer = [[None] * N for _ in range(N)]
    stamp(layer, BURGER, 8, 3)
    stamp(layer, TROPHY, 7, 14)
    stamp(layer, SPARK, 3, 5)
    stamp(layer, SPARK, 26, 9)
    return layer


def rays():
    g = [[None] * N for _ in range(N)]
    for y in range(N):
        for x in range(N):
            a = math.atan2(y + 0.5 - RAY_CY, x + 0.5 - RAY_CX) % (2 * math.pi)
            g[y][x] = RAYS[int(a / (2 * math.pi) * RAY_COUNT) % len(RAYS)]
    return g


def scene():
    """Trophy Burger over LA Triple victory rays, with a 1px black edge around the subject."""
    g, sub = rays(), subject()
    for y in range(N):
        for x in range(N):
            if sub[y][x] is None and any(
                0 <= y + dy < N and 0 <= x + dx < N and sub[y + dy][x + dx] is not None
                for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1))
            ):
                g[y][x] = K
    for y in range(N):
        for x in range(N):
            if sub[y][x] is not None:
                g[y][x] = COLORS[sub[y][x]]
    return g


def silhouette():
    """White trophy-burger shape on transparent, for the notification icon."""
    sub = subject()
    return [[W if sub[y][x] not in (None, "W") else None for x in range(N)] for y in range(N)]


def hex_rgb(h):
    return tuple(int(h[i:i + 2], 16) for i in (1, 3, 5))


def render(grid, size, inner=None):
    inner = inner or size
    if inner % N:
        raise SystemExit(f"art size {inner} is not a whole multiple of {N}")
    off = (size - inner) // 2
    rows = []
    for Y in range(size):
        row = bytearray()
        for X in range(size):
            ix, iy = X - off, Y - off
            c = grid[iy * N // inner][ix * N // inner] if 0 <= ix < inner and 0 <= iy < inner else None
            row += bytes(hex_rgb(c) + (255,)) if c else b"\x00\x00\x00\x00"
        rows.append(bytes(row))
    return rows


def write_png(path, rows, rgb=False):
    h, w = len(rows), len(rows[0]) // 4
    if rgb:  # App Store icons must be opaque (no alpha channel)
        rows = [bytes(b for i, b in enumerate(r) if i % 4 != 3) for r in rows]
    raw = b"".join(b"\x00" + r for r in rows)

    def chunk(tag, data):
        return struct.pack(">I", len(data)) + tag + data + struct.pack(">I", zlib.crc32(tag + data) & 0xFFFFFFFF)

    png = b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", struct.pack(">IIBBBBB", w, h, 8, 2 if rgb else 6, 0, 0, 0))
    png += chunk(b"IDAT", zlib.compress(raw, 9)) + chunk(b"IEND", b"")
    with open(path, "wb") as f:
        f.write(png)
    print(f"wrote {path} ({w}x{h})")


def write_logoart(path, grid):
    """Write the in-app logo (src/components/pixel/logoArt.ts) from the icon grid."""
    keys, palette = {}, {}
    letters = iter("kwabcdefghijmnopqrstuvxyz")
    for c in (K, W):
        keys[c] = next(letters)
    rows = []
    for row in grid:
        out = ""
        for c in row:
            if c not in keys:
                keys[c] = next(letters)
            out += keys[c]
        rows.append(out)
    for c, k in keys.items():
        palette[k] = c
    body = "\n".join(f'    "{r}",' for r in rows)
    pal = "\n".join(f"    {k}: '{c}'," for k, c in palette.items())
    with open(path, "w") as f:
        f.write(
            "// Logo art only. Generated by .claude/skills/design/scripts/icons.py logoart, so the\n"
            "// in-app logo matches the app icon exactly. Its extra NES colors are logo-only, not UI\n"
            "// tokens (see docs/ui.md \"Logo art\"). Do not edit by hand.\n"
            "import type { Bitmap } from './art';\n\n"
            "export const LOGO_ICON: Bitmap = {\n  width: 32,\n  height: 32,\n  pixels: [\n"
            f"{body}\n  ],\n  palette: {{\n{pal}\n  }},\n}};\n"
        )
    print(f"wrote {path}")


def main():
    if len(sys.argv) != 3 or sys.argv[1] not in ("assets", "preview", "logoart"):
        raise SystemExit(__doc__)
    s = scene()
    if sys.argv[1] == "logoart":
        write_logoart(sys.argv[2], s)
        return
    if sys.argv[1] == "preview":
        write_png(sys.argv[2], render(s, 32 * 16))
        return
    out = sys.argv[2]
    write_png(f"{out}/icon.png", render(s, 1024), rgb=True)          # 32x, full bleed, opaque
    write_png(f"{out}/adaptive-icon.png", render(s, 1024))          # Android foreground
    write_png(f"{out}/splash-icon.png", render(s, 1024, 768))       # 24x with transparent margin
    write_png(f"{out}/favicon.png", render(s, 64))                  # 2x
    write_png(f"{out}/notification-icon.png", render(silhouette(), 96))  # 3x, white on clear


if __name__ == "__main__":
    main()
