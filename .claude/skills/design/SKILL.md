---
name: design
description:
  Design the Freebies app in its 8-bit (NES-style) look. Covers building and reviewing app screens
  and components in apps/mobile (NES palette tokens, Press Start 2P plus system-font type, pixel
  borders, ▶ menus, dialog boxes, cut-ins, motion, accessibility, UI copy) and drawing, previewing
  and shipping pixel sprites such as deal food items, sport badges and glyphs, which the API serves
  so they reach the app without a release. Use when adding or changing UI, auditing a screen,
  writing UI copy, or creating, changing or mapping a sprite.
compatibility:
  Designed for Claude Code in the freebies repo (Expo app in apps/mobile, Go API in services/api).
  Needs Python 3, mise, curl, and grep.
metadata:
  author: bitsugar
  version: "1.0"
---

# design

The app is dark-only and drawn as if for an NES sports game, but built for touch and accessibility.
It must feel retro without copying any game. Sprites are data (rows of palette keys plus a
key-to-color map) in the `sprites` table; the app draws them with `PixelSprite` and the API renders
them to PNG. The source of truth for sprites is `scripts/sprites.py`.

## Route

| Need                                                                           | Read                                                           |
| ------------------------------------------------------------------------------ | -------------------------------------------------------------- |
| UI tokens, type, grid and component rules                                      | [references/ui-rules.md](references/ui-rules.md)               |
| Turning an NES pattern into an app component (menus, cut-ins, HUD)             | [references/ui-patterns.md](references/ui-patterns.md)         |
| Checking a screen before calling it done                                       | [references/ui-review.md](references/ui-review.md)             |
| Sprite drawing rules (size, outline, light, palettes, what's forbidden)        | [references/sprite-rules.md](references/sprite-rules.md)       |
| NES color indices and hex values                                               | [references/sprite-palette.md](references/sprite-palette.md)   |
| Adding, changing or mapping a sprite, step by step                             | [references/sprite-workflow.md](references/sprite-workflow.md) |
| What the approved screens and art look like                                    | `docs/pivot/la-jumbotron-vs-8bit.html` (open in a browser)     |
| Where deal, team, sprite and badge data must live (never hardcoded in the app) | `docs/architecture.md#server-driven-content`                   |
| Anything that conflicts with the above                                         | `docs/ui.md` wins                                              |

## Workflow

1. Pick the job (UI or sprite) and read its rules file from the table above.
2. **UI:** reuse the components in `apps/mobile/src/components/pixel/`. Use theme tokens only; Press
   Start 2P only for short display text.
3. **Sprite:** draw a `SPRITES` entry in `scripts/sprites.py`, run `scripts/preview.sh` and look at
   the PNG, map deals, then `scripts/ship.sh <slug>`.
4. Check: `scripts/audit-colors.sh`, `scripts/contrast.py` for new color pairs, and every item in
   `references/ui-review.md`.
5. Verify on screen: typecheck `apps/mobile`, then look at it in the simulator (the `ops` skill
   starts the API and the app). For sprites, `scripts/fetch.sh <id>`.
6. Walk the owner through it as you go: one plain line per step on what it does and why, then a
   short summary of what changed, what you looked at and what is left.

## Scripts

- `scripts/audit-colors.sh [path]`: lists hex colors in app source that aren't NES tokens.
- `scripts/contrast.py <fg> <bg>`: prints the WCAG contrast ratio and whether it passes for small or
  large text.
- `scripts/preview.sh [out.png]`: validates every sprite and renders an 8× contact sheet.
- `scripts/ship.sh <slug>`: writes the next numbered migration from `sprites.py` and runs the sprite
  tests.
- `scripts/fetch.sh <id> [scale] [out.png]`: downloads a sprite PNG from the local API.
- `scripts/sprites.py`: sprite source of truth (`sql` and `sheet` subcommands).
- `scripts/icons.py`: the app icon (Trophy Burger over victory rays). `assets <dir>` writes the Expo
  icon files, `logoart <file>` the in-app logo, `preview <png>` a 16x preview.
- `scripts/logo.py <repo root>`: the README logo, with the icon mark clipped to rounded corners.
