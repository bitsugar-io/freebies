# NES patterns for app UI

Each classic NES pattern below is mapped to the app component that carries it, plus the touch and
accessibility adaptation it needs. Source: `docs/pivot/nes-style-research.md` section 4.

## Menus, boxes and screens

| NES pattern                  | App use                                                                                             | Adaptation                                                                                                                           |
| ---------------------------- | --------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| ▶ cursor menu                | Settings, claim actions, team picker                                                                | Whole row tappable and at least 44pt. The ▶ is decorative and hidden from VoiceOver. It moves in steps with no tween.                |
| Bordered text box            | Deal message, alerts, onboarding                                                                    | 3pt white border with padding of 1 tile. Long text uses the system font.                                                             |
| Typewriter reveal            | Deal-unlocked message                                                                               | Finishes in 1s or less. A tap completes it. Off under Reduce Motion. VoiceOver reads the full text at once.                          |
| Title screen, "TAP TO START" | First run                                                                                           | Blinks at 2 Hz or slower, and not at all under Reduce Motion. A real button sits underneath.                                         |
| Team select grid             | Teams tab: SELECT SPORT, then a sport's teams                                                       | Team names as text, a sport badge on each, multi-select with ✓. Off-season teams in grey. Fixed font size, names wrap, never shrink. |
| Blinking selection           | ▶ on the active tab and focused menu row, `NEW!`, `ENDS SOON`, `PAUSED`, unlocked count, followed ✓ | One shared 1 Hz `useBlink()`. Solid under Reduce Motion. Never body text, never more than 3 blinking at once.                        |
| Game HUD tab bar             | WON, TODAY, TEAMS, YOU                                                                              | Black bar, 3pt white top border, stat line, text tabs with a blinking ▶. Stepped wipe between tabs.                                  |
| START / pause menu           | START button on every HUD bar                                                                       | PAUSE overlay: pause alerts, alert settings, about, resume. Transparent modal route, no animation.                                   |

## Game moments

| NES pattern                    | App use                                              | Adaptation                                                                                                  |
| ------------------------------ | ---------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| HUD bar (score, clock, period) | Live game card                                       | The accessibility label is a full sentence, e.g. "Kings 2, Sharks 1, third period, win means free burrito". |
| Score tally count-up           | Deals won this season                                | 4-frame steps. Skipped under Reduce Motion.                                                                 |
| "GOAL!" / "TOUCHDOWN!" cut-in  | Push landing when a deal triggers                    | Full screen for 2s or less, then the deal. Dismissible.                                                     |
| Item held overhead             | Claim screen                                         | The sprite rises 8px over 8 frames.                                                                         |
| "CONTINUE?"                    | Deal missed, e.g. "NO FREEBIE TODAY. NEXT GAME: FRI" | Neutral, never punitive. YES and NO are real buttons.                                                       |
| Pause menu                     | Overflow actions                                     | A native sheet with pixel styling inside.                                                                   |
