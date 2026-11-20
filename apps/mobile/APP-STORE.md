# App Store Listing and Submission Checklist

Listing for version 2.0.1, the relaunch as **Freebies: LA Sports Deals**. The bundle ID stays
`io.bitsugar.freebies`, so the existing App Store record, ratings and installs carry over.

Copy rules for everything below (from [docs/ui.md](../../docs/ui.md#legal-guardrails)):

- Never "Nintendo", "NES", "Famicom" or "Nintendo Entertainment System". Say "8-bit", "retro" or
  "pixel art".
- No team or restaurant names in the app name, subtitle or icon. Name them in the description only,
  as facts about which games and deals the app tracks.
- Only list deals that are active and verified in the app at submission time.

## How to use this file

The sections follow App Store Connect page by page, in the order the fields appear. Go top to
bottom. Copy only what's inside a gray box into the field with the same name; the notes under a box
are for us and are never pasted.

1. [Version page](#1-version-page-ios-app--201): screenshots, text, build, review notes
2. [App Information](#2-app-information): name, subtitle, category, content rights, age rating
3. [App Privacy](#3-app-privacy): privacy policy URL and data answers
4. [Pricing and Availability](#4-pricing-and-availability)

Name and subtitle can only change while a version is editable ("Prepare for Submission"). Create the
version first with **+ Version**, using the exact version in `app.json` (**2.0.1**, digits only, no "v"), and those
fields unlock. Everything except promotional text needs Apple's review, and it all goes live
together when 2.0.1 is approved.

## 1. Version page (iOS App → 2.0.1)

Open it from the left sidebar under **iOS App**. The fields below are in page order.

### Previews and Screenshots

Fill **both** iPhone slots, from **View All Sizes in Media Manager**: the six PNGs in
`assets/screenshots/6.9-inch/` go in **6.9" Display** and the six in `assets/screenshots/6.3-inch/`
go in **6.3" Display** (iPhone with Dynamic Island, medium). A slot left over from an older version
is not replaced by the 6.9" set: 2.0.0 shipped with the 1.x pictures still in 6.3", and every
6.1"/6.3" phone showed them. Screenshots lock once a version is approved.

### What's New in This Version

```text
Freebies is now all LA.

NEW
• Every LA team: Dodgers, Angels, Lakers, Clippers, Sparks, Rams, Chargers, Kings, Ducks, LAFC, Galaxy, Angel City FC, USC and UCLA
• A new 8-bit look, with pixel food for every deal
• Four tabs: Won for the deals you've unlocked, Today for deals anyone can grab now and tonight's games, Teams to pick who you follow, and You for your trophy case
• The Win Board on your lock screen and Dynamic Island: every win's food and a live countdown
• A lock-screen widget
• At-the-game deals, for fans in the stands
• Pushes now say what happened and where to claim

CHANGED
• Deals outside Los Angeles are retired
• The app is dark only

Thanks for playing.
```

### Promotional Text (170 characters max, no review needed)

```text
Pick your LA teams. When they hit the mark, a free or cheap meal unlocks and we ping you with where to claim it. A Lakers win means 2 free tacos.
```

145 characters. This is the only listing field you can change without sending the app back to
review, so swap the last sentence for whatever deal is hot that week.

Before naming a deal here, check two things: it is active in the app, and it is still running on the
team's or restaurant's own page. The Ducks deal was verified on 2026-11-15
(`034_deal_verification.sql`) and the NHL season is on, so it can trigger now. The Lakers deal above
was active in the app on 2026-11-17, with the NBA season on. Don't feature a deal whose season is
over (for example a Dodgers deal in winter); it is active but can't unlock.

### Description (4000 characters max)

```text
When LA wins, you win.

Freebies watches LA's teams for you. When something happens on the field that unlocks a food deal, you get a push with the reward, where to claim it and when it ends.

Some of the deals Freebies tracks:
• Lakers win: 2 free tacos at Jack in the Box
• Ducks win at home: $2 McGriddles at McDonald's the next morning
• Rams intercept a pass: a free Famous Star at Carl's Jr.
• Kings score in the last minute of the 2nd period: a free McFlurry at McDonald's
• Chargers make a 40+ yard field goal: a free medium sub at Firehouse Subs
• At the game: Ducks score 5 at home and fans in the arena get a free Chick-fil-A sandwich

Deals change every season. The app only shows deals we've checked against the restaurant's or team's own page.

HOW IT WORKS
1. Pick your teams on the team select screen.
2. Follow all of a team's deals, or just the ones you want.
3. Get a push the moment a deal unlocks, with the real stat ("9 strikeouts").
4. Open the deal, follow the claim steps, and grab your food before the countdown runs out.

WHAT YOU GET
• Every LA team in one place: Dodgers, Angels, Lakers, Clippers, Sparks, Rams, Chargers, Kings, Ducks, LAFC, Galaxy, Angel City FC, USC and UCLA
• Pushes only for the deals you follow, plus one reminder before a deal ends
• Today: deals anyone can grab right now, and tonight's LA games with a deal on the line
• The Win Board on your lock screen and Dynamic Island: every win's food with a live countdown
• A trophy case of what you've unlocked and claimed
• A countdown bar on every unlocked deal so you know how long you have
• Pixel art all the way through: a retro 8-bit look with original food sprites
• No account, no email, no location. The app uses an anonymous device ID

Not every team has a deal running at any given time. Teams without one show up so you're ready when a sponsor adds one.

Freebies: LA Sports Deals is not affiliated with, endorsed by or sponsored by any team, league, school or restaurant. Team and restaurant names are used only to describe the deals they offer. Deals are set by each restaurant and can change or end without notice. Check the restaurant's terms before you order.
```

Before each submission, list only deals that are active in the app and in season (the API's
`GET /api/v1/events` lists the active ones). As of 2026-11-17 the six bullets above are active, with
the NBA, NFL and NHL seasons on. Also active but left out: the LAFC lunch (the MLS season may be
over by review), UCLA's two Fat Sal's deals and the Clippers' Chick-fil-A deal (at the game). The
Dodgers and Angels deals return in spring; add their bullets back then. Swap the Chargers and Rams
bullets out after the NFL season.

The disclaimer wording (last paragraph) was approved by the owner on 2026-11-15. Use the same text
on the in-app About screen.

### Keywords (100 characters max, comma-separated, no spaces after commas)

```text
free food,alerts,reward,game day,coupon,promo,taco,burger,breakfast,los angeles,socal,retro,8-bit
```

97 characters. Words already in the name and subtitle (la, freebies, sports, deals, wins) are left
out because Apple indexes those fields anyway.

**Decision (2026-11-15): no team names in keywords.** Team names are trademarks and guideline 2.3.7
bars trademarked terms in keywords. The description names the teams factually, which is indexed.

The 1.x keywords
(`sports,deals,free food,promotions,MLB,NBA,NFL,NHL,baseball,basketball,football,hockey,alerts`) are
replaced on purpose: league abbreviations are trademarks too, and "sports" and "deals" are now in
the name.

### Support URL

```text
https://github.com/retr0h/freebies/issues
```

### Marketing URL (optional)

```text
https://github.com/retr0h/freebies
```

### Version

```text
2.0.1
```

Must match `version` in `app.json` exactly, or the build won't attach.

### Copyright

Keep the holder that's already there from 1.x and set the year to 2026, for example `2026 <holder>`.
The holder isn't recorded in this repo.

### Build

Pick the 2.0.1 build after `eas submit` uploads it and Apple finishes processing (usually 10 to 30
minutes; you get an email). The build number is set by EAS (`appVersionSource: remote`).

### App Review Information

**Sign-in required:** leave **unchecked**. The app has no login; it creates an anonymous account
from a random device ID, so there is no demo account to give.

**Contact Information:** the first name, last name, phone and email Apple's reviewer should use.
Keep what's there from 1.x. It's personal, so it isn't stored in this repo.

**Notes:** paste the block below, replacing the 1.x notes (bell icon, "My Deals" tab). The help text
about Chinese permits (news, religious, book and magazine apps) doesn't apply to this app, so leave
permit details out.

```text
Freebies: LA Sports Deals tells users when an LA sports result unlocks a restaurant's promotional deal (for example, "$2 McGriddles the morning after a Ducks home win").

The deals are run by the restaurants. The app is not affiliated with any team, league, school or restaurant, and names them only to describe the deals. There are no team or restaurant logos in the app.

No login is required. The app creates an anonymous account from a random device ID.

To test:
1. Open the app and tap through the title screen.
2. Open the Teams tab, pick a sport, and follow a few teams (the Lakers and Ducks have deals running now).
3. Open the Won tab to see the deals your teams have unlocked (the Lakers deal unlocks after any Lakers win).
4. Open the Today tab to see deals anyone can claim now and tonight's games with a deal on the line.
5. Tap any deal to see its trigger and claim steps.

While a followed team has a live win, the Win Board Live Activity shows it on the lock screen and in the Dynamic Island. Push notifications need a physical device.
```

**Attachment:** none needed.

### Version Release

Choose **Manually release this version**. After approval, you press Release when you're ready, so
the new name, listing and app don't go live at a surprise time.

## 2. App Information

Left sidebar → **General → App Information**. Save at the top right after editing.

### Name (30 characters max)

```text
Freebies: LA Sports Deals
```

25 characters. "Freebies" is the brand and the home-screen name in `app.json`; "LA Sports Deals"
says what it is and where. The pixel-font title screen and top bar show "Freebies" alone (docs/ui.md
caps Press Start 2P at 3 words). Apple checks that the name isn't taken when you save.

### Subtitle (30 characters max)

```text
When LA wins, you win
```

21 characters. It is the tagline, and it doesn't say "sports", so it still fits when non-sports LA
triggers arrive.

### Category

- **Primary:** Food & Drink
- **Secondary:** Sports

**Decision (2026-11-15).** The 1.x listing used Sports first. Food & Drink matches "you win Y"
better and stays right when non-sports triggers arrive.

### Content Rights

Choose **No, it does not contain, show, or access third-party content**. Team and restaurant names
appear only as factual descriptions. The app has no team or restaurant logos, and all art is
original.

### Age Rating

**4+**. Answer **None** to every content question in the questionnaire.

## 3. App Privacy

Left sidebar → **General → App Privacy**.

### Privacy Policy URL

```text
https://github.com/retr0h/freebies/blob/main/apps/mobile/PRIVACY-POLICY.md
```

The policy must be on `main` before you submit, or this link shows the 1.x text.

### Data collection

These answers follow [PRIVACY-POLICY.md](PRIVACY-POLICY.md). If 1.x already has answers, check they
still match; nothing new is collected in 2.0.

- **Do you collect data?** Yes.
- **Identifiers → Device ID:** used for App Functionality. Linked to the user: yes (it keys the
  anonymous account). Used for tracking: no.
- **Usage Data → Product Interaction** (which deals you follow, dismiss or claim): used for App
  Functionality. Linked to the user: yes. Used for tracking: no.
- Nothing else: no name, email, location, contacts, purchases or analytics SDK.

## 4. Pricing and Availability

- **Price:** Free.
- **Availability:** every deal is in Los Angeles, so the United States is the only place it's
  useful. Keep the 1.x availability unless you decide to narrow it to the US.

## Screenshots

Capture on a 6.9" iPhone (1320×2868). App Store Connect scales them for smaller devices. Use real,
verified LA deals on screen. No team logos, and no "NES" or "Nintendo" anywhere, including captions.
See [`assets/screenshots/README.md`](assets/screenshots/README.md) for capture steps.

Captions are set in Press Start 2P on `bg.night`, at most 6 words, sentence case. The shot list
changed on 2026-11-17 for what 2.0 shipped: TODAY and the Win Board replace the trophy case and the
push.

| #   | Filename          | Screen                                                          | Caption                       |
| --- | ----------------- | --------------------------------------------------------------- | ----------------------------- |
| 1   | `01-unlocked.png` | Victory cut-in: "5 WINS!" with the five meals on the shelf      | When LA wins, you win         |
| 2   | `02-won.png`      | Won tab: unlocked deals with sprites and countdown bars         | Your free food, counting down |
| 3   | `03-today.png`    | Today tab: LIVE NOW, deals anyone can claim                     | Free food anyone can grab     |
| 4   | `04-teams.png`    | Football's teams, Rams and Chargers followed                    | Pick your LA teams            |
| 5   | `05-claim.png`    | Lakers tacos deal: countdown, claim steps with the code, ▶ menu | Claim it before time runs out |
| 6   | `06-winboard.png` | Lock screen with the Win Board: 5 wins, three clocks, +2 MORE   | Your wins on the lock screen  |

Captured 2026-11-17 from a development build against the local API with real, active deals (the dev
tools button hidden), then framed at 1320×2868 with the caption above: the caption in Press Start 2P
gold with a red shadow, the LA ray colors as a stripe, the screenshot at 83% below. The first five
have a 9:41 status bar; the lock screen shows the real date.

The first two carry the pitch, since most people only see those.

## Build checklist

### Before submitting

- [ ] Every deal named in the description and screenshots is active and verified in the app
- [ ] Privacy policy updated for the Freebies: LA Sports Deals name and merged to `main`
- [ ] App icon is the Trophy Burger over LA rays, 1024×1024, no text, no team marks
- [ ] Display name in `app.json` is "Freebies"
- [x] Six screenshots captured from the 8-bit build (2026-11-17)
- [ ] In-app About screen shows the not-affiliated disclaimer and the Press Start 2P OFL license
- [ ] Version 2.0.1 in `app.json`, build number bumped

### EAS build commands

```bash
# Build and upload to TestFlight (the same build goes to the App Store)
eas build --platform ios --profile production --auto-submit

# Upload an existing build
eas submit --platform ios --profile production
```

## Apple Developer account requirements

1. **Apple Developer Program** ($99/year): https://developer.apple.com/programs/enroll/
2. **App Store Connect access:** the app record already exists under `io.bitsugar.freebies`.
3. **Certificates and provisioning:** EAS handles the distribution certificate and the App Store
   provisioning profile.

## TestFlight first

Only `production` builds can go to TestFlight; `preview` is internal distribution.

1. Build and submit in one step: `eas build -p ios --profile production --auto-submit`
2. In App Store Connect, add internal testers
3. Test on a device, including a real push and the Win Board, before submitting for review
