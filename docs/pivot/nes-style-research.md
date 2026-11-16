# NES-Style Art Direction Research (LA Freebies redesign)

Research date: 2026-11-15. Status: input for `docs/ui.md`. Not legal advice.

Legend: **[V]** verified against a cited source during this research. **[U]** unverified, from
general knowledge or memory of the games. Check these before treating them as fact. Sports-game
descriptions are written in words only as references for conventions. Nothing here should be traced
or ripped.

---

## 1. NES hardware constraints that define the look

### 1.1 Facts

| Constraint            | Detail                                                                                                                                                            | Source                                                                                                                                                        |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Resolution            | 256×240 picture region **[V]**. NTSC TVs show about the center 256×224; title-safe area is 224×192 **[V]**                                                        | [NESdev PPU rendering](https://www.nesdev.org/wiki/PPU_rendering), [NESdev Overscan](https://www.nesdev.org/wiki/Overscan)                                    |
| Tiles                 | 8×8 px, 2 bits per pixel (4 indices, index 0 = transparent/backdrop) **[V]**                                                                                      | [NESdev PPU OAM](https://www.nesdev.org/wiki/PPU_OAM)                                                                                                         |
| Sprites               | 64 in OAM; 8×8 or 8×16 mode (8×16 = two stacked tiles) **[V]**                                                                                                    | [NESdev PPU OAM](https://www.nesdev.org/wiki/PPU_OAM)                                                                                                         |
| Sprites per scanline  | 8 (secondary OAM holds 8). Extra sprites drop out, which causes the familiar flicker **[V]**                                                                      | [NESdev PPU OAM](https://www.nesdev.org/wiki/PPU_OAM)                                                                                                         |
| Palettes              | 4 BG + 4 sprite palettes. Each has 3 colors plus a shared backdrop/transparent entry ($3F00) **[V]**                                                              | [NESdev PPU palettes](https://www.nesdev.org/wiki/PPU_palettes)                                                                                               |
| Attribute granularity | Each attribute byte covers 32×32 px, split into four 16×16 quadrants. Each quadrant picks one BG palette **[V]**                                                  | [NESdev Attribute tables](https://www.nesdev.org/wiki/PPU_attribute_tables)                                                                                   |
| Master palette        | 64 indices with about 52–55 distinct colors (duplicate blacks and whites). $0D is "blacker than black" and should not be used. $0F is the canonical black **[V]** | [NESdev PPU palettes](https://www.nesdev.org/wiki/PPU_palettes), [Wikipedia palettes list](https://en.wikipedia.org/wiki/List_of_video_game_console_palettes) |
| Frame rate            | About 60 Hz NTSC (262 scanlines/frame) **[V for 262 lines; 60.0988 Hz figure U]**                                                                                 | [NESdev PPU rendering](https://www.nesdev.org/wiki/PPU_rendering)                                                                                             |
| Metasprites           | Characters are built from several 8×8 or 8×16 hardware sprites. Common sizes are 16×16, 16×24, and 16×32 **[U: convention, widely observed]**                     | n/a                                                                                                                                                           |

### 1.2 Master palette (sRGB approximation)

The NES has **no canonical RGB palette**. The PPU outputs composite video, and the colors you see
depend on the TV's decoder, its hue and saturation settings, and phosphors. NESdev says "no single
composite palette can match the intended display of every possible game title" **[V]**. FirebrandX
publishes measured alternatives ("Composite Direct", "Smooth", "NES Classic") as downloadable files,
without hex values on the page **[V]** ([FirebrandX](https://www.firebrandx.com/nespalette.html)).

We use the widely circulated 55-color set published on Lospec
([Lospec "Nintendo Entertainment System"](https://lospec.com/palette-list/nintendo-entertainment-system)).
The hex set is **[V]** against Lospec. The mapping to PPU index below follows the common
emulator-era table; it matches the Lospec set color-for-color, but the index assignment is **[U]**.
The Lospec page does not attribute the palette to anyone.

| Row | x0     | x1     | x2     | x3     | x4     | x5     | x6     | x7     | x8     | x9     | xA     | xB     | xC     | xD     |
| --- | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ |
| $0x | 7C7C7C | 0000FC | 0000BC | 4428BC | 940084 | A80020 | A81000 | 881400 | 503000 | 007800 | 006800 | 005800 | 004058 | 000000 |
| $1x | BCBCBC | 0078F8 | 0058F8 | 6844FC | D800CC | E40058 | F83800 | E45C10 | AC7C00 | 00B800 | 00A800 | 00A844 | 008888 | 000000 |
| $2x | F8F8F8 | 3CBCFC | 6888FC | 9878F8 | F878F8 | F85898 | F87858 | FCA044 | F8B800 | B8F818 | 58D854 | 58F898 | 00E8D8 | 787878 |
| $3x | FCFCFC | A4E4FC | B8B8F8 | D8B8F8 | F8B8F8 | F8A4C0 | F0D0B0 | FCE0A8 | F8D878 | D8F878 | B8F8B8 | B8F8D8 | 00FCFC | F8D8F8 |

Columns xE and xF are black in every row ($0F is the black to use).

### 1.3 Adopt vs relax for a phone UI

| Constraint                        | Decision                                                                               | Why                                                                                           |
| --------------------------------- | -------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| Master palette                    | **Adopt strictly.** Every color in code is a token mapped to a palette entry           | Color does more for the "NES" read than anything else, and the restriction costs nothing      |
| 3 colors + transparent per sprite | **Adopt for sprites and icons**                                                        | This is what makes art look native. 3 colors per sprite also keeps drawing fast               |
| 8×8 tile grid                     | **Adopt as a layout unit** (1 "tile" = 8 art px × scale)                               | Borders, panels, and spacing on a tile grid read as authentic                                 |
| Integer scaling                   | **Adopt.** Art renders at 2×/3×/4× with nearest-neighbor, never fractional             | Fractional scaling blurs or distorts pixels, which kills the look immediately                 |
| 16×16 attribute color clash       | **Relax.** Mimic it only in decorative backgrounds                                     | Text and data need free color placement                                                       |
| 8 sprites per scanline / flicker  | **Relax entirely.** Never simulate flicker                                             | Flicker is an accessibility hazard (photosensitivity) and has no benefit                      |
| 256×240 canvas                    | **Relax.** Use native phone resolution. Art is pixel-scaled; layout follows safe areas | Phones are 390–440pt wide. Shrinking the UI to 256 px would break readability and tap targets |
| Uppercase-only 8×8 font           | **Relax for body text.** Use it for display text only                                  | Dynamic Type and readability come first                                                       |
| 60 Hz frame stepping              | **Adopt in spirit.** Animate on 2/4/8-frame holds (see §7)                             | Stepped motion is a strong era signal                                                         |
| Hard pixel grid for everything    | **Relax for system UI** (sheets, keyboard, share sheet, maps)                          | Native controls stay native                                                                   |

---

## 2. Sports game visual conventions (references only)

General NES sports conventions **[U unless noted]**:

- Players are stubby 16×24 to 16×32 metasprites, usually about 2.5–3 heads tall. Heads are big and
  read at a glance; limbs are 2–3 px thick.
- The playfield is a large flat color: green turf, white ice, tan or brown court. Line markings are
  1 px. The camera scrolls along the long axis.
- Score, clock, and period sit in a black strip at the top or bottom, in an 8×8 uppercase font.
  Numerals are often in a contrasting color (yellow or white on black).
- Big moments trigger a **cut-in**: the game pauses and shows a large close-up, or full-screen text
  ("TOUCHDOWN!", "GOAL!", "HOME RUN!") with palette flashing or crowd animation.
- Crowds are a repeating 8×8 or 16×16 pattern tile. Animating them by palette cycling or 2-frame
  swaps reads as cheering.
- Team-select screens are grids or lists of cities with a ▶ cursor and a stat bar or letter grade
  per team. Lacking league licenses, many games used city names only.

### Hockey (Kings, Ducks)

- _Ice Hockey_ (1988) offered three skater types: light/fast, average, and heavy/strong. Teams had 5
  per side including the goalie, and contested pucks could turn into fights **[V]**
  ([Wikipedia](<https://en.wikipedia.org/wiki/Ice_Hockey_(1988_video_game)>)). Showing body type as
  a gameplay stat through silhouette (thin, medium, round) is a memorable convention **[U: visual
  depiction]**.
- _Blades of Steel_ (1988) had a separate full-screen fight scene with 5-hit health bars, voice
  samples ("Faceoff", "Fight"), penalty shots, and shootouts **[V]**
  ([Wikipedia](https://en.wikipedia.org/wiki/Blades_of_Steel)).
- View: the rink scrolls vertically in a high three-quarter view, with goals at the top and bottom
  of the scroll **[U]**. Ice is white or light grey with red and blue lines. The scoreboard sits in
  a top strip.
- Translation for us: a 16×24 skater leaning forward with a stick extending 4–6 px past the body and
  a 2-px puck. The ice is $30 or $3x tints with a $16 red line. A "GOAL!" flash on win.

### Basketball (Lakers, Clippers, Sparks)

- _Double Dribble_ (1987) is remembered for a full-screen dunk close-up cut-scene and a side-view
  scrolling court **[U]**.
- Typical view: side-on with slight depth, scrolling horizontally between hoops. Players are taller
  (16×32) to sell height. The court is tan or orange wood with 1 px lines **[U]**.
- HUD: score per team, game clock, and quarter. The shot clock was often absent **[U]**.
- Translation: a 16×32 player mid-dribble (2-frame bounce). A dunk "cut-in" card on win, with the
  rim and net drawn large and stepped at 4 frames.

### Football (Rams, Chargers, USC, UCLA)

- _Tecmo Bowl_ (1989): 4 plays per team (usually 2 run, 2 pass). Rosters were licensed through the
  NFLPA and teams were named after their cities, because LJN held the NFL league license **[V]**
  ([Wikipedia](https://en.wikipedia.org/wiki/Tecmo_Bowl)). This is a useful precedent: show the team
  name factually and use no league or team marks.
- _Tecmo Bowl_ and _Tecmo Super Bowl_ are remembered for cinematic cut-scenes (QB drop-back
  close-up, big-hit and touchdown celebration screens) and a horizontally scrolling field **[U]**.
- Typical view: side-on, scrolling horizontally; green field with white yard lines every 5 yards; 9
  players per side on NES **[V for 9]**. Players are 16×24 with helmets that read as a large rounded
  head shape.
- Translation: a 16×24 QB in a throwing pose (3 frames). A "TOUCHDOWN!" banner with spiral-ball
  sprite. Yard-line stripes as a background pattern.

### Baseball (Dodgers, Angels)

- _R.B.I. Baseball_, _Bases Loaded_, _Baseball Stars_ **[U for all details]**:
  - The batting view is from behind the plate (or behind the pitcher on some titles), centered on
    the pitcher. Fielding cuts to a top-down or high three-quarter overview of the diamond.
  - _Bases Loaded_ used a broadcast-style center-field camera with larger sprites.
  - _Baseball Stars_ added team creation and naming (a "customize your team" screen) and a
    team-roster stat layout.
  - Line-score scoreboards (innings 1–9 across, R/H/E) as a between-innings screen. A "HOME RUN"
    text flash with fireworks.
- Translation: a 16×32 batter with the bat held up (bat as a 1–2 px diagonal). An inning-style line
  score is a strong motif for showing a game's progress toward a deal (e.g. strikeout count).

### Soccer (LAFC, Galaxy, Angel City)

- _Goal!_ (Jaleco, 1989) used a top-down or high three-quarter scrolling pitch **[U]**.
- _Nintendo World Cup_ (Technos, 1990) used super-deformed, chibi players (about 2 heads tall) with
  exaggerated "special shots" **[U]**. Do not copy its character designs (Kunio-kun lineage).
- Typical palette: green pitch in two alternating stripe shades ($1A/$2A or $19/$29) with white 1 px
  lines. A "GOAL!" full-screen flash and net ripple.
- Translation: a 16×24 player mid-kick with a 6×6 ball. Women's teams (Angel City, Sparks) should
  get their own sprite variants (ponytail or hair silhouette, same proportions) rather than palette
  swaps of the men's.

### Recommended unified convention for our app

To keep the app cohesive across sports, all athletes use **one shared body template**: 16 px wide,
24 px tall (32 for basketball and batter "tall" poses), about 2.5 heads, a 3-color palette plus
outline, facing right. Each sport changes only equipment and pose.

---

## 3. Typography

### How NES text worked

- Glyphs were 8×8 tiles from the same pattern table as the graphics, usually 7×7 or 6×7 glyphs with
  a 1 px gutter. Text was monospaced on the 8 px grid. Many games had uppercase only to save tiles;
  numerals 0–9 were always present and often bolder. Punctuation was minimal, and ©/! were custom
  tiles **[U: convention]**.
- Text was 1 color on the backdrop, sometimes with a drop shadow done as a second sprite or palette
  color.

### Open-license candidates

| Font                                  | Designer                    | License                                               | Notes                                                                                                                                                                                     | Rip risk                                                                                                                                                                                                                                                                                                     |
| ------------------------------------- | --------------------------- | ----------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Press Start 2P**                    | Cody "CodeMan38" Boisclair  | SIL OFL 1.1, RFN "Press Start 2P" **[V]**             | Described as "based on the font design from 1980s Namco arcade games". Sources say its glyphs derive from the _Return of Ishtar_ ROM font **[V: per third-party sources; U: provenance]** | **Flag: derived from a game font.** OFL covers the font software. In the US, typeface designs are generally not copyrightable, so commercial risk is low **[U, not legal advice]**. But it is the most common "retro game" font, so it reads as generic and can look like cliché or a rip. Avoid as primary. |
| **Pixelify Sans**                     | Stefie Justprince           | OFL 1.1, variable weight 400–700 **[V]**              | Original design "inspired by 1980s computer graphics and video games". Has lowercase and is readable                                                                                      | Low                                                                                                                                                                                                                                                                                                          |
| **Silkscreen**                        | Jason Kottke (1999)         | OFL 1.1, 400/700 **[V]**                              | Original, made for tiny web graphics text. Very small x-height: caps-like                                                                                                                 | Low                                                                                                                                                                                                                                                                                                          |
| **Tiny5**                             | Stefan Schmidt              | OFL 1.1, single weight **[V]**                        | Original 5-px-tall pixel sans. Good for micro labels                                                                                                                                      | Low                                                                                                                                                                                                                                                                                                          |
| **Micro 5**                           | Sarah Cadigan-Fried         | OFL **[V via directory; check on Google Fonts]**      | Original small pixel face                                                                                                                                                                 | Low                                                                                                                                                                                                                                                                                                          |
| **Jersey 10 / Jersey 15 / Jersey 20** | Sarah Cadigan-Fried **[U]** | OFL **[U]**                                           | Sports-jersey numeral feel. Strong candidate for scores                                                                                                                                   | Low                                                                                                                                                                                                                                                                                                          |
| **Bytesized**                         | Baltdev                     | OFL, on Google Fonts since 2025 **[V via directory]** | Tiny pixel face                                                                                                                                                                           | Low                                                                                                                                                                                                                                                                                                          |
| **Handjet**                           | Rosetta                     | OFL **[V via directory]**                             | Variable "pixel element" font, multi-script                                                                                                                                               | Low                                                                                                                                                                                                                                                                                                          |

Avoid any free-font-site download named after a game ("NES font", "Super Mario font", "Tecmo font",
etc.). These are usually direct tile rips and almost never properly licensed.

### Recommended pairing

- **Display / HUD / scores:** **Pixelify Sans Bold (700)** in UPPERCASE with +1 px letter spacing.
  Render only at sizes that are integer multiples of its pixel grid. Alternative: Jersey 10/15 for
  big score numerals **[verify license and metrics first]**.
- **Body / long text:** a non-pixel humanist sans system font (SF Pro / Roboto via the platform
  default). It respects Dynamic Type and stays readable at all sizes. If the team wants pixel body
  text anyway, use Pixelify Sans Regular at ≥16pt with Dynamic Type enabled and a non-pixel fallback
  above the accessibility size categories.
- **Micro labels (≤11pt):** Silkscreen or Tiny5. Decorative only, never for essential info.
- Press Start 2P: optional for one-off "PUSH START"-style moments only, if at all.

All OFL fonts may be bundled in a commercial app. Ship the OFL text in the app's licenses screen,
and do not sell the font by itself or rename a modified version with the Reserved Font Name.

---

## 4. NES UI patterns mapped to app UI

| NES pattern                                                                                  | App use                                                             | Touch/a11y adaptation                                                                                                                                                                                                       |
| -------------------------------------------------------------------------------------------- | ------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **▶ cursor menu**                                                                            | Settings lists, team picker                                         | The whole row is tappable (≥44pt tall). The ▶ marks the focused or selected row and moves on press (stepped, 0 frames of tweening). VoiceOver reads the row label. The cursor is decorative (`accessibilityElementsHidden`) |
| **Bordered text box** (2–3 px white or double border, black or navy fill, 8 px tile corners) | Deal cards, alerts, onboarding copy                                 | Use 9-slice border image or `borderWidth` multiples of the pixel scale. Padding = 1 tile. Text inside uses system font when long                                                                                            |
| **Typewriter text reveal**                                                                   | Deal-triggered message                                              | ≤1 s total; tapping completes it instantly; disabled under Reduce Motion; VoiceOver gets full text immediately                                                                                                              |
| **Title screen + "PUSH START"**                                                              | Splash / first-run                                                  | "TAP TO START" blinks at 2 Hz (≤3 flashes/sec stays below seizure thresholds); not blinking under Reduce Motion. Real button underneath                                                                                     |
| **Team select grid**                                                                         | Pick your teams                                                     | Grid of team cards using **text names**, sport icon sprites, and neutral palette, no logos. Multi-select with a ✓ and a ▶                                                                                                   |
| **HUD bar** (score, clock, period)                                                           | Live game card: score, inning/quarter, progress toward deal trigger | Numbers in display font; give the full sentence as the accessibility label ("Dodgers 4, Giants 2, bottom 7th, 6 of 7 strikeouts")                                                                                           |
| **Score tally / bonus count-up**                                                             | "Deals earned this season"                                          | Count-up ticks at 4-frame steps with a short sound; skip under Reduce Motion                                                                                                                                                |
| **Password screen**                                                                          | Not needed for auth. Could be a fun promo-code entry                | Use the native keyboard; never a custom glyph grid for real input                                                                                                                                                           |
| **Pause menu**                                                                               | Bottom sheet / overflow menu                                        | Native sheet with pixel styling inside it                                                                                                                                                                                   |
| **"GAME OVER / CONTINUE?"**                                                                  | Deal missed ("NO FREEBIE TODAY. NEXT GAME: FRI")                    | Neutral tone, not punitive. YES/NO become real buttons                                                                                                                                                                      |
| **Stage clear / "TOUCHDOWN!" cut-in**                                                        | Push-notification landing screen when a deal triggers               | Full-screen moment ≤2 s, then the deal details. Dismissible                                                                                                                                                                 |
| **Item get jingle + item held overhead**                                                     | Redeem / claim flow                                                 | Item sprite rises 8 px over 8 frames                                                                                                                                                                                        |

General a11y rules for the pattern set: honor Dynamic Type (pixel display text may cap at a set size
and switch to system font at accessibility sizes), Reduce Motion, Reduce Transparency (not used
anyway), and "Differentiate Without Color". Do not rely on palette color alone for state; pair color
with an icon or text.

---

## 5. Food/item sprites

NES item conventions **[U: observed convention]**:

- 16×16 metasprite (four 8×8 tiles), sometimes 8×8 for tiny pickups. Our food icons: **16×16 canvas,
  14×14 or less drawn area** so a 1 px margin remains.
- 3 colors + transparent per sprite palette. Common recipe: **dark outline** (black $0F or a darkest
  hue), **base color**, **highlight**. Since outline uses one of the 3, items often look
  "posterized". That's the look.
- Light source upper-left. Highlight is a 1–3 px cluster or short line on the top-left of the shape,
  never a gradient. No dithering at 16×16, except maybe a 2-px checker for texture (sesame seeds,
  crumbs).
- Silhouette must read at 1×. Exaggerate the identifying feature (burger layers, taco shell curve,
  fries sticking out of a box).
- A 2-frame "bob" (up 1 px, down 1 px, 16 frames each) or a 4-frame palette "glint" signals
  collectibility.
- Multi-color items (burger) cheat with two overlapping sprites (two palettes). Allowed for us, max
  2 palettes per item.

Proposed palettes (indices from §1.2):

| Item                                                                                      | Palette (outline / base / highlight) + optional 2nd palette       |
| ----------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| Burger                                                                                    | $0F / $17 bun / $27 highlight; 2nd: $0F / $1A lettuce / $06 patty |
| Taco                                                                                      | $0F / $28 shell / $38; 2nd: $1A lettuce / $16 tomato              |
| Burrito                                                                                   | $0F / $36 tortilla / $30; foil end $10                            |
| McGriddle-style sandwich (generic "griddle cake sandwich", not the trademark name in art) | $0F / $18 cake / $28; 2nd: $27 egg yolk / $30 egg white           |
| Chicken sandwich                                                                          | $0F / $27 fried crust / $38; bun $17                              |
| Plate lunch                                                                               | $0F / $30 plate & rice / $10; 2nd: $07 meat / $1A greens          |
| Fries                                                                                     | $0F / $28 fries / $38; carton $16                                 |
| Nuggets                                                                                   | $0F / $27 / $38 (3 blobs)                                         |
| Pizza slice                                                                               | $0F / $28 crust & cheese / $38; 2nd: $16 pepperoni                |
| Hot dog                                                                                   | $0F / $17 bun / $27; 2nd: $06 sausage / $28 mustard               |
| Drink                                                                                     | $0F / $16 cup / $30; straw $30, lid $10                           |
| Gift                                                                                      | $0F / $21 box / $31; 2nd: $16 ribbon / $26                        |

Deal names in text can use the brand's product name factually ("Free McGriddles"). The sprite itself
should be generic food, not a reproduction of a brand's packaging, logo, or wrapper art.

---

## 6. Legal guardrails (not legal advice)

**Generally safe** **[U: general understanding]**:

- Palette values, resolution, tile sizes, and the general "8-bit" style. Hardware limits and style
  are not protectable.
- Original sprites drawn from scratch in a retro style, including generic athletes.
- Phrases like "8-bit", "retro", "pixel art" in marketing.
- OFL fonts (Pixelify Sans, Silkscreen, Tiny5, etc.) with license text included.
- Team names used factually to say which games we track ("Dodgers home win"). Tecmo's NFLPA-only
  approach is a historical example of showing team identity by name/city without league marks.

**Risky / avoid:**

- "Nintendo", "NES", "Famicom", "Nintendo Entertainment System" in the app name, icon, App Store
  metadata, screenshots, or marketing. These are Nintendo trademarks. Describe the look as "8-bit"
  or "retro".
- Trade dress: the NES controller (rectangle, D-pad + red A/B circles on black stripe), the grey
  console or cartridge shape, the "Seal of Quality", the black-box game packaging art style. Avoid
  controller-shaped UI or a cartridge app icon.
- Any recognizable character, sprite, or tile from a real game (Mario, Tecmo players, the _Ice
  Hockey_ skaters, _Nintendo World Cup_ players, _Punch-Out!!_ style figures, etc.), traced or
  "inspired by" closely enough to be recognizable.
- Ripped game fonts, music, or sound effects (including "close" re-creations of famous jingles).
- Team logos, wordmarks, mascots, uniforms, and team color schemes used as identity. League marks
  (NFL shield, MLB silhouette, NBA logoman, etc.). Do not draw an athlete in a recognizable team
  uniform. Keep athletes in neutral colors, and refer to teams with plain text names.
- Partner (restaurant) logos and packaging art in sprites. State the deal in plain text.

**App Store 5.2.1** **[V]**: "Don't use protected third-party material such as trademarks,
copyrighted works, or patented ideas in your app without permission, and don't include misleading,
false, or copycat representations, names, or metadata in your app bundle or developer name." Also
**4.1(c)**: "You cannot use another developer's icon, brand, or product name in your app's icon or
name, without approval from the developer." Source:
[App Store Review Guidelines](https://developer.apple.com/app-store/review/guidelines/). Expect
review to look closely at metadata and screenshots mentioning team names. Add a "not affiliated
with..." disclaimer in the description and in-app about screen **[U: common practice, not a
guarantee]**.

---

## 7. Proposed `docs/ui.md` rulebook outline

### 7.1 Palette (16 tokens, all from §1.2)

Contrast ratios computed with the WCAG 2.x formula against the listed background.

| Token                 | Index | Hex     | Role                                                                       | Contrast notes                                  |
| --------------------- | ----- | ------- | -------------------------------------------------------------------------- | ----------------------------------------------- |
| `bg.night`            | $0F   | #000000 | App background                                                             | —                                               |
| `bg.panel`            | $0C   | #004058 | Cards, text boxes                                                          | white 10.9:1                                    |
| `bg.panelAlt`         | $02   | #0000BC | Headers, HUD bars, selected rows                                           | white 11.9:1                                    |
| `text.primary`        | $30   | #FCFCFC | Text, borders                                                              | 20.5:1 on night                                 |
| `text.secondary`      | $10   | #BCBCBC | Secondary text                                                             | 11.1 night / 5.9 panel                          |
| `text.muted`          | $00   | #7C7C7C | Disabled, hints **on night only**                                          | 5.0 night; **2.7 on panel: fail**               |
| `accent.primary`      | $28   | #F8B800 | Primary actions, deal highlights, cursor ▶                                 | 11.8 night / 6.3 panel; black text on it 11.8   |
| `accent.primaryLight` | $38   | #F8D878 | Highlights, sprite shine                                                   | 15.1 night                                      |
| `state.success`       | $2A   | #58D854 | "Deal triggered", wins                                                     | 11.4 night / 6.1 panel                          |
| `state.successDeep`   | $09   | #007800 | Turf, success fills under white text                                       | white 5.6                                       |
| `state.alert`         | $16   | #F83800 | Errors, "missed", live dot                                                 | 5.6 night; **3.0 panel: icons/large text only** |
| `state.alertDeep`     | $06   | #A81000 | Alert fills under white text                                               | white 7.5                                       |
| `state.info`          | $21   | #3CBCFC | Info, links, "upcoming"                                                    | 9.8 night / 5.2 panel                           |
| `state.timer`         | $27   | #FCA044 | Countdown, "in progress"                                                   | 10.3 night / 5.5 panel                          |
| `ice.light`           | $31   | #A4E4FC | Ice/sky surfaces, cool highlights                                          | 15.1 night                                      |
| `skin.neutral`        | $36   | #F0D0B0 | Sprite skin base (add $27 and $17 as alternates for a range of skin tones) | sprite use only                                 |

Rules: no colors outside the master palette; no opacity or blend modes on art (opacity only for
scrim behind modal, using `bg.night`); team identity never encoded by color.

### 7.2 Grid and scale

- 1 art pixel = **P** points, where P = 2 (compact), 3 (default), or 4 (hero art). Integer only.
- 1 tile = 8 art px = 24pt at P=3. Spacing tokens: ½ tile, 1 tile, 2 tiles.
- Borders: 1 art px (=P pt). Panel corners are a notched 1-px-cut corner, never `borderRadius`.
- Image scaling: nearest-neighbor only. In RN/Expo: export sprites pre-scaled at 1×/2×/3× PNG via
  integer scaling, or use `expo-image` with no smoothing, and check that sprite positions land on
  whole points to avoid sub-pixel blur **[U: verify in Expo on both platforms]**.
- Layout honors safe areas; full-width elements are fine. No 256 px canvas.

### 7.3 Sprite rules

- Palette: 3 colors + transparent, from the master palette. Max 2 palettes per object.
- Sizes: items 16×16; athletes 16×24 (basketball/batter 16×32); badges 8×8; hero cut-ins up to 64×64
  built from the same rules.
- Outline: dark 1 px outline on sprites shown over busy backgrounds. Light upper-left. No
  anti-aliasing, gradients, or partial transparency.
- Every sprite is original; keep source `.aseprite`/`.png` files in the repo with author noted.

### 7.4 Type scale

| Token        | Font                             | Size (pt)                    | Use                       |
| ------------ | -------------------------------- | ---------------------------- | ------------------------- |
| `display.xl` | Pixelify Sans 700, uppercase     | 32                           | Cut-in banners ("GOAL!")  |
| `display.l`  | Pixelify Sans 700, uppercase     | 24                           | Screen titles, big scores |
| `display.m`  | Pixelify Sans 600/700, uppercase | 16                           | Section headers, HUD      |
| `body`       | System font                      | 17 (Dynamic Type `body`)     | All sentences             |
| `caption`    | System font                      | 13 (Dynamic Type `footnote`) | Secondary                 |
| `micro`      | Silkscreen / Tiny5               | 8 multiples                  | Decorative labels only    |

Pixel font sizes should be multiples of the font's pixel unit so the glyphs stay crisp **[U: verify
unit size per font]**. Display text scales with Dynamic Type up to a cap, then switches to system
bold.

### 7.5 Components

`PixelPanel` (bordered box), `MenuRow` (▶ cursor list row, ≥44pt), `PixelButton` (2-state: up and
pressed shifts 1 art px down-right, no tween), `HudBar` (score/clock/progress), `DealCard` (food
sprite + partner + condition + status), `ProgressPips` (e.g. 7 strikeout pips filling), `CutIn`
(full-screen celebration), `TeamTile` (text name + sport sprite), `Toast` ("ITEM GET" style),
`TabBar` (pixel icons 16×16, labels in system font).

### 7.6 Motion

- Frame unit = 1/60 s. Animations are frame-stepped holds, not easing: 2, 4, 8, or 16 frames per
  pose. Cursor/button: instant. Blink: 30 frames on/off (1 Hz) or 15 (2 Hz) max.
- Celebration cut-ins: ≤120 frames (2 s), then auto-advance.
- Screen transitions: hard cut, or a quick "wipe" in 8 px bands over 8–16 frames.
- Palette cycling (e.g. a crowd or "LIVE" text alternating two palette colors) instead of
  translating pixels when possible.
- Never full-screen flash faster than 3×/sec; no flicker simulation. Reduce Motion: replace with
  static frames.
- RN implementation hint: drive sprite frames with `setInterval`/Reanimated on a frame counter;
  avoid `Animated.timing` easing curves for pixel elements **[U: implementation suggestion]**.

### 7.7 Sound (optional)

- Original chiptune-style SFX only (2 pulse + triangle + noise style is fine to emulate). Cursor
  blip, confirm, "deal triggered" fanfare (≤1.5 s), missed (2-note down).
- Respect silent mode; custom notification sound ≤30 s per platform limits **[U: verify Apple's
  current limit]**. Never re-create recognizable game jingles.

### 7.8 Accessibility

- All text meets WCAG AA (4.5:1; 3:1 for ≥24pt bold display). Use the contrast column in §7.1.
- Tap targets ≥44×44pt (iOS) / 48×48dp (Android), regardless of sprite size.
- Dynamic Type supported; pixel display text has a cap and fallback.
- Reduce Motion: no blinking, typewriter, cut-in animation, or count-ups.
- State never conveyed by color alone. Sprites are decorative unless they carry info, in which case
  they have an `accessibilityLabel`.
- No flicker; flashes ≤3/sec.

### 7.9 Per-sport original athlete sprites

Shared template: 16 px wide, round head about 8×8 (≈1/3 of a 24-px body), 2 px limbs, facing right,
3 colors + transparent: outline/hair ($0F), jersey ($16 for "home" or $21 for "away", generic), skin
($36 or alternates). No numbers that match real players. No team logos, stripes, or uniform
patterns.

| Sport             | Size  | Pose / frames                                  | Distinguishing prop                                                  | Scene                                                      |
| ----------------- | ----- | ---------------------------------------------- | -------------------------------------------------------------------- | ---------------------------------------------------------- |
| Hockey skater     | 16×24 | Lean-forward glide, 2-frame stride; shot frame | Stick extends 6 px past body; helmet line                            | White ice ($30) + red center line ($16) + blue line ($21)  |
| Basketball player | 16×32 | 2-frame dribble; jump/dunk frame               | 6×6 orange ball ($27)                                                | Wood court ($27/$17), rim drawn as 1 px orange + white net |
| QB / football     | 16×24 | Drop-back, throw (3 frames)                    | Helmet as round head with 1 px facemask in $10; 6×4 brown ball ($17) | Green turf $1A with white yard lines                       |
| Batter            | 16×32 | Stance, swing (3 frames)                       | Bat as 1–2 px diagonal ($27); helmet                                 | Dirt $17/$27, grass $1A, white base                        |
| Soccer player     | 16×24 | Run (2), kick (1)                              | 6×6 white ball with 1 px black patches                               | Striped pitch $1A/$2A                                      |
| Variants          | same  | —                                              | Women's teams: hair silhouette variant (ponytail 2–3 px)             | —                                                          |

College football (USC, UCLA): same QB sprite; names in text only. Do not use school marks or mascot
characters.

---

## Sources

- NESdev wiki: [PPU palettes](https://www.nesdev.org/wiki/PPU_palettes),
  [PPU OAM](https://www.nesdev.org/wiki/PPU_OAM),
  [PPU attribute tables](https://www.nesdev.org/wiki/PPU_attribute_tables),
  [PPU rendering](https://www.nesdev.org/wiki/PPU_rendering),
  [Overscan](https://www.nesdev.org/wiki/Overscan)
- [Wikipedia: List of video game console palettes](https://en.wikipedia.org/wiki/List_of_video_game_console_palettes)
- [Lospec: Nintendo Entertainment System palette](https://lospec.com/palette-list/nintendo-entertainment-system)
- [FirebrandX NES palettes](https://www.firebrandx.com/nespalette.html)
- [Wikipedia: Ice Hockey (1988)](<https://en.wikipedia.org/wiki/Ice_Hockey_(1988_video_game)>),
  [Blades of Steel](https://en.wikipedia.org/wiki/Blades_of_Steel),
  [Tecmo Bowl](https://en.wikipedia.org/wiki/Tecmo_Bowl)
- Fonts: [Fontsource Press Start 2P](https://fontsource.org/fonts/press-start-2p/about),
  [Font Library Press Start 2P](https://fontlibrary.org/font/press-start-2p),
  [Fontsource Pixelify Sans](https://fontsource.org/fonts/pixelify-sans),
  [Fontsource Silkscreen](https://fontsource.org/fonts/silkscreen),
  [Fontsource Tiny5](https://fontsource.org/fonts/tiny5),
  [Kottke: Silkscreen launched](https://kottke.org/99/08/silkscreen-launched),
  [Serbyte: Pixelify Sans](https://www.serbyte.net/fonts/pixelify-sans),
  [Serbyte: Bytesized](https://www.serbyte.net/fonts/bytesized),
  [Fontmeme Google pixel fonts](https://fontmeme.com/jfont/googleweb-pixel-fonts/)
- [Apple App Store Review Guidelines](https://developer.apple.com/app-store/review/guidelines/)

### Open items to verify before finalizing `docs/ui.md`

1. The exact PPU index ↔ hex mapping (diff against an emulator `.pal` file, e.g. Mesen's default).
2. Jersey 10/15/20 and Micro 5 designer/license on fonts.google.com.
3. Sports-game details marked [U] (views, cut-scenes). Watch gameplay footage for reference only.
4. Pixel-unit sizes for Pixelify Sans and Silkscreen for crisp rendering.
5. Nearest-neighbor rendering behavior in Expo (`expo-image`) on iOS and Android.
6. Trademark review by counsel for name, icon, and store metadata.
