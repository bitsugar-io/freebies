# App Store Submission Checklist & Metadata

## App Store Connect Metadata

### App Name

```
Freebie - Sports Deals Alert
```

### Subtitle (30 characters max)

```
Never Miss a Free Food Deal
```

### Keywords (100 characters max, comma-separated)

```
sports,deals,free food,promotions,MLB,NBA,NFL,NHL,baseball,basketball,football,hockey,alerts
```

### Description

```
Never miss out on free food and promotional deals from your favorite sports teams!

Freebie alerts you when your team triggers promotional offers - like free tacos when they score a certain number of runs, or free pizza when they win by a specific margin.

HOW IT WORKS:
1. Browse available deals from MLB, NBA, NFL, and NHL teams
2. Subscribe to the offers you want to track
3. Get notified the moment a deal is triggered
4. Claim your freebie before it expires!

FEATURES:
• Real-time push notifications when deals are triggered
• Track multiple teams and offers
• See countdown timers for expiring deals
• One-tap access to claim your freebies
• Support for all major sports leagues

Stop missing out on free tacos, pizza, wings, and more. Download Freebie today and start saving!
```

### Promotional Text (170 characters max, can be updated without app review)

```
New deals added weekly! Get notified instantly when your favorite teams trigger free food promotions.
```

### What's New (Version 1.0.0)

```
Initial release of Freebie!

• Browse promotional deals from MLB, NBA, NFL, and NHL teams
• Subscribe to track your favorite offers
• Get push notifications when deals are triggered
• Track active deals and expiration times
```

---

## Required URLs

### Support URL (Required)

```
https://github.com/retr0h/freebies/issues
```

### Privacy Policy URL (Required)

```
https://github.com/retr0h/freebies/blob/main/apps/mobile/PRIVACY-POLICY.md
```

### Marketing URL (Optional)

```
https://github.com/retr0h/freebies
```

---

## App Information

### Category

**Primary**: Sports **Secondary**: Food & Drink

### Age Rating

**4+** (No objectionable content)

### Content Rights

- [x] This app does not contain third-party content that requires rights

### App Review Information

#### Contact Information

```
First Name: [Your first name]
Last Name: [Your last name]
Phone: [Your phone]
Email: [Your email]
```

#### Demo Account

```
Not required - app uses anonymous device-based accounts
```

#### Notes for Reviewer

```
This app helps users discover promotional deals offered by sports teams (e.g., "Free tacos when the team scores 5+ runs").

No login is required - the app creates an anonymous account based on the device ID.

To test:
1. Open the app and browse available deals
2. Subscribe to any deal by tapping the bell icon
3. View your subscriptions in the "My Deals" tab

Note: Push notifications require a physical device to test.
```

---

## Screenshots Required

Capture on iPhone 15 Pro Max (6.7" - 1290x2796). App Store Connect will auto-scale for smaller
devices.

See [`assets/screenshots/README.md`](assets/screenshots/README.md) for capture instructions.

| #   | Filename              | Description                       |
| --- | --------------------- | --------------------------------- |
| 1   | `01-home.png`         | Home screen - deals by league     |
| 2   | `02-deal-detail.png`  | Deal detail with subscribe button |
| 3   | `03-my-deals.png`     | My Deals tab - subscribed offers  |
| 4   | `04-active-deal.png`  | Triggered deal with countdown     |
| 5   | `05-notification.png` | Push notification (can be mocked) |

---

## Build Checklist

### Before Submitting

- [x] Privacy Policy hosted at public URL
- [x] Support page/URL set up
- [x] App icon is final (1024x1024 for App Store)
- [x] Screenshots captured (`assets/screenshots/`)
- [x] Version number updated in app.json (1.0.0)
- [x] Build number set (1)

### EAS Build Commands

```bash
# Build for TestFlight (internal testing)
eas build --platform ios --profile preview

# Build for App Store submission
eas build --platform ios --profile production

# Submit to App Store
eas submit --platform ios --profile production
```

---

## Apple Developer Account Requirements

1. **Apple Developer Program** ($99/year)
   - https://developer.apple.com/programs/enroll/

2. **App Store Connect Access**
   - Create app record before first submission
   - https://appstoreconnect.apple.com

3. **Certificates & Provisioning** (EAS handles this automatically)
   - Distribution certificate
   - App Store provisioning profile

---

## TestFlight Setup (Recommended First)

1. Build with preview profile: `eas build --platform ios --profile preview`
2. Submit to TestFlight: `eas submit --platform ios`
3. In App Store Connect, add internal testers
4. Test thoroughly before production submission
