# LA deal verification (15 Nov 2026)

Web research only. No code or database changes. Every deal below was checked against a **primary
source**: the team's own partner or promotions page, the sponsor's own page, or an official team or
sponsor post or press release. Meal Steals, lahomewin.com, Dodger Blue, Dodgers Nation, Slickdeals
and Doctor of Credit were used **only as leads**. Nothing here copies their wording, and none of
them is cited as a source.

## Status key

| Status       | Meaning                                                                               |
| ------------ | ------------------------------------------------------------------------------------- |
| VERIFIED     | Primary page fetched on 15 Nov 2026, live and tied to the current season              |
| VERIFIED (u) | Primary page live today but **undated**. Its season is inferred from context          |
| LIKELY       | Primary source for a prior season or earlier in this season, with no change announced |
| UNVERIFIED   | No primary source found, or the primary page blocked our fetch                        |
| DISCONTINUED | The primary page was removed or no longer lists the deal                              |

Season context on 15 Nov 2026: NHL, NBA and NFL 2026-27 are in season. MLB 2026 ended, so nothing
from MLB fires until about late March 2027. MLS 2026 is in its playoffs or over, and NWSL and WNBA
2026 are over.

**Metric names** in the trigger mapping reuse the `trigger_rule` metrics already in
`services/api/internal/db/migrations` (`home_win`, `win`, `runs`, `home_runs_scored`, `strikeouts`,
`home_double_plays`, `home_stolen_bases`, `home_first_half_goals`, `score`). Names marked **new** do
not exist yet.

**In-venue** means only fans at the venue can claim the reward. The App Store listing promises that
every deal is checked against the restaurant's or team's own page, so these deals must carry a clear
"at the game only" label, or be left out.

---

## 1. Summary table

| #   | Team       | Sponsor          | Trigger                                   | Status               | In-venue | Action                    |
| --- | ---------- | ---------------- | ----------------------------------------- | -------------------- | -------- | ------------------------- |
| 1   | Ducks      | McDonald's       | Home win                                  | VERIFIED             | No       | **Keep active**           |
| 2   | Ducks      | Chick-fil-A      | 5 goals at home                           | VERIFIED             | **Yes**  | Keep active, labeled      |
| 3   | Kings      | El Pollo Loco    | Home win                                  | DISCONTINUED?        | No       | **Remove**                |
| 4   | LAFC       | Ono Hawaiian BBQ | First-half goal at home (wording differs) | VERIFIED (u)         | No       | Seed inactive (offseason) |
| 5   | LAFC       | Chick-fil-A      | Home win                                  | VERIFIED             | **Yes**  | Seed inactive, labeled    |
| 6   | Clippers   | Chick-fil-A      | Visitor misses 2 straight FTs in Q4       | VERIFIED (u)         | **Yes**  | Keep active, labeled      |
| 7   | Dodgers    | Panda Express    | Home win                                  | UNVERIFIED           | No       | Seed inactive             |
| 8   | Dodgers    | Jack in the Box  | 7+ strikeouts                             | UNVERIFIED (trigger) | No       | Seed inactive             |
| 9   | Dodgers    | McDonald's       | 6+ runs                                   | UNVERIFIED           | No       | Seed inactive             |
| 10  | Dodgers    | ampm             | Stolen base at home                       | UNVERIFIED           | No       | Seed inactive             |
| 11  | Dodgers    | The Habit        | Double play at home                       | LIKELY (2027)        | No       | Seed inactive             |
| 12  | Dodgers    | Uber Eats        | Save on the road                          | UNVERIFIED (terms)   | No       | Seed inactive             |
| 13  | Angels     | Del Taco         | 5+ runs at home                           | LIKELY / page gone   | No       | Seed inactive             |
| 14  | Angels     | McDonald's       | Any win (home or road)                    | VERIFIED (2026)      | No       | Seed inactive             |
| 15  | Angels     | Mountain Mike's  | Shutout                                   | UNVERIFIED           | No       | **Remove**                |
| 16  | Rams       | Carl's Jr.       | Interception                              | VERIFIED (u)         | No       | **Keep active** (new)     |
| 17  | Galaxy     | NORMS            | 1+ goal at home                           | VERIFIED (u)         | No       | Seed inactive (new)       |
| 18  | Angel City | Once Upon a Farm | "When We Win"                             | UNVERIFIED           | ?        | Remove                    |
| 19  | Lakers     | Jack in the Box  | Win (terms unknown)                       | UNVERIFIED           | ?        | Do not seed               |
| 20  | Kings      | McDonald's       | Goal in last minute of 2nd period         | UNVERIFIED           | ?        | Do not seed               |
| 21  | Chargers   | Firehouse Subs   | 40+ yd field goal                         | UNVERIFIED           | ?        | Do not seed               |
| 22  | LAFC       | DoorDash         | Win                                       | UNVERIFIED           | No       | Do not seed               |

**Active right now (in-season, verified, not in-venue):** Ducks/McDonald's and Rams/Carl's Jr. That
is the full list. Everything else is offseason, in-venue only, or unconfirmed.

---

## 2. Deal detail: candidates from the earlier research

### 1. Anaheim Ducks: McDonald's, $2 McGriddles after a home win

- **Trigger (source wording):** "If the Ducks win at home, get $2 Sausage, Egg & Cheese McGriddles
  the next day on the McDonald's app."
- **Reward:** Sausage, Egg & Cheese McGriddles for $2.
- **Redemption:** McDonald's app, "Rewards & Deals" section. Download, registration and rewards
  opt-in required. "Offer valid 1x at participating McDonald's the day after the Anaheim Ducks win
  at home." "Breakfast hours only." Valid only in the Los Angeles/Orange County and Palm Springs
  DMAs. No code.
- **Season:** The page sits under the "2026-27 Season Promotions" heading.
- **Sources:** https://www.nhl.com/ducks/fans/partner-promotions and
  https://www.nhl.com/ducks/fans/mcdonalds-ducks-win (both fetched 15 Nov 2026).
- **Status:** VERIFIED.
- **Trigger metric:** `home_win == 1`, `redemption_window: next_day`, with
  `expires_at_local: "10:30"` for breakfast hours. The existing seed uses `same_day` plus 10:30, so
  check how the worker counts "same day" against "the day after".
- **In-venue:** No.
- **Action:** Keep active.

### 2. Anaheim Ducks: Chick-fil-A "Goals for Chicken"

- **Trigger:** "If the Ducks score 5 goals at home, everyone at the Honda Center will receive an
  Original Chick-fil-A Chicken Sandwich." The sponsor's page says "5 or more goals."
- **Reward:** One Original Chick-fil-A Chicken Sandwich.
- **Redemption:** Chick-fil-A app, "My Rewards." Location services must be on. "Claim period is from
  when the trigger activates, until 11:59 PM Pacific time that night (while at the Honda Center)."
  The reward expires after 3 days, not counting Sundays. Participating Greater LA/OC restaurants
  only.
- **Season:** "2026-27 Season Promotions".
- **Sources:** https://www.nhl.com/ducks/fans/partner-promotions and https://www.cfasocal.com/
- **Status:** VERIFIED.
- **Trigger metric:** `home_goals >= 5` (**new**: NHL source, home game only).
- **In-venue:** **Yes.** The reward can only be claimed at Honda Center.
- **Action:** Keep active only with an "At Honda Center only" label. Better: send no push to people
  who aren't at the game, and show it in the feed as information.

### 3. LA Kings: El Pollo Loco, free BRC Burrito after a home win

- **Trigger (last primary wording):** "El Pollo Loco is giving fans a FREE BRC Burrito with the
  purchase of any large drink after each LA Kings victory at home!"
- **Reward:** Free BRC Burrito with the purchase of a large drink.
- **Redemption:** Loco Rewards app, coupon code "KINGS".
- **Season:** The last capture of the page is from Aug 2025, with a 2024-25 yearbook in the nav
  (Wayback Machine,
  `web.archive.org/web/20250813175203/https://www.nhl.com/kings/fans/partner-promotions`). **The
  live page now returns the Kings "Oops! Error 404" page.** Search engines still show the old
  snippet. No 2026-27 mention appears on any Kings or El Pollo Loco page, the Kings 2026-27
  promotional schedule (published 17 Aug 2026), or in El Pollo Loco 2026 releases.
- **Source:** https://www.nhl.com/kings/fans/partner-promotions (404 today). Archived copy above.
- **Status:** DISCONTINUED, or at least not confirmed for 2026-27. Treat it as gone.
- **Trigger metric:** `home_win == 1`, next day.
- **Action:** **Remove.** Do not seed it, even as inactive, until a 2026-27 primary page appears.
  This changes the plan in la-pivot-options.html, which proposed seeding this deal.

### 4. LAFC: Ono Hawaiian BBQ, $5.99 chicken plate

- **Trigger, LAFC wording:** "When LAFC scores in the first half—YOU WIN!" ... "available the next
  business day following the LAFC home game."
- **Trigger, Ono wording:** the headline reads "When LAFC Scores First in the First Half". Terms:
  "valid the next business day after a MLS LAFC home game only."
- **The two primary pages disagree** on "scores in the first half" versus "scores first in the first
  half". Use the stricter reading, Ono's, because Ono honors the deal.
- **Reward:** Chicken Plate Lunch for $5.99. Hawaiian BBQ Chicken, Chicken Katsu or Teriyaki Chicken
  plate, while supplies last.
- **Redemption:** Online only, for registered users, with code **LAFCSCORES**. "Not valid on
  third-party platforms." Limit one per order, per registered user, per day. Participating Greater
  LA locations.
- **Season:** Undated. The LAFC promotions page is the 2026 page (it advertises a "2026 Ford
  Maverick" sweepstakes). Earlier notes said "playoffs only". Neither page says that; both say "MLS
  LAFC home game".
- **Sources:** https://www.lafc.com/club/promotions/ and https://onohawaiianbbq.com/lafc/
- **Status:** VERIFIED (u) for 2026. Likely to return in 2027, but not announced.
- **Trigger metric:** `home_first_half_goals >= 1` (already exists). Under Ono's reading it needs
  **new** play-by-play: LAFC scores the match's first goal, and that goal comes in the first half.
  Redemption window: next business day, which is **new** (skip weekends).
- **In-venue:** No.
- **Action:** Seed inactive, then turn it on for the 2027 MLS home opener after rechecking both
  pages.

### 5. LAFC: Chick-fil-A after a home win

- **Trigger:** "When LAFC wins a home game, all fans at BMO Stadium are eligible to receive a free
  Chick-fil-A Chicken Sandwich by going into the CFA app and claiming their reward before leaving
  the stadium."
- **Reward:** Original Chick-fil-A Chicken Sandwich.
- **Redemption:** Claim in the CFA app before leaving the stadium, then redeem within 3 business
  days.
- **Season:** 2026 page.
- **Sources:** https://www.lafc.com/club/promotions/ and https://www.cfasocal.com/ ("if LAFC wins
  ... everyone at BMO Stadium receives an Original Chick-fil-A® Sandwich in the Chick-fil-A® App").
- **Status:** VERIFIED (2026).
- **Trigger metric:** `home_win == 1`.
- **In-venue:** **Yes.**
- **Action:** Seed inactive with an in-venue label, and recheck in 2027.

### 6. LA Clippers: Chick-fil-A, visitors miss two straight free throws in the 4th quarter

- **Trigger (sponsor wording):** "if the visiting team misses 2 consecutive free throws in the 4th
  quarter, everyone at Intuit Dome receives an Original Chick-fil-A® Sandwich in the Chick-fil-A®
  App".
- **Reward:** Original Chick-fil-A Sandwich.
- **Redemption:** Chick-fil-A app, for fans at Intuit Dome. The claim window is not stated on the
  primary page.
- **Season:** The page is undated. It names Intuit Dome and is live during the 2026-27 NBA season.
  No Clippers-hosted page was found. nba.com/clippers/promotions has not been captured since 2021.
- **Source:** https://www.cfasocal.com/ (the sponsor's SoCal page).
- **Status:** VERIFIED (u).
- **Trigger metric:** **new** NBA play-by-play metric: `q4_opp_consecutive_ft_misses >= 2`, home
  only. "Consecutive" needs care: two misses in a row by the visitors, any shooters, Q4 only.
- **In-venue:** **Yes.** Older versions paid out to the whole SoCal app region, but the current page
  says "everyone at Intuit Dome".
- **Action:** Keep active only with an in-venue label. It also needs NBA play-by-play, so it is low
  priority.

### 7. Dodgers: Panda Express, $7 Panda Plate after a home win

- **What the leads say:** a 2-entrée plate for $7 the day after a home win, for Panda Rewards
  members with code DODGERSWIN, one per account per day, extended into the 2026 postseason.
- **Primary source:** https://www.pandaexpress.com/dodgerswin returned **HTTP 403** to our fetcher,
  and no archived copy exists. **Not confirmed on a primary page.** A Fox LA story says its
  information came from Panda Express and the Dodgers. That is a news report, not primary.
- **Status:** UNVERIFIED (the primary page is blocked). A human should open the page in a browser.
- **Trigger metric:** `home_win == 1`, next day, until 11:59 pm.
- **Action:** Seed inactive. Before 2027 Opening Day, a person checks pandaexpress.com/dodgerswin
  and records `verified_at`.

### 8. Dodgers: Jack in the Box, 7+ strikeouts

- **Primary:** The official Dodgers X account, 27 Mar 2026: "Stop by @JackBox and get a free Jumbo
  Jack® with a large Coca-Cola purchase, available in-store or on the Jack App with promo code
  GODODGERS26." (https://x.com/Dodgers/status/2037387288746049649, read through a public embed
  mirror.)
- **The trigger (7+ strikeouts, home or away, next day) is not in that post.** It comes only from
  leads. No jackinthebox.com page was found; /dodgers returns 404.
- **Status:** Reward and code VERIFIED for 2026. Trigger UNVERIFIED. The code is year-specific, so
  2027 will need a new code.
- **Trigger metric:** `strikeouts >= 7` (scope `team_pitchers`), next day.
- **Action:** Seed inactive. In 2027, confirm the trigger on a Jack in the Box or Dodgers page.

### 9. Dodgers: McDonald's, 6+ runs, free 6-piece McNuggets

- **Primary:** none fetched. Leads cite a @mcdonaldswestcoast TikTok post ("6 or more"). One lead
  says home games only, another says any game. The two disagree.
- **Status:** UNVERIFIED.
- **Trigger metric:** `runs >= 6`, or `home_runs_scored >= 6` if it is home only. Next day.
- **Action:** Seed inactive. Confirm the home/away scope from McDonald's in 2027.

### 10. Dodgers: ampm, stolen base at home

- **Primary:** none found for 2026. The newest material is 2025 lead coverage, which says hot dog
  plus 12 oz Coca-Cola through the ampm app. A Dodgers 2026 release mentions only an ampm
  bobblehead.
- **Status:** UNVERIFIED.
- **Trigger metric:** `home_stolen_bases >= 1`, next day.
- **Action:** Seed inactive.
- **Note:** "Dodger Dog" is a Dodgers trademark. Describe the reward as "a hot dog", unless the
  sponsor page uses the mark.

### 11. Dodgers: The Habit, double play at home

- **Trigger:** "the Los Angeles Dodgers turn a double play during a 2026 regular season Dodgers home
  game." Qualifying games run from 26 Mar 2026 to 27 Sep 2026.
- **Reward:** "FREE Double Char with any qualifying $8+ purchase!" Cart subtotal must be at least $8
  before discounts and tax.
- **Redemption:** Code **DODGERS26** in the "My Offers" tab of a Habit account. "Offer becomes
  available the day immediately following the Qualifying Game" and is "only valid until close of
  business that day."
- **Source:** https://www.habitburger.com/dodgers/. Background: the 2025 Dodgers–Habit partnership
  press release,
  https://www.mlb.com/press-release/los-angeles-dodgers-and-habit-enter-into-partnership
- **Status:** VERIFIED for 2026, which has now ended. LIKELY for 2027: the partnership is
  multi-year, but the code and dates are year-specific.
- **Trigger metric:** `home_double_plays >= 1` (scope `team_fielders`), next day.
- **Action:** Seed inactive, and update the code and dates when the 2027 page goes up.

### 12. Dodgers: Uber Eats, save on the road

- **Primary:** Dodgers X, 20 May 2026: "A save on the road means Dodger fans save at home with Uber
  Eats!" (https://x.com/Dodgers/status/2056964408845414471).
- **Not confirmed:** the reward ($15 off $20+) and the code (SAVEAWAY) come from leads only. The
  post text has neither.
- **Status:** Trigger VERIFIED for 2026. Terms UNVERIFIED.
- **Trigger metric:** **new** `away_saves >= 1` (MLB boxscore has saves per pitcher).
- **Action:** Seed inactive. This is a delivery discount, not free food, so consider whether it fits
  the brand.

### 13. Angels: Del Taco, 5+ runs at home

- **Trigger (primary, archived 15 Mar 2026):** "When the Angels score five (5) or more runs at a
  home game, get two (2) free The Del Tacos with any purchase via the Del Yeah! Rewards app the
  following day."
- **Redemption:** "valid from 10am PT to 11:59pm PT the day after." Participating Del Taco
  restaurants in LA, Orange, Riverside, Ventura and San Bernardino counties. Limit one per guest.
  Not valid with third-party delivery.
- **Odd detail:** the page title reads "Del Taco 7 Run Promotion", but the body says five. Use the
  body text.
- **Source:** https://www.mlb.com/angels/tickets/specials/homestand-partner-offers/del-taco. It was
  live in March 2026, with the 2026 spring schedule in its nav (Wayback 20260315000122). **It now
  returns 404**, and the Angels' partner-offers hub no longer lists it (absent from the May 2026
  capture and from today's page).
- **Status:** LIKELY ran in early 2026, then removed mid-season. 2027 unknown. Treat as DISCONTINUED
  until it reappears.
- **Trigger metric:** `home_runs_scored >= 5`, next day, 10:00–23:59.
- **Action:** Seed inactive at most. Do not activate without a fresh primary page.

### 14. Angels: McDonald's "Wins Promotion"

- **Trigger:** "Whenever the Angels win (at home and on the road), fans get the World Famous medium
  fries for FREE the following day with a $2 minimum purchase from the McDonald's App."
- **Redemption:** McDonald's app, after download and registration. "Only valid within the Greater
  LA/Orange County DMA."
- **Source:** https://www.mlb.com/angels/tickets/specials/homestand-partner-offers ("Season Long
  Offers"). Live today, with 2026 nav and copyright.
- **Status:** VERIFIED for 2026, which has ended. LIKELY for 2027.
- **Correction:** the earlier inventory said "fries/McGriddles". The current page mentions only
  fries.
- **Trigger metric:** `win == 1` (any venue), next day.
- **Action:** Seed inactive, and reactivate at 2027 Opening Day after a recheck.

### 15. Angels: Mountain Mike's, shutout pizza

- **Primary:** Mountain Mike's press releases describe a "Shutout Challenge" for 2022–2024 (a free
  mini one-topping pizza the day after an Angels shutout, valid "for the 2024 Angels Baseball season
  only"). Its April 2025 release ("fourth year as an Official Pizza Partner") **does not mention a
  shutout offer**. mountainmikespizza.com/angels shows no Angels promotion today, and
  mountainmikespizza.com/kings returns 404.
- **Status:** UNVERIFIED for 2025 and 2026. Probably discontinued.
- **Trigger metric:** **new** `shutout` (opponent runs == 0).
- **Action:** Remove.

### Lakers, Rams, Chargers, Galaxy, Angel City, Sparks, USC, UCLA

See section 3 for new finds (Rams, Galaxy) and section 4 for unconfirmed leads. Results by team:

- **Lakers.** https://www.nba.com/lakers/promotions lists no result-triggered food offer for
  2026-27.
- **Chargers.** https://www.chargers.com/promotions has no game-event offer.
- **Sparks.** No result-triggered deal. 2026 partnerships such as Zespri are about community
  programs.
- **UCLA.** The promotional calendars (https://uclabruins.com/ucla-athletics-promotional-calendars)
  list only unconditional giveaways, such as hot-dog vouchers and rooter-bus Chick-fil-A sandwiches.
  None are tied to results.
- **USC.** No result-triggered deal found.
- **Angel City FC.** The club's 18 Jun 2024 release says Once Upon a Farm has "Ownership of the
  'When We Win' promotion on game days". No terms were published anywhere we could find. UNVERIFIED,
  so remove.

---

## 3. New LA deals found (verified on primary pages)

### 16. LA Rams: Carl's Jr. "Famous Star Picks", interception

- **Trigger:** "When your Los Angeles Rams snag an interception you win a free Famous Star®."
- **Reward:** Free Famous Star with Cheese. "No purchase necessary."
- **Redemption:** "Only available in the Carl's Jr. App!" ... "Redeemable only in the Carl's Jr.
  app. online ordering platform or in-restaurant by registering my rewards users. Promotional code
  available for unlock 2 days following the qualifying game. Offer expires in app 1 week from
  unlock." Code **RAMS6617**. Participating LA-area Carl's Jr. only.
- **Season:** Undated, but it is the live Rams gameday promotions page during the 2026 NFL season.
- **Source:** https://www.therams.com/game-day/promotions
- **Status:** VERIFIED (u).
- **Trigger metric:** **new** NFL source: `interceptions >= 1` (defense, any venue). Redemption
  window: unlocks 2 days after the game and runs 7 days after unlock. Both are **new** window types.
- **In-venue:** No.
- **Action:** **Keep active.** It is in season and region-wide. The app needs an NFL source first.

### 17. LA Galaxy: NORMS, a goal at home

- **Trigger:** "Offer valid when LA Galaxy score at least one goal at home games only."
- **Reward:** "NORMS Rewards members win a FREE stack of Buttermilk Hotcakes when the LA Galaxy
  scores a goal!"
- **Redemption:** Present the NORMS Rewards app at checkout at any Southern California NORMS,
  dine-in or to-go. "NOT valid on NORMS.com, NORMS Rewards App orders, Third-Party Delivery or
  Catering orders." "Valid for 3 days after goal is scored." One per person, one-time use.
- **Season:** Undated, live on lagalaxy.com. The MLS 2026 regular season has ended.
- **Source:** https://www.lagalaxy.com/norms
- **Status:** VERIFIED (u) for 2026. LIKELY for 2027.
- **Trigger metric:** `home_goals >= 1` (MLS; the existing `score >= 1` with a home-only scope).
  3-day window.
- **In-venue:** No.
- **Action:** Seed inactive now, then activate for the 2027 Galaxy home opener after a recheck.

### Not LA-specific, out of scope

MLB-wide offers, such as Mountain Dew Baja Blast home runs and Firehouse Subs MLB extra innings,
were seen as leads only. Neither was verified, and neither is an LA team deal.

---

## 4. Leads not confirmed on any primary page (do not ship)

| Lead (from)                                                              | What we checked                                                                       | Result                                                                                                                                                                                |
| ------------------------------------------------------------------------ | ------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| Lakers / Jack in the Box, 2 tacos after a win (lahomewin, Meal Steals)   | nba.com/lakers/promotions, Jack in the Box releases                                   | UNVERIFIED                                                                                                                                                                            |
| Kings                                                                    | McDonald's "McFlurry Minute", goal in the last minute of P2                           | LIKELY+: nhl.com/kings/fans/redeem (indexed text: "all games this season, home and away"); Kings on X 27 Mar 2026; lahomewin.com tracks it for 2026-27. Metric `p2_last_minute_goals` | Active |
| Chargers / Firehouse Subs, 40+ yd Dicker field goal (lahomewin)          | chargers.com/promotions; Firehouse's offer-terms page is a JS app we could not render | UNVERIFIED. The same mechanic exists for the Jaguars                                                                                                                                  |
| LAFC / DoorDash, $10 off after a win (lahomewin cites Instagram Stories) | lafc.com promotions page does not list it                                             | UNVERIFIED                                                                                                                                                                            |
| Dodgers / Mr. Ramen (Meal Steals)                                        | Not checked                                                                           | UNVERIFIED                                                                                                                                                                            |
| Angels / Chick-fil-A, 7+ runs (old Angels page)                          | Angels CFA page 404. cfasocal.com now lists only "unannounced giveaway dates"         | DISCONTINUED as a trigger deal                                                                                                                                                        |

---

## 5. Trademark and IP notes

1. **Team names and marks.** "Dodgers", "Ducks", "LAFC", "Rams" and the rest are trademarks, and
   league marks (NHL Shield, MLS shield and so on) are registered. Using a team name as plain text
   to say what triggers a deal is generally fine as nominative use. Use **no team logos, wordmark
   fonts, uniforms or team colors**, and never imply the app is official, endorsed or a partner. Add
   a footer line: "Not affiliated with any team, league or restaurant."
2. **Product names with ®.** These are registered marks: Jumbo Jack®, Famous Star®, Original
   Chick-fil-A® Chicken Sandwich, Chicken McNuggets, McGriddles, Double Char, Panda Plate, BRC
   Burrito, The Del Taco and Buttermilk Hotcakes (descriptive). Use them only to name the reward, in
   plain text, without logos or product photos, unless the sponsor grants permission. Draw our own
   pixel sprites; don't trace product art.
3. **"Dodger Dog"** is a Dodgers mark. Write "hot dog" unless a sponsor's own primary page uses the
   term.
4. **Campaign names are sponsor copy.** "Goals for Chicken", "Famous Star Picks" and "Fowl Shot"
   belong to the sponsors. Don't use them as our headlines. Write our own neutral line, for example
   "Ducks win at home → $2 breakfast sandwich tomorrow (McDonald's app)", and link to the source.
5. **Our tagline.** "When LA wins, you win" is close to LAFC's and Ono's "YOU WIN!" copy and to Rams
   copy ("you win a free..."). The phrase is common, but keep it out of any single team's or
   sponsor's deal card, so it doesn't read as their slogan.
6. **Copy rule.** Don't copy wording from competitor apps. The quotes in this report are from
   primary pages, for our records only. App copy should paraphrase them and link to `offer_url` (the
   primary page).
7. **Promo codes.** Showing a code that a sponsor publishes publicly, such as LAFCSCORES or
   RAMS6617, is low risk, but each one should link back to the sponsor's own page. Codes change
   every year (GODODGERS26, DODGERS26), so store a `valid_through` date with each one.

---

## 6. Backend implications

- **New metrics needed:**
  - NHL `home_goals`.
  - NFL `interceptions`.
  - MLB `away_saves` and `shutout`.
  - NBA `q4_opp_consecutive_ft_misses` (needs play-by-play).
  - MLS: a "first goal of match, in the first half" variant (needs play-by-play).
- **New redemption windows:** `next_business_day` (Ono), `unlock_after_days: 2, expires_days: 7`
  (Carl's Jr.), `valid_days: 3` (NORMS), and a start time (Del Taco from 10:00).
- **Schema** (already proposed): `in_venue_only`, `verified_at`, plus `source_url` and
  `valid_through` for each event.
- **Re-verification checkpoints:**
  - Before 2027 MLB Opening Day (about late March 2027): Panda, Jack in the Box, McDonald's, ampm,
    Habit, Uber Eats, Angels McDonald's, Del Taco.
  - Before the 2027 MLS opener: Ono, LAFC Chick-fil-A, NORMS.
  - Monthly for in-season NHL, NBA and NFL deals.

## 7. Couldn't confirm: needs a human with a browser or app

- pandaexpress.com/dodgerswin (blocked our fetcher with HTTP 403).
- The Jack in the Box Dodgers strikeout trigger, and the McDonald's Dodgers scope (home only or all
  games).
- Uber Eats reward and code.
- ampm for 2026 or 2027.
- Any 2026-27 El Pollo Loco deal for the Kings (check the Loco Rewards app).
- The Clippers claim window (check the Chick-fil-A app after a trigger).
- Lakers, Chargers, Kings McDonald's and LAFC DoorDash leads (check the sponsor apps and team social
  accounts).

---

## 8. Second pass (15 Nov 2026, evening): in-season teams without a deal

Leads came from lahomewin.com and mealsteals.co (including their per-team pages), which the owner
pointed at. Every status below rests on a team, school or sponsor page or post, never on those
sites. Migration `038_in_venue_deals.sql` seeds the result.

| Team     | Deal                                                                        | Status                                                                                                                                                                                     | In 038              |
| -------- | --------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------- |
| Chargers | Firehouse Subs, free medium sub on a 40+ yd Dicker field goal               | VERIFIED: firehousesubs.com/offer-terms names the 2026-27 regular season (read from the indexed text; the page is a JS app, so a person should open it once)                               | Active              |
| Clippers | Chick-fil-A "Fowl Shot", visitors miss 2 FTs in a row in Q4                 | VERIFIED (undated): cfasocal.com/promotion-details/fowl-shot. At the game; claim by 11:59 PM that night                                                                                    | Active, at the game |
| Ducks    | Chick-fil-A, 5+ goals at home                                               | VERIFIED (section 2). At the game; claim that night                                                                                                                                        | Active, at the game |
| UCLA     | Fat Sal's "Lucky 21", free fries when football scores 21+                   | VERIFIED (undated): uclabruins.com/jmi-promotions. Ticket is the proof, valid the next business day                                                                                        | Active, at the game |
| Lakers   | Jack in the Box, 2 free tacos after any win, LETSGOLAKERS                   | LIKELY+: Jack in the Box on X (12 Nov 2024); mealsteals.co and lahomewin.com both log it firing after 2026-27 games (22 Oct 2026; a 131-123 win, so no 111-point cap)                      | Active              |
| Kings    | McDonald's "McFlurry Minute", goal in the last minute of P2                 | LIKELY+: nhl.com/kings/fans/redeem (indexed text: "all games this season, home and away"); Kings on X 27 Mar 2026; lahomewin.com tracks it for 2026-27. Metric `p2_last_minute_goals`      | Active              |
| UCLA     | Fat Sal's "Missed Fry Throws", visitors miss 2 FTs in a row in the 2nd half | Confirmed for 2025-26; Fat Sal's runs UCLA promos in 2026. Metric `basketball_home_h2_opponent_ft_miss_streak`                                                                             | Active, at the game |
| Rams     | El Torito "Rampage Nachos", 4th-down stop                                   | DISCONTINUED: eltorito.com/rampagenachos returns 404; therams.com lists only Carl's Jr.                                                                                                    | Not seeded          |
| USC      | none                                                                        | Third pass too: usctrojans.com promotions (giveaways and theme nights only), sponsors (Fatburger, Rex, Chipotle NIL, DoorDash), USC Village, Daily Trojan. No result-triggered deal exists | n/a                 |

Chick-fil-A's arena deals use the `game_night` window: they never alert (the claim closes before the
morning check) and appear in the app's TODAY tab on game days with an AT THE GAME tag.

The owner asked for a deal on every in-season team. Lakers, Kings and UCLA basketball went active on
LIKELY+ evidence: the business's or team's own post from last season, plus two independent trackers
logging the deal firing this season. Recheck each in the sponsor's app when possible. USC has no
deal anywhere; it stays empty rather than carry an invented one.
