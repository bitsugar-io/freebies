"""Render the README logo (palm mark + block-letter FREEBIES) as SVG, sibling-repo style.

Writes asset/logo-dark.svg, asset/logo-light.svg and asset/mark.svg at the repo root.
The mark is the app icon scene from icons.py; the wordmark is a 5x7 pixel font drawn
as square blocks, like opensesame and hocusfocus. Pure standard library.

Usage: python3 -I logo.py <repo root>
"""
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import icons  # noqa: E402

GLYPHS = {
    "F": ["11111", "10000", "10000", "11110", "10000", "10000", "10000"],
    "R": ["11110", "10001", "10001", "11110", "10100", "10010", "10001"],
    "E": ["11111", "10000", "10000", "11110", "10000", "10000", "11111"],
    "B": ["11110", "10001", "10001", "11110", "10001", "10001", "11110"],
    "I": ["01110", "00100", "00100", "00100", "00100", "00100", "01110"],
    "S": ["01111", "10000", "10000", "01110", "00001", "00001", "11110"],
}
WORD = "FREEBIES"
BLOCK = 12          # px per letter pixel
GAP = 1             # letter pixels between glyphs
SHADOW = 4          # drop-shadow offset in px
MARK_SCALE = 3      # 32x32 mark -> 96px
PAD = 8
VARIANTS = {
    # name: (letters, shadow)
    "dark": ("#F8B800", "#4428BC"),   # gold on dark backgrounds
    "light": ("#4428BC", "#F8B800"),  # purple on light backgrounds
}


def runs(rows, color, x0, y0, size):
    """Merge horizontal runs of lit pixels into rects."""
    out = []
    for y, row in enumerate(rows):
        x = 0
        while x < len(row):
            if row[x] == "1":
                e = x
                while e < len(row) and row[e] == "1":
                    e += 1
                out.append(f'<rect x="{x0 + x * size}" y="{y0 + y * size}" width="{(e - x) * size}" height="{size}" fill="{color}"/>')
                x = e
            else:
                x += 1
    return out


def mark_rects(x0, y0, scale):
    grid = icons.scene()
    out = []
    for y, row in enumerate(grid):
        x = 0
        while x < len(row):
            c, e = row[x], x + 1
            while e < len(row) and row[e] == c:
                e += 1
            out.append(f'<rect x="{x0 + x * scale}" y="{y0 + y * scale}" width="{(e - x) * scale}" height="{scale}" fill="{c}"/>')
            x = e
    return out


def word_rows():
    rows = [""] * 7
    for i, ch in enumerate(WORD):
        for r in range(7):
            rows[r] += ("0" * GAP if i else "") + GLYPHS[ch][r]
    return rows


def logo(letters, shadow):
    mark = icons.N * MARK_SCALE
    rows = word_rows()
    word_w, word_h = len(rows[0]) * BLOCK, 7 * BLOCK
    x_word = PAD + mark + 24
    height = mark + 2 * PAD
    y_word = (height - word_h) // 2
    width = x_word + word_w + SHADOW + PAD
    parts = [f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} {height}" shape-rendering="crispEdges" role="img" aria-label="Freebies: LA Sports Deals"><title>Freebies: LA Sports Deals</title>']
    parts += mark_rects(PAD, PAD, MARK_SCALE)
    parts += runs(rows, shadow, x_word + SHADOW, y_word + SHADOW, BLOCK)
    parts += runs(rows, letters, x_word, y_word, BLOCK)
    return "".join(parts) + "</svg>\n"


def mark_only():
    size = icons.N * MARK_SCALE
    return (f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {size} {size}" shape-rendering="crispEdges" role="img" aria-label="Freebies: LA Sports Deals"><title>Freebies: LA Sports Deals</title>'
            + "".join(mark_rects(0, 0, MARK_SCALE)) + "</svg>\n")


def main():
    if len(sys.argv) != 2:
        raise SystemExit(__doc__)
    out = os.path.join(sys.argv[1], "asset")
    os.makedirs(out, exist_ok=True)
    for name, (letters, shadow) in VARIANTS.items():
        path = os.path.join(out, f"logo-{name}.svg")
        with open(path, "w") as f:
            f.write(logo(letters, shadow))
        print(f"wrote {path}")
    path = os.path.join(out, "mark.svg")
    with open(path, "w") as f:
        f.write(mark_only())
    print(f"wrote {path}")


if __name__ == "__main__":
    main()
