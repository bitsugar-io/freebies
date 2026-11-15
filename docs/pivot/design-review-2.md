# Design review 2: LA redesign (feat/la-redesign)

Independent review using the `design` skill (ui-rules, ui-patterns, ui-review, sprite-rules),
`docs/ui.md` and `docs/architecture.md#server-driven-content`, against the owner's round-2 picks
(logo 2, sport-first Teams with two-tone tiles, game-HUD tab bar WON / TEAMS / YOU, MENU → PAUSE,
title screen, blinking ▶).

Live pass: iPhone 16e simulator (iOS 26.2), Expo Go on Metro :8082, API on :8080, user `16420984…`
(follows Ducks; one active Ducks deal). Screenshots are in
`/private/tmp/claude-501/-Users-john-git-bitsugar-io-freebies/c180eadd-302f-461a-ac5b-78bf906bba17/scratchpad/review2/`
(below, `review2/<name>.png`). I couldn't tap, so I reviewed pressed, blinking, wipe, title screen,
cut-in, offline and large-text states in the code. Those are marked "code".

`audit-colors.sh`: `ok: only NES tokens`. Every text/background pair I checked passes `contrast.py`
(timer on panel 5.47, success on panel 6.06, muted on night 5.03, ice on panelAlt 8.74).

## Findings

### High

**H1. The app ignores `availableAt`. A locked Rams code shows as claimable now.** (WON, DEAL, code)

- Evidence: `availableAt` doesn't appear anywhere in `apps/mobile/src`. `ActiveDeal`
  (`src/api/client.ts:89-97`) has no field for it. The API sends it
  (`services/api/internal/api/handlers/types.go:89-105`), and the push copy already says "Code
  unlocks Tue" (`internal/worker/copy.go:94`). The live Rams rule is
  `unlock_after_days: 2, valid_days: 7`. I couldn't create a delayed deal live because
  `deals trigger` only takes `--hours`, so this is from the code.
- What the user sees: the WON row says "9 DAYS LEFT" and "Because Rams …". The deal screen says "You
  got Free Famous Star from Carl's Jr.", shows a 9-day countdown, and puts ▶ OPEN CARL'S JR and GOT
  IT, CLAIMED first. The push said the code is locked. Someone who goes to order early gets a code
  that doesn't work.
- Fix:
  - Add `availableAt?: string` to `ActiveDeal`.
  - While it's in the future:
    - WON row: replace the time-left line with a `LOCKED` tag plus "Unlocks Tue 12:00am PT" in
      `info`.
    - Deal dialog: "Rams got a pick. Your Free Famous Star code unlocks Tue 12:00am PT."
    - Countdown: count down to the unlock, not the expiry.
    - Hide OPEN and GOT IT, and offer "Remind me when it unlocks" instead.
    - Bar: start the ExpiryBar at `availableAt`.
  - Add a VoiceOver label: "Locked until Tuesday."

**H2. The HUD "WATCH" count is wrong, and the same bug turns off alert onboarding.** (all tabs)

- Evidence:
  - The HUD reads `WATCH 10` (`review2/won.png`), but the user follows one team: Teams says
    `1 PICKED` and `HOCKEY 1/2`, and You says "Following Ducks".
  - The DB has 10 subscriptions. 9 of them are on retired events (`is_active = 0`: Dodgers, Angels
    and LAFC offers).
  - `subscribedCount = subscriptions.size` (`src/hooks/useSubscriptions.ts:81`) counts them all.
  - The same number drives `wasEmpty` in `setSubscribed` (`src/context/AppDataContext.tsx:179`). Any
    user carrying stale subscriptions never sees the alerts primer, and this user's `DEAL ALERTS` is
    `OFF` (`review2/you.png`).
- Fix: count only subscriptions whose event is in the active `events` list, and use that same count
  for `wasEmpty`. Better still, have the API drop subscriptions to retired events from
  `GET /subscriptions`.

**H3. Team colors hide season state and make teams look alike.** (Teams › Hockey, Soccer, Football)

- Evidence:
  - TeamTile swaps primary and secondary when the primary is "dark" (luminance < 0.1,
    `src/components/pixel/TeamTile.tsx:63-67`).
  - Kings (`#000000` / `#BCBCBC`) get a light-grey field (`review2/teams_hockey.png`). That looks
    just like the greyscale off-season treatment (Angels, `review2/teams_baseball.png`). Season
    state is supposed to be the color channel, and here it fails.
  - `#0000BC` counts as dark too, so Galaxy and Rams also swap to gold fields. LAFC and Galaxy end
    up as two identical gold tiles, told apart only by a 6pt stripe that is black on one and dark
    blue on the other (`review2/teams_soccer.png`).
- Fix:
  - Don't swap. Draw the primary field and add a 3pt `text.primary` inner rule when the primary is
    near black.
  - Give off-season tiles a second cue on the field itself: a 2px checker dither over the grey, or a
    black `OFF` corner tag. Then a grey team color can't be mistaken for "off season".
  - Ask the API to serve a `tile` color pair per team, so these choices stay server-side.

### Medium

**M1. The followed ✓ blinks off into a black hole.** (Teams tiles) `TeamTile.tsx:117-127` puts the ✓
inside a black square. `Blink` only sets the child's `opacity: 0` (`useBlink.tsx:143`), so on the
off phase a 24pt black notch sits in the team color (`review2/teams_hockey.png`, Ducks tile, top
right). It reads as a rendering bug. Fix: wrap the whole badge (background included) in `<Blink>`,
or keep the square and alternate the ✓ between `accent` and `accentLight` (a palette swap, which the
rules prefer).

**M2. Press Start 2P is used below 12pt for text people have to read.** The `micro` variant (10pt,
`useTheme.tsx:73`) is used for:

- tile status "NO DEAL", "BACK MAY", "1 DEAL" (`TeamTile.tsx:112`)
- the partial count "1/2" (`TeamTile.tsx:124`)
- "BACK MAR" on sport rows (`TeamsScreen.tsx` SportRow)
- the HUD status "1 PICKED", "SUN NOV 15" (`HudBar.tsx:41`)
- the WON tab badge count (`HudTabBar.tsx:98`)

This breaks ui.md Type rule 3. Fix: use `displayM` (12pt) or the system `caption`, and retire
`micro` or keep it for decorative text only.

**M3. "SLIDE TO STOP" is unclear and the gesture is fragile.** (DEAL,
`review2/deal_0be2de9a-c5ff-4.png`)

- "Stop" what? The deal, the reminders, the game?
- Code: the pan handlers sit on the 48pt thumb only (`PixelSlide.tsx:68`). Tapping the track label
  does nothing, even though the doc comment says a track tap asks for confirmation.
- `accessibilityRole="adjustable"` (`PixelSlide.tsx:58`) tells VoiceOver to swipe up and down. This
  is a one-shot action, so the role should be `button`.
- Fix: replace the slider with a third ▶ row, "NO MORE REMINDERS", that opens the existing confirm
  dialog. If you keep the slider, label it "SLIDE: MUTE DEAL", attach the responder to the track and
  change the role to `button`.

**M4. One idea has several names.**

| Idea      | Names used                                                                                                                                                                  |
| --------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Following | `WATCH` (HUD), `PICKED` (Teams header), "following" (Deals by team), "Watching" (deal preview switch), "Following Ducks" (You)                                              |
| Won deals | `WON` (tab), `CLAIM 1` (HUD), "1 unlocked" (tab a11y label), "is back in Unlocked" and "Move back to unlocked" (`DealScreen.tsx` toast and claimed row), "You got" (dialog) |

"Unlocked" is the name of a screen that no longer exists. Fix: use FOLLOW for teams and deals, WON
for unlocked, and CLAIMED for done. For example, HUD `WON 1  FOLLOW 1`, toast "Back in Won", dialog
"Ducks won. $2 McGriddles unlocked." (the ui.md voice).

**M5. The PAUSE menu copy and ABOUT row.** (`review2/pause.png`, `PauseScreen.tsx`)

- While paused, `RESUME ALERTS` and `RESUME` sit in the same menu but do different things.
- The title `PAUSE` plus the row `PAUSE ALERTS` repeats the word.
- `ABOUT` jumps to the top of YOU, where About is below the fold.
- Fix: rename `RESUME` to `CLOSE` (or `BACK TO GAME`). Have ABOUT open YOU scrolled to the About
  anchor, or a small About modal.

**M6. Most team tiles say NO DEAL but look just as tappable as the rest.** 9 of 14 teams have no
deal (`review2/teams_basketball.png`, `review2/teams_soccer.png`). The hint says "Tap a team to
follow all its deals", and tapping one only shows a toast. A first-time user will tap Lakers, get
nothing to follow, and wonder what went wrong. Fix:

- Sort teams with deals first.
- Change the hint when a sport has none: "No Lakers or Clippers deals yet. We'll add them when a
  sponsor runs one."
- Or let people follow a team for future deals. The API log shows a team-subscriptions backfill, so
  team-level follows may already exist on the server.

**M7. SELECT SPORT always highlights the first row.** `selected={i === 0}` (`TeamsScreen.tsx`) makes
BASKETBALL accent-colored with a blinking ▶ (`review2/teams.png`). Accent is also the color for "you
follow this" (the hockey `1/2`, DUCKS in Deals by team). A new user reads basketball as picked. Fix:
put the ▶ on the last-visited sport, or on the first sport with a follow, and keep the label white.
The cursor alone shows focus.

**M8. The sport screen has no loading, empty or error state.** `/teams/curling` shows a header,
BACK, the hint and nothing else (`review2/teams_unknown.png`). The same blank screen appears during
a cold-start deep link before `/teams` loads, because `SportTeamsScreen` has no loading branch. Fix:
show skeleton tiles while loading. For an unknown or empty sport, show "NO TEAMS HERE" and a BACK
button. On failure, show `ErrorPanel`.

**M9. The deal HUD title is the offer name in Press Start 2P.** `HudBar title={event.offerName}`
with `numberOfLines={2}` (`DealScreen.tsx`). Offer names run past 3 words ("Free Famous Star", or
longer names like "Free Original Chicken Sandwich") and will truncate. Fix: put "DEAL WON!" or the
partner name in the HUD. The offer name is already the accent key word in the dialog.

**M10. The YOU season card's "0" looks like a coin.** At 24pt, Press Start 2P's "0" reads as a gold
ring icon beside "DEALS CLAIMED" (`review2/you.png`). Fix: show `0 CLAIMED` on one line, or replace
a zero with "First win pending" in caption text.

### Low

- **Deal countdown color:** always `warning` (`DealScreen.tsx` countdown). The rule is the timer
  color only under 6h. Use `info` above 6h, as `TimeLeft` already does.
- **WON row headline:** "ENDS SOON 2H 17M LEFT" is 5 words in Press Start 2P and fills the row
  beside NEW! (`review2/won.png`). Make it an `ENDS SOON` tag plus `2H 17M`.
- **Opacity:** `MenuRow` uses `opacity: 0.5` when disabled (`PixelMenu.tsx:79`), which breaks
  palette rule 6 (no opacity). Use the `textMuted` color instead.
- **Dynamic Type:** body and caption are capped at `maxScale: 2` (`useTheme.tsx:74-76`). ui.md says
  unbounded.
- **Server-driven rule:**
  - `CITY_PREFIXES` is a hardcoded city list (`utils/format.ts:1`). Use the API's `team.name`
    (already short) for event rows too.
  - Sport labels are `sport.toUpperCase()` of the raw id, so an id like `womens-soccer` would render
    badly. Add a `sportName` to `/teams`.
- **Tab bar layout:** `space-between` puts YOU against the right edge at 4pt while the screen gutter
  is 16. The stat line `CLAIM 1 WATCH 10` has no separator and reads as one phrase
  (`review2/won.png`).
- **Sport badges:** the hockey badge (stick and puck) is thin strokes, not a filled silhouette, and
  reads as stray lines at 32pt. The football reads as a red pill (`review2/teams.png`). Redraw both
  with the design skill: a 1px outline and a filled shape.
- **Roles:** `SportRow` uses role `button` inside a `menu`; use `menuitem`. The HudTabBar comment
  says "four text tabs", but there are three.
- **Title screen:** it auto-continues after 4s even under VoiceOver (`TitleScreen.tsx:17`). Skip the
  timer when a screen reader is running.
- **YOU alerts caption:** indented 24pt while the other section text sits at 0 (`review2/you.png`).

## What's working

- Only NES tokens, and every text pair passes contrast. Text on team tiles always sits on the black
  strip.
- Sports, teams, badges and colors come from `/teams`, so an unknown sport still renders (it just
  needs an empty state).
- One shared 1 Hz blink clock with a top-3 budget, solid under Reduce Motion. The wipe and the
  cut-in check Reduce Motion (code).
- Display text falls back to system bold at large text sizes (code). Tile names wrap at one size,
  and tiles keep equal heights (Angel City).
- The WON empty state has a PICK TEAMS button. There's an error panel with details, and a stale
  banner with the saved time. Offline is covered (code).
- The deal screen hierarchy is clear: sprite, dialog with the reason, a big PT countdown, how to
  claim, ▶ actions, and a confirm step before stopping reminders.
- The HUD tab bar has 48pt tabs with role `tab` and `selected`, an accessible stat summary, and a
  MENU target of at least 44pt. PAUSE opens instantly and moves VoiceOver focus to its title.

## Checklist (ui-review.md)

P = pass, F = fail, U = unverified (needs a device interaction), — = not applicable.

| Item                                         | Title                           | WON                     | Select sport            | Sport teams                        | YOU        | PAUSE | DEAL                    |
| -------------------------------------------- | ------------------------------- | ----------------------- | ----------------------- | ---------------------------------- | ---------- | ----- | ----------------------- |
| Theme tokens only                            | P                               | P                       | P                       | P                                  | P          | P     | P                       |
| PS2P ≤3 words, ≥12pt                         | F (tagline 5 words, owner spec) | F (row headline)        | F (BACK MAR 10pt)       | F (tile status 10pt)               | P          | P     | F (HUD offer name)      |
| Pixel borders, no radius                     | P                               | P                       | P                       | P                                  | P          | P     | P (native sheet corner) |
| Integer sprite scales                        | P                               | P                       | P                       | P                                  | P          | —     | P                       |
| Matches reference                            | U                               | P                       | F (cursor on first row) | F (color swap)                     | P          | P     | P                       |
| ≥44pt, non-gesture alternative               | P                               | P                       | P                       | P                                  | P          | P     | F (slider tap target)   |
| Roles and labels                             | P                               | P                       | F (menuitem)            | P                                  | P          | P     | F (adjustable)          |
| Contrast                                     | P                               | P                       | P                       | P                                  | P          | P     | P                       |
| Largest Dynamic Type                         | U                               | U                       | U                       | U                                  | F (cap 2×) | U     | U                       |
| Reduce Motion                                | P (code)                        | P (code)                | P                       | P                                  | P          | P     | P (code)                |
| Loading/empty/error/offline                  | —                               | P                       | P                       | F                                  | P          | —     | P                       |
| No logos or "NES"; team colors only on tiles | P                               | P                       | P                       | F (H3 legibility; owner exception) | P          | P     | P                       |
| Deal copy: reason, reward, PT deadline       | —                               | P                       | —                       | —                                  | —          | —     | F (locked deals, H1)    |
| Team selection flow                          | —                               | —                       | P                       | F (NO DEAL tiles, blink hole)      | —          | —     | —                       |
| Your deals flow                              | —                               | F (wrong HUD count, H2) | —                       | —                                  | —          | —     | F (dismiss UX, M3)      |
