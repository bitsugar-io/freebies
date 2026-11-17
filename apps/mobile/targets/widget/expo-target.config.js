/** Lock-screen widget: the newest win, or today's game for a followed team. See index.swift. */
/** @type {import('@bacons/apple-targets/app.plugin').ConfigFunction} */
module.exports = (config) => ({
  type: 'widget',
  name: 'FreebiesWidget',
  displayName: 'Freebies',
  bundleIdentifier: '.widget',
  deploymentTarget: '16.4',
  frameworks: ['SwiftUI', 'WidgetKit'],
  entitlements: {
    'com.apple.security.application-groups':
      config.ios.entitlements['com.apple.security.application-groups'],
  },
});
