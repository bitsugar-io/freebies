# App Store Listing and Submission Checklist

Listing for version 2.0.0, the relaunch as **Freebies: LA Sports Deals**. The bundle ID stays
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

1. [Version page](#1-version-page-ios-app--200): screenshots, text, build, review notes
2. [App Information](#2-app-information): name, subtitle, category, content rights, age rating
3. [App Privacy](#3-app-privacy): privacy policy URL and data answers
4. [Pricing and Availability](#4-pricing-and-availability)

Name and subtitle can only change while a version is editable ("Prepare for Submission"). Create the
version first with **+ Version**, using the exact version in `app.json` (**2.0.0**), and those
fields unlock. Everything except promotional text needs Apple's review, and it all goes live
together when 2.0.0 is approved.

## 1. Version page (iOS App → 2.0.0)

Open it from the left sidebar under **iOS App**. The fields below are in page order.

### Previews and Screenshots

Upload the six PNGs listed in [Screenshots](#screenshots) to the **6.9" Display** slot. App Store
Connect scales them for smaller iPhones.

### What's New in This Version

```text
Freebies is now all LA.

NEW
• Every LA team: Dodgers, Angels, Lakers, Clippers, Sparks, Rams, Chargers, Kings, Ducks, LAFC, Galaxy, Angel City FC, USC and UCLA
• A new 8-bit look, with pixel food for every deal
• Three tabs: Won for deals you've unlocked, Teams to pick who you follow, and You for your trophy case and settings
• Follow a whole team in one tap
• Pushes now say what happened and where to claim

CHANGED
• Deals outside Los Angeles are retired
• The app is dark only

Thanks for playing.
```

### Promotional Text (170 characters max, no review needed)

```text
Pick your LA teams. When they hit the mark, a free or cheap meal unlocks and we ping you with where to claim it. Ducks home wins mean $2 McGriddles.
```

148 characters. This is the only listing field you can change without sending the app back to
review, so swap the last sentence for whatever deal is hot that week.

Before naming a deal here, check two things: it is active in the app, and it is still running on the
team's or restaurant's own page. The Ducks deal above was verified on 2026-11-15
(`034_deal_verification.sql`) and the NHL season is on, so it can trigger now. Don't feature a deal
whose season is over (for example a Dodgers deal in winter); it is active but can't unlock.

### Description (4000 characters max)

```text
When LA wins, you win.

Freebies watches LA's teams for you. When something happens on the field that unlocks a food deal, you get a push with the reward, where to claim it and when it ends.

Some of the deals Freebies tracks:
• Ducks win at home: $2 McGriddles at McDonald's the next morning
• Rams intercept a pass: a free Famous Star with Cheese at Carl's Jr.

Deals change every season. The app only shows deals we've checked against the restaurant's or team's own page.

HOW IT WORKS
1. Pick your teams on the team select screen.
2. Follow all of a team's deals, or just the ones you want.
3. Get a push the moment a deal unlocks, with the real stat ("9 strikeouts").
4. Open the deal, follow the claim steps, and grab your food before the countdown runs out.

WHAT YOU GET
• Every LA team in one place: Dodgers, Angels, Lakers, Clippers, Sparks, Rams, Chargers, Kings, Ducks, LAFC, Galaxy, Angel City FC, USC and UCLA
• Pushes only for the deals you follow, plus one reminder before a deal ends
• An inventory of what you've unlocked, what you're watching and what you've claimed
• A countdown bar on every unlocked deal so you know how long you have
• Pixel art all the way through: a retro 8-bit look with original food sprites
• No account, no email, no location. The app uses an anonymous device ID

Not every team has a deal running at any given time. Teams without one show up so you're ready when a sponsor adds one.

Freebies: LA Sports Deals is not affiliated with, endorsed by or sponsored by any team, league, school or restaurant. Team and restaurant names are used only to describe the deals they offer. Deals are set by each restaurant and can change or end without notice. Check the restaurant's terms before you order.
```

Before each submission, list only deals that are active in the app (`034_deal_verification.sql` has
the current list). As of 2026-11-15 that is the Ducks and Rams deals. The Dodgers, Angels and LAFC
deals return when their seasons start; add their bullets back then. The Kings and El Pollo Loco deal
is left out until it is verified for 2026-27.

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
2.0.0
```

Must match `version` in `app.json` exactly, or the build won't attach.

### Copyright

Keep the holder that's already there from 1.x and set the year to 2026, for example `2026 <holder>`.
The holder isn't recorded in this repo.

### Build

Pick the 2.0.0 build after `eas submit` uploads it and Apple finishes processing (usually 10 to 30
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
2. Open the Teams tab, pick a sport, and follow a few teams (the Ducks have a deal running now).
3. Open the Won tab to see unlocked deals and what's on the line in upcoming games.
4. Tap any deal to see its trigger and claim steps.

Push notifications need a physical device.
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

Captions are set in Press Start 2P on `bg.night`, at most 6 words, sentence case.

| #   | Filename              | Screen                                              | Caption                       |
| --- | --------------------- | --------------------------------------------------- | ----------------------------- |
| 1   | `01-unlocked.png`     | Deal unlocked: "DUCKS WIN!" cut-in into the sprite  | When LA wins, you win         |
| 2   | `02-won.png`          | Won tab: what's on the line in a live game, HUD bar | See what's on the line        |
| 3   | `03-teams.png`        | Team select grid with 4 teams picked                | Pick your LA teams            |
| 4   | `04-trophies.png`     | You tab trophy case with sprites and claimed deals  | Every deal you've unlocked    |
| 5   | `05-claim.png`        | Deal detail: dialog box, claim steps, ▶ menu        | Claim it before time runs out |
| 6   | `06-notification.png` | Lock screen push "Ducks won. $2 McGriddles..."      | Know the moment it unlocks    |

The first two carry the pitch, since most people only see those.

## Build checklist

### Before submitting

- [ ] Every deal named in the description and screenshots is active and verified in the app
- [ ] Privacy policy updated for the Freebies: LA Sports Deals name and merged to `main`
- [ ] App icon is the 8-bit sunset palm, 1024×1024, no text, no team marks
- [ ] Display name in `app.json` is "Freebies"
- [ ] Six screenshots captured from the 8-bit build
- [ ] In-app About screen shows the not-affiliated disclaimer and the Press Start 2P OFL license
- [ ] Version 2.0.0 in `app.json`, build number bumped

### EAS build commands

```bash
# Build for TestFlight (internal testing)
eas build --platform ios --profile preview

# Build for App Store submission
eas build --platform ios --profile production

# Submit to App Store
eas submit --platform ios --profile production
```

## Apple Developer account requirements

1. **Apple Developer Program** ($99/year): https://developer.apple.com/programs/enroll/
2. **App Store Connect access:** the app record already exists under `io.bitsugar.freebies`.
3. **Certificates and provisioning:** EAS handles the distribution certificate and the App Store
   provisioning profile.

## TestFlight first

1. Build with the preview profile: `eas build --platform ios --profile preview`
2. Submit to TestFlight: `eas submit --platform ios`
3. In App Store Connect, add internal testers
4. Test on a device, including a real push, before submitting for review
