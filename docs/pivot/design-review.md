# Freebies: Design Review and 8-bit Redesign Proposal

Method: `mobile-design-review` (mobile-taste-skill 0.2.0), report-only. Rubric, tell catalog and
redesign protocol from `skills/_shared/references/`. Section numbers like 6.4 refer to
`skills/mobile-taste/SKILL.md`. The design commitments doc used for drift checks is
`docs/pivot/la-jumbotron-vs-8bit.html` (8-bit column). There is no `MOBILE-DESIGN.md` and no earlier
`design-baseline.json`, so this is a first run, not a regression run. The baseline JSON is inlined
in section 3 instead of written to disk, as requested.

**Pass run:** static only. No simulator, emulator or Expo web session was used.

**Code state:** line numbers refer to commit `759d3b7` (HEAD of `feat/la-redesign`). Another agent
was editing `apps/mobile` during this review, adding `react-native-svg`, a `Sprite` type on `Event`
(`src/api/client.ts`) and a `PixelSprite` component used in `ActiveDealCard`, `DealModal` and
`EventCard`. Line numbers in those three files have since shifted by about 5 to 10 lines. Where it
matters, the findings take that work into account.

---

## Contents

1. Inventory and user flow map
2. Findings
3. Scores and baseline
4. What works and must be preserved
5. 8-bit redesign proposal (App Read, dials, tokens, components, Nav Read, every screen and state)
6. Server-driven blocks: keep, rename, replace
7. Implementation phases
8. Unverified list

---

## 1. Inventory and user flow map

### 1.1 Stack sniff

| Item                      | Finding                                                                                                                             |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| Expo SDK                  | 58 (`expo ~58.0.21`), RN 0.88.1, React 19.3                                                                                         |
| Router                    | React Navigation 7 bottom tabs (`App.tsx:16`). No Expo Router, no native stack, no `linking` config                                 |
| Styling                   | `StyleSheet.create` plus a hand-rolled theme context (`src/hooks/useTheme.tsx`)                                                     |
| Icons                     | Emoji only (tab bar, empty states, badges). No icon family                                                                          |
| Fonts                     | System only. No `expo-font`, no type tokens. 88 `fontSize` literals                                                                 |
| Animation                 | RN `Animated` + `PanResponder` in one component (`SwipeButton`). No Reanimated or Gesture Handler                                   |
| Haptics                   | None (`expo-haptics` not installed)                                                                                                 |
| Network state             | None (no NetInfo)                                                                                                                   |
| Splash                    | `expo-splash-screen` plugin, but `preventAutoHideAsync` is never called                                                             |
| Theme                     | Light, dark and system modes, persisted in AsyncStorage (`useTheme.tsx:78-133`)                                                     |
| Server-driven UI          | `GET /api/v1/config` returns `features` (flags) and `screens` (ordered blocks). Defaults in `src/context/AppConfigContext.tsx:5-31` |
| In progress (other agent) | `react-native-svg 15.16.0`, `Sprite { id, name, width, height, pixels[], palette{} }`, `PixelSprite`                                |

### 1.2 Screens

| Screen                             | File                             | What it does                                                                                                                                  | Data                                                   |
| ---------------------------------- | -------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| Boot gate                          | `App.tsx:31-44`                  | Full-screen `ActivityIndicator` until user and deals load                                                                                     | `userLoading`, `dealsLoading`                          |
| Maintenance                        | `App.tsx:46-59`                  | Blocks the whole app when `features.maintenance_mode` is true. Wrench emoji, "We'll be right back"                                            | `config.features`                                      |
| Deals tab ("Active Deals")         | `src/screens/DealsScreen.tsx`    | Renders `screens.deals` blocks. Owns the `DealModal` that push taps open, registering its setters in context (`:26-28`)                       | `undismissedDeals`, `dismissDeal`, subscriptions       |
| Discover tab ("Discover Events")   | `src/screens/DiscoverScreen.tsx` | Renders `screens.discover` blocks (league filter and team-grouped events). Owns a second `DealModal` instance                                 | `events`, `leagues`, subscriptions, `undismissedDeals` |
| Profile tab ("Profile & Settings") | `src/screens/ProfileScreen.tsx`  | Renders `screens.profile` blocks (stats, subscriptions, settings)                                                                             | user, stats, subscriptions, permission status          |
| Deal detail                        | `src/components/DealModal.tsx`   | `Modal` with `pageSheet`. Two modes: an un-triggered offer (subscribe toggle) and a triggered deal ("Got It!" plus "Slide to Stop Reminding") | `Event`, `ActiveDeal`, `screens.deal_detail` blocks    |

Initial tab is chosen from data: Deals when there are unclaimed deals, otherwise Discover
(`App.tsx:65`). The Deals tab shows a badge with the unclaimed count (`App.tsx:94`).

### 1.3 Blocks (server-driven, `src/components/blocks/`)

| Type                | Component                   | Used on               | What it renders                                                                                             | Data                                                |
| ------------------- | --------------------------- | --------------------- | ----------------------------------------------------------------------------------------------------------- | --------------------------------------------------- |
| `banner`            | `BannerBlock.tsx`           | any                   | Text strip, optional ✕ (session-only dismiss)                                                               | `config.text/backgroundColor/textColor/dismissible` |
| `active_deals`      | `ActiveDealsBlock.tsx`      | deals                 | Non-scrolling FlatList of `ActiveDealCard`, emoji empty state                                               | `undismissedDeals`, `openDealModal`                 |
| `league_filter`     | `LeagueFilterBlock.tsx`     | discover              | Horizontal pill row: All plus leagues enabled by flag                                                       | `leagues`, `features.enable_*`                      |
| `event_list`        | `EventListBlock.tsx`        | discover              | Team cards: badge, name, "N offers available • M following", Follow/Following, expandable offer rows with ✓ | `events`, `isSubscribed`, `toggleSubscription`      |
| `promo_card`        | `PromoCardBlock.tsx`        | deal_detail (and any) | Tappable card that opens a URL. Colors come from the server                                                 | `config.title/subtitle/url/colors`                  |
| `user_stats`        | `UserStatsBlock.tsx`        | profile               | Three equal stat tiles: Claimed / Active / Subscribed                                                       | `GET /users/{id}/stats`, context counts             |
| `subscription_list` | `SubscriptionListBlock.tsx` | profile               | Flat list of subscribed offers with Remove. Renders nothing when empty                                      | `events`, `isSubscribed`                            |
| `settings`          | `SettingsBlock.tsx`         | profile               | Appearance (Light/Dark/System), push status row, User ID, dev push token, version, footer                   | permissions, theme, user                            |
| (renderer)          | `BlockRenderer.tsx`         | all                   | `ScrollView` + `RefreshControl`, shows `ErrorCard` for user/events errors, maps blocks                      | `config.screens[screenId]`                          |

### 1.4 Components (`src/components/`)

| Component            | What it does                                                                                                                                                                | Status                                                                  |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| `ActiveDealCard.tsx` | Unclaimed-deal card: expiry pill, league•team pill, partner pill, offer name (plus emoji or, in the new work, a sprite), team name, green trigger box, "Redeem" pill button | Live                                                                    |
| `DealModal.tsx`      | Detail sheet (see 1.2)                                                                                                                                                      | Live, instantiated twice                                                |
| `SwipeButton.tsx`    | Slide-to-confirm track for "Stop Reminding" (`PanResponder`, native-driver `Animated`)                                                                                      | Live                                                                    |
| `SponsoredCard.tsx`  | Affiliate "Team Gear" card. Hidden because `show_affiliate_links` is 0 (migration 010)                                                                                      | Live, flagged off                                                       |
| `ErrorCard.tsx`      | "Something went wrong", expandable raw message, Try Again                                                                                                                   | Live                                                                    |
| `EventCard.tsx`      | Old per-event card with a `Switch`                                                                                                                                          | **Dead code**, imported nowhere (the sprite work added to it is wasted) |
| `src/types/index.ts` | Old `Offer/Rule/FreebieEvent` types                                                                                                                                         | **Dead code**                                                           |

### 1.5 Context and hooks

| File                            | Role                                                                                                                                                           |
| ------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `context/AppConfigContext.tsx`  | Fetches `/config` on mount and on foreground. Falls back to `DEFAULT_CONFIG`                                                                                   |
| `context/AppDataContext.tsx`    | Composes user, leagues, events, subscriptions, deals and push. Holds a ref to the DealsScreen modal setters for push taps (`:96-124`). Refreshes on foreground |
| `hooks/useUser.tsx`             | Anonymous device user. Token in SecureStore, user in AsyncStorage. Re-creates the user if the stored one fails                                                 |
| `hooks/useEvents.ts`            | `GET /events`. Sets `isLoading` true on every refetch                                                                                                          |
| `hooks/useSubscriptions.ts`     | Set of subscribed event IDs, optimistic toggle with revert. `error` is never exposed upward                                                                    |
| `hooks/useActiveDeals.ts`       | `GET /users/{id}/active-deals`, dismiss/undo, `formatExpiresAt`, `isExpiringSoon(6h)`. `error` is never exposed upward                                         |
| `hooks/usePushNotifications.ts` | Requests permission and registers the token **on mount**. Foreground and tap listeners. Dev test-notification helper                                           |
| `hooks/useTheme.tsx`            | Light and dark palettes, mode persistence. Renders `null` until the stored mode loads                                                                          |
| `api/client.ts`                 | Fetch wrapper with no timeout. Bearer token for user-scoped routes                                                                                             |

### 1.6 API shape (from `services/api`)

- `GET /config`: `{ features: {key: bool}, screens: {screen: [{type, key, config}]} }` from the
  `feature_flags` and `screen_blocks` tables (`internal/api/handlers/config.go`, migration
  `009_server_driven_ui.sql`). Sent with `Cache-Control: public, max-age=60`. **Every installed app
  version reads the same config.**
- `GET /events`: an `EventResponse` per team+offer pairing (`id, teamId, teamName, league`
  (lowercase, e.g. `mlb`),
  `teamColor, icon (emoji), partnerName, offerName, offerDescription, triggerCondition, regionName, offerUrl, isActive`).
  The sprite fields are being added.
- `GET /users/{id}/active-deals`:
  `{ id (triggered event), eventId, triggeredAt, expiresAt, event, isDismissed, dismissalType }`.
  `triggered_events.payload` stores game data that is not exposed.
- `POST/DELETE /users/{id}/subscriptions[/eventId]`, `POST/DELETE /users/{id}/dismissals`.
- Push (`internal/worker/service.go:146-147, 207-208`): title `"{emoji} Deal Unlocked!"` or
  `"{emoji} Don't forget your deal!"`, body `"{Partner}: {Offer}"`, data
  `{eventId, triggeredEventId}`.
- There is no games or schedule endpoint. Sources only fetch yesterday's game
  (`internal/sources/source.go:27-33`), so nothing live exists today.

### 1.7 User flow map (as built)

```
FIRST LAUNCH
  splash (#222034) -> blank frame (ThemeProvider returns null) -> full-screen spinner
  -> anonymous user created -> SYSTEM PUSH PERMISSION DIALOG (on mount, no context)
  -> no deals yet -> lands on DISCOVER
SUBSCRIBE
  Discover -> [league pill] -> team card -> [Follow] (subscribes every offer of that team)
           -> or [See offers] -> offer row -> DealModal -> [Subscribe to this Deal]
DEAL TRIGGERS (server, morning after a game)
  push "🍔 Deal Unlocked!" / "Jack in the Box: Free Jumbo Jack"
  -> tap -> openDealModal(eventId) -> works only if DealsScreen has mounted (FINDING-001)
  -> or open app -> boot lands on DEALS (badge count) -> ActiveDealCard -> [Redeem] -> DealModal
CLAIM / DISMISS
  DealModal -> [Got It!] (dismissal got_it, counts as claimed)
            -> [Slide to Stop Reminding] (dismissal stop_reminding, no more reminder pushes)
  -> card leaves the list. dismissedDeals and undoDismissal exist in context but no UI uses them
REMINDER
  push "🍔 Don't forget your deal!" for undismissed deals
SETTINGS
  Profile -> stats triplet -> Your Subscriptions [Remove] -> Appearance -> Push status -> User ID -> Version
```

---

## 2. Findings

Format per `review-rubric.md`. Impact: **high** breaks usability or trust, **medium** is visibly
wrong but recoverable, **polish** is taste-level (listed, not scored).

### Code-reconstructed narration and Trunk Test

These are reconstructed from code, not from rendered screens.

- **Deals (code-reconstructed):** The eye goes to the orange or blue time pill, then to three more
  pills before reaching the offer name, which is the biggest text but sits fourth in the stack. The
  green trigger box and the full-width green "Redeem" pill compete. White on green is hard to read.
- **Discover (code-reconstructed):** A column of white rounded team cards with colored circles. The
  green "Follow" pill is the only color on each card, so the eye scans down the right edge. "▼ See
  offers" at the bottom of every card reads as repeated noise.

| Screen                           | Q1 app?                                                   | Q2 screen?                                                    | Q3 where?       | Q4 back?                                                | Result                              |
| -------------------------------- | --------------------------------------------------------- | ------------------------------------------------------------- | --------------- | ------------------------------------------------------- | ----------------------------------- |
| Deals (tab root)                 | No brand mark anywhere; generic emoji tabs                | "Active Deals" header                                         | Tab bar         | n/a                                                     | PARTIAL                             |
| Discover (tab root)              | Same                                                      | "Discover Events" (internal vocabulary)                       | Tab bar         | n/a                                                     | PARTIAL                             |
| Profile (tab root)               | "Freebies" only in the footer at the bottom of the scroll | "Profile & Settings"                                          | Tab bar         | n/a                                                     | PARTIAL                             |
| DealModal from push (deep entry) | n/a (interior)                                            | "Deal Details" header is generic, but the offer name is large | Sheet over tabs | "Close" top-left, Android back works (`onRequestClose`) | PASS (if it opens; see FINDING-001) |

No screen fails. The PARTIALs are all Q1 (no brand presence), which the 8-bit wordmark and palm
sprite fix.

### High

```
FINDING-001 [high | navigation] Push tap can't open the deal unless the Deals tab has mounted
  where: src/context/AppDataContext.tsx:96-124, src/screens/DealsScreen.tsx:26-28, App.tsx:65,125
  evidence: openDealModal() returns early when modalHandlersRef is null. Only DealsScreen
    registers it, and bottom tabs mount lazily. When the app booted onto Discover (no unclaimed
    deals at launch) and the user never opened Deals, a notification tap does nothing. On cold
    start the response listener is attached in an effect while the boot gate is still showing
    the spinner, and getLastNotificationResponseAsync / useLastNotificationResponse is never
    used, so a tap that launches the app can be lost. If the deal list hasn't refreshed yet, the
    fallback opens the un-triggered "Subscribe" view instead of the claim view. The deal detail
    is component state, not a route: no linking config, no freebie://deal/:id, and the modal is
    duplicated in two screens. This is the product's core loop (push -> claim).
  tell: none (proposed new tell below)
  fix: Make the deal a route (deal/[id], modal presentation) at the root stack. Add linking
    config for freebie://deal/:id. Handle cold start with useLastNotificationResponse. When the
    deal isn't in memory, fetch it (or wait for refetchDeals) before rendering, and show a
    skeleton rather than the subscribe view. Delete the duplicated DealModal instances and
    setModalHandlers.
```

```
FINDING-002 [high | navigation] Notification permission requested on mount, before any value
  where: src/hooks/usePushNotifications.ts:37-52,147-153, src/context/AppDataContext.tsx:161-166
  evidence: registerForPushNotificationsAsync() runs in a mount effect and calls
    requestPermissionsAsync() with no priming. The first thing a new user sees after the spinner
    is the iOS system dialog, before they've followed a team. A "Don't Allow" here is near
    permanent and kills the product's only delivery channel. 6.6.
  fix: Ask after the first value moment (first team followed), on a priming screen that says
    why (deals unlock the morning after and expire fast), then call the system prompt. Move
    token registration to after a grant. Keep the You-tab row as the recovery path.
```

```
FINDING-003 [high | states] A failed deals fetch shows "No Active Deals"
  where: src/hooks/useActiveDeals.ts:30-33,93; src/context/AppDataContext.tsx:87-94;
         src/components/blocks/BlockRenderer.tsx:42-44
  evidence: useActiveDeals sets `error` but AppDataContext never passes it on. BlockRenderer
    only shows ErrorCard for userError || eventsError. When /active-deals fails (offline, 5xx,
    expired token), the Deals tab renders the empty state "No Active Deals", which tells someone
    holding a free McGriddle that they have nothing. There's no offline banner and no cache, so
    a fan at the counter with one bar of signal can't see their claim steps.
  tell: #38
  fix: Expose dealsError. Show an inline error with retry on the Deals screen. Cache the last
    good active-deals response in AsyncStorage and show it under an "Offline, showing deals
    saved at 8:47pm" banner. Add NetInfo for the banner.
```

```
FINDING-004 [high | color] White text on the accent fails contrast in both modes
  where: src/hooks/useTheme.tsx:42,61; white literals at ActiveDealCard.tsx:162,
    DealModal.tsx:178,386, EventListBlock.tsx:91, LeagueFilterBlock.tsx:38,
    SettingsBlock.tsx:53, ErrorCard.tsx:112
  evidence: Light accent #4CAF50 with #FFFFFF text is 2.8:1. Dark accent #03DAC6 with
    #FFFFFF is 1.8:1. Both fail WCAG AA (4.5:1 for 14-16pt text, 3:1 even for large text in the
    dark case). These are the primary CTAs: Redeem, Got It!, Follow, the active league pill,
    the active theme option and Try Again.
  fix: Add an `onAccent` token and use it everywhere '#fff' sits on the accent. In the 8-bit
    palette, night (#222034) on coin (#FBF236) is 13.5:1.
```

```
FINDING-005 [high | interaction] "Stop Reminding" is swipe-only, and no touchable has an accessibility role
  where: src/components/SwipeButton.tsx:33-110, src/components/DealModal.tsx:198-208;
         repo-wide: 0 accessibilityRole/Label/State across 21 TouchableOpacity uses
  evidence: The pan handlers sit on the 48pt thumb only. There's no accessibilityActions or
    onAccessibilityAction, so VoiceOver and TalkBack users can't stop reminders. Selected state
    for league pills, the theme options, Follow/Following and the per-offer ✓ is purely visual
    (no accessibilityState). The ✓ glyph and emoji tab icons get read out literally.
  fix: Keep the slide (it's a signature interaction, see section 4), but give the track
    accessibilityRole="adjustable" or an accessibilityActions [{name:'activate', label:'Stop
    reminders'}] that fires the same handler. Add role, label and state to every custom
    pressable. Mark sprite images accessible={false} when the text label sits next to them.
```

### Medium

```
FINDING-006 [medium | states] "Try Again" can't recover from the error it's showing
  where: src/components/blocks/BlockRenderer.tsx:55; src/context/AppDataContext.tsx:126-128
  evidence: ErrorCard's onRetry is the screen's onRefresh -> refreshAll(), which only refetches
    events and deals. When the card is showing userError (user init failed), refreshUser() is
    never called, so every user-scoped call keeps failing and the button does nothing. The
    message is raw (e.g. "HTTP 502") behind a "Tap to see details" toggle.
  fix: Retry dispatches by failure source (user -> refreshUser, then refreshAll). Write plain
    copy: "Couldn't reach Freebies. Check your connection."
```

```
FINDING-007 [medium | states] Full-screen spinner over the tab bar, and Discover flashes "No Teams Found"
  where: App.tsx:38-44; src/components/blocks/EventListBlock.tsx:133-139; src/hooks/useEvents.ts:11
  evidence: The boot gate is an ActivityIndicator covering everything. Events aren't part of
    the gate, so Discover renders its empty state while /events is in flight. useEvents also
    sets isLoading on every refetch.
  tell: #37
  fix: Gate only on splash-level needs (fonts, stored user). Render the tabs right away with
    skeletons that match the final layout. Keep initial-load and refresh state separate.
```

```
FINDING-008 [medium | states] Subscribe, follow and stats failures are silent
  where: src/hooks/useSubscriptions.ts:59-72 (error never surfaced);
         src/components/blocks/EventListBlock.tsx:62-70; src/components/blocks/UserStatsBlock.tsx:16-18
  evidence: A failed toggle silently reverts the optimistic state, so the user sees Follow flip
    back with no message. Following a team runs N sequential requests with no in-progress
    state. Stats errors go to console only and show "0 Claimed".
  tell: #38
  fix: A toast or snackbar on failed toggle ("Couldn't follow the Kings. Try again."), a busy
    state on the team cell while a batch runs, and a hidden stats tile rather than a fake 0.
```

```
FINDING-009 [medium | navigation] Emoji as the icon system, including the tab bar
  where: App.tsx:18-29,50; ActiveDealsBlock.tsx:31; EventListBlock.tsx:135; ErrorCard.tsx:23;
         SettingsBlock.tsx:35-37,92; LeagueFilterBlock.tsx:18; DealModal.tsx:96,149,227;
         ActiveDealCard.tsx:44; SponsoredCard.tsx:45; server: events.icon, leagues.icon, push titles
  evidence: 25 emoji hits in JSX. Tab icons ignore tabBarActiveTintColor, so the active state is
    only 26pt vs 24pt plus an opacity change. They render differently on iOS and Android and
    read aloud as "wrapped gift" / "magnifying glass".
  tell: #1
  fix: One icon family. In 8-bit that's a sprite set rendered by the same Sprite primitive
    as the food items (tab icons 12x12 at 2x).
```

```
FINDING-010 [medium | typography] No type scale: 88 fontSize literals, no tokens, no brand fonts
  where: every component; e.g. ActiveDealCard.tsx:110-165, DealModal.tsx:260-399, App.tsx:144-153
  evidence: Sizes 10,12,13,14,15,16,17,18,20,24,26,28,32,48 and weights 500/600/700/bold mixed
    per file. The maintenance hint uses fontFamily 'monospace'. The 8-bit doc commits to Press
    Start 2P + Pixelify Sans, and neither is loaded (drift from the design doc).
  fix: Type tokens (section 5.3), loaded with expo-font before the splash hides.
```

```
FINDING-011 [medium | typography] Fixed 56pt-high pills will clip at large text sizes
  where: ActiveDealCard.tsx:155-160, DealModal.tsx:375-381, SwipeButton.tsx:115-121
  evidence: height: 56 with unbounded font scaling and no maxFontSizeMultiplier. "Slide to Stop
    Reminding" in a half-width track already wraps at default size on a 375pt phone. 6.1.
  fix: minHeight instead of height, maxFontSizeMultiplier around 1.3 on button labels, and
    full-width rows for long labels (the PixelMenu pattern).
```

```
FINDING-012 [medium | color] No accent lock; Material baseline palettes
  where: src/hooks/useTheme.tsx:32-68
  evidence: Light accent green #4CAF50, dark accent teal #03DAC6, dark primary purple #BB86FC
    (the Material 2 dark-theme defaults). success == accent in both modes, so the "you're
    subscribed" state and the main CTA are the same color. 6.2.
  fix: Replace with the 8-bit semantic tokens. Coin is the single accent; clear (green) only
    means unlocked or followed.
```

```
FINDING-013 [medium | color] Inline colors outside the theme, invisible dividers in dark mode
  where: UserStatsBlock.tsx:27 (#E91E63), EventListBlock.tsx:9 (#666666),
         SettingsBlock.tsx:137 and DealModal.tsx:404 (rgba(0,0,0,0.1)),
         PromoCardBlock.tsx:9-10 and BannerBlock.tsx:16-17 (server-supplied hex),
         SwipeButton.tsx:102, EventCard.tsx:53, usePushNotifications.ts:114
  evidence: 20 color literals outside useTheme. Black-at-10% dividers vanish on #1E1E1E. Server
    blocks can inject any hex, so /config can break the palette.
  tell: #6
  fix: Tokens only. Server blocks name a palette token ("tone": "coin") instead of passing hex.
```

```
FINDING-014 [medium | surfaces] Profile is a card stack
  where: UserStatsBlock.tsx:44, SubscriptionListBlock.tsx:38, SettingsBlock.tsx:131
  evidence: Five or six rounded surface sections, each with its own 18pt title, on a gray
    background. Settings rows sit inside cards inside a scroll.
  tell: #3
  fix: Grouped menus (PixelMenu sections with Press Start labels), not cards.
```

```
FINDING-015 [medium | surfaces] Stat triplet on Profile
  where: src/components/blocks/UserStatsBlock.tsx:22-39
  evidence: Three equal tiles (Claimed pink, Active green, Subscribed blue). "Active" duplicates
    the Deals badge and "Subscribed" duplicates the list below it.
  tell: #8
  fix: One hero number ("12 deals claimed this season") with team badges collected.
```

```
FINDING-016 [medium | surfaces] Pill overload on the deal card
  where: src/components/ActiveDealCard.tsx:29-83, DealModal.tsx:80-114
  evidence: An expiry pill, a league•team pill and a partner pill, then a tinted trigger box and a
    pill button: five rounded tinted shapes per card. The information is right (section 4
    explains why), but the treatment buries the offer name.
  tell: #10
  fix: Keep the information order. Turn expiry into the DealBar, make team and partner one line
    of plain text, and present the trigger as the reason line ("Ducks won at home").
```

```
FINDING-017 [medium | interaction] Opacity-only press feedback, no haptics anywhere
  where: 21 TouchableOpacity uses; no Pressable, no android_ripple, no expo-haptics
  evidence: Every control fades to 0.7. The claim moment ("Got It!") has no confirmation beyond
    the sheet closing.
  tell: #35
  fix: One branded press (pixel "press-in", section 5.6) and the haptics vocabulary from 6.5:
    selection on team toggle, success notification on Stage Clear.
```

```
FINDING-018 [medium | anatomy] Tap targets under 44pt
  where: BannerBlock.tsx:24,38 (✕ ~20pt), EventListBlock.tsx:153 (Follow ~34pt),
         EventListBlock.tsx:117,162 (See offers ~34pt), SubscriptionListBlock.tsx:46 (Remove ~29pt),
         DealModal.tsx:72,257 (Close: the text height, ~20pt)
  tell: #19
  fix: 44pt minimum height, hitSlop on glyph buttons, full-width menu rows.
```

```
FINDING-019 [medium | anatomy] Profile sections run edge to edge with rounded corners
  where: src/components/blocks/BlockRenderer.tsx:47 (no contentContainerStyle);
         SettingsBlock.tsx:131, UserStatsBlock.tsx:44, SubscriptionListBlock.tsx:38 (no marginHorizontal)
  evidence: Deals and Discover blocks pad themselves (16). Profile blocks don't, so their 12pt-radius
    surfaces touch the screen edges. Padding is per block, not per screen.
  fix: The screen owns horizontal padding (16), set once in BlockRenderer.
```

```
FINDING-020 [medium | anatomy] Bottom insets: magic numbers and an un-inset modal
  where: ActiveDealsBlock.tsx:41, EventListBlock.tsx:145 (paddingBottom: 100); DealModal.tsx:79
  evidence: Lists hard-code 100pt instead of using the inset. The modal ScrollView has no bottom
    inset, and on Android (edge-to-edge is mandatory on SDK 58) its last element can sit under
    the gesture bar.
  tell: #23
  fix: contentContainerStyle paddingBottom = insets.bottom + 16 on every scrollable.
```

```
FINDING-021 [medium | copy] Em-dash in a visible string; exclamation voice
  where: SettingsBlock.tsx:80 ("Disabled in Settings — tap to open"); "Got It!" DealModal.tsx:196;
         "Check back shortly!" App.tsx:55; "✓ You claimed this deal!" DealModal.tsx:217;
         fallback "...when deals drop!" ActiveDealsBlock.tsx:15
  tell: #33
  fix: Copy in section 5.
```

```
FINDING-022 [medium | copy] Verbs that don't match the action; internal vocabulary
  where: ActiveDealCard.tsx:86-91 ("Redeem" only opens the detail sheet); App.tsx:103 ("Discover
         Events": an "event" is the database row, users think in teams and deals);
         DealModal.tsx:181 ("✓ Subscribed - Tap to Unsubscribe" puts state and instruction in one
         label); SettingsBlock.tsx:101-102 (raw User ID shown to users)
  fix: The button names what happens next ("How to claim"). Teams/Deals vocabulary. A toggle
    whose label is the state. The User ID moves behind a "Support code" row.
```

```
FINDING-023 [medium | copy] Push copy has no reason and no deadline
  where: services/api/internal/worker/service.go:146-147,207-208
  evidence: "🍔 Deal Unlocked!" / "Jack in the Box: Free Jumbo Jack". It doesn't say which team
    did what, or when the deal ends, and those are the two facts that make someone act.
  fix: The push copy table in section 5.17.
```

```
FINDING-024 [medium | copy] League flags are never applied to the team list (casing mismatch)
  where: src/components/blocks/EventListBlock.tsx:31-37 vs services/api/internal/api/handlers/handlers.go:98
         and migrations (league stored as 'mlb', 'nhl', ...)
  evidence: leagueFlags is keyed 'MLB','NBA',... but event.league is lowercase, so flagKey is
    always undefined. With enable_nhl=0 the NHL pill is hidden (LeagueFilterBlock lowercases
    correctly) but NHL teams still appear under "All". The content shown doesn't match the
    filter. The LA pivot removes league filtering, but the same bug would hit any per-team flag.
  fix: Normalize case in one helper, or have the server filter /events by flags.
```

```
FINDING-025 [medium | performance] Following a team fires 2N requests in series
  where: src/components/blocks/EventListBlock.tsx:62-70; src/context/AppDataContext.tsx:131-135
  evidence: Each toggleSubscription awaits a POST and then calls refetchDeals(). Following the
    Dodgers (2 deals) makes 4 requests. A future "follow all LA" would make dozens.
  fix: A batch endpoint (PUT /users/{id}/subscriptions with a list of eventIds, or
    POST /users/{id}/teams/{teamId}/follow), with one refetch afterwards.
```

```
FINDING-026 [medium | performance] Splash to first screen: blank frame, then spinner, then a light-mode flash
  where: src/hooks/useTheme.tsx:125-127; App.tsx:38-44; app.json splash backgroundColor #222034
  evidence: preventAutoHideAsync is never called, so the native splash hides right away,
    ThemeProvider renders null while AsyncStorage loads, then a spinner appears on #F5F5F5
    (light mode). The new splash color is #222034, so light-mode users get dark, blank, light.
    8.
  fix: Hold the splash until fonts, theme and the stored user are ready. The first screen's
    background is night (#222034).
```

### Polish (listed, not scored)

```
FINDING-027 [polish | performance] FlatLists with scrollEnabled={false} nested in a ScrollView
  where: ActiveDealsBlock.tsx:18-19, EventListBlock.tsx:127-128, BlockRenderer.tsx:47
  evidence: This turns off virtualization and logs the RN nested-VirtualizedList warning. The
    data is bounded (14 LA teams, a handful of deals), so the tell #13 hard fail does not apply.
  tell: #13 (triaged: bounded)
  fix: Let the screen own one list. Blocks become sections of a single FlatList/SectionList.
```

```
FINDING-028 [polish | surfaces] Radius drift: 8, 12, 14, 16, 20, 24, 28
  tell: #11. Moot in 8-bit (radius 0, stepped corners).
```

```
FINDING-029 [polish | interaction] SwipeButton runs the gesture on the JS thread and captures a stale callback
  where: src/components/SwipeButton.tsx:33 (PanResponder created once in useRef, closing over
         the first onSwipeComplete)
  evidence: It works today only because the button remounts per deal.
  fix: Gesture Handler + Reanimated, or a ref to the latest callback.
```

```
FINDING-030 [polish | copy] Dev-only alert() and a test notification that names a stale task
  where: usePushNotifications.ts:252,262,276 ("task backend:deals" is now "task api:deals")
  tell: #32 (dev-only, triaged)
```

```
FINDING-031 [polish | states] Maintenance screen is a dead end
  where: App.tsx:46-59; AppConfigContext.tsx:65-73
  evidence: Emoji wrench, no status, and it only re-checks when the app comes back to the
    foreground. Unclaimed deals (with their claim steps) are hidden too.
  fix: Re-poll /config every 60s while showing. Keep cached unlocked deals readable.
```

```
FINDING-032 [polish | navigation] Dead code: EventCard.tsx and src/types/index.ts
  evidence: Nothing imports them. The in-progress sprite work was also added to EventCard.
```

```
FINDING-033 [polish | anatomy] Off-grid spacing (2, 3, 6, 10, 14) alongside the 4/8 scale
```

### Grep checklist triage (every hit)

| Check                         | Hits                                                                                                                             | Triage                                                                                        |
| ----------------------------- | -------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| Em/en-dash                    | `App.tsx:46` (comment), `SettingsBlock.tsx:80` (visible)                                                                         | Comment dismissed (not visible). SettingsBlock -> FINDING-021                                 |
| Emoji in JSX                  | 25                                                                                                                               | -> FINDING-009 (the `usePushNotifications.ts` test titles are dev-only, folded in)            |
| Indigo hex                    | 0                                                                                                                                | Clean                                                                                         |
| LinearGradient                | 0                                                                                                                                | Clean                                                                                         |
| Inline hex/rgba outside theme | 20                                                                                                                               | -> FINDING-004 (`#fff` on accent), FINDING-013 (the rest). `EventCard.tsx` hits are dead code |
| ScrollView + .map             | `BlockRenderer.tsx` (blocks, bounded), `LeagueFilterBlock.tsx` (leagues, bounded), `DealModal.tsx` (deal_detail blocks, bounded) | All dismissed as bounded static maps. The nested FlatList issue is FINDING-027                |
| paddingTop 40-69              | 0                                                                                                                                | Clean                                                                                         |
| useNativeDriver: false        | 0                                                                                                                                | Clean (SwipeButton uses the native driver)                                                    |
| Alert.alert / alert()         | 2 (`usePushNotifications.ts:262,276`)                                                                                            | Dev-only test path -> FINDING-030                                                             |
| shadowColor                   | 1 (`EventCard.tsx:132`)                                                                                                          | Has sibling `elevation: 3`. Dead code anyway. Dismissed                                       |
| GestureHandlerRootView        | 0                                                                                                                                | Not needed today (no RNGH). It will be needed if the slide moves to RNGH                      |
| accessibilityRole/Label       | 0                                                                                                                                | -> FINDING-005                                                                                |
| Tell #41 dead setting         | Appearance toggle                                                                                                                | Dismissed: persisted (`useTheme.tsx:101-108`) and consumed (`App.tsx:63`)                     |
| Tell #42 destructive residue  | none                                                                                                                             | N/A: no delete-account or erase flow                                                          |

**Proposed new tell (append to tells.md as #43): THE ORPHANED DEEP LINK.** A notification-tap or
deep-link handler that depends on a specific screen having mounted (handler refs registered in a
screen effect, lazy tabs). Check: grep `addNotificationResponseReceivedListener` and confirm that
the handler navigates to a route, not a ref set by a screen. It maps to FINDING-001.

---

## 3. Scores and baseline

Per category: start at 4.0, minus 1.0 per high, minus 0.5 per medium. Polish isn't scored.

| #   | Category             | Findings scored                   | Value | Letter |
| --- | -------------------- | --------------------------------- | ----- | ------ |
| 1   | Navigation & IA      | 001 H, 002 H, 009 M               | 1.5   | D      |
| 2   | Screen anatomy       | 018 M, 019 M, 020 M               | 2.5   | C+     |
| 3   | Typography           | 010 M, 011 M                      | 3.0   | B      |
| 4   | Color & theming      | 004 H, 012 M, 013 M               | 2.0   | C      |
| 5   | Surfaces & grouping  | 014 M, 015 M, 016 M               | 2.5   | C+     |
| 6   | States               | 003 H, 006 M, 007 M, 008 M        | 1.5   | D      |
| 7   | Interaction & motion | 005 H, 017 M                      | 2.5   | C+     |
| 8   | Forms & input        | none (the app has no text inputs) | 4.0   | A      |
| 9   | Performance signals  | 025 M, 026 M                      | 3.0   | B      |
| 10  | Copy & content       | 021 M, 022 M, 023 M, 024 M        | 2.0   | C      |

**Design Score:** mean = 24.5 / 10 = **2.45 -> C+**. The B cap for open high findings doesn't come
into play because the mean is already below B.

**AI Slop Score** (only findings that map to a numbered tell): navigation 3.5 (#1), anatomy 3.0
(#19, #23), typography 4.0, color 3.5 (#6), surfaces 2.5 (#3, #8, #10), states 2.0 (#38 H, #37,
#38), interaction 3.5 (#35), forms 4.0, performance 4.0 (#13 is polish), copy 3.5 (#33). Mean = 33.5
/ 10 = **3.35 -> B+**.

Reading the two together: the app isn't template slop (no indigo, no gradients, no five-tab FAB).
Its real problems are correctness in the push-to-claim loop, honest states, contrast and
accessibility.

### Baseline (inline, not written to disk)

```json
{
  "tool": "mobile-design-review",
  "version": 1,
  "date": "2026-11-15",
  "commit": "759d3b7",
  "designScore": "C+",
  "slopScore": "B+",
  "categories": {
    "navigation": "D",
    "anatomy": "C+",
    "typography": "B",
    "color": "C",
    "surfaces": "C+",
    "states": "D",
    "interaction": "C+",
    "forms": "A",
    "performance": "B",
    "copy": "C"
  },
  "findings": [
    {
      "id": "FINDING-001",
      "impact": "high",
      "category": "navigation",
      "summary": "Push tap is dropped unless Deals tab mounted; deal detail not a route",
      "status": "open"
    },
    {
      "id": "FINDING-002",
      "impact": "high",
      "category": "navigation",
      "summary": "Push permission requested on mount without priming",
      "status": "open"
    },
    {
      "id": "FINDING-003",
      "impact": "high",
      "category": "states",
      "summary": "Deals fetch error renders empty state; no offline/cache",
      "status": "open"
    },
    {
      "id": "FINDING-004",
      "impact": "high",
      "category": "color",
      "summary": "White on accent 2.8:1 light / 1.8:1 dark",
      "status": "open"
    },
    {
      "id": "FINDING-005",
      "impact": "high",
      "category": "interaction",
      "summary": "Swipe-only Stop Reminding; zero accessibility roles/labels",
      "status": "open"
    },
    {
      "id": "FINDING-006",
      "impact": "medium",
      "category": "states",
      "summary": "Retry does not re-init user",
      "status": "open"
    },
    {
      "id": "FINDING-007",
      "impact": "medium",
      "category": "states",
      "summary": "Full-screen spinner; Discover empty flash during load",
      "status": "open"
    },
    {
      "id": "FINDING-008",
      "impact": "medium",
      "category": "states",
      "summary": "Silent subscribe/follow/stats failures",
      "status": "open"
    },
    {
      "id": "FINDING-009",
      "impact": "medium",
      "category": "navigation",
      "summary": "Emoji icon system incl. tab bar",
      "status": "open"
    },
    {
      "id": "FINDING-010",
      "impact": "medium",
      "category": "typography",
      "summary": "88 fontSize literals, no tokens, design-doc fonts not loaded",
      "status": "open"
    },
    {
      "id": "FINDING-011",
      "impact": "medium",
      "category": "typography",
      "summary": "Fixed-height buttons clip under Dynamic Type",
      "status": "open"
    },
    {
      "id": "FINDING-012",
      "impact": "medium",
      "category": "color",
      "summary": "Accent differs per mode; Material baseline palettes",
      "status": "open"
    },
    {
      "id": "FINDING-013",
      "impact": "medium",
      "category": "color",
      "summary": "20 inline colors; server can inject hex",
      "status": "open"
    },
    {
      "id": "FINDING-014",
      "impact": "medium",
      "category": "surfaces",
      "summary": "Profile card stack",
      "status": "open"
    },
    {
      "id": "FINDING-015",
      "impact": "medium",
      "category": "surfaces",
      "summary": "Stat triplet",
      "status": "open"
    },
    {
      "id": "FINDING-016",
      "impact": "medium",
      "category": "surfaces",
      "summary": "Pill overload on deal card",
      "status": "open"
    },
    {
      "id": "FINDING-017",
      "impact": "medium",
      "category": "interaction",
      "summary": "Opacity-only press, no haptics",
      "status": "open"
    },
    {
      "id": "FINDING-018",
      "impact": "medium",
      "category": "anatomy",
      "summary": "Tap targets under 44pt",
      "status": "open"
    },
    {
      "id": "FINDING-019",
      "impact": "medium",
      "category": "anatomy",
      "summary": "Profile sections edge-to-edge",
      "status": "open"
    },
    {
      "id": "FINDING-020",
      "impact": "medium",
      "category": "anatomy",
      "summary": "Magic bottom padding; modal lacks inset",
      "status": "open"
    },
    {
      "id": "FINDING-021",
      "impact": "medium",
      "category": "copy",
      "summary": "Em-dash and exclamation voice",
      "status": "open"
    },
    {
      "id": "FINDING-022",
      "impact": "medium",
      "category": "copy",
      "summary": "Misleading verbs, internal vocabulary",
      "status": "open"
    },
    {
      "id": "FINDING-023",
      "impact": "medium",
      "category": "copy",
      "summary": "Push copy lacks reason and deadline",
      "status": "open"
    },
    {
      "id": "FINDING-024",
      "impact": "medium",
      "category": "copy",
      "summary": "League flag casing mismatch shows hidden leagues",
      "status": "open"
    },
    {
      "id": "FINDING-025",
      "impact": "medium",
      "category": "performance",
      "summary": "Follow team = 2N serial requests",
      "status": "open"
    },
    {
      "id": "FINDING-026",
      "impact": "medium",
      "category": "performance",
      "summary": "Splash -> blank -> spinner -> light flash",
      "status": "open"
    },
    {
      "id": "FINDING-027",
      "impact": "polish",
      "category": "performance",
      "summary": "Nested non-virtualized FlatLists",
      "status": "open"
    },
    {
      "id": "FINDING-028",
      "impact": "polish",
      "category": "surfaces",
      "summary": "Radius drift",
      "status": "open"
    },
    {
      "id": "FINDING-029",
      "impact": "polish",
      "category": "interaction",
      "summary": "SwipeButton JS-thread pan + stale closure",
      "status": "open"
    },
    {
      "id": "FINDING-030",
      "impact": "polish",
      "category": "copy",
      "summary": "Dev alert() and stale task name",
      "status": "open"
    },
    {
      "id": "FINDING-031",
      "impact": "polish",
      "category": "states",
      "summary": "Maintenance screen dead end",
      "status": "open"
    },
    {
      "id": "FINDING-032",
      "impact": "polish",
      "category": "navigation",
      "summary": "Dead EventCard and types/index.ts",
      "status": "open"
    },
    {
      "id": "FINDING-033",
      "impact": "polish",
      "category": "anatomy",
      "summary": "Off-grid spacing",
      "status": "open"
    }
  ],
  "unverified": [
    "no device: splash/white flash",
    "cold-start push tap",
    "1.3x and max font scale",
    "Android edge-to-edge insets",
    "VoiceOver/TalkBack",
    "scroll perf",
    "keyboard (N/A, no inputs)",
    "contrast computed from hex, not measured on screen"
  ]
}
```

---

## 4. What works and must be preserved

The owner likes two things: how you pick what to follow, and how your deals show up. Both are good,
and here is what specifically makes them work. In redesign-protocol terms, these are **signature
interactions**. They survive the redesign, and any change to them needs approval.

### 4.1 Team and deal selection (`EventListBlock.tsx`, `DealModal.tsx` subscribe mode)

1. **Grouped by team, not by offer.** People think "I'm a Kings fan", not "I want El Pollo Loco".
   The team is the unit you scan (`EventListBlock.tsx:39-54`), with the offers underneath.
2. **One tap commits a whole team.** The Follow button on the team row subscribes every offer of
   that team (`:62-70`) without opening anything. That's the fastest path from launch to value.
3. **Partial state is visible.** "2 offers available • 1 following" (`:83-85`), and the team shows
   "Following" only when all its offers are followed (`:51-53`). You can be selective without losing
   the summary.
4. **Progressive disclosure.** The per-offer list stays collapsed until you ask for it (`:97-115`),
   so the team list stays scannable. Expanded rows read offer, partner and trigger in one line
   ("Free Jumbo Jack / Jack in the Box • Dodgers pitchers record 7+ strikeouts").
5. **Per-offer opt-in.** A ✓ on each followed offer, and tapping an offer opens its detail with its
   own toggle. Two levels of control: team (bulk) and deal (granular).
6. **Optimistic toggles.** The state flips immediately (`useSubscriptions.ts:42-51`), so the UI
   never waits on the network to acknowledge a choice.
7. **Team color as identity.** A colored badge with the team abbreviation (`:77-79`) is the only
   strong color on the card, and it's categorical, not decorative.

**In the 8-bit version these map to:** TeamSelectGrid cells (team as the unit, one tap follows the
whole team, lit when followed), a fill indicator per cell ("1/2" when partially followed, half-lit
border), a pushed Team screen that lists every deal with its own toggle (the expanded rows, now with
the sprite), optimistic toggles plus selection haptics, and a team-color stripe on each cell. The
behavior is the same; only the skin changes. Section 5.9 and 5.10.

### 4.2 How your deals are shown (`ActiveDealCard`, `DealModal` claim mode, tab behavior, `SubscriptionListBlock`)

1. **Urgency comes first.** Time left is the first thing on the card (`ActiveDealCard.tsx:29-47`)
   and changes color under 6 hours (`isExpiringSoon(…, 6)`). Free food that expires is a race.
2. **The reason is on the card.** The trigger condition sits in its own highlighted box (`:78-83`).
   It makes the deal feel earned ("because the Dodgers struck out 9") instead of like an ad.
3. **The offer is the headline.** The offer name is the largest text (20pt on the card, 28pt in the
   sheet). Team and partner are secondary.
4. **One action per card.**
5. **The app comes to you.** The tab badge counts unclaimed deals (`App.tsx:94`), and the app opens
   on Deals when there's something to claim (`App.tsx:65`). The right first screen depends on state.
6. **Two dismiss intents.** "Got it" (claimed, counts toward stats) and "Stop reminding" (no more
   pushes) are separate server-side dismissal types, and the second needs a deliberate slide so a
   stray tap can't turn off reminders.
7. **Done means gone.** Dismissed deals leave the list at once (local state update in
   `useActiveDeals.ts:52-58`).
8. **Your subscriptions in one flat list.** Team badge, offer, "team • partner", Remove inline
   (`SubscriptionListBlock.tsx`). No drill-down needed to stop watching something.

**In the 8-bit version these map to:** an Inventory row whose DealBar is the expiry clock (drains,
turns timer-orange under 6h, same threshold), a reason line in plain English, the sprite and offer
name as the headline, a single row tap opening the claim screen, the tab badge and state-chosen
landing tab kept as is, StageClear with "GOT IT" and a pixel slide track for "STOP REMINDERS" (same
deliberate gesture, plus an accessibility action), immediate removal with a short step animation,
and the Watching list on the Deals tab with an inline ✕. Section 5.7 and 5.8.

### 4.3 Other things to keep

- The server-driven block system. It's what lets the LA content and layout change without a release.
  Section 6 covers keeping it compatible with old clients.
- Optimistic subscribe with revert.
- Foreground refresh of data and config (`AppDataContext.tsx:138-147`,
  `AppConfigContext.tsx:65-73`).
- Anonymous device users. No sign-up wall.
- `onRequestClose` on the modal (Android back parity).
- `dismissedDeals` / `undoDismissal` are already built but unused. They power a "Claimed" log with
  undo for free.

---

## 5. The 8-bit redesign

### 5.1 App Read and dials

```
APP READ: Location-based sports-rewards utility for LA fans (14 pro teams + USC/UCLA), an
arcade-nostalgia 8-bit language (winning food is a power-up, game night is a level), built on
Expo 58 + StyleSheet tokens + react-native-svg sprites.
platforms: iOS first-class · Android first-class · web best-effort (dev preview only)
posture: unified-brand (one look on both platforms, native navigation underneath)
offline: NOT N/A. Claims happen at a counter, so cached deals must be readable offline
```

| Dial              | Value | Why                                                                                                                         |
| ----------------- | ----- | --------------------------------------------------------------------------------------------------------------------------- |
| DESIGN_EXPRESSION | 9     | Full 8-bit language is the chosen brief. At 8 or above, every inset, press state and accessibility role is owned by the app |
| MOTION_INTENSITY  | 5     | All motion is stepped (2 to 4 frames). Motion is tied to moments: unlock, cursor, timer. No easing curves, no parallax      |
| VISUAL_DENSITY    | 4     | One-handed, glanceable. Pixel fonts need room                                                                               |

**Chrome lags content, even at 9.** Keep the navigator's tab bar and headers, and skin them through
`tabBarStyle`, `tabBarIcon` (sprites), `headerStyle` and `headerTitleStyle`. Do not write a custom
tab bar component. The native one gives back gestures, safe areas, accessibility and badges for
free, and that's where custom-chrome apps usually break.

**Signature element per screen (one bold move):** Tonight has the DealBar, Deals has the sprite
inventory, the claim screen has the STAGE CLEAR drop, Teams has the select grid, You has the season
collection. Everything else stays quiet.

### 5.2 Color tokens

Keep the existing token **names** in `useTheme.tsx`, because they're consumed everywhere
(redesign-protocol: never rename silently), and change the values. Add the new ones.

| Token (existing name) | 8-bit value     | Role                                                 |
| --------------------- | --------------- | ---------------------------------------------------- |
| `background`          | night `#222034` | Screen background, splash                            |
| `surface`             | panel `#3F3F74` | PixelBox fill, grid cells                            |
| `surfaceSecondary`    | deep `#141021`  | DialogBox, tab bar, header, skeleton base            |
| `text`                | `#FFFFFF`       | Body                                                 |
| `textSecondary`       | `#CBDBFC`       | Secondary body (DB32 light blue, used in the sketch) |
| `textMuted`           | muted `#9BADB7` | Captions, only on night or deep (see contrast)       |
| `border`              | `#FFFFFF`       | PixelBox stepped border                              |
| `primary`             | coin `#FBF236`  | Primary action fill, selected cell, cursor           |
| `primaryText`         | night `#222034` | Text on coin                                         |
| `accent`              | coin `#FBF236`  | Single accent (accent lock)                          |
| `success`             | clear `#6ABE30` | Unlocked, followed, DealBar fill                     |
| `successBackground`   | `#141021`       | Deep box behind a clear message                      |
| `warning`             | timer `#DF7126` | Expiring under 6h, countdown                         |
| `warningBackground`   | `#141021`       |                                                      |
| `info`                | sky `#5FCDE4`   | Times, schedule                                      |
| `infoBackground`      | `#141021`       |                                                      |
| **new** `onAccent`    | `#222034`       | Replaces every `'#fff'` on accent (FINDING-004)      |
| **new** `danger`      | hit `#D95763`   | Live dot, errors, offline icon. Never body text      |
| **new** `cellOff`     | `#2B2A40`       | Off-season team cell                                 |
| **new** `barEmpty`    | `#45283C`       | Unlit DealBar cell                                   |

**Contrast, computed from the hex values (WCAG 2.x):**

| Pair                                   | Ratio     | Use                                                                            |
| -------------------------------------- | --------- | ------------------------------------------------------------------------------ |
| white on night                         | 15.9:1    | Body                                                                           |
| coin on night                          | 13.5:1    | Headings, cursor, selected                                                     |
| night on coin                          | 13.5:1    | Primary button label                                                           |
| sky on night                           | 8.6:1     | Times                                                                          |
| muted on deep                          | 8.0:1     | Tab labels (inactive), captions in dialogs                                     |
| muted on night                         | 6.8:1     | Captions                                                                       |
| clear on night                         | 6.8:1     | Status words ("CLEAR", "ON")                                                   |
| timer on night                         | 4.9:1     | Countdown (AA for normal text)                                                 |
| **muted on panel**                     | **4.2:1** | **Fails AA for small text. Use `textSecondary` or white inside panel boxes**   |
| **hit on night**                       | **4.2:1** | **Large or bold labels and icons only (the 3:1 UI threshold). Not small text** |
| **white on hit**                       | **3.8:1** | **The sketch's 6px "P1" tag fails. Use night on coin for that tag**            |
| **#696A6A tab label on deep (sketch)** | **3.4:1** | **Fails. Use muted #9BADB7 for inactive tab labels**                           |

**Dark only for v1.** Set `userInterfaceStyle: "dark"` and remove the Appearance setting. This is a
declared change to an existing setting: users who chose Light lose that choice. A "day level" light
palette can come later as a token set, never as a per-screen fork.

**Server colors:** `banner` and `promo_card` stop accepting hex and take a `tone` token
(`"coin" | "clear" | "sky" | "timer" | "hit" | "panel"`). The client ignores unknown tones. This
closes FINDING-013's server path.

### 5.3 Type tokens

Fonts: `@expo-google-fonts/press-start-2p` and `@expo-google-fonts/pixelify-sans` (both OFL), loaded
with `useFonts` before `SplashScreen.hideAsync()`.

| Token        | Font              | Size / line | Use                                                        | maxFontSizeMultiplier |
| ------------ | ----------------- | ----------- | ---------------------------------------------------------- | --------------------- |
| `display`    | Press Start 2P    | 24 / 32     | STAGE CLEAR!, score digits                                 | 1.2                   |
| `title`      | Press Start 2P    | 16 / 24     | Screen headings (3 words max)                              | 1.2                   |
| `label`      | Press Start 2P    | 12 / 20     | Section labels (INVENTORY, UP NEXT), menu rows, cell codes | 1.3                   |
| `body`       | Pixelify Sans 400 | 17 / 24     | Everything people read as a sentence                       | unbounded             |
| `bodyStrong` | Pixelify Sans 600 | 17 / 24     | Offer names in rows                                        | unbounded             |
| `caption`    | Pixelify Sans 400 | 14 / 20     | Secondary lines, tab labels (12)                           | 1.5                   |

Rules carried over from the design doc, adjusted for device size: Press Start 2P only for 3 words or
fewer, never for sentences, never for numbers people compare except the score and countdown. The
sketch's 6.5 to 8px sizes were drawn on a 272px-wide mock phone (about 0.7x scale). On a device,
nothing in Press Start goes below 12pt. 8/16/24pt render crisp at both 2x and 3x pixel densities;
12pt is crisp at 2x and acceptable at 3x. Android: `includeFontPadding: false` and an explicit
`lineHeight` on Press Start text. Tab labels use Pixelify 12, not Press Start, so they survive 1.3x
scaling.

### 5.4 Spacing, shape, surfaces

- 4pt grid only: 4, 8, 12, 16, 24, 32. Screen gutter 16, set once by the screen owner (FINDING-019).
- Radius 0 everywhere. The stepped corner is the shape.
- No shadows, no blur, no gradients. Elevation is a 4px offset "drop" in deep (#141021) under
  pressable boxes.
- Group with whitespace and dotted 2px rules (panel on night) inside lists. One PixelBox per
  self-contained object (a deal, a dialog), not per section.

### 5.5 Component set (new primitives and their RN implementation)

| Component                                          | What it is                                             | RN implementation notes                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| -------------------------------------------------- | ------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `Sprite` (rename of the in-progress `PixelSprite`) | Renders `{width, height, pixels[], palette{}}` bitmaps | `react-native-svg`. Merge horizontal runs per row into one `<Rect>`, set `shapeRendering="crispEdges"`, memoize by `sprite.id`. **Take an integer `scale`, not an arbitrary `size`**: the current `size={32}` / `size={64}` only stays crisp when width divides it (a 12px or 24px sprite would smear). Missing sprite falls back to a bundled "plate" sprite, never emoji. `accessible={false}` when a text label is adjacent, otherwise `accessibilityLabel` = the sprite's `name` |
| `PixelBox`                                         | Panel with a 4px stepped white border                  | Four absolutely positioned 4px bars inset 4px from each corner, so corners stay notched. No images, no SVG. Variants: `panel`, `deep` (dialog), `coin` (selected)                                                                                                                                                                                                                                                                                                                    |
| `PixelButton`                                      | Primary action                                         | PixelBox in coin, night label (Press Start 12), 4px deep drop. Press = move 2px down-right and remove the drop (instant, no spring). `minHeight: 48`. Role `button`                                                                                                                                                                                                                                                                                                                  |
| `PixelMenu`                                        | Pause-menu list                                        | Full-width rows, 48pt minimum, Press Start 12. Focused or pressed row shows a ▶ cursor (1s stepped blink, solid under reduced motion). Role `menu` / `menuitem`. This replaces button stacks                                                                                                                                                                                                                                                                                         |
| `DialogBox`                                        | RPG text box                                           | Deep box, white border, Pixelify 17 body, coin for key words, bobbing ▼ (2-frame) when there's more text. Tap advances. All text is in the accessibility label at once (screen readers don't step through it)                                                                                                                                                                                                                                                                        |
| `DealBar`                                          | 12-cell HP-style bar                                   | Two modes. **expiry**: cells = time left / (expiresAt - triggeredAt), drains, clear until 6h then timer orange (same threshold as today). **progress**: metric / target (e.g. 5 of 7 K), needs live data. `accessibilityRole="progressbar"` with `accessibilityValue {min, max, now, text: "3 hours 28 minutes left"}`                                                                                                                                                               |
| `TeamCell`                                         | Character-select cell                                  | Team code (Press Start 12), name (Pixelify 14), league + season line (caption). States: off, followed (coin fill), partial ("1/2", coin border only), off-season (cellOff, still selectable). A 4px team-color stripe at the top. Role `checkbox` with `accessibilityState {checked: true/false/'mixed'}`                                                                                                                                                                            |
| `StageClear`                                       | Unlock hero                                            | Sprite at 6x, "STAGE CLEAR!" display text drops in over 4 frames, DialogBox with the reason, countdown, PixelMenu of actions                                                                                                                                                                                                                                                                                                                                                         |
| `PixelSlide`                                       | Re-skin of `SwipeButton`                               | Deep track, coin pixel thumb with ▶, "SLIDE TO STOP" label. Same 70% threshold. Plus an `accessibilityActions` activate. Gesture Handler + Reanimated if RNGH gets added, otherwise keep the native-driver Animated                                                                                                                                                                                                                                                                  |
| `PixelToast`                                       | Transient message                                      | Deep strip above the tab bar, 3s, `accessibilityLiveRegion="polite"` / `announceForAccessibility`                                                                                                                                                                                                                                                                                                                                                                                    |
| `PixelSkeleton`                                    | Loading placeholder                                    | Panel-outline boxes filled with a 2x2 checker dither (deep/night). Dither shifts 1px every 500ms (stepped). Static under reduced motion                                                                                                                                                                                                                                                                                                                                              |
| `SpriteIcon` set                                   | Tab and UI icons                                       | 12x12 bitmaps at 2x: moon (Tonight), bag (Deals), controller-free "flag" (Teams), face (You), bell, gear, ✕, ▶. Drawn for this app, nothing from a known game                                                                                                                                                                                                                                                                                                                        |

### 5.6 Motion and haptics

| Moment                                 | Motion                                                        | Haptic                                                          | Reduced motion                |
| -------------------------------------- | ------------------------------------------------------------- | --------------------------------------------------------------- | ----------------------------- |
| Press any PixelButton or menu row      | 2px press-in, instant                                         | none (6.5: never on plain taps)                                 | same (not decorative)         |
| Toggle a team cell                     | Cell flips to coin in 1 frame                                 | `selectionAsync`                                                | same                          |
| Claim screen opens on an unlocked deal | STAGE CLEAR drops in 4 frames over 240ms, sprite blinks twice | `notificationAsync(Success)` once per deal (remember shown IDs) | No drop or blink, haptic kept |
| DealBar update                         | Cells change one at a time, 40ms stagger                      | none                                                            | Instant                       |
| Menu cursor                            | ▶ blinks, 1s stepped                                          | none                                                            | Solid                         |
| Dialog ▼                               | 2-frame bob, 1s                                               | none                                                            | Static                        |
| "GOT IT"                               | Row collapses in 3 steps, toast "Claimed: $2 McGriddles"      | `impactAsync(Light)`                                            | Instant removal, toast kept   |
| Countdown                              | Seconds tick only on the claim screen, minutes elsewhere      | none                                                            | Same (it's data)              |

Use `useReducedMotion()` from Reanimated if it's added, otherwise
`AccessibilityInfo.isReduceMotionEnabled` plus its change event. Stepped animation in RN:
`Animated.timing` with a stepped `interpolate` (or `withSequence` of
`withTiming(…, {duration: 0})` + `withDelay` in Reanimated). Never `Easing.linear` tweens on
transform.

### 5.7 Proposed navigation (Nav Read)

Recommended implementation is Expo Router (file tree = Nav Read, typed routes, linking for free).
The equivalent in React Navigation 7 is a root native stack (onboarding group, tabs, `deal/[id]`
modal) with a `linking` config. Either works. **This whole section is approval-gated** (tab order,
labels, routes and deep links change).

```
NAV READ
platforms: iOS first-class · Android first-class · web best-effort (dev)
tabs (4): Tonight · Deals · Teams · You
app/
  _layout.tsx               root Stack. Holds the splash until fonts + user + cached deals load.
                            Redirects to (onboarding) when no onboarding flag is stored
  (onboarding)/             fullscreen group, own stack, no tab bar
    teams.tsx               SELECT YOUR TEAMS (grid)
    deals.tsx               CHOOSE REWARDS for the picked teams (all on by default)
    alerts.tsx              permission prime -> system prompt -> (tabs)
  (tabs)/
    _layout.tsx             Tabs (native), sprite icons, Deals badge = unclaimed count
    index.tsx       Tonight -> deal/[id]
    deals/index.tsx Deals   segments: Unlocked · Watching · Claimed -> deal/[id]
    teams/index.tsx Teams   -> push teams/[teamId]
    teams/[teamId].tsx      push (bar stays): that team's deals with toggles -> deal/[id]
    you/index.tsx   You     -> push you/alerts, you/about
  deal/[id].tsx             modal (pageSheet, ✕ CLOSE in header). id = triggered deal id, or
                            offer id for a preview. Covers the tab bar
entry points: deal/[id] <- push tap, Inventory row, Tonight row, Team deal row
              teams/[teamId] <- lineup row on Teams, team name on deal/[id]
sheets: none (stop-reminders confirm is inline on the claim screen)
deep links: freebie://deal/:id, freebie://teams/:teamId, freebie://deals, freebie://tonight
android back: default pop. In (onboarding), back steps through the group. On tab roots, back exits
initial tab: Deals if unclaimed deals > 0, else Tonight (keeps today's state-aware landing)
max taps to any core screen: claim = 1 (push) or 1 (Deals row); follow a team = 1; toggle one deal = 2
```

Why four tabs and why these: **Tonight** answers "can I win food today?" (time-bound). **Deals**
answers "what do I have and what am I watching?" (keeps the current tab, its badge and the owner's
favorite display). **Teams** is the selection flow the owner likes, promoted from "Discover" to a
peer destination. **You** holds stats and settings. There's no Settings tab (4.C), no FAB. Tab label
"Deals" stays rather than the sketch's "ITEMS": it's clearer, it's the word in every push, and it
keeps today's tab. "INVENTORY" appears as the in-screen section label, which is where the game word
belongs ("game words in moderation", design doc).

**Tonight depends on data the API doesn't have yet** (no games or schedule endpoint, sources only
read yesterday's game). Until `GET /games?date=` exists, Tonight shows the unlocked deals and a
"what's on the line" list built from followed deals plus `season` info, with no live score. See 5.9.

### 5.8 Screen by screen

Sample data used throughout (Sunday Nov 15, 2026: NHL in season, LAFC in the MLS playoffs, MLB in
the off-season):

| Team    | Partner          | Reward           | Trigger                        |
| ------- | ---------------- | ---------------- | ------------------------------ |
| Ducks   | McDonald's       | $2 McGriddles    | Home win                       |
| Kings   | El Pollo Loco    | Free BRC burrito | Home win                       |
| LAFC    | Ono Hawaiian BBQ | $5.99 plate      | Goal in the first half at home |
| Dodgers | Jack in the Box  | Free Jumbo Jack  | Pitchers strike out 7+         |
| Dodgers | Panda Express    | $7 plate         | Home win                       |
| Angels  | Del Taco         | 2 free tacos     | 5+ runs at home                |

Kings/El Pollo Loco is unverified per the 8-bit doc. Redemption windows below are placeholders to
confirm against each sponsor's terms.

---

#### 5.8.1 Splash and first launch

```
┌──────────────────────────────┐
│                              │   night #222034, the same color as the first screen
│                              │
│         [palm sprite 8x]     │   pixel sunset palm, 24x24 at 8x = 192pt
│                              │
│        LA FREEBIES           │   Press Start 16, coin, 2px hit drop shadow
│                              │
└──────────────────────────────┘
```

- **Components:** native splash image (palm + wordmark exported from the same bitmap), held by
  `SplashScreen.preventAutoHideAsync()` until fonts, theme, stored user and cached deals load. No
  spinner screen (FINDING-007, 026).
- **New users** go to `(onboarding)/teams`. Returning users go to tabs.
- **Accessibility:** the splash has no interactive content. Hide it within 2s even when the network
  is slow; data loads behind skeletons.

#### 5.8.2 Onboarding 1: Select your teams

```
┌──────────────────────────────┐
│ SELECT YOUR TEAMS            │  title, coin
│ Pick who you root for. 3 on  │  body, white
│ ┌────┐┌────┐┌────┐┌────┐     │
│ │LAK ││ANA ││LAFC││LAD │     │  TeamCell grid, 4 columns
│ │King││Duck││LAFC││Dodg│     │  coin = picked, panel = not picked
│ │NHL ││NHL ││MLS ││Mar │     │  dim = off-season (still pickable)
│ └────┘└────┘└────┘└────┘     │
│  ...LAL LAC LAR LAC(NFL)     │
│  ...LAA LAG ACFC LAS         │
│  ...USC UCLA                 │
│                              │
│ Dim teams are between        │  caption, muted
│ seasons. We'll let you know  │
│ when they're back.           │
│ ┌──────────────────────────┐ │
│ │      ▶ NEXT: 3 TEAMS     │ │  PixelButton, safe-area footer
│ └──────────────────────────┘ │
│        Skip for now          │  text button, 44pt
└──────────────────────────────┘
```

- **Copy:** title "SELECT YOUR TEAMS". Sub-line "Pick who you root for. You can change this any
  time." Button "NEXT: 3 TEAMS" (count live), disabled state "PICK A TEAM". Skip goes to Tonight
  with empty states that point back to Teams.
- **Interaction:** tap a cell to toggle (selection haptic). The first pick gets the P1 tag (night on
  coin, not white on hit) and becomes the first team on Tonight. Not a three-slide tour (tell #31):
  each onboarding screen collects a real input.
- **Accessibility:** each cell is a checkbox: "Kings, NHL, in season, selected". The grid reflows to
  3 columns at font scale 1.3 and above. The footer button stays visible (safe-area footer, not
  `bottom: 30`).

#### 5.8.3 Onboarding 2: Choose rewards

```
┌──────────────────────────────┐
│ ‹                            │  native header back
│ CHOOSE REWARDS               │
│ All on. Turn off any you     │
│ don't want.                  │
│ ── KINGS ─────────────────── │  label
│ [burrito] Free BRC burrito ■ │  sprite 2x, bodyStrong, toggle (coin square = on)
│           El Pollo Loco      │  caption
│           Home win           │  caption, sky
│ ── DUCKS ─────────────────── │
│ [mcgrdl]  $2 McGriddles    ■ │
│           McDonald's · Home win
│ ── DODGERS · BACK IN MARCH ─ │
│ [jumbo]   Free Jumbo Jack  ■ │
│           Jack in the Box · 7+ strikeouts
│ [panda]   $7 plate         ■ │
│           Panda Express · Home win
│ ┌──────────────────────────┐ │
│ │     ▶ WATCH 4 DEALS      │ │
│ └──────────────────────────┘ │
└──────────────────────────────┘
```

- This is the expanded-offers view from today's Discover (section 4.1, points 4 and 5), shown once
  with everything pre-checked, so "follow the team" stays the default.
- **Interaction:** one batch request on WATCH (FINDING-025), optimistic.
- **Accessibility:** each row is a switch: "$2 McGriddles from McDonald's, when the Ducks win at
  home, on".

#### 5.8.4 Onboarding 3: Notification prime

```
┌──────────────────────────────┐
│                              │
│      [bell sprite 6x]        │
│ ╔══════════════════════════╗ │
│ ║ Deals unlock after the   ║ │  DialogBox
│ ║ game, and most are gone  ║ │
│ ║ by the next morning.     ║ │
│ ║ Want a ping when one     ║ │
│ ║ drops?                 ▼ ║ │
│ ╚══════════════════════════╝ │
│   ▶ TURN ON ALERTS           │  PixelMenu row 1 -> requestPermissionsAsync()
│     NOT NOW                  │  row 2 -> tabs, You tab shows "Alerts: off"
└──────────────────────────────┘
```

- Fixes FINDING-002. The system prompt only appears after TURN ON ALERTS. Register the push token
  only on grant.
- If the user skipped team selection, the prime moves to right after their first follow from the
  Teams tab (the same DialogBox in a modal).
- **Accessibility:** the dialog text is read in one go. Menu rows are buttons.

#### 5.8.5 Tonight (home)

With live data (needs `GET /games`):

```
┌──────────────────────────────┐
│ [palm] LA FREEBIES           │  header: wordmark (Q1 of the Trunk Test)
│        Sunday · 3 teams      │
│ ┌──────────────────────────┐ │
│ │ KINGS  2  VS  1  SHARKS  │ │  PixelBox, score in display type
│ │      3RD PERIOD · 8:14   │ │  label, muted
│ │ WIN ■■■■■■■■■□□□         │ │  DealBar progress mode
│ │ [burrito] Hold the lead: │ │  body
│ │ a win means a free BRC   │ │
│ │ burrito.                 │ │
│ └──────────────────────────┘ │
│ INVENTORY                    │  label
│ [mcg] $2 McGriddles   NEW    │  row -> deal/[id]
│       Ducks won · til 10:30am│
│ UP NEXT                      │
│ LAFC vs Austin      7:00pm   │  sky time
│ [ono] Goal in 1st half,      │
│       $5.99 plate            │
│ Lakers vs Suns      7:30pm   │
│       No deal this game      │  muted
└──────────────────────────────┘
│ ☾Tonight  ▣Deals  ⚑Teams  ☺You│  tab bar: deep, sprite icons, Pixelify 12 labels
```

Before live data exists (ship this first):

```
┌──────────────────────────────┐
│ [palm] LA FREEBIES           │
│ INVENTORY                    │
│ [mcg] $2 McGriddles          │
│       Ducks won 4-2 at home  │
│       ■■■■■■■■□□□□ 3h left   │  DealBar expiry mode
│ ON THE LINE                  │
│ [burr] Free BRC burrito      │
│       If the Kings win at    │
│       home                   │
│ [ono]  $5.99 Ono plate       │
│       If LAFC scores in the  │
│       first half at home     │
│ OFF-SEASON                   │
│ Dodgers, Angels · back in    │
│ March                        │
└──────────────────────────────┘
```

- **Components:** PixelBox (live game, max one), DealBar, Sprite, PixelMenu-style rows.
- **Data:** active deals, subscriptions, new `season` fields on teams, `GET /games` (later).
- **Copy rule:** the condition is always written as "If the Kings win at home", never as raw rule
  text.
- **Empty (no teams followed):** see 5.8.13.
- **Accessibility:** the live box reads as a single element: "Kings 2, Sharks 1, third period, 8:14
  left. A win means a free BRC burrito." Live score changes announce politely at most once a minute.

#### 5.8.6 Deals: Unlocked (inventory)

```
┌──────────────────────────────┐
│ DEALS                        │  header title (Press Start 16)
│ [UNLOCKED 2][WATCHING 5][CLAIMED]│  segmented, Press Start 12, 44pt
│ ┌──────────────────────────┐ │
│ │[mcgriddle 4x] $2         │ │  PixelBox per deal (a discrete tappable object)
│ │               McGriddles │ │  bodyStrong 17 (the headline)
│ │ Ducks won 4-2 at home    │ │  body, the reason (the trigger box today)
│ │ McDonald's               │ │  caption, muted
│ │ ■■■■■■■■□□□□  3h 28m left│ │  DealBar expiry, clear
│ └──────────────────────────┘ │
│ ┌──────────────────────────┐ │
│ │[burrito 4x]  Free BRC    │ │
│ │              burrito     │ │
│ │ Kings won 3-1 at home    │ │
│ │ El Pollo Loco            │ │
│ │ ■■□□□□□□□□□□  1h 10m left│ │  under 6h: timer orange, label "ENDS SOON"
│ └──────────────────────────┘ │
└──────────────────────────────┘
```

- Preserves section 4.2: urgency (the bar is the first strong color), the reason, the offer as
  headline, one action (tap the box), the badge, Deals as the landing tab when non-empty.
- Card order changes from "pills, then name, then trigger" to "sprite and name, reason, partner,
  bar". Same facts, one shape instead of five (FINDING-016).
- "Redeem" goes away. The whole box opens the claim screen, and the claim screen names what to do.
- **Data:** `ActiveDeal` plus a new `game` summary (`opponent, score, home`) from
  `triggered_events.payload`, so the reason line can say "won 4-2 at home". Fallback when missing:
  the trigger text.
- **Accessibility:** box role button, label "$2 McGriddles from McDonald's. Ducks won 4-2 at home. 3
  hours 28 minutes left." The bar's own value is included in that label, not read twice.

#### 5.8.7 Deals: Watching (your selected deals)

```
┌──────────────────────────────┐
│ [UNLOCKED 2][WATCHING 5][CLAIMED]
│ KINGS                    ▶   │  label, team row -> teams/[teamId]
│ [burr] Free BRC burrito   ✕  │  ✕ = stop watching, 44pt hit area
│        El Pollo Loco · Home win
│ DUCKS                    ▶   │
│ [mcg]  $2 McGriddles      ✕  │
│        McDonald's · Home win │
│ LAFC                     ▶   │
│ [ono]  $5.99 plate        ✕  │
│        Ono · 1st-half goal at home
│ DODGERS · BACK IN MARCH  ▶   │
│ [jumbo] Free Jumbo Jack   ✕  │
│ [panda] $7 plate          ✕  │
│ ┌──────────────────────────┐ │
│ │      ▶ ADD TEAMS         │ │
│ └──────────────────────────┘ │
└──────────────────────────────┘
```

- This is today's "Your Subscriptions" (section 4.2, point 8) moved next to the inventory, where
  people look for it, and grouped by team (section 4.1, point 1).
- **Interaction:** ✕ removes optimistically and shows a toast with UNDO for 4 seconds ("Stopped
  watching $2 McGriddles. UNDO"). No confirm dialog (it's reversible).
- **Accessibility:** ✕ label "Stop watching $2 McGriddles". The team header is a heading.

#### 5.8.8 Deals: Claimed

```
│ THIS SEASON: 12 CLAIMED      │  label
│ [mcg] $2 McGriddles    NOV 9 │
│ [burr] Free BRC burrito NOV 2│
│ [ono] $5.99 plate      OCT 25│
```

- Uses the already-built `dismissedDeals` / `undoDismissal`. Tapping a row opens the claim screen in
  "claimed" state with "MOVE BACK TO UNLOCKED" while the deal hasn't expired.
- Only `got_it` dismissals are listed here. `stop_reminding` ones stay in Unlocked (reminders are
  off but the deal is still claimable), labeled "Reminders off".

#### 5.8.9 Deal detail and claim (`deal/[id]`, with sprite)

Unlocked state:

```
┌──────────────────────────────┐
│ $2 McGriddles       ✕ CLOSE  │  native header in deep, title = the thing
│                              │
│       STAGE CLEAR!           │  display 24, clear, drops in (4 frames)
│     [mcgriddle sprite 6x]    │  16x16 at 6x = 96pt, centered
│ ╔══════════════════════════╗ │
│ ║ You got $2 McGriddles.   ║ │  DialogBox, "$2 McGriddles" in coin
│ ║ The Ducks beat Seattle   ║ │
│ ║ 4-2 at Honda Center.   ▼ ║ │
│ ╚══════════════════════════╝ │
│        03:28:11              │  display, timer orange, ticks per second
│   Breakfast ends 10:30am     │  caption
│ ■■■■■■■■□□□□                 │  DealBar
│ HOW TO CLAIM                 │  label
│ 1 Open the McDonald's app    │  body, numbered steps from the API
│ 2 Deals tab, Ducks deal      │
│ 3 Order at the counter or    │
│   drive-thru                 │
│   ▶ OPEN MCDONALD'S          │  PixelMenu: offerUrl / app link
│     GOT IT, CLAIMED          │  dismissal got_it -> toast + back
│     REMIND ME AT 9:00        │  later (needs a scheduled local notification)
│ ┌[▶]═══ SLIDE TO STOP ════┐  │  PixelSlide: stop_reminding
│ └──────────────────────────┘ │
└──────────────────────────────┘
```

Preview state (deal not triggered, opened from Teams or Watching):

```
│ $2 McGriddles       ✕ CLOSE  │
│     [mcgriddle sprite 6x]    │  no STAGE CLEAR, sprite at 50% (dim, not locked iconography)
│ ╔══════════════════════════╗ │
│ ║ If the Ducks win at home,║ │
│ ║ you get $2 McGriddles    ║ │
│ ║ the next morning.        ║ │
│ ╚══════════════════════════╝ │
│ NEXT CHANCE                  │
│ Ducks vs Seattle · Tue 7:00pm│  needs schedule data, hidden until it exists
│   ▶ WATCHING  ■              │  toggle row, the label is the state
│ Ducks · McDonald's ▶         │  -> teams/ANA
```

Expired state: a gray (cellOff) sprite, "GAME OVER" in Press Start is too cute here, so use the
title "EXPIRED" and the body "This one ended at 10:30am. The next Ducks home game is Tuesday."

- **Components:** StageClear, Sprite, DialogBox, DealBar, PixelMenu, PixelSlide.
- **Data needed:** `event.sprite` (in progress), `activeDeal.game` (new), `event.claimSteps[]` and
  `event.redemptionNote` ("Breakfast ends 10:30am") (new; today these live inside `offerDescription`
  prose), `offerUrl` (exists).
- **Push entry:** this screen is the push target. While the deal loads it shows a skeleton of this
  same layout, never the subscribe view (FINDING-001).
- **Accessibility:** the countdown is announced as "3 hours 28 minutes left", updated once a minute
  for screen readers, not every second. The slide has an "activate" action labeled "Stop reminders
  for $2 McGriddles". STAGE CLEAR is decorative (`accessibilityElementsHidden`); the dialog text
  carries the meaning. Header ✕ is a 44pt button labeled "Close".

#### 5.8.10 Teams (selection, kept and re-skinned)

```
┌──────────────────────────────┐
│ TEAMS                        │
│ 4 followed · tap to add      │  caption
│ ┌────┐┌────┐┌────┐┌────┐     │
│ │P1  ││    ││    ││1/2 │     │  P1 tag (night on coin), partial "1/2"
│ │LAK ││ANA ││LAFC││LAD │     │
│ │Kngs││Dcks││LAFC││Dodg│     │
│ │NHL ││NHL ││PLYF││MAR │     │
│ └────┘└────┘└────┘└────┘     │
│ ... 14 teams, 4 columns ...  │
│ YOUR LINEUP                  │  label
│ Kings       1 of 1 deals  ▶  │  -> teams/LAK
│ Ducks       1 of 1 deals  ▶  │
│ LAFC        1 of 1 deals  ▶  │
│ Dodgers     1 of 2 deals  ▶  │
└──────────────────────────────┘
```

- **Tap a cell** = follow or unfollow every deal for that team (today's Follow button, section 4.1,
  point 2). Unfollowing a team that has an unlocked deal asks inline: "Stop watching the Ducks? Your
  unlocked $2 McGriddles stays." with ▶ STOP WATCHING / KEEP.
- **Partial follow** shows as a coin border plus "1/2" (section 4.1, point 3). Accessibility state
  `checked: 'mixed'`.
- **YOUR LINEUP** rows push the team screen: that's the "See offers" expansion, promoted to a screen
  because it now holds sprites and per-deal toggles.
- **Busy state:** while a batch follow runs, the cell shows a 2-frame dither and ignores taps.
  Failure shows a toast and reverts (FINDING-008).
- League filter is removed: 14 teams fit on one screen. League is a caption on each cell.

#### 5.8.11 Team detail (`teams/[teamId]`)

```
┌──────────────────────────────┐
│ ‹ Teams          DODGERS     │  native header, title = the team
│ ████ (team-color stripe)     │
│ MLB · Season starts Mar 26   │  caption, sky
│   ▶ FOLLOW ALL 2 DEALS       │  PixelMenu, toggles to "UNFOLLOW TEAM"
│ DEALS                        │
│ [jumbo 4x] Free Jumbo Jack  ■│  switch
│   Jack in the Box            │
│   If Dodgers pitchers strike │
│   out 7 or more              │
│ [panda 4x] $7 plate         □│
│   Panda Express              │
│   If the Dodgers win at home │
│ LAST UNLOCKED                │
│ Free Jumbo Jack · Sep 28     │  history, optional
└──────────────────────────────┘
```

- Tapping a deal row (outside the switch) opens `deal/[id]` in preview state.
- **Accessibility:** the switch label includes the condition. The header back is native.

#### 5.8.12 You (profile and settings)

```
┌──────────────────────────────┐
│ YOU                          │
│ ┌──────────────────────────┐ │
│ │ 12 DEALS CLAIMED         │ │  one hero number (FINDING-015)
│ │ 2026-27 season           │ │
│ │ [ANA][LAK][LAFC] badges  │ │  team badge sprites, collected per season
│ └──────────────────────────┘ │
│ ALERTS                       │  label
│   Deal alerts         ON     │  clear "ON" / hit "OFF > Open Settings"
│   Morning reminder    ON     │  (only if the reminder push becomes optional)
│ ABOUT                        │
│   How Freebies works     ▶   │
│   Privacy policy         ▶   │
│   Support code   7F3A-…      │  replaces the raw User ID, copy on tap
│   Version 2.0.0              │
└──────────────────────────────┘
```

- Grouped PixelMenu sections, no card stack (FINDING-014). The Appearance setting is removed (dark
  only, declared).
- **Tell #41 check:** each row must have a read site. If "Morning reminder" can't be backed by a
  server setting, don't ship it.
- Dev-only rows (test push, push token) stay behind `__DEV__`.

#### 5.8.13 Empty states

| Where                             | Sprite      | Title (Press Start) | Body (Pixelify)                                                           | Action        |
| --------------------------------- | ----------- | ------------------- | ------------------------------------------------------------------------- | ------------- |
| Deals > Unlocked, following teams | Empty plate | NO ITEMS YET        | Deals land here the morning after a game. The Kings play tonight at 7:30. | ▶ SEE TONIGHT |
| Deals > Unlocked, no teams        | Empty plate | NO ITEMS YET        | Follow a team and you'll get free food when they win.                     | ▶ PICK TEAMS  |
| Deals > Watching                  | Palm        | NOTHING WATCHED     | Pick your teams and we'll watch their games for you.                      | ▶ PICK TEAMS  |
| Deals > Claimed                   | Empty plate | NONE CLAIMED        | Claimed deals show up here so you can track your season.                  | none          |
| Tonight, no games                 | Moon        | QUIET NIGHT         | None of your teams play tonight. Next up: Kings vs Sharks, Tue 7:30pm.    | ▶ ALL TEAMS   |
| Tonight, everything off-season    | Palm        | OFF-SEASON          | Your teams are between seasons. The Dodgers are back March 26.            | ▶ ADD A TEAM  |

Centered layout is correct here (tell #22 allows it for empty states). One sprite, one line, one
action. No emoji.

#### 5.8.14 Error, offline, maintenance

```
OFFLINE BANNER (top of Tonight and Deals, under the header)
┌──────────────────────────────┐
│ [x] OFFLINE · Saved 8:47pm   │  deep strip, hit icon, caption white
└──────────────────────────────┘
```

- **Offline:** show the cached inventory and claim steps (claims happen at counters). Toggles queue
  or are disabled with the caption "Reconnect to change teams". Copy: "Offline. Showing deals saved
  at 8:47pm."
- **Load error (with cache):** the same banner, worded as "Couldn't refresh. Showing deals saved at
  8:47pm." plus ▶ RETRY.
- **Load error (no cache):** a centered DialogBox: "Couldn't reach Freebies. Check your connection
  and try again." ▶ TRY AGAIN (retries the failed source: user first, then data; FINDING-006). A
  details row ("Error: HTTP 502") behind "SHOW DETAILS" for support.
- **Toggle failure:** a toast: "Couldn't follow the Kings. Try again."
- **Maintenance:**

```
┌──────────────────────────────┐
│       [palm sprite 6x]       │
│          PAUSED              │  title, coin
│ Freebies is down for a quick │  body
│ fix. Your deals are safe.    │
│ ┌ YOUR ITEMS ──────────────┐ │  cached unlocked deals, read-only
│ │ $2 McGriddles · 3h left  │ │
│ └──────────────────────────┘ │
│ Checking again in 0:42       │  caption, re-polls /config every 60s
└──────────────────────────────┘
```

#### 5.8.15 Loading skeletons

- Inventory: 2 PixelBox outlines at the final height (sprite square + 2 text bars + a 12-cell empty
  bar) filled with the 2x2 deep/night dither.
- Teams grid: 14 cells in `cellOff` with no text.
- Claim screen: sprite square + dialog outline + 3 step bars.
- Pull to refresh: native `RefreshControl` tinted coin (no custom pull animation, per section 7).
- No full-screen spinner anywhere. Button-level work uses a 3-frame "..." in the button label.

#### 5.8.16 Notification permission recovery

If permission is denied, the You tab's Alerts row shows "OFF" in hit and the caption "Turn on in
Settings to hear about deals". Tapping it calls `Linking.openSettings()`. Tonight shows a one-time
DialogBox banner after the first unlocked deal arrives while alerts are off: "You got $2 McGriddles,
but alerts are off. Turn them on so you don't miss the next one." ▶ TURN ON / NOT NOW. That's the
second value moment, and the only other place the app asks.

#### 5.8.17 Push copy

Server change in `internal/worker/service.go`. No emoji (the pixel brand can't render in system
notifications; the Android notification icon is the white palm silhouette). Title: who did what and
the reward. Body: where and by when. Data adds `url: "freebie://deal/{triggeredEventId}"`.

| Deal                      | Unlock push title                      | Unlock push body                                                  | Reminder (once, ~90 min before end)     |
| ------------------------- | -------------------------------------- | ----------------------------------------------------------------- | --------------------------------------- |
| Ducks / McDonald's        | Ducks won. $2 McGriddles unlocked      | McDonald's app, Deals tab. Good until 10:30am.                    | Last call: $2 McGriddles end at 10:30am |
| Kings / El Pollo Loco     | Kings won. Free BRC burrito            | Claim in the Loco Rewards app. Good through {date}.               | Your free BRC burrito ends {time}       |
| LAFC / Ono                | LAFC scored early. $5.99 Ono plate     | Order at any LA-area Ono Hawaiian BBQ {window}.                   | $5.99 Ono plate ends {time}             |
| Dodgers / Jack in the Box | 9 strikeouts. Free Jumbo Jack          | Code GODODGERS25 in the Jack app, with a large drink. Today only. | Free Jumbo Jack ends tonight            |
| Dodgers / Panda Express   | Dodgers won. $7 Panda plate            | Order in the Panda app {window}.                                  | $7 Panda plate ends {time}              |
| Angels / Del Taco         | Angels scored 6. 2 free Del Taco tacos | Show the Del Taco app at checkout {window}.                       | 2 free tacos end {time}                 |

Rules: use the real stat when the rule is numeric ("9 strikeouts", "scored 6"). Keep titles to about
40 characters so they don't truncate on the lock screen. No exclamation marks. The windows in braces
come from a new `redemption_note` field per event.

### 5.9 Accessibility summary (applies to every screen)

- Every pressable gets `accessibilityRole`, a label, and `accessibilityState` where it has state.
- 44pt minimum targets (48 for PixelMenu rows).
- Press Start text has `maxFontSizeMultiplier` 1.2 to 1.3. Body text is unbounded and every layout
  wraps (no fixed heights).
- Contrast pairs from 5.2 only. muted is never used on panel, and hit is never used for small text.
- Reduced motion: every stepped animation has a static fallback (5.6).
- Sprites are decorative when the text label sits next to them.
- Countdown and live-score updates are rate-limited for screen readers.

---

## 6. Server-driven blocks: keep, rename, replace

**The compatibility rule.** `/config` is shared by every installed version. Old clients ignore
unknown block types (`BlockRenderer.tsx:57-58`), so adding types is safe. Removing or retyping
blocks under the existing screen keys (`deals`, `discover`, `profile`, `deal_detail`) would blank
those screens on old builds. So:

1. New screens use new keys: `tonight`, `deals_v2`, `teams`, `team_detail`, `you`, `deal_v2`.
2. Old keys keep their current blocks until the minimum supported version passes the redesign.
3. Add `min_app_version` / `max_app_version` columns to `screen_blocks` (one goose migration),
   filter them in `GetConfig` using an `X-App-Version` header, and the client sends
   `Constants.expoConfig.version`.

| Today                           | Decision               | New type / screen                                                                 | Notes                                                                                                                                        |
| ------------------------------- | ---------------------- | --------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `banner`                        | **Keep type, restyle** | `banner` on any screen                                                            | Renders as a DialogBox strip. Config gets `tone` (palette token) and an optional `cta {label, url}`. Hex config is ignored by the new client |
| `active_deals`                  | **Keep type, restyle** | `active_deals` on `deals_v2` (segment Unlocked) and `tonight` (`config.limit: 2`) | Inventory rows with Sprite and DealBar. `emptyTitle/emptySubtitle` still honored                                                             |
| `subscription_list`             | **Keep type, move**    | `subscription_list` on `deals_v2` (segment Watching)                              | Grouped by team, ✕ with undo. No longer on the profile screen                                                                                |
| `event_list`                    | **Replace**            | `team_select_grid` on `teams`                                                     | TeamCell grid + YOUR LINEUP. `config.columns` (default 4)                                                                                    |
| (expanded rows of `event_list`) | **Replace**            | `team_deals` on `team_detail`                                                     | Per-deal switches with sprites                                                                                                               |
| `league_filter`                 | **Retire**             | none                                                                              | LA-only, 14 teams. Keep the component registered so old configs don't crash; it renders null for the new keys                                |
| `user_stats`                    | **Replace**            | `season_card` on `you`                                                            | One hero number + collected team badges                                                                                                      |
| `settings`                      | **Keep type, restyle** | `settings` on `you`                                                               | PixelMenu groups. `showThemeToggle` is ignored (dark only)                                                                                   |
| `promo_card`                    | **Keep type, restyle** | `promo_card` on `deal_v2`, `you`                                                  | PixelBox. `tone` instead of hex                                                                                                              |
| (none)                          | **New**                | `tonight_board` on `tonight`                                                      | Live game box (when `/games` exists) or "On the line" list                                                                                   |
| (none)                          | **New**                | `stage_clear`, `claim_steps`, `deal_actions` on `deal_v2`                         | Lets the server reorder the claim screen or insert a promo between the steps and the actions                                                 |
| (none)                          | **New**                | `claimed_log` on `deals_v2` (segment Claimed)                                     | Uses `dismissedDeals`                                                                                                                        |
| `SponsoredCard` (not a block)   | **Keep, flagged off**  | n/a                                                                               | If re-enabled, it goes through `promo_card` instead of hard-coded copy                                                                       |

The renderer itself also changes. One screen-level `FlatList` whose sections are blocks, with screen
padding owned once (FINDING-019, 027) and the screen's error and offline banners rendered by the
renderer from `dealsError` and network state (FINDING-003).

**API additions this design needs:**

| Field / endpoint                                                                                       | For                                     | Priority                                       |
| ------------------------------------------------------------------------------------------------------ | --------------------------------------- | ---------------------------------------------- |
| `event.sprite` `{id, name, width, height, pixels[], palette{}}`                                        | Every deal surface                      | In progress                                    |
| Sprite size rule: width and height multiples of 8 (16x16 default), palette limited to the 16-color set | Crisp integer scaling, palette lock     | With the sprite work                           |
| `activeDeal.game {opponent, teamScore, oppScore, home, venue}` from `triggered_events.payload`         | Reason line, dialog copy                | High                                           |
| `event.claimSteps[]`, `event.redemptionNote`                                                           | Claim screen, push body                 | High                                           |
| `GET /teams` `{id, code, name, league, color, season {status, startsOn}}`                              | Grid, dim cells, off-season copy        | High (teams are currently derived from events) |
| Batch subscribe (`PUT /users/{id}/subscriptions` with an eventIds list, or follow-team)                | FINDING-025                             | High                                           |
| Push data `url`                                                                                        | FINDING-001                             | High                                           |
| `GET /games?date=` with live state                                                                     | Tonight live box, DealBar progress mode | Later                                          |
| `screen_blocks.min_app_version/max_app_version`                                                        | Compatibility                           | Before the first new screen ships              |

---

## 7. Implementation phases

Order follows the redesign protocol's lever order (type, color, states, motion, recomposition,
navigation last), with a correctness phase first, because three of the five high findings are bugs
the redesign shouldn't wait for. Mode is **overhaul** (new visual language, same IA until phase 7).
Per protocol: one finding or component per atomic commit (`style(design): FINDING-NNN …`), re-run
the relevant grep after each, risk governor checked every 5 fixes (stop above 20%, cap 40 fixes per
run for app scope).

**Phase 0: correctness (no visual change, current nav kept)**

- FINDING-001: lift the deal sheet to one root-level instance driven by context state, not a ref set
  by DealsScreen. Add `useLastNotificationResponse` for cold start. Fetch the deal before rendering.
  (The full route conversion happens in phase 7.)
- FINDING-003: expose `dealsError`, add an AsyncStorage cache of active deals, add NetInfo.
- FINDING-006: retry by source.
- FINDING-008: surface toggle errors.
- FINDING-025: batch subscribe endpoint plus client.
- FINDING-024: normalize league case (moot after LA-only, but cheap).
- Delete `EventCard.tsx` and `src/types/index.ts` (FINDING-032) so the sprite work isn't spread over
  dead code.
- Server: push copy (FINDING-023), push `url`, `redemption_note`, `game` summary.

**Phase 1: type and spacing (lever 1)**

- `expo-font` + Press Start 2P + Pixelify Sans. `SplashScreen.preventAutoHideAsync` held until
  fonts, theme and user are ready (FINDING-026).
- Type tokens (5.3) and spacing tokens (4pt). Replace the 88 `fontSize` literals (FINDING-010).
  `minHeight` and `maxFontSizeMultiplier` on buttons (FINDING-011).
- Screen padding owned by BlockRenderer (FINDING-019). Insets on every scrollable (FINDING-020).

**Phase 2: color (lever 2)**

- New values under the existing token names, plus `onAccent`, `danger`, `cellOff`, `barEmpty` (5.2).
- `userInterfaceStyle: "dark"`, remove the Appearance block (declared change).
- Remove every inline color (FINDING-004, 012, 013). Server `tone` support in `banner`/`promo_card`.
- Grep gate: zero hex outside `useTheme.tsx` and sprite data.

**Phase 3: primitives**

- `Sprite` (integer scale, run-merged rects, memo, fallback), `PixelBox`, `PixelButton`,
  `PixelMenu`, `DialogBox`, `DealBar` (expiry mode), `TeamCell`, `PixelSlide`, `PixelToast`,
  `PixelSkeleton`, and the sprite icon set. Each with role, label and state from the start.
- Replace emoji everywhere, including the tab bar (FINDING-009). Server: stop sending emoji `icon`
  once sprites cover every event.

**Phase 4: states (lever 3)**

- Skeletons for Deals, Teams and the claim screen. Remove the full-screen spinner (FINDING-007).
- Offline banner and cached reads, error and empty states from 5.8.13 and 5.8.14, maintenance with
  cached items and a 60s re-poll (FINDING-031).

**Phase 5: press and motion (lever 4)**

- Pixel press-in on every pressable (FINDING-017). `expo-haptics` with the vocabulary in 5.6.
- Stepped animations with reduced-motion fallbacks. StageClear drop.
- `PixelSlide` with an accessibility action (FINDING-005, 029).

**Phase 6: screen recomposition (lever 5), inside the current 3 tabs**

- Deals: Unlocked rows (FINDING-016), Watching and Claimed segments.
- The deal sheet becomes the StageClear claim layout.
- Discover becomes the TeamSelectGrid + lineup. The team detail opens as a push from it.
- Profile becomes You: season card, grouped menus (FINDING-014, 015), support code (FINDING-022).
- Copy pass across every string (FINDING-021, 022).
- Server: `screen_blocks` version columns, then new screen keys and block types (section 6).

**Phase 7: navigation restructure (lever 6, approval required)**

- 4 tabs (Tonight · Deals · Teams · You) per the Nav Read. Deal and team as routes. Linking config
  for `freebie://deal/:id` and the others. Onboarding group with the permission prime (FINDING-002).
  Initial-tab rule kept.
- Decide Expo Router vs React Navigation native stack + linking at this point.
- Tonight ships in the "no live data" form first. The live box and DealBar progress mode follow
  `GET /games`.

**Phase 8: brand surfaces and verification**

- Icon (pixel sunset palm, exported at 1024 from the same bitmap), splash, a white-silhouette
  notification icon (the file exists as `assets/notification-icon.png`; confirm it's a pure-alpha
  silhouette), store screenshots from the new screens.
- Re-run `mobile-design-review` against the baseline in section 3. Run the live pass this review
  couldn't: VoiceOver/TalkBack, font scale 1.3x and max, Android edge-to-edge, cold-start push tap,
  splash continuity.

---

## 8. Unverified (static-only review)

No simulator or device was reachable. These were not verified:

- Splash to first screen continuity and font swap (FINDING-026 is from code)
- Cold-start and background push-tap behavior (FINDING-001 is reconstructed from code; confirm on a
  device before and after the fix)
- Tab bar and buttons at 1.3x and at maximum font scale (FINDING-011 is inferred)
- Android edge-to-edge insets in the deal sheet (FINDING-020)
- VoiceOver and TalkBack reading order
- Scroll performance
- Keyboard handling (N/A: the app has no text inputs)
- Contrast figures are computed from hex values, not measured on rendered pixels
- Narration and Trunk Test are code-reconstructed, not first impressions of rendered screens
- Line numbers are from commit `759d3b7`. `ActiveDealCard.tsx`, `DealModal.tsx` and `EventCard.tsx`
  were being edited by another agent during the review
