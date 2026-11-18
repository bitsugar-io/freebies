# Deals

A deal is one row in the `events` table, seeded by a goose migration that runs when the API starts.
The `trigger_rule` JSON is in [deals-trigger-rule.md](deals-trigger-rule.md).

## Steps

1. Research the deal ([Finding deals](#finding-deals)), then confirm it on the sponsor's or team's
   own page. If you can't, it ships with `is_active = 0`.
2. Create the migration with `scripts/new-migration.sh <slug>` and fill in every `TODO`.
3. Write `trigger_rule` from a metric the league's source emits
   ([deals-trigger-rule.md](deals-trigger-rule.md)).
4. Give it its own sprite with the `design` skill: reuse one that fits, or draw a new one and ship
   it with `ship.sh` (a migration, so it reaches the app without a release). Every LA deal has its
   own sprite; `gift` is only a placeholder while a new sprite is being drawn, never what ships.
5. Word the offer and condition by the Copy voice section of `docs/ui.md`.
6. Add the event to `TestOnlyLAEventsActive`, run `scripts/check.sh api`, restart the API and check
   `GET /api/v1/events/<id>`.
7. Run `deals audit` before shipping (below). An active deal that can never fire fails it.

## Can the deal ever fire? `deals audit`

```bash
cd services/api && mise exec -- go run . deals audit
```

Lists every event in the database as `OK`, `NO_SOURCE` (no source registered for the league),
`UNKNOWN_METRIC` (the source never emits the rule's metric), `UNKNOWN_TEAM` (the source cannot map
`team_id`), `BAD_RULE` (missing or invalid `trigger_rule`, unknown operator) or `INACTIVE`. For an
inactive event, the detail says what would be wrong if you turned it on. It exits non-zero if any
**active** event is not `OK`. `TestSeededActiveDealsCanFire` runs the same audit over every
migration, so a deal migration that ships an unfireable active deal fails the tests. At run time the
checker logs `WARN deal <id> can never fire` for such a deal and skips it, without failing the run.

A deal you cannot compute yet (a metric no source has) ships with `is_active = 0`.

An at-the-game deal (only fans at the venue can claim it) sets `in_venue = 1` and `venue` (for
example `Intuit Dome`); the app tags it AT THE GAME and names the venue. If it is claimed in the
venue that night, use the `game_night` window; if the ticket is the proof and you redeem later (UCLA
Fat Sal's), use a normal window and it alerts like any other deal.

## Finding deals

Research every source below, every time, and say which ones you checked. Leads from any of them
still need the sponsor's or team's page before a deal goes active.

| Source             | How                                                                                                                                                                                                                                                 | What it's good for                                                                                                                            |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| Google             | In the owner's Chrome (the `claude-in-chrome` tools): "free food when <team> win 2026", "<team> promo code", "<chain> <team> deal"                                                                                                                  | Current-season news, sponsor pages, Instagram and Facebook deal posts                                                                         |
| Reddit             | In the owner's Chrome, through Google: `site:reddit.com/r/FoodLosAngeles <team>`, `site:reddit.com <team> free food` (Reddit itself is blocked to the browser tools and to scripts; Google shows each thread's text)                                | Fans confirming a deal fired, codes, changed terms. Subs: r/FoodLosAngeles, r/ChickFilAforSoCal, the team subs, r/LosAngeles, r/AskLosAngeles |
| Meal Steals        | `https://mealsteals.co/deals/` and its team guides (`/deals/dodgers/`, `/deals/rams/`); the dated archive `/deals/YYYY-MM-DD/` lists what fired each day; its public feed `getOffers?city=Los%20Angeles` (the one its page loads) has the live list | The closest competitor: what's running now and what fired on past dates                                                                       |
| Other trackers     | lahomewin.com, didmyteamearnmefreebies.com (NHL), dealsonmeals.app                                                                                                                                                                                  | Cross-checks and teams Meal Steals skips                                                                                                      |
| Sponsors and teams | The restaurant's promo page and the team's promotions page                                                                                                                                                                                          | The only source that makes a deal active                                                                                                      |

Then check each lead's terms against [deals-trigger-rule.md](deals-trigger-rule.md) ("available for
unlock for 2 days" is a deadline, not a delay), and draw its sprite (step 4).

## Verifying a deal

### The source must be the sponsor or the team

A deal is active only after someone has read it on the restaurant's or the team's own page for the
current season. Fan sites, deal blogs and last season's page don't count.

Record what you found, in a comment above the INSERT:

```sql
-- Anaheim Ducks - McDonald's McGriddles. Verified for 2026-27:
-- https://www.nhl.com/ducks/fans/partner-promotions
```

Put the same page in `offer_url` so users can check it too.

### Unverified means inactive

If you can't confirm it, insert the row with `is_active = 0` and say why:

```sql
-- LA Kings - El Pollo Loco. NOT verified for 2026-27, so inactive.
```

Later, when it's confirmed, a new migration sets `is_active = 1` and updates `offer_url` and the
comment.

### What to capture from the page

- The exact trigger ("win at home", "7 or more strikeouts", "score in the first half at home").
- The reward and its price ("$2", "free with a large drink").
- How to claim: app, promo code, in store, participating locations.
- The deadline: same day, next day, a clock time ("breakfast hours", "until 10:30am"), a week.
- Limits: membership, age, residency, purchase minimum.

### LA only

The product covers Los Angeles and Orange County, nothing else.

| `region_code` | Use for                                                                                                           |
| ------------- | ----------------------------------------------------------------------------------------------------------------- |
| `us-ca-la`    | LA County teams: Dodgers, Lakers, Clippers, Sparks, Rams, Chargers, Kings, LAFC, Galaxy, Angel City FC, USC, UCLA |
| `us-ca-oc`    | Orange County teams: Angels, Ducks                                                                                |

- An active deal with any other region fails `TestOnlyLAEventsActive`.
- Deals from outside LA are deactivated (`is_active = 0`), never deleted, so past `triggered_events`
  still resolve.

### Retiring a deal

Never delete a deal row. A new migration sets `is_active = 0`; its Down sets it back to 1. Remove
the event from the `want` list in `TestOnlyLAEventsActive`.

## The migration

### Create it

```bash
.claude/skills/ops/scripts/new-migration.sh dodgers-habit     # writes NNN_dodgers_habit.sql
```

The script takes the next free number in `services/api/internal/db/migrations/`. Never edit a
migration that has shipped; add a new one. A migration that only ran on your laptop can be edited,
then `just clean` and restart.

Good examples: `026_la_only.sql` (several LA deals, an inactive one, `expires_at_local`) and
`022_dodgers_habit.sql` (one deal).

### `events` columns

| Column                               | Example                                     | Notes                                                                       |
| ------------------------------------ | ------------------------------------------- | --------------------------------------------------------------------------- |
| `id`                                 | `event-nhl-ana-mcdonalds`                   | `event-<league>-<team>-<partner>`, lowercase                                |
| `offer_id`                           | `ana-mcdonalds`                             | `<team>-<partner>`                                                          |
| `team_id`                            | `ANA`                                       | The code the league's source uses (`LAD`, `LAA`, `LAK`, `LAFC`)             |
| `team_name`                          | `Anaheim Ducks`                             | Full name, plain text                                                       |
| `league`                             | `nhl`                                       | Lowercase league id; must match a `leagues` row and the source's `League()` |
| `team_color`                         | `#F47A38`                                   | Legacy; may be `''`. The 8-bit UI never shows team colors (`docs/ui.md`)    |
| `icon`                               | `🍳`                                        | Legacy; may be `''`. The app shows the sprite; pushes carry no emoji        |
| `partner_name`                       | `McDonald's`                                | Escape quotes as `''`                                                       |
| `offer_name`                         | `$2 McGriddles`                             | The reward, short                                                           |
| `offer_description`                  | `Get a $2 ... at participating McDonald's.` | Reward, trigger, claim steps, deadline, limits                              |
| `trigger_condition`                  | `Ducks win at home`                         | Shown in the app and pushes as the reason                                   |
| `trigger_rule`                       | `{"metric": "home_win", ...}`               | See [deals-trigger-rule.md](deals-trigger-rule.md)                          |
| `region_code`                        | `us-ca-oc`                                  | `us-ca-la` or `us-ca-oc` only (see "LA only" above)                         |
| `offer_url`                          | sponsor or team page                        | The page you verified it on                                                 |
| `affiliate_url`, `affiliate_tagline` | `''`                                        | Unused for LA deals                                                         |
| `is_active`                          | `1`                                         | `0` until verified                                                          |
| `sprite_id`                          | `griddle-sandwich`                          | A `sprites` row; `gift` if none fits                                        |

`created_at` and `updated_at` default to now.

### Copy

`offer_name`, `offer_description` and `trigger_condition` follow the Copy voice section of
`docs/ui.md`: outcome and reward first, plain sentences, real numbers, no emoji, no exclamation
marks. Name the restaurant and team in plain text only.

### Down

Every Up has a Down that undoes it exactly: `DELETE` the rows you inserted, or flip `is_active` back
for an UPDATE.

### Followers

On start the API subscribes everyone who follows a team to that team's new active deals, so a
migration needs no subscription rows.

## Testing a deal

### Unit tests

`services/api/internal/api/handlers/la_only_test.go` (`TestOnlyLAEventsActive`) runs every migration
into an in-memory database and checks that:

- every active event is in `us-ca-la` or `us-ca-oc`,
- every active event has a sprite,
- the set of active event ids equals its `want` list.

Add a new active deal to `want`, and remove a retired one. Keep the list sorted. An inactive deal
can get its own assertion, like the Kings check at the end of the test.

Run:

```bash
cd services/api && mise exec -- go test ./internal/api/handlers/ ./internal/triggers/ ./internal/worker/
```

Or everything with `.claude/skills/ops/scripts/check.sh api`.

A new rule shape (a new window or an `expires_at_local` case) gets a row in `TestExpirationFor`
(`internal/triggers/checker_test.go`). A change to push wording gets a case in
`internal/worker/copy_test.go`.

### End to end

```bash
.claude/skills/ops/scripts/api.sh restart         # runs the new migration
curl -s localhost:8080/api/v1/events/event-mlb-lad-habit | python3 -m json.tool
sqlite3 services/api/freebie.db "SELECT id, is_active, region_code, sprite_id FROM events WHERE id = 'event-mlb-lad-habit'"
```

Check the fields read the way the copy rules want, the sprite is right, and `is_active` is what you
meant.

To see the unlocked deal and its push in the app without waiting for a game:

```bash
cd services/api && mise exec -- go run . deals trigger event-mlb-lad-habit --hours 6 --notify=false
```

`--notify=true` sends a real push to every subscriber of that deal in the local database.

### Before the pull request

- [ ] `deals audit` shows every active deal `OK`
- [ ] Verified on the sponsor's or team's page, URL in the comment and in `offer_url`
- [ ] `region_code` is `us-ca-la` or `us-ca-oc`
- [ ] `trigger_rule` uses a metric the source emits, or a NOTE says it can't trigger yet
- [ ] Its own sprite, shipped by migration (`ship.sh`); not `gift`
- [ ] Copy follows `docs/ui.md`
- [ ] Down undoes Up
- [ ] `TestOnlyLAEventsActive` updated and the tests pass
- [ ] If the deal appears in store copy, `apps/mobile/APP-STORE.md` and the README examples match
