# Game data sources

Which provider feeds each league's `Source` (`services/api/internal/sources/`), why, what its terms
say, and what to switch to if it breaks. Researched on 2026-11-15 by calling each candidate from a
server-side client (no browser).

## Decision table

| League           | Chosen source                                                                                                     | Why                                                                                                                                                                                                                                  | Terms note                                                                     | Fallback                                                                                        |
| ---------------- | ----------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------- |
| MLB              | MLB Stats API (`statsapi.mlb.com`) via `mlb-sdk`, `sources/mlb`                                                   | Official feed, full box scores (strikeouts, steals, double plays)                                                                                                                                                                    | Unofficial public API; MLB terms restrict commercial use. Needs the 406 retry  | None needed                                                                                     |
| NHL              | NHL web API (`api-web.nhle.com/v1/score/{date}`), `sources/nhl`                                                   | Official feed, no key, final state and OT/SO flags                                                                                                                                                                                   | Unofficial and undocumented; terms restrict commercial automated use           | ESPN `hockey/nhl` scoreboard                                                                    |
| NBA (LAL, LAC)   | ESPN site API `basketball/nba`, `sources/nba`                                                                     | `cdn.nba.com` liveData answers **403 Access Denied** to server traffic, even with browser headers; `stats.nba.com` times out. ESPN returns final scores for past dates reliably                                                      | ESPN publishes no API terms; its site terms cover personal, non-commercial use | Sportradar or the NBA's licensed feed; `cdn.nba.com` only from a residential or allow-listed IP |
| WNBA (LAS)       | ESPN `basketball/wnba`, `sources/wnba`                                                                            | `cdn.wnba.com` has the same 403 wall                                                                                                                                                                                                 | As NBA                                                                         | Sportradar                                                                                      |
| NFL (LAR, LACH)  | ESPN `football/nfl` scoreboard plus `summary?event=` for interceptions, `sources/nfl`                             | No public official feed (`static.www.nfl.com` scorestrip is gone, 404). The summary box score carries per-team interceptions                                                                                                         | As NBA                                                                         | Sportradar, NFL licensed data                                                                   |
| MLS (LAFC, LAG)  | ESPN `soccer/usa.1`, `sources/mls`                                                                                | MLS's own stats API is not publicly reachable. ESPN's scoreboard lists every goal with its minute, so first-half goals need no second call. Checked on 256 completed MLS matches: goal events matched the final score in all of them | As NBA                                                                         | Opta or Stats Perform feed; MLS licensed data                                                   |
| NWSL (ACFC)      | ESPN `soccer/usa.nwsl`, `sources/nwsl`                                                                            | NWSL's own API (`api-sdp.nwslsoccer.com`) returned 500; ESPN has the same goal-event format as MLS                                                                                                                                   | As NBA                                                                         | As MLS                                                                                          |
| NCAA (USC, UCLA) | ESPN `football/college-football` (groups=80) and `basketball/mens-college-basketball` (groups=50), `sources/ncaa` | `ncaa-api.henrygd.me` is a third-party proxy, `data.ncaa.com` returned 404; ESPN has both sports with scores for past dates                                                                                                          | As NBA                                                                         | NCAA statistics licensed data                                                                   |

ESPN is one shared client for six leagues (`sources/espn`), so one fix covers all of them.

## What to know about ESPN

- The API is
  `https://site.api.espn.com/apis/site/v2/sports/{sport}/{league}/scoreboard?dates=YYYYMMDD` and
  `/summary?event=ID`. It needs no key and no special headers. It is unofficial and undocumented,
  and its shape can change without notice.
- ESPN files a game under its own Eastern-style day. The client fetches the requested day and the
  next, then keeps the games whose start time falls on the requested **Pacific** date.
- Date ranges (`dates=A-B`) fail for some leagues, so the client always asks for one day.
- A game is final only when `status.type.completed` is true and `state` is `post`. Postponed,
  canceled and live games never count.
- A neutral-site game is never a home game (`home_*` metrics are 0), whichever side ESPN labels
  "home".
- Soccer: a goal is a `details` entry with `scoringPlay` (shootout kicks excluded). An own goal
  belongs to the team it benefits. If the goal events do not add up to the final score, the game is
  an error rather than a guess.
- The default college basketball scoreboard lists only featured games; `groups=50&limit=400` returns
  all of Division I.

## College basketball

The `teams` table lists USC and UCLA as sport `football`. The `ncaa` source reads both football and
men's basketball, and for each game emits the generic metrics plus the same metrics prefixed with
the sport (`football_win`, `basketball_points`, ...). A deal that names `basketball_home_win` can
never fire on a football game. The `teams` row's `sport` only picks the badge and season dates; if
USC or UCLA basketball deals ship, set the row's season to cover winter or add a basketball team
row.

## Adding or changing a source

1. Curl the candidates (official feed first), for a past date and for a date with no game.
2. Write the decision here.
3. For an ESPN league, add a package like `sources/nba`: team-id map (ESPN ids, from `/teams`),
   `Metrics`, and the builder. For anything else, implement `sources.Source` including `Metrics()`.
4. Fixtures: `services/api/test/fixtures/espn_{sport}_{league}_{YYYYMMDD}.json`, trimmed from real
   responses, served by `espntest.Server`. Cover home win, away win, loss, draw (soccer), not-final
   and no game.
5. Blank-import the package in `cmd/serve.go` and in the audit test
   (`internal/triggers/audit_test.go`).
6. Check against the live provider:

   ```bash
   cd services/api && mise exec -- go run . verify-source --league mls --team LAFC --date 2026-09-09
   ```
