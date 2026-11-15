"""Render the Freebies: LA Sports Deals app icon (logo 2: palm sunset in LA colors) into Expo assets.

The icon is a 32x32 pixel grid in NES palette colors, scaled by whole numbers
(nearest neighbor). Pure standard library; no PIL.

Usage:
  python3 -I icons.py assets <apps/mobile/assets dir>   # writes all icon files
  python3 -I icons.py preview <out.png>                 # 32x preview of the icon only
"""
import math
import struct
import sys
import zlib

N = 32
K, W = "#000000", "#FCFCFC"
PURPLE, VIOLET, MAGENTA = "#4428BC", "#6844FC", "#940084"
GOLD, GOLD_LIGHT, BLUE = "#F8B800", "#F8D878", "#0058F8"

PALM = [
    "...kk.kk...",
    ".kk..kk..kk",
    "k...kkkk..k",
    "...k.kk.k..",
    "..k..k...k.",
]
TRUNK = [(16, 9), (16, 10), (16, 11), (15, 12), (15, 13), (15, 14), (14, 15), (14, 16), (14, 17)]


def palm(grid, dx, dy, color):
    for y, row in enumerate(PALM):
        for x, ch in enumerate(row):
            if ch == "k":
                grid[4 + dy + y][11 + dx + x] = color
    for x, y in TRUNK:
        if y + dy < N:
            grid[y + dy][x + dx] = color


def scene():
    """Purple sky, gold sun, blue Pacific with white sparkle, black palm."""
    g = [[None] * N for _ in range(N)]
    for y in range(N):
        for x in range(N):
            if y <= 21 and math.hypot(x + 0.5 - 15, y + 0.5 - 21) <= 8:
                c = GOLD_LIGHT if (y % 3 == 0 and y > 16) else GOLD
            elif y <= 8:
                c = PURPLE
            elif y <= 15:
                c = VIOLET
            elif y <= 21:
                c = MAGENTA
            else:
                c = W if ((x + y) % 6 == 0 and y % 2 == 0) else BLUE
            g[y][x] = c
    palm(g, 5, 3, K)
    palm(g, 5, 4, K)
    return g


def silhouette():
    """White palm and horizon on transparent, for the Android notification icon."""
    g = [[None] * N for _ in range(N)]
    palm(g, 5, 3, W)
    palm(g, 5, 4, W)
    for x in range(6, 26):
        g[23][x] = W
    return g


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


def main():
    if len(sys.argv) != 3 or sys.argv[1] not in ("assets", "preview"):
        raise SystemExit(__doc__)
    s = scene()
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
