"""Render the README logo (rounded app-icon mark + FREEBIES in Press Start 2P) as SVG.

Writes asset/logo-dark.svg, asset/logo-light.svg and asset/mark.svg at the repo root.
The mark is the app icon scene from icons.py, clipped to iOS-style rounded corners so it reads as
an app icon on GitHub (the uploaded icon file itself stays square). The wordmark is Press Start 2P,
the app's display font (SIL Open Font License), drawn from its 8x8 glyph bitmaps as squares so the
SVG needs no web font. Pure standard library.

Usage: python3 -I logo.py <repo root>
"""
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import icons  # noqa: E402

# Press Start 2P glyphs (rows 0-6 of the 8x8 cell; row 7 and column 7 are the font's own spacing).
GLYPHS = {
    " ": ["00000000", "00000000", "00000000", "00000000", "00000000", "00000000", "00000000"],
    "A": ["00111000", "01101100", "11000110", "11000110", "11111110", "11000110", "11000110"],
    "B": ["11111100", "11000110", "11000110", "11111100", "11000110", "11000110", "11111100"],
    "D": ["11111000", "11001100", "11000110", "11000110", "11000110", "11001100", "11111000"],
    "E": ["11111110", "11000000", "11000000", "11111100", "11000000", "11000000", "11111110"],
    "F": ["11111110", "11000000", "11000000", "11111100", "11000000", "11000000", "11000000"],
    "I": ["01111110", "00011000", "00011000", "00011000", "00011000", "00011000", "01111110"],
    "L": ["01100000", "01100000", "01100000", "01100000", "01100000", "01100000", "01111110"],
    "O": ["01111100", "11000110", "11000110", "11000110", "11000110", "11000110", "01111100"],
    "P": ["11111100", "11000110", "11000110", "11000110", "11111100", "11000000", "11000000"],
    "R": ["11111100", "11000110", "11000110", "11001110", "11111000", "11011100", "11001110"],
    "S": ["01111100", "11000110", "11000000", "01111100", "00000110", "11000110", "01111100"],
    "T": ["01111110", "00011000", "00011000", "00011000", "00011000", "00011000", "00011000"],
}
WORD, TAGLINE = "FREEBIES", "LA SPORTS DEALS"
BLOCK = 11          # px per font pixel, wordmark
TAG_BLOCK = 4       # px per font pixel, tagline
SHADOW = 4          # drop-shadow offset in px
MARK_SCALE = 3      # 32x32 mark -> 96px
RADIUS = 0.2237     # iOS app-icon corner radius as a fraction of the side
PAD = 8

# Official team Pantone colors (owner decision, 2026-11-15), paired the way the teams wear them:
# Lakers purple with Lakers gold (the wordmark and its drop shadow), Dodger blue with white (the
# tagline and its drop shadow). On GitHub's light theme a white shadow vanishes, so the tagline's
# shadow there is Dodger silver, the Dodgers' other official color.
LAKERS_PURPLE, LAKERS_GOLD, DODGER_BLUE, WHITE = "#552583", "#FDB927", "#005A9C", "#FCFCFC"
DODGER_SILVER = "#A5ACAF"
TAG_SHADOWS = {"dark": WHITE, "light": DODGER_SILVER}  # README <picture> source -> tagline shadow
TAG_SHADOW = 2      # tagline drop shadow: straight down, half a font pixel. A diagonal offset
                    # leaves bright slivers inside the letters, since the shadow is lighter than them.


def rects(rows, color, x0, y0, size):
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


def text(word, color, x0, y0, size):
    out = []
    for i, ch in enumerate(word):
        out += rects(GLYPHS[ch], color, x0 + i * 8 * size, y0, size)
    return out


def mark_rects(x0, y0, scale):
    """The icon's pixels as merged rects inside a rounded-corner clip."""
    grid = icons.scene()
    side = icons.N * scale
    out = [f'<clipPath id="icon-{x0}-{y0}"><rect x="{x0}" y="{y0}" width="{side}" height="{side}" rx="{side * RADIUS:.1f}"/></clipPath>',
           f'<g clip-path="url(#icon-{x0}-{y0})">']
    for y, row in enumerate(grid):
        x = 0
        while x < len(row):
            c, e = row[x], x + 1
            while e < len(row) and row[e] == c:
                e += 1
            out.append(f'<rect x="{x0 + x * scale}" y="{y0 + y * scale}" width="{(e - x) * scale}" height="{scale}" fill="{c}"/>')
            x = e
    out.append("</g>")
    return out


def logo(tag_shadow):
    mark = icons.N * MARK_SCALE
    word_w = (len(WORD) * 8 - 1) * BLOCK
    word_h, tag_h = 7 * BLOCK, 7 * TAG_BLOCK
    gap = 3 * TAG_BLOCK
    text_h = word_h + gap + tag_h + TAG_SHADOW
    height = max(mark, text_h) + 2 * PAD
    x_word = PAD + mark + 24
    y_word = (height - text_h) // 2
    width = x_word + word_w + SHADOW + PAD
    y_tag = y_word + word_h + gap
    parts = ['<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {} {}" shape-rendering="crispEdges" role="img" aria-label="Freebies: LA Sports Deals"><title>Freebies: LA Sports Deals</title>'.format(width, height)]
    parts += mark_rects(PAD, (height - mark) // 2, MARK_SCALE)
    parts += text(WORD, LAKERS_GOLD, x_word + SHADOW, y_word + SHADOW, BLOCK)
    parts += text(WORD, LAKERS_PURPLE, x_word, y_word, BLOCK)
    parts += text(TAGLINE, tag_shadow, x_word + 2, y_tag + TAG_SHADOW, TAG_BLOCK)
    parts += text(TAGLINE, DODGER_BLUE, x_word + 2, y_tag, TAG_BLOCK)
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
    for name, tag_shadow in TAG_SHADOWS.items():
        path = os.path.join(out, f"logo-{name}.svg")
        with open(path, "w") as f:
            f.write(logo(tag_shadow))
        print(f"wrote {path}")
    path = os.path.join(out, "mark.svg")
    with open(path, "w") as f:
        f.write(mark_only())
    print(f"wrote {path}")


if __name__ == "__main__":
    main()
