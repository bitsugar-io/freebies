# Deals: trigger_rule

How a deal decides it has unlocked and when it expires. The rest of the deal workflow is in
[deals.md](deals.md).

## trigger_rule

`trigger_rule` is a JSON string parsed into `Rule` in `services/api/internal/triggers/checker.go`.

```json
{
  "metric": "home_win",
  "scope": "team",
  "operator": "==",
  "value": 1,
  "redemption_window": "same_day",
  "expires_at_local": "10:30"
}
```

| Field               | Values                                                                                         | Effect                                                                                        |
| ------------------- | ---------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| `metric`            | a key the league's source emits (below)                                                        | The number compared                                                                           |
| `operator`          | `>=` `>` `<=` `<` `==` `=` `!=`                                                                | Anything else never triggers                                                                  |
| `value`             | integer                                                                                        | Compared with the metric                                                                      |
| `redemption_window` | `same_day`, `next_day`, `week`, `next_business_day`, `through_next_business_day`, `game_night` | When the deal expires (below)                                                                 |
| `expires_at_local`  | `"HH:MM"`, 24h Pacific, optional                                                               | Clock time on the last day instead of 23:59:59                                                |
| `unlock_after_days` | integer, default 0                                                                             | Code unlocks at midnight Pacific that many days after the game date; stored as `available_at` |
| `valid_days`        | integer, default 0                                                                             | Length of the window (below); replaces `redemption_window` when set                           |
| `scope`             | `team`, `team_batters`, `team_pitchers`, `team_fielders`                                       | Descriptive only; the checker ignores it. Set it anyway                                       |

### When it's checked

The `check-triggers` CronJob runs each morning (Pacific) and checks **yesterday's** games. One game
triggers a deal at most once.

### Deadlines

Every window is computed from the **game's date** in `America/Los_Angeles`, never from when the
check happens to run, and ends at 23:59:59 Pacific on its last valid day (or at `expires_at_local`
on that day where it applies). Day 1 is the day after the game, when the morning check runs:

| Window                          | Expires                                                         |
| ------------------------------- | --------------------------------------------------------------- |
| `same_day`                      | Day 1 at 23:59:59, or at `expires_at_local` on day 1            |
| `next_day` (and unknown values) | Day 2 at 23:59:59, or at `expires_at_local` on day 2            |
| `week`                          | Day 7 at 23:59:59; `expires_at_local` is ignored                |
| `game_night`                    | The game date itself, at 23:59:59                               |
| `next_business_day`             | Only the first weekday after the game (locked until then)       |
| `through_next_business_day`     | Usable now, through the end of the first weekday after the game |

"The morning after the game, breakfast hours only" is `same_day` with `"expires_at_local": "10:30"`.

`game_night` is for at-the-game deals claimed in the venue that night (Chick-fil-A's arena deals).
The morning check is always too late for them, so they never record a win or send a push; the app
shows them in Today on game days. Set `in_venue = 1` and `venue` on the event as well.

A deal is never recorded, and never pushed, if its deadline has already passed when the run reaches
it (a late run or a backfill after 10:30am for the breakfast deal). The run logs it at WARN with
`event_id`, `status=EXPIRED` and `reason`, and the `check-triggers` response lists it under
`skipped` (`skippedEvents` counts them). Deals that can never fire are listed there too.

| Rule fields                             | Unlocks                                                                      | Expires                                                                                                               |
| --------------------------------------- | ---------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `redemption_window: next_business_day`  | Midnight of the first weekday after the game (open at once if that is day 1) | 23:59:59 that weekday, or `expires_at_local`                                                                          |
| `valid_days: N` (no unlock)             | Immediately                                                                  | 23:59:59 on day N, or `expires_at_local`                                                                              |
| `unlock_after_days: U`, `valid_days: N` | Midnight Pacific, U days after the game date                                 | 23:59:59 on the Nth day counting the unlock day as the first (`valid_days` defaults to 1; `expires_at_local` ignored) |

Example: a code that "unlocks 2 days after the game, expires 1 week after unlock" is
`"unlock_after_days": 2, "valid_days": 7`. For a Sunday Nov 15 game it unlocks Tue Nov 17 and ends
Mon Nov 23 at 11:59pm. The push says "Code unlocks Tue. Claim by Mon Nov 23 11:59pm." and the app
reads `availableAt` from the active and claimed deal JSON to show "Code unlocks Tue".

Read "unlock" carefully. "Code available for unlock for 2 days following the game" (the Rams and
Carl's Jr.) is a deadline, not a delay: the fan adds the code within 2 days, so it is `next_day`.
What happens after they add it ("expires in app one week from unlock") belongs in the description,
not the rule.

### Reminders

`send-reminders` runs three times a day (11am, 3pm and 5pm Pacific), but a user gets at most one
reminder per deal per Pacific day, and only on the day the deal first becomes claimable (it unlocks)
and on its last day.

A deal with `available_from_local` gets its unlock-day reminder even when it starts after the last
run: on the unlock day a run before the start time sends it, worded with the start ("Your Free Taco
code starts at 6pm"). A run after the start says the code works now. The same rule words the "you
won" push: a code that starts later the same Pacific day reads "Code starts at 10am.", one that
starts on a later day reads "Code unlocks Mon." (or "tomorrow").

### Push delivery

The `notifications` table has one row per push (`kind` is `trigger` or `reminder`). The worker
writes it as `pending` just before the send and then sets the outcome from that user's own Expo
ticket:

| Status    | Meaning                                                                          | Retried? |
| --------- | -------------------------------------------------------------------------------- | -------- |
| `sent`    | Expo returned an ok ticket                                                       | No       |
| `failed`  | Expo rejected it, or never got it (connection refused, 429 or 503 after retries) | Yes      |
| `unknown` | Expo may have it: timeout or reset after the request was sent, other 5xx         | No       |
| `pending` | The process died during the send                                                 | No       |

The rule is that a missed push is better than a duplicate one, so only `failed` is retried. The Expo
client retries a request only when Expo cannot have processed it (connection refused, 429, 503),
never after a timeout or reset once the request was written. A `DeviceNotRegistered` ticket clears
that push token.

The "you won" push is owed to everyone who followed the deal's event before it triggered and has no
`sent`, `unknown` or `pending` trigger row for it. Every `check-triggers` run (7am and 9am) sends
what is owed for every live deal, so a push that failed at 7am (Expo down, or the process died after
recording the deal) goes out at 9am or on a later run before the deal expires. Paused users are
skipped and get it on a later run if they resume while it is live. A user never gets two trigger
pushes for one deal. Deals made by `deals create` or `deals trigger --notify=false` are never
pushed, reminders included; `deals trigger` (with `--notify`, the default) sends through the same
code as `check-triggers`.

To make a delayed deal locally,
`go run . deals trigger <event-id> --unlock-after-days 3 --valid-days 7 --notify=false` builds the
window from yesterday's game with the same code the checker uses.

### Metrics by league

Only leagues with a registered source can trigger. Sources live in `services/api/internal/sources/`;
every source lists its metrics in `Metrics()`, and `go run . deals audit` rejects an active deal
whose metric is not on that list. League names are matched case-insensitively. The data provider for
each league, its terms and its fallback are in [data-sources.md](data-sources.md).

**MLB** (`sources/mlb/source.go`), for the deal's team in that game:

| Metric                                                                              | Meaning                              |
| ----------------------------------------------------------------------------------- | ------------------------------------ |
| `win`, `home_win`                                                                   | 1 if won (at home), else 0           |
| `runs`, `hits`, `home_runs`, `rbi`, `stolen_bases`                                  | Batting totals                       |
| `home_runs_scored`, `home_stolen_bases`                                             | Same, but 0 for away games           |
| `strikeouts`, `pitching_hits`, `pitching_runs`, `pitching_walks`, `pitching_homers` | Pitching totals                      |
| `double_plays`, `home_double_plays`                                                 | Fielding; the home variant is 0 away |

MLB does not emit `away_saves` or `shutout` yet (the deals that would use them are inactive until
2027). Check a real game before trusting a metric:

```bash
cd services/api && mise exec -- go run . verify-mlb --team LAD --date 2026-09-01
```

**NHL** (`sources/nhl/source.go`), for the deal's team in a finished game (FINAL/OFF only; live,
future and postponed games are ignored). An NHL game cannot end tied, so OT and shootout wins count
as wins:

| Metric                 | Meaning                                                            |
| ---------------------- | ------------------------------------------------------------------ |
| `win`, `home_win`      | 1 if won (at home), else 0                                         |
| `goals`                | Goals scored                                                       |
| `home_goals_scored`    | Same, but 0 for away games                                         |
| `shutout`              | 1 if the opponent scored 0 (never in a shootout)                   |
| `overtime`, `shootout` | 1 if the game was decided in overtime / a shootout, win or lose    |
| `p2_last_minute_goals` | Goals scored at 19:00 or later in the 2nd period (McFlurry Minute) |

The data is the NHL's unofficial public web API (`api-web.nhle.com`), whose terms restrict
commercial automated use; budget a licensed feed before monetizing. Check a real game:

```bash
cd services/api && mise exec -- go run . verify-nhl --team ANA --date 2026-11-08
```

**NBA, WNBA** (`sources/nba`, `sources/wnba`, ESPN; teams LAL, LAC, LAS). Final games only, overtime
included, on the game's Pacific date:

| Metric            | Meaning                    |
| ----------------- | -------------------------- |
| `win`, `home_win` | 1 if won (at home), else 0 |
| `points`          | Points scored              |
| `home_points`     | Same, but 0 for away games |

The NBA source also emits `home_q4_opponent_ft_miss_streak`: at home, the most free throws the
visitors missed in a row in the 4th quarter (overtime excluded), from the play-by-play; 0 away.

**NFL** (`sources/nfl`, ESPN; teams LAR, LACH). Same as basketball, plus a box-score metric. A tie
is neither a win nor a loss; a neutral-site game (international games, the Super Bowl) is never a
home game:

| Metric                  | Meaning                                                             |
| ----------------------- | ------------------------------------------------------------------- |
| `win`, `home_win`       | 1 if won (at home), else 0                                          |
| `points`, `home_points` | Points scored, and the same but 0 for away games                    |
| `interceptions`         | Passes the team's defense intercepted, any venue (pick-sixes count) |
| `long_field_goal`       | The team's longest made field goal in yards, 0 if none              |

**MLS, NWSL** (`sources/mls`, `sources/nwsl`, ESPN; teams LAFC, LAG, ACFC). Goals come from the
scoreboard's goal events; `first_half` counts goals up to the 45th minute plus first-half stoppage
time (a goal shown as 46' is second half). Own goals count for the team they benefit. A match
decided on penalties after a level score is a draw (`draw` 1, `win` 0):

| Metric                                      | Meaning                                                                  |
| ------------------------------------------- | ------------------------------------------------------------------------ |
| `win`, `home_win`, `draw`                   | 1 if won (at home) / level at full time, else 0                          |
| `goals`, `home_goals_scored`                | Goals scored; the home variant is 0 away                                 |
| `first_half_goals`, `home_first_half_goals` | First-half goals; the home variant is 0 away                             |
| `scored_first_in_first_half`                | 1 if the team scored the match's first goal and it was in the first half |
| `home_scored_first_in_first_half`           | Same, but 0 for away games (LAFC and Ono's stricter wording)             |

**NCAA** (`sources/ncaa`, ESPN; teams USC, UCLA). Football and men's basketball. Each game emits
`win`, `home_win`, `points`, `home_points` and the same four prefixed with its sport
(`football_win`, `basketball_home_points`, ...); use a prefixed metric so a deal can never fire on
the other sport. Neutral-site games (tournaments, bowls) are never home games.

Men's basketball also emits `basketball_home_h2_opponent_ft_miss_streak`: at a home game, the most
free throws the visitors missed in a row in the 2nd half (play-by-play), else 0.

The ESPN sources use an unofficial, undocumented API with no commercial license; see
[data-sources.md](data-sources.md). Check a real game for any league:

```bash
cd services/api && mise exec -- go run . verify-source --league mls --team LAFC --date 2026-09-09
```

**No source yet**: a deal for a metric no source emits (for example `q4_opp_consecutive_ft_misses`)
can be seeded but **must be inactive**: `deals audit` and the audit test fail an active deal that
can never fire. Add a NOTE comment saying what is missing.

### Examples

| Deal                                                  | Rule                                                                                                                                               |
| ----------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| Dodgers pitchers strike out 7+                        | `{"metric": "strikeouts", "scope": "team_pitchers", "operator": ">=", "value": 7, "redemption_window": "next_day"}`                                |
| Angels score 5+ at home                               | `{"metric": "home_runs_scored", "scope": "team_batters", "operator": ">=", "value": 5, "redemption_window": "same_day"}`                           |
| Dodgers double play at home                           | `{"metric": "home_double_plays", "scope": "team_fielders", "operator": ">=", "value": 1, "redemption_window": "next_day"}`                         |
| Rams interception, code unlocks in 2 days             | `{"metric": "interceptions", "scope": "team", "operator": ">=", "value": 1, "redemption_window": "week", "unlock_after_days": 2, "valid_days": 7}` |
| LAFC score first in the first half, next business day | `{"metric": "home_scored_first_in_first_half", "scope": "team", "operator": "==", "value": 1, "redemption_window": "next_business_day"}`           |
| Ducks home win, breakfast                             | `{"metric": "home_win", "scope": "team", "operator": "==", "value": 1, "redemption_window": "same_day", "expires_at_local": "10:30"}`              |

## Two more fields

| Field                  | Values               | Meaning                                                                                                        |
| ---------------------- | -------------------- | -------------------------------------------------------------------------------------------------------------- |
| `available_from_local` | `"HH:MM"` Pacific    | The deal starts at that time on its first usable day ("valid 10am-11:59pm", "not during breakfast": `"10:30"`) |
| `season`               | `"regular"` or empty | `"regular"`: playoff games never fire it ("during the regular season"). Empty: any game                        |

`deals audit` (and the seeded-deals test) rejects a misspelled window, a time that isn't `HH:MM` and
any other `season` value. An empty window still means `next_day`.

## Read the terms against the rule

Before a deal goes active, map every timing phrase in its terms to a field, and say so in the
migration comment:

| The terms say                                                      | The rule needs                            |
| ------------------------------------------------------------------ | ----------------------------------------- |
| "the day after", "valid until midnight"                            | `same_day`                                |
| "breakfast hours", "until 10:30am"                                 | `same_day` + `expires_at_local`           |
| "from 10am", "not during breakfast"                                | `available_from_local`                    |
| "within 3 days", "valid 3 consecutive days"                        | `valid_days: 3`                           |
| "code unlocks 2 days after the game", "expires 1 week from unlock" | `unlock_after_days: 2`, `valid_days: 7`   |
| "available for unlock for 2 days following the game"               | `next_day` (a deadline to add the code)   |
| "the next business day"                                            | `next_business_day`                       |
| "through the following business day"                               | `through_next_business_day`               |
| "claim at the arena that night"                                    | `game_night`, plus `in_venue` and `venue` |
| "regular season"                                                   | `season: "regular"`                       |
| "home games", "at home"                                            | a `home_*` metric                         |

Then trigger it locally with `deals trigger <event-id> --notify=false`: without `--hours` it uses
the deal's own rule as if the game were last night, so the app shows what a real win would (a locked
code, a breakfast deadline, a start time).
