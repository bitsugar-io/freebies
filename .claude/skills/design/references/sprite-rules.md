# Sprite rules

Condensed from `docs/ui.md` and `docs/pivot/nes-style-research.md` (sections 1, 5, 6, 7.3).
`docs/ui.md` wins on any conflict.

## Shape

- **Canvas:** 16×16 for food items and sport badges. Athletes are 16×24 (16×32 for tall poses). Hero
  cut-ins are at most 64×64.
- Keep a 1 px transparent margin where the art allows. `.` is transparent.
- The silhouette must read at 1× (16 px). Exaggerate the identifying feature: burger layers, the
  curve of a taco shell, fries poking out of a carton.

## Color

- Only NES master-palette colors, referenced by index in `sprites.py` (see `palette.md`).
- **Black 1 px outline** (`K` = `0F`) around the silhouette.
- At most **two 3-color palettes** per item: 6 fill colors plus the shared black. `sprites.py`
  rejects more.
- Light comes from the **upper left**: a 1–3 px highlight cluster. No gradients, no anti-aliasing,
  no partial transparency. The only dithering allowed is a 2 px checker for texture.

## Content

- **Generic food only.** No restaurant packaging, logos, wrappers or mascots. Deal text may name the
  product ("$2 McGriddles"); the sprite stays generic ("griddle cake sandwich").
- No team logos, uniforms or team colors. Athletes wear neutral colors.
- Nothing traced from, or recognizable as, a real game's sprite or character.
- User-facing copy never says "Nintendo" or "NES"; say "8-bit" or "retro".
- Women's teams get their own athlete silhouettes, not recolors.

## Naming

- `id` is a kebab-case noun: `plate-lunch`, `hot-dog`.
- `name` is plain English: "Plate lunch".
- `gift` is the fallback for any deal without a sprite. Never delete it.
