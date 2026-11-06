[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.27.3+-00ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
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

<table align="center">
  <tr>
    <td align="center" width="20%"><a href="apps/mobile/assets/screenshots/01-home.png"><img src="apps/mobile/assets/screenshots/01-home.png" width="100%" alt="Discover"></a></td>
    <td align="center" width="20%"><a href="apps/mobile/assets/screenshots/02-deal-detail.png"><img src="apps/mobile/assets/screenshots/02-deal-detail.png" width="100%" alt="Deal details"></a></td>
    <td align="center" width="20%"><a href="apps/mobile/assets/screenshots/03-my-deals.png"><img src="apps/mobile/assets/screenshots/03-my-deals.png" width="100%" alt="My deals"></a></td>
    <td align="center" width="20%"><a href="apps/mobile/assets/screenshots/04-active-deal.png"><img src="apps/mobile/assets/screenshots/04-active-deal.png" width="100%" alt="Active deal"></a></td>
    <td align="center" width="20%"><a href="apps/mobile/assets/screenshots/05-notification.png"><img src="apps/mobile/assets/screenshots/05-notification.png" width="100%" alt="Notification"></a></td>
  </tr>
  <tr>
    <td align="center"><sub>Discover</sub></td>
    <td align="center"><sub>Deal details</sub></td>
    <td align="center"><sub>My deals</sub></td>
    <td align="center"><sub>Active deal</sub></td>
    <td align="center"><sub>Notification</sub></td>
  </tr>
</table>

## Components

- [Mobile app](apps/mobile/): React Native app built with Expo.
- [API](services/api/): Go server for leagues, offers, subscriptions, and push notifications.
- [Scheduler](services/scheduler/): Go CronJob that polls live games and triggers deals.

## Documentation

- [Development guide](docs/development.md)
- [Contributing](docs/contributing.md)
- [Architecture](docs/architecture.md)
- [Operations](docs/operations.md)
- Mobile: [architecture](apps/mobile/docs/architecture.md),
  [development](apps/mobile/docs/development.md), [App Store checklist](apps/mobile/APP-STORE.md),
  [privacy policy](apps/mobile/PRIVACY-POLICY.md)
- API: [architecture](services/api/docs/architecture.md),
  [development](services/api/docs/development.md), [API reference](services/api/docs/api.md),
  [CLI reference](services/api/docs/cli.md), [deployment](services/api/docs/deployment.md)

## License

The [MIT](LICENSE) License.
