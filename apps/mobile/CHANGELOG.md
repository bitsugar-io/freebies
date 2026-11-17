# Changelog

All notable changes to the Freebies iOS app will be documented in this file.

Format based on [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

## [2.0.0] - 2026-11-16

Freebies is now Freebies: LA Sports Deals, for every LA team.

### Added

- Deals for the Lakers, Clippers, Kings, Ducks, Rams, Chargers and UCLA, with more LA teams as their
  seasons start
- TODAY: deals anyone can claim right now and today's LA games with the deals riding on them
- AT THE GAME tag for deals you claim at the venue
- A team-color stripe and "RAMS WON!" when your team wins, and team-color rays on the deal screen
- Locked codes that unlock later, with a countdown; pause alerts from the menu; a trophy case of
  claimed deals
- Lock-screen widget: your newest win, or today's game with a deal on it

### Changed

- New 8-bit look and the Trophy Burger icon
- Opens on TEAMS until you follow a team, then on TODAY, or on WON when you have a win to claim
- Upgrade to Expo SDK 58 and React Native 0.88
- Require iOS 16.4 or later, the minimum for Expo SDK 58

### Fixed

- Saved deals show without a signal, and a slow connection no longer drops your follows
- Deals you already won stay yours if you unfollow the team

## [1.1.1] - 2026-04-05

### Fixed

- Fix tab jump bug where following a team on Discover would flash the loading spinner and land on
  the Deals tab
- Auto-subscribe existing team followers when new deals are added via migration
- Fix tab bar overlap on home indicator devices (bottom padding)

### Added

- LA Dodgers deals: Free 6pc McNuggets (McDonald's), Free Dodger Dog + Coke (ampm)
- LA Angels deals: $2 McGriddles (McDonald's), 2 Free Del Tacos (Del Taco)
- San Diego Padres deals: Free Jumbo Jack, Free Medium Fries, 25% Off at Petco
- Stolen base tracking for ampm deal trigger

## [1.1.0] - Skipped

Not released.

## [1.0.0] - 2026-03-19

### Added

- Initial App Store release
- Live game monitoring with push notifications for triggered deals
- Discover tab to browse and follow teams
- Deals tab showing active deals for followed teams
- Support for MLB deals (Dodgers Jack in the Box, Panda Express)
- Dark mode support
- Expo push notifications
