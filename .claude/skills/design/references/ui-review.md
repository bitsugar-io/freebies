# Screen review checklist

A screen is done only when every item passes. Fix failures; don't waive them.

## Look

- [ ] Only theme tokens. `scripts/audit-colors.sh` reports nothing new.
- [ ] Press Start 2P only on short display text (3 words or fewer, 12pt or larger). Body text uses
      the system font.
- [ ] Pixel borders and notched corners. No `borderRadius`, gradients, blur or soft shadows on pixel
      UI.
- [ ] Sprites and badges render at whole-number scales.
- [ ] Matches the reference screens in `docs/pivot/la-jumbotron-vs-8bit.html`.

## Use

- [ ] Every tap target is at least 44pt, and every action has a non-gesture alternative.
- [ ] Every interactive element has an accessibility role and label. Decorative ▶ and sprites have
      labels or are hidden.
- [ ] Text contrast is at least 4.5:1, or 3:1 for 18pt and larger. Check new color pairs with
      `scripts/contrast.py`.
- [ ] Works at the largest Dynamic Type size, with display text falling back to the system font.
- [ ] Reduce Motion removes cut-ins, blinking and count-ups. Flashes stay at 3 per second or fewer.
- [ ] Loading, empty, error and offline states exist. An error is never shown as an empty state.

## Content

- [ ] No team logos, team colors, restaurant logos, or "Nintendo"/"NES" in copy.
- [ ] Deal copy gives the reason, the reward and the deadline in Pacific time.

## The two flows the owner loves

These must behave exactly as before:

- [ ] **Team selection:** grouped by team, one-tap follow, partial state ("1 following"), expandable
      per-offer ✓, toggles that flip instantly.
- [ ] **Your deals:** time left first (timer color under 6h), the reason shown, the offer as
      headline, one action, tab badge, and two dismiss choices.
