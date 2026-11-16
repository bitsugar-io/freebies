# UI rules

The binding rulebook for the Freebies: LA Sports Deals interface. Every screen, component, sprite
and push in the mobile app follows it. Where it conflicts with an older doc (including the files in
`docs/pivot/`), this file wins. A change to a rule here is a PR to this file first, then the code.

The look is NES-era 8-bit: the real NES master palette, square pixels at integer scale, stepped
motion, sprites with three colors and a black outline. "NES" is fine in this internal doc. It never
appears in the app, the App Store listing or marketing (see [Legal guardrails](#legal-guardrails)).

Background research: [nes-style-research.md](pivot/nes-style-research.md) and
[design-review.md](pivot/design-review.md). The DB32 colors in the design review are replaced by the
NES tokens below, and the font choice in [Type](#type) overrides both docs.

## Palette

Sixteen tokens, all from the NES master palette (Lospec 55-color set). No other colors exist in the
app. Code references tokens, never hex values.

| Token                 | NES | Hex       | Role                                                   |
| --------------------- | --- | --------- | ------------------------------------------------------ |
| `bg.night`            | $0F | `#000000` | App background, splash, sprite outline                 |
| `bg.panel`            | $0C | `#004058` | PixelBox and card fill                                 |
| `bg.panelAlt`         | $02 | `#0000BC` | Header bar, HUD strip, selected row                    |
| `text.primary`        | $30 | `#FCFCFC` | Text, borders                                          |
| `text.secondary`      | $10 | `#BCBCBC` | Secondary text                                         |
| `text.muted`          | $00 | `#7C7C7C` | Disabled, hints. On `bg.night` only                    |
| `accent.primary`      | $28 | `#F8B800` | Primary action, ▶ cursor, selected cell, key words     |
| `accent.primaryLight` | $38 | `#F8D878` | Highlights, sprite shine                               |
| `state.success`       | $2A | `#58D854` | Deal unlocked, wins                                    |
| `state.successDeep`   | $09 | `#007800` | Success fills under white text, turf                   |
| `state.alert`         | $16 | `#F83800` | Live dot, errors, offline icon                         |
| `state.alertDeep`     | $06 | `#A81000` | Alert fills under white text                           |
| `state.info`          | $21 | `#3CBCFC` | Times, links, "upcoming"                               |
| `state.timer`         | $27 | `#FCA044` | Countdowns, "in progress", expiring under 6 hours      |
| `ice.light`           | $31 | `#A4E4FC` | Cool highlights, ice and sky surfaces                  |
| `skin.neutral`        | $36 | `#F0D0B0` | Sprite skin only. $27 and $17 are alternate skin tones |

### Contrast

Ratios are WCAG 2.x, computed from the hex values. Small text needs 4.5:1. Display text at 24pt bold
or larger, and icons, need 3:1.

| Foreground            | on `bg.night` | on `bg.panel` | on `bg.panelAlt` |
| --------------------- | ------------- | ------------- | ---------------- |
| `text.primary`        | 20.5          | 10.9          | 11.9             |
| `text.secondary`      | 11.1          | 5.9           | 6.4              |
| `text.muted`          | 5.0           | **2.7 fail**  | **2.9 fail**     |
| `accent.primary`      | 11.8          | 6.3           | 6.9              |
| `state.success`       | 11.4          | 6.1           | 6.6              |
| `state.alert`         | 5.6           | **3.0 large** | **3.2 large**    |
| `state.info`          | 9.8           | 5.2           | 5.7              |
| `state.timer`         | 10.3          | 5.5           | 5.9              |
| `ice.light`           | 15.1          | 8.0           | 8.7              |
| `accent.primaryLight` | 15.1          | 8.0           | 8.7              |

Fills with text on them: `bg.night` on `accent.primary` 11.8, `bg.night` on `state.success` 11.4,
`text.primary` on `state.successDeep` 5.6, `text.primary` on `state.alertDeep` 7.5.

Rules:

1. `text.muted` is used on `bg.night` only.
2. `state.alert` is never small text on a panel. Use it for icons, the live dot and large labels.
3. Text on `accent.primary` is `bg.night`, never white.
4. `state.successDeep` and `state.alertDeep` are fills. They are never text on black (3.7 and 2.7).
5. State is never shown by color alone. Pair color with a word or an icon ("ON", "LIVE", ✓).
6. No opacity, blend modes or gradients on art. The only opacity in the app is the modal scrim,
   which is `bg.night`.
7. Team identity is never encoded by color. The `events.team_color` column is not rendered.
8. Dark only. There is no light theme in 2.0. A future light set would be a second token map, never
   a per-screen fork.

## Grid and pixel scale

- One art pixel renders as **P** points, and P is an integer: 2, 3, 4, 6 or 8. Never fractional,
  never computed from a size that the sprite width doesn't divide.
- Images scale nearest-neighbor only. Sprites render as `react-native-svg` rects with
  `shapeRendering="crispEdges"`, and every sprite lands on whole points.
- Layout spacing is a 4pt grid: 4, 8, 12, 16, 24, 32. The screen gutter is 16, set once by the
  screen, not by each child.
- Borders are 4pt (one art pixel at P=4), in `text.primary` unless the component says otherwise.
- Corners are notched (one art pixel cut out), never `borderRadius`. Radius is 0 everywhere.
- No shadows or blur. A pressable box gets a 4pt `bg.night` drop offset down-right instead.
- Layout follows the device and its safe areas. Nothing is drawn on a fixed 256×240 canvas.
- Native chrome stays native: headers, sheets, keyboard, share sheet and system alerts are skinned
  through their style props, not rebuilt. The tab bar is the one exception: it is the custom HUD tab
  bar (see [Navigation](#navigation)).

| Use                       | Sprite  | Scale | Rendered |
| ------------------------- | ------- | ----- | -------- |
| UI icon (tab, bell, gear) | 12×12   | 2     | 24pt     |
| Deal row                  | 16×16   | 3     | 48pt     |
| Deal card, team detail    | 16×16   | 4     | 64pt     |
| Deal detail hero          | 16×16   | 6     | 96pt     |
| Cut-in                    | ≤ 64×64 | 2–8   | ≤ 128pt  |
| Splash palm               | 24×24   | 8     | 192pt    |

## Sprites

Food sprites are data, not image files. Each one is rows of palette keys plus a key-to-hex map,
stored in the `sprites` table and served by `GET /api/v1/sprites` (all, for caching),
`GET /api/v1/sprites/{id}` (JSON) and `GET /api/v1/sprites/{id}.png?scale=N` (PNG, for push
attachments and sharing). Events point at a sprite through `events.sprite_id`. A deal with no sprite
gets `gift`.

Sprites are drawn and shipped with the `design` Claude skill (`.claude/skills/design/`), whose
script is the source of truth. Never edit the generated migration by hand.

Rules for every sprite:

1. Canvas is 16×16 for food and items. Draw within 14×14 so a 1 px margin remains. Athletes are
   16×24 (16×32 for basketball and batters). Badges are 8×8.
2. Colors come only from the NES master palette (the `NES` map in the design skill).
3. Each palette is 3 colors plus transparent. An object uses at most 2 palettes, sharing the black
   outline, so at most 6 fill colors. The skill's script rejects more.
4. A 1 px black ($0F) outline around the silhouette.
5. Light comes from the upper left. Highlights are 1–3 px clusters on the top-left of a shape. No
   gradients, no anti-aliasing, no partial transparency, no dithering at 16×16 beyond a 2 px checker
   for texture.
6. The silhouette must read at 1×. Exaggerate the identifying feature (burger layers, taco curve).
7. Generic food only. Never a restaurant's packaging, wrapper, logo, mascot or signature product
   shape. The deal text can name the product ("$2 McGriddles"); the sprite is a "griddle cake
   sandwich".
8. Athletes wear neutral generic jerseys ($16 home, $21 away). No team colors, stripes, logos,
   uniform patterns or real player numbers. Women's teams get their own sprite variants, not
   recolors of the men's.
9. Every sprite is original. Nothing traced, ripped or "close to" a real game's art.
10. A missing sprite falls back to `gift`, never to an emoji.

## Type

| Token        | Font                      | Size / line | Use                                                  | Max scale |
| ------------ | ------------------------- | ----------- | ---------------------------------------------------- | --------- |
| `display`    | Press Start 2P, UPPERCASE | 24 / 32     | Cut-in banners ("GOAL!"), scores, hero countdown     | 1.2       |
| `title`      | Press Start 2P, UPPERCASE | 16 / 24     | Screen titles, HUD bar                               | 1.2       |
| `label`      | Press Start 2P, UPPERCASE | 12 / 20     | ▶ menu items, tab labels, section labels, short tags | 1.3       |
| `body`       | System font               | 17 / 24     | Every sentence                                       | unbounded |
| `bodyStrong` | System font, semibold     | 17 / 24     | Offer names in rows                                  | unbounded |
| `caption`    | System font               | 13 / 18     | Secondary lines                                      | unbounded |

The display font is **Press Start 2P**. This is the owner's decision (November 2026). It is licensed
under the SIL Open Font License 1.1, so it may be bundled in a commercial app; ship the OFL text on
the in-app licenses screen, and don't sell the font by itself or rename a modified copy "Press Start
2P". Provenance note from the research: the font is reportedly based on a 1980s Namco arcade ROM
font. In the US, typeface designs are generally not copyrightable, so the risk is low. That is not
legal advice.

Rules:

1. Two fonts only: Press Start 2P for display text and the platform system font for everything else.
   No other pixel fonts, and no font downloaded from a site that names it after a game.
2. Press Start 2P is for 3 words or fewer in uppercase, plus numbers people read at a glance
   (scores, countdowns). Sentences are always the system font.
3. Anything in Press Start 2P that a person must read is at least 12pt on device. Sizes are 12, 16,
   24 or 32 (multiples of its 8 px grid render crisp at 2× and 3×; 12 is crisp at 2× and acceptable
   at 3×).
4. On Android set `includeFontPadding: false` and an explicit `lineHeight` on Press Start 2P text.
5. Body and caption text scale with Dynamic Type without limit, and every layout wraps. No fixed
   heights on text containers.
6. Press Start 2P text scales up to its cap, then switches to the system font bold at accessibility
   text sizes.

## Components

| Component        | What it is                             | Rules                                                                                                                                                                                                                                                            |
| ---------------- | -------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `PixelSprite`    | Draws a sprite from API data           | Takes an integer `scale`. Merges row runs into rects, `crispEdges`, memoized by sprite id. `accessible={false}` when a text label is adjacent, otherwise the label is the sprite's `name`                                                                        |
| `PixelBox`       | Bordered panel                         | 4pt `text.primary` border, notched corners, `bg.panel` fill, 16pt padding. Variants: `panel`, `deep` (`bg.night`), `selected` (`accent.primary` border). One box per object, not per section                                                                     |
| `PixelButton`    | Primary action                         | `accent.primary` fill, `bg.night` label in `label`, 4pt drop. Pressed: moves 4pt down-right and loses the drop, instantly. Min height 48. Role `button`                                                                                                          |
| `DialogBox`      | RPG-style text box                     | `deep` PixelBox, `body` text, key words in `accent.primary`. A 2-frame bobbing ▼ when more text follows. Tap advances or completes. The accessibility label holds the full text at once                                                                          |
| `PixelMenu`      | ▶ cursor list                          | Full-width rows, min 48pt, `label` text. ▶ marks the focused or pressed row and is decorative (`accessibilityElementsHidden`). Roles `menu` and `menuitem`. Replaces stacks of buttons                                                                           |
| `DealBar`        | 12-cell bar                            | Mode `expiry` drains with time left, `state.success` until 6 hours, then `state.timer`. Mode `progress` fills toward a numeric trigger (5 of 7 strikeouts). Role `progressbar` with a text value                                                                 |
| `TeamSelectGrid` | Team picker                            | 4 columns, one cell per team: team name in text, league or season caption, a generic sport sprite. No logos, no team colors. States: off, followed, partial ("1/2"), off-season (grey, still selectable). Role `checkbox` with `checked: true / false / 'mixed'` |
| Cut-in           | Full-screen moment when a deal unlocks | Banner in `display` ("DUCKS WIN!", "GOAL!"), then the food sprite rises 8 art px over 8 frames ("item get"). At most 2 seconds, then the deal detail. Tap skips. Decorative; the DialogBox carries the meaning                                                   |
| `PixelToast`     | Transient message                      | `deep` strip above the tab bar, 3 seconds, announced politely to screen readers                                                                                                                                                                                  |
| `PixelSkeleton`  | Loading placeholder                    | Box outlines at final size with a 2×2 `bg.panel` / `bg.night` checker. No full-screen spinners                                                                                                                                                                   |

Art and team/league data come from the API; bundled art is offline fallback only. A new deal, team,
sport or sprite must never need an app release, so the app never hardcodes a team list, league map
or sprite list.

### Navigation

Four tabs in a custom HUD tab bar (React Navigation `tabBar` prop): **WON**, **TODAY**, **TEAMS**,
**YOU**. TODAY lists deals anyone can claim right now (any LA team) and today's LA games with the
deals riding on each; it comes from `GET /today`. The bar is black with a 3pt white top border, a
stat line (`WON n`, `FOLLOWING n`, separated by thin pixel bars, and a blinking `PAUSED` when alerts
are paused), then the tabs in Press Start 2P. The active tab is accent with a blinking ▶. The won
count is a blinking badge on WON. `FOLLOWING n` counts only followed deals that are still active.
Targets are at least 44pt, tabs have role `tab` with `selected` state, and the bar sits above the
bottom safe-area inset.

- **WON** (home) shows only deals the user has won and can still claim. Nothing grey, no not-yet-won
  deals. Empty state: "NOTHING WON YET" and a PICK TEAMS button.
- **TEAMS** is sport first. SELECT SPORT is a ▶ menu with one row per sport, derived from
  `GET /teams` (group by `sport`), then a sport screen with that sport's team tiles and a per-offer
  ✓ list. Leaving the tab or pressing TEAMS again returns to SELECT SPORT. Deep link
  `/teams/<sport>` opens a sport directly.
- **YOU** has the season card, TROPHY CASE (claimed deals), alerts, About and licenses.
- Every screen's HUD bar has a START button (label "Menu"). It opens the PAUSE overlay: PAUSE ALERTS
  (1 HOUR, UNTIL TOMORROW 8am PT, UNTIL I RESUME), ALERT SETTINGS, ABOUT (opens YOU scrolled to
  About), BACK. While paused the first row reads RESUME ALERTS; BACK only closes the menu. It is a
  transparent modal route with no animation.
- Old `/tonight` and `/deals` links open WON.
- First launch only, a title screen shows "FREEBIES" with "LA SPORTS DEALS" under it, the logo, a
  blinking "TAP TO START" and "WHEN LA WINS, YOU WIN". Tap, or 4 seconds, continues (no timer while
  a screen reader is on). Under Reduce Motion it still shows, solid.

### Logo art

The palm sunset logo (`pixel/logoArt.ts`, generated from the icon script's `scene()`) uses four
extra NES master-palette entries: `#4428BC`, `#6844FC`, `#940084`, `#0058F8`. They are logo-only art
colors, not UI tokens. `audit-colors.sh` allows them in that one file.

### Team tiles and team colors

A team tile is a primary-color field holding the sport badge, a 6pt secondary stripe, then a black
label strip with the name and status ("1 DEAL", "NO DEAL YET", "BACK MAR", in system bold). The
primary and secondary swap only when the primary is effectively black (contrast against `#000000`
under 1.4: Kings, LAFC, Angel City); navy and other dark-but-visible colors keep their order. Text
is always on the black strip, so contrast never depends on team colors. A 3pt border is always
drawn. Off season: the two colors and the badge go to greyscale, a 2px black checker covers the
field and an `OFF` tag sits in the corner, so a grey or silver in-season team never reads as off
season. A team with no deal shows "NO DEAL YET" in muted text and a toast on tap. Followed: white
border and a blinking gold ✓ in a black badge that blinks as a whole, background included (partial:
a solid "1/2"). One font size for every tile name; long names wrap to two centered lines in a fixed
box.

Team colors are allowed only on team tiles, to tell teams apart. They are never the app's brand or
chrome, they are always NES-snapped, and they are served by the API (`team.colors`), never
hardcoded, so `audit-colors.sh` has nothing to allow. Open question for the owner: this sits against
legal guardrail 4 and palette rule 7 ("no team color schemes", "team identity is never encoded by
color"); see the note under [Legal guardrails](#legal-guardrails).

## Motion

- Frame unit is 1/60 s. Motion is stepped holds of 2, 4, 8 or 16 frames. No easing curves, no
  springs, no tweened transforms, no parallax.
- Buttons, cursor moves and toggles are instant (0 frames).
- Cut-ins last at most 2 seconds (120 frames) and auto-advance.
- Blinks run at 1 Hz or 2 Hz. Nothing flashes more than 3 times per second, and nothing simulates
  sprite flicker.
- Screen transitions are hard cuts or an 8-band wipe over 8–16 frames. Native stack transitions are
  fine for pushes. Switching tabs plays a stepped black wipe (8 steps, about 240ms); it is off under
  Reduce Motion.
- **Blinking** runs on ONE shared clock (`useBlink()` and `<Blink>` in `hooks/useBlink.tsx`): 1 Hz,
  stepped, a single interval for the whole app, solid under Reduce Motion. Rules: blink only
  selection cursors (▶), tags (`NEW!` on deals unlocked in the last 24 hours, `LIVE` on a game in
  progress; `AT THE GAME` is a fact and never blinks), status words (`ENDS SOON`, `PAUSED`,
  `TAP TO START`), the unlocked count badge, the followed ✓, and the countdown colons. Never blink
  body text. Never flash content faster than 1 Hz. Never more than 3 blinking things on screen at
  once: each blinker has a priority and only the top 3 on the focused screen blink, the rest stay
  solid (priorities are listed in `useBlink.tsx`). Loading skeleton tones and the ▼ bob run on the
  raw clock and are not counted.
- Prefer palette swaps (two colors alternating) to moving pixels.
- Reduce Motion: every animation has a static fallback. No blinking, typewriter, cut-in, bob or
  count-up. Haptics stay.
- Haptics only at moments: `selectionAsync` on team toggle, `notificationAsync(Success)` once per
  unlocked deal. Never on plain taps.

## Accessibility

1. Tap targets are at least 44×44pt on iOS and 48×48dp on Android, whatever the sprite size. Menu
   rows are 48pt.
2. Small text meets 4.5:1. Display text at 24pt bold and icons meet 3:1. Use only the pairs in
   [Contrast](#contrast).
3. Every pressable has `accessibilityRole`, a label, and `accessibilityState` when it has state.
4. VoiceOver labels are full sentences: "Kings 2, Sharks 1, third period. A home win unlocks a free
   burrito." Decorative pixels (▶, banners, crowd) are hidden.
5. Countdowns read as "3 hours 28 minutes left" and update for screen readers once a minute.
6. Dynamic Type is supported as in [Type](#type).
7. Reduce Motion is honored as in [Motion](#motion).
8. Color never carries state alone (see contrast rule 5).

## Per-sport conventions

Each sport gets a cut-in word, a progress motif and a generic background. All art is original.

| Sport      | Teams                       | Cut-in       | Progress motif                       | Background                          |
| ---------- | --------------------------- | ------------ | ------------------------------------ | ----------------------------------- |
| Baseball   | Dodgers, Angels             | `HOME RUN!`  | Line score, pips toward a stat (7 K) | Dirt $17/$27, grass $1A             |
| Basketball | Lakers, Clippers, Sparks    | `SLAM DUNK!` | Quarter tracker                      | Wood court $27/$17                  |
| Football   | Rams, Chargers, USC, UCLA   | `TOUCHDOWN!` | Quarter tracker                      | Turf $1A with white yard lines      |
| Hockey     | Kings, Ducks                | `GOAL!`      | Period lights                        | Ice $30, red $16 and blue $21 lines |
| Soccer     | LAFC, Galaxy, Angel City FC | `GOAL!`      | Half tracker                         | Striped pitch $1A/$2A               |

Win cut-ins can also use `<TEAM> WIN!` with the team name in text. Non-sports triggers, when they
come, get their own cut-in word and use the `gift` sprite until they have art.

## Copy voice

One idea, one word:

| Idea                              | Word          | Where                                                                                         |
| --------------------------------- | ------------- | --------------------------------------------------------------------------------------------- |
| Teams or deals you subscribed to  | `FOLLOWING`   | Stat line `FOLLOWING n`, team screens `n FOLLOWING` or `1/2 FOLLOWING`, YOU "Following Ducks" |
| Deals you won and can still claim | `WON`         | Tab, stat line `WON n`, "back in Won"                                                         |
| Deals you tapped GOT IT on        | `CLAIMED`     | Trophy case, "DEALS CLAIMED"                                                                  |
| A won deal whose code works later | `LOCKED`      | Row: `CODE UNLOCKS TUE` (Pacific day name, or `NOV 24` past 6 days)                           |
| Only fans at the venue can claim  | `AT THE GAME` | Black-on-ice tag on rows and the deal screen, plus "Only for fans at Intuit Dome" in words    |
| A game in progress                | `LIVE`        | White on deep red tag in TODAY, blinking                                                      |

Never "Watching", "Picked", "Unlocked" (as a place) or "Claim n". "Unlock" only describes a code
becoming usable.

A locked deal (the API's `availableAt` is in the future) shows the sprite and headline, the status
`CODE UNLOCKS TUE` instead of time left (no ENDS SOON or NEW), and on its screen "Your code unlocks
Tue Nov 17 at 12:00am PT" with a countdown to the unlock. OPEN and GOT IT stay hidden until then;
SLIDE TO STOP REMINDERS stays. It flips to the normal state on its own at `availableAt`.

The pitch is "if X happens, you win Y". The tagline is **When LA wins, you win.** Core copy talks
about LA moments and rewards, not only sports, so other LA triggers fit later.

- Lead with the outcome and the reward: "Ducks won. $2 McGriddles unlocked."
- Use the real number when the trigger is numeric: "9 strikeouts", "scored 6".
- Say where and by when: "McDonald's app, Deals tab. Good until 10:30am."
- Short and plain. No exclamation marks in pushes or body text. Display banners may use one.
- Game words in moderation: "unlocked", "item", "stage clear" in display text. Sentences stay plain.
- Missed deals are neutral, not punitive: "No freebie today. Next chance: Tue 7:00pm."
- No emoji anywhere in the app or in pushes.
- Push titles stay under about 40 characters so the lock screen doesn't cut them off.

## Legal guardrails

Not legal advice. When in doubt, leave it out and ask the owner.

1. Never use "Nintendo", "NES", "Famicom" or "Nintendo Entertainment System" in the app, its name,
   icon, screenshots, App Store metadata or marketing. Say "8-bit", "retro" or "pixel art".
2. No NES controller, console, cartridge or Seal of Quality shapes. No controller-shaped UI and no
   cartridge icon.
3. No characters, sprites, tiles, fonts, music or sounds from any real game, traced or close enough
   to recognize.
4. No team or league logos, wordmarks, mascots, uniforms or color schemes, and no school marks.
   Teams are named in plain text, as a fact about which games we track.
5. No restaurant logos, packaging or wrapper art. Deals are described in plain text.
6. The App Store description and the in-app About screen carry this disclaimer: "Freebies: LA Sports
   Deals is not affiliated with, endorsed by or sponsored by any team, league, school or restaurant.
   Team and restaurant names are used only to describe the deals they offer. Deals are set by each
   restaurant and can change or end without notice."
7. Team names never go in the app name or icon (App Store guidelines 4.1 and 5.2.1).
8. Only verified deals are active. A deal we can't confirm on the sponsor's or team's own page ships
   with `is_active = 0`.

Owner exception, pending legal review: team tiles show two NES-snapped colors per team from the API
(`team.colors`). Guardrail 4 ("no team color schemes") and palette rule 7 still apply everywhere
else, and this exception needs a legal check before release.
