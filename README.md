[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.25+-00ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![React Native](https://img.shields.io/badge/React_Native-20232A?style=for-the-badge&logo=react&logoColor=61DAFB)](https://reactnative.dev)
[![Expo](https://img.shields.io/badge/expo-000020?style=for-the-badge&logo=expo&logoColor=white)](https://expo.dev)
[![App Store](https://img.shields.io/badge/App_Store-0D96F6?style=for-the-badge&logo=app-store&logoColor=white)](https://apps.apple.com)
[![Turso](https://img.shields.io/badge/turso-4FF8D2?style=for-the-badge&logo=turso&logoColor=black)](https://turso.tech)
[![MLB](https://img.shields.io/badge/MLB-002D72?style=for-the-badge&logo=mlb&logoColor=white)](https://mlb.com)
[![NBA](https://img.shields.io/badge/NBA-17408B?style=for-the-badge&logo=nba&logoColor=white)](https://nba.com)
[![NFL](https://img.shields.io/badge/NFL-013369?style=for-the-badge&logo=nfl&logoColor=white)](https://nfl.com)
[![NHL](https://img.shields.io/badge/NHL-000000?style=for-the-badge&logo=nhl&logoColor=white)](https://nhl.com)

# Freebies

Get notified about free offers when your favorite sports teams win or hit milestones.

## Screenshots

<p align="center">
  <img src="apps/mobile/assets/screenshots/01-home.png" width="160" alt="Home">
  <img src="apps/mobile/assets/screenshots/02-deal-detail.png" width="160" alt="Deal Detail">
  <img src="apps/mobile/assets/screenshots/03-my-deals.png" width="160" alt="My Deals">
  <img src="apps/mobile/assets/screenshots/04-active-deal.png" width="160" alt="Active Deal">
  <img src="apps/mobile/assets/screenshots/05-notification.png" width="160" alt="Notification">
</p>

## Quick Start

Requires [Task](https://taskfile.dev/) runner.

```bash
task setup          # First-time setup

task api:serve  # Terminal 1: Start backend
task mobile:serve   # Terminal 2: Start mobile app
```

Then press `i` for iOS simulator, `w` for web, or scan QR with Expo Go.

## Documentation

- [Development Guide](docs/development.md) - Local setup and development workflow

### Architecture

- [Backend Architecture](services/api/docs/architecture.md) - DOKS, Turso, Go decisions
- [Mobile Architecture](apps/mobile/docs/architecture.md) - Expo, EAS, state management

## Apps

### Mobile App (`apps/mobile/`)

React Native app built with Expo. Features:

- Browse freebie offers by league (MLB, NBA, NFL, NHL)
- Subscribe to deals you're interested in
- Get notified when deals trigger
- Track active deals with expiration timers

#### Documentation

- [Architecture](apps/mobile/docs/architecture.md)
- [Development Guide](apps/mobile/docs/development.md)
- [App Store Checklist](apps/mobile/APP-STORE.md)
- [Privacy Policy](apps/mobile/PRIVACY-POLICY.md)

### Backend Service (`services/api/`)

Go-based API server that:

- Serves league/team/offer data
- Manages user subscriptions
- Tracks triggered deals and dismissals
- Sends push notifications

#### Documentation

- [Architecture](services/api/docs/architecture.md)
- [Deployment Guide](services/api/docs/deployment.md)
- [CLI Reference](services/api/docs/cli.md)
- [Development Guide](services/api/docs/development.md)
- [API Reference](services/api/docs/api.md)

### Scheduler (`services/scheduler/`)

Lightweight Go CLI that runs as a Kubernetes CronJob. Polls live game data and calls API internal
endpoints to trigger notifications when deal conditions are met.

## Data

All offer data is managed through SQL migrations in `services/api/internal/db/migrations/`:

- `001_schema.sql` - Database schema
- `002_initial_leagues.sql` - MLB, NBA, NFL, NHL leagues
- `003_mlb_data.sql` - MLB teams and offers
- `004_nba_data.sql` - NBA teams and offers
- `005_nfl_data.sql` - NFL teams and offers
- `006_nhl_data.sql` - NHL teams and offers

To add new offers, create a new migration file (e.g., `007_add_new_team.sql`).

## Prerequisites

- Node.js 20+ (for mobile app)
- Go 1.25+ (for backend)
- [Task](https://taskfile.dev/) (task runner)
- [EAS CLI](https://docs.expo.dev/eas/) (for mobile builds)
- [mise](https://mise.jdx.dev/) (optional, for version management)

## License

The [MIT](LICENSE) License.
