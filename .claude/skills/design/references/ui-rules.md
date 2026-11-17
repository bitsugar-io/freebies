# UI rules

Condensed from `docs/ui.md` and `docs/pivot/nes-style-research.md` (sections 1.3, 3, 7).
`docs/ui.md` wins on any conflict.

## Color tokens

Every value comes from the NES master palette. App code references tokens from
`src/hooks/useTheme.tsx`, never raw hex values.

| Token          | Hex     | Role                                   | Contrast limits                   |
| -------------- | ------- | -------------------------------------- | --------------------------------- |
| bg.night       | #000000 | App background                         | —                                 |
| bg.panel       | #101010 | Boxes, cards (graphite)                | —                                 |
| bg.panelAlt    | #2A2A2A | Selected, card header strips           | —                                 |
| line.divider   | #3A3A3A | Row dividers                           | —                                 |
| line.border    | #A5ACAF | Pixel borders (silver)                 | —                                 |
| accent.shadow  | #8A6410 | Drop under gold buttons                | —                                 |
| text.primary   | #FCFCFC | Text, pixel borders                    | OK on every background            |
| text.secondary | #BCBCBC | Secondary text                         | OK on night and panel             |
| text.muted     | #7C7C7C | Hints, disabled                        | 4.6:1 on panel, 4.8:1 on night    |
| accent         | #F8B800 | Actions, ▶ cursor, scores              | Black text on it, never white     |
| accent.light   | #F8D878 | Shine                                  | —                                 |
| success        | #58D854 | Deal won                               | —                                 |
| success.deep   | #007800 | Fills under white text, turf           | —                                 |
| alert          | #F83800 | Live dot, errors                       | Icons or large text only on panel |
| alert.deep     | #A81000 | Fills under white text                 | —                                 |
| info           | #D8B8F8 | Lilac: times, upcoming, section labels | —                                 |
| timer          | #FCA044 | Countdowns, expiring soon              | —                                 |
| ice            | #A4E4FC | Small labels, ice                      | —                                 |
| skin           | #F0D0B0 | Sprites only                           | —                                 |

- No color carries meaning on its own; pair it with text or an icon.
- Team identity is never shown by color, with one owner-approved exception: team tiles show the
  API's NES-snapped `team.colors` (never hardcoded, never app chrome). See `docs/ui.md`.

## Type

- **Press Start 2P**, uppercase, for display text: screen titles, HUD, scores, cut-ins, tab labels
  and ▶ menu items.
  - At most about 3 words.
  - At least 12pt for anything a user must read.
  - Switch to the system font at accessibility text sizes.
  - The owner chose this font. It is OFL-licensed, so ship the license text in the app.
- **System font** for body, descriptions, deal terms and anything longer than a label. Honor Dynamic
  Type.

## Grid and shape

- One art pixel is P points, where P is 2, 3 or 4. Integers only.
- One tile is 8 art pixels. Spacing comes in ½, 1 and 2 tiles.
- Pixel borders are 3pt white on panels. Corners are notched or square; never use `borderRadius` on
  pixel UI.
- Sprites and badges render at whole-number scales only (see `PixelSprite`).
- One font size per component: wrap, don't shrink (no `adjustsFontSizeToFit`, no per-item sizes).
- State on separate channels: color/fill says season, border and ✓ say followed.
- Art and team/league data come from the API; bundled art is offline fallback only.
- Native system UI (sheets, keyboard, share sheet) stays native, with pixel styling inside it.

## Components

These live in `apps/mobile/src/components/pixel/`; reuse them before writing new ones.

- PixelBox
- DialogBox
- PixelMenu (▶ rows)
- HudBar
- DealBar / period lights
- CutIn
- SportBadge
- TeamTile
- PixelSprite

## Motion

- Stepped animation on 2, 4 or 8-frame holds, with no easing.
- Cut-ins last at most 2 seconds and can be dismissed.
- Flashes stay at 3 or fewer per second.
- Blinking: one shared 1 Hz `useBlink()`, never body text, at most 3 blinking things on screen.
- Never simulate sprite flicker.
- Everything has a Reduce Motion path.

## Voice

- Short and plain: "Ducks won! $2 McGriddles." Deadlines in Pacific time.
- Game words in moderation ("GOT IT!", "SELECT TEAMS"). Deal terms stay in plain English.
- Never write "Nintendo" or "NES" in user-facing copy; say "8-bit" or "retro".
- Never show team logos, and never imply affiliation.
