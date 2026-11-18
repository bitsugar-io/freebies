<p align="center">
  <picture>
    <source srcset="asset/logo-dark.svg" media="(prefers-color-scheme: dark)">
    <source srcset="asset/logo-light.svg" media="(prefers-color-scheme: light)">
    <img src="asset/logo-dark.svg" alt="Freebies: LA Sports Deals" width="584">
  </picture>
</p>

<p align="center">When LA wins, you win.</p>

<p align="center">
  <a href="LICENSE"><img alt="MIT license" src="https://img.shields.io/badge/license-MIT-FDB927?style=for-the-badge&amp;labelColor=552583&amp;color=FDB927"></a>
  <a href="https://go.dev"><img alt="Go" src="https://img.shields.io/badge/Go-1.27-FCFCFC?style=for-the-badge&amp;labelColor=005A9C&amp;color=FCFCFC&amp;logo=go&amp;logoColor=white"></a>
  <a href="https://expo.dev"><img alt="Expo and React Native" src="https://img.shields.io/badge/Expo-React_Native-FDB927?style=for-the-badge&amp;labelColor=552583&amp;color=FDB927&amp;logo=expo&amp;logoColor=white"></a>
  <img alt="iOS" src="https://img.shields.io/badge/iOS-App_Store-FCFCFC?style=for-the-badge&amp;labelColor=005A9C&amp;color=FCFCFC&amp;logo=apple&amp;logoColor=white">
  <a href="https://conventionalcommits.org"><img alt="conventional commits" src="https://img.shields.io/badge/Conventional_Commits-1.0.0-FDB927?style=for-the-badge&amp;labelColor=552583&amp;color=FDB927"></a>
  <a href="https://just.systems"><img alt="built with just" src="https://img.shields.io/badge/Built_with-Just-FCFCFC?style=for-the-badge&amp;labelColor=005A9C&amp;color=FCFCFC&amp;logo=just&amp;logoColor=white"></a>
  <a href="https://github.com/tekk/hovnokod-badge"><img alt="hovnokod" src="https://raw.githubusercontent.com/tekk/hovnokod-badge/main/assets/badges/hovnokod-for-the-badge.svg"></a>
</p>

I built this for myself because LA restaurants give away food when LA teams win, and I kept finding
out after the deal had expired. The app watches the games for the teams I follow and sends a push
the moment a deal unlocks, with where to claim it and when it ends.

<table align="center">
  <tr>
    <td align="center" width="50%"><a href="asset/screenshots/won.png"><img src="asset/screenshots/won.png" width="100%" alt="Two deals won after Lakers and Rams wins, under a Lakers team-color stripe"></a></td>
    <td align="center" width="50%"><a href="asset/screenshots/deal.png"><img src="asset/screenshots/deal.png" width="100%" alt="A won deal over Rams-color victory rays, with a countdown and how to claim it"></a></td>
  </tr>
  <tr>
    <td align="center"><sub>Won</sub></td>
    <td align="center"><sub>You won</sub></td>
  </tr>
  <tr>
    <td align="center" width="50%"><a href="asset/screenshots/today.png"><img src="asset/screenshots/today.png" width="100%" alt="Today: deals anyone can claim now and today's LA games"></a></td>
    <td align="center" width="50%"><a href="asset/screenshots/teams.png"><img src="asset/screenshots/teams.png" width="100%" alt="Basketball teams to follow, in their team colors"></a></td>
  </tr>
  <tr>
    <td align="center"><sub>Today</sub></td>
    <td align="center"><sub>Teams</sub></td>
  </tr>
</table>

## Features

| Feature              | What it does                                                                                  |
| -------------------- | --------------------------------------------------------------------------------------------- |
| LA teams only        | Follow LA and Orange County teams; every deal is tied to one of them.                         |
| Game-triggered deals | A Lakers win, a Rams interception or a Kings goal in the last minute of the 2nd unlocks food. |
| Push alerts          | A notification arrives when a deal unlocks, with where to claim it and when it ends.          |
| Today                | Deals anyone can claim right now, and today's LA games with the deal riding on each.          |
| Lock-screen widget   | Your newest win, or today's game with a deal on it, on the lock screen. It refreshes itself.  |
| Verified deals       | A deal goes live only after it is checked on the sponsor's or team's own page.                |
| Pause alerts         | Mute pushes for a while without unfollowing your teams.                                       |
| 8-bit look           | Pixel sprites and a retro palette, built for touch, screen readers and reduced motion.        |
| Server-driven        | New deals, teams, seasons and sprites reach the app from the API, without an app release.     |

Freebies: LA Sports Deals is not affiliated with, endorsed by or sponsored by any team, league,
school or restaurant.

## Quickstart

Run from a checkout. Requires [mise](https://mise.jdx.dev), which installs Go, Node, Bun and
[just](https://just.systems).

```sh
mise install
just setup        # first time: install the app's dependencies
just api-serve    # API on :8080; migrates the database
just mobile-serve # Expo dev server, in a second terminal
```

Press `i` in the Expo terminal to open the iOS simulator.

## Skills

Every workflow lives in a [Claude skill](https://agentskills.io) under `.claude/skills/`; people can
read them too. [ops](.claude/skills/ops/SKILL.md) sets up, runs and tests the API and the app, adds
deals, deploys the backend and releases iOS. [design](.claude/skills/design/SKILL.md) builds and
reviews 8-bit screens and draws and ships pixel sprites.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development, testing and releases.

## License

[MIT](LICENSE)
