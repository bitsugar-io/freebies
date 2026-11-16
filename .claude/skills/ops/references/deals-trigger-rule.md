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

| Field               | Values                                                   | Effect                                                                                        |
| ------------------- | -------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| `metric`            | a key the league's source emits (below)                  | The number compared                                                                           |
| `operator`          | `>=` `>` `<=` `<` `==` `=` `!=`                          | Anything else never triggers                                                                  |
| `value`             | integer                                                  | Compared with the metric                                                                      |
| `redemption_window` | `same_day`, `next_day`, `week`, `next_business_day`      | When the deal expires (below)                                                                 |
| `expires_at_local`  | `"HH:MM"`, 24h Pacific, optional                         | Clock time on the last day instead of 23:59:59                                                |
| `unlock_after_days` | integer, default 0                                       | Code unlocks at midnight Pacific that many days after the game date; stored as `available_at` |
| `valid_days`        | integer, default 0                                       | Length of the window (below); replaces `redemption_window` when set                           |
| `scope`             | `team`, `team_batters`, `team_pitchers`, `team_fielders` | Descriptive only; the checker ignores it. Set it anyway                                       |

### When it's checked

The `check-triggers` CronJob runs each morning (Pacific) and checks **yesterday's** games. One game
triggers a deal at most once.

### Deadlines

Every window is computed from the **game's date** in `America/Los_Angeles`, never from when the
check happens to run, and ends at 23:59:59 Pacific on its last valid day (or at `expires_at_local`
on that day where it applies). Day 1 is the day after the game, when the morning check runs:

| Window                          | Expires                                              |
| ------------------------------- | ---------------------------------------------------- |
| `same_day`                      | Day 1 at 23:59:59, or at `expires_at_local` on day 1 |
| `next_day` (and unknown values) | Day 2 at 23:59:59, or at `expires_at_local` on day 2 |
| `week`                          | Day 7 at 23:59:59; `expires_at_local` is ignored     |

"The morning after the game, breakfast hours only" is `same_day` with `"expires_at_local": "10:30"`.

A deal is never recorded, and never pushed, if its deadline has already passed when the run reaches
it (a late run or a backfill after 10:30am for the breakfast deal). The run logs it at WARN with
`event_id`, `status=EXPIRED` and `reason`, and the `check-triggers` response lists it under
`skipped` (`skippedEvents` counts them). Deals that can never fire are listed there too.

| Rule fields                             | Unlocks                                                                      | Expires                                                                                                               |
| --------------------------------------- | ---------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `redemption_window: next_business_day`  | Midnight of the first weekday after the game (open at once if that is day 1) | 23:59:59 that weekday, or `expires_at_local`                                                                          |
| `valid_days: N` (no unlock)             | Immediately                                                                  | 23:59:59 on day N, or `expires_at_local`                                                                              |
| `unlock_after_days: U`, `valid_days: N` | Midnight Pacific, U days after the game date                                 | 23:59:59 on the Nth day counting the unlock day as the first (`valid_days` defaults to 1; `expires_at_local` ignored) |

Example: the Rams and Carl's Jr. code "unlocks 2 days after the game, expires 1 week after unlock"
is `"unlock_after_days": 2, "valid_days": 7`. For a Sunday Nov 15 game it unlocks Tue Nov 17 and
ends Mon Nov 23 at 11:59pm. The push says "Code unlocks Tue. Claim at Carl's Jr. by Mon Nov 23
11:59pm." and the app reads `availableAt` from the active and claimed deal JSON to show "Code
unlocks Tue".

### Reminders

`send-reminders` runs three times a day, but a user gets at most one reminder per deal per Pacific
day, and only on the day the deal first becomes claimable (it unlocks) and on its last day. The
`notifications` table records what was sent (`kind` is `trigger` or `reminder`); failed sends do not
count and are retried on the next run.

To make a delayed deal locally, `go run . deals trigger <event-id> --unlock-after-days 3
--valid-days 7 --notify=false` builds the window from yesterday's game with the same code the
checker uses.

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

| Metric                 | Meaning                                                         |
| ---------------------- | --------------------------------------------------------------- |
| `win`, `home_win`      | 1 if won (at home), else 0                                      |
| `goals`                | Goals scored                                                    |
| `home_goals_scored`    | Same, but 0 for away games                                      |
| `shutout`              | 1 if the opponent scored 0 (never in a shootout)                |
| `overtime`, `shootout` | 1 if the game was decided in overtime / a shootout, win or lose |

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

**NFL** (`sources/nfl`, ESPN; teams LAR, LACH). Same as basketball, plus a box-score metric. A tie
is neither a win nor a loss; a neutral-site game (international games, the Super Bowl) is never a
home game:

| Metric                  | Meaning                                                             |
| ----------------------- | ------------------------------------------------------------------- |
| `win`, `home_win`       | 1 if won (at home), else 0                                          |
| `points`, `home_points` | Points scored, and the same but 0 for away games                    |
| `interceptions`         | Passes the team's defense intercepted, any venue (pick-sixes count) |

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
