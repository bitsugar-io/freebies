import React, { useCallback, useEffect, useRef, useState } from 'react';
import { StatusBar } from 'expo-status-bar';
import * as Linking from 'expo-linking';
import * as Notifications from 'expo-notifications';
import * as SplashScreen from 'expo-splash-screen';
import { useFonts, PressStart2P_400Regular } from '@expo-google-fonts/press-start-2p';
import { DarkTheme, LinkingOptions, NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { SafeAreaProvider } from 'react-native-safe-area-context';

import { WonScreen } from './src/screens/WonScreen';
import { TodayScreen } from './src/screens/TodayScreen';
import { TeamsNavigator } from './src/screens/TeamsNavigator';
import { YouScreen } from './src/screens/YouScreen';
import { DealScreen } from './src/screens/DealScreen';
import { AlertsScreen } from './src/screens/AlertsScreen';
import { MaintenanceScreen } from './src/screens/MaintenanceScreen';
import { HudTabBar } from './src/components/pixel/HudTabBar';
import { ScreenWipeProvider } from './src/components/pixel/ScreenWipe';
import { VictoryCutInHost } from './src/components/pixel/VictoryCutIn';
import { LaunchRays } from './src/components/pixel/LaunchRays';
import { useReducedMotion } from './src/hooks/useReducedMotion';
import { PauseScreen } from './src/screens/PauseScreen';
import { ToastProvider } from './src/components/pixel/Toast';
import { UserProvider } from './src/hooks/useUser';
import { colors, fonts } from './src/hooks/useTheme';
import { urlFromNotification } from './src/hooks/usePushNotifications';
import { launchCorrection, launchDataSettled, launchTab, navSnapshot } from './src/navigation/launch';
import { serverNow } from './src/utils/serverClock';
import { AppDataProvider, useAppData } from './src/context/AppDataContext';
import { AppConfigProvider, useAppConfig } from './src/context/AppConfigContext';
import { navigationRef } from './src/navigation/ref';
import { LINK_CONFIG } from './src/navigation/linkConfig';
import { attachLinkListener, deliverLink, firstSighting, initialLink } from './src/navigation/pendingLink';
import type { RootStackParamList, TabParamList } from './src/navigation/types';

SplashScreen.preventAutoHideAsync().catch(() => {});

const Tab = createBottomTabNavigator<TabParamList>();
const Stack = createNativeStackNavigator<RootStackParamList>();


/** Notification taps become freebie:// links, so cold start, background and foreground all share one path. */
function notificationUrl(response: Notifications.NotificationResponse | null | undefined) {
  return urlFromNotification(response?.notification);
}

// Listen from launch, not from when the navigator mounts: the splash, title and maintenance
// screens render no navigator, and a link that lands then must wait for one (see pendingLink.ts).
const linkGlobal = globalThis as { __freebieStopLinks?: () => void };
linkGlobal.__freebieStopLinks?.(); // Fast Refresh re-runs this module; drop the old listeners
const urlSub = Linking.addEventListener('url', ({ url }) => deliverLink(url));
let pushSub: Notifications.EventSubscription | undefined;
try {
  pushSub = Notifications.addNotificationResponseReceivedListener((response) => {
    const url = notificationUrl(response);
    if (url && firstSighting(response.notification.request.identifier)) deliverLink(url);
  });
} catch {
  // Notifications unavailable (web, Expo Go): links still work.
}
linkGlobal.__freebieStopLinks = () => {
  urlSub.remove();
  pushSub?.remove();
};

async function readLaunchUrl(): Promise<string | null> {
  const url = await Linking.getInitialURL();
  if (url) return url;
  // Cold start from a push tap.
  try {
    const response = await Notifications.getLastNotificationResponseAsync();
    const fromPush = notificationUrl(response);
    return fromPush && firstSighting(response?.notification.request.identifier) ? fromPush : null;
  } catch {
    return null;
  }
}

const linking: LinkingOptions<RootStackParamList> = {
  prefixes: [Linking.createURL('/'), 'freebie://'],
  config: LINK_CONFIG,
  getInitialURL: () => initialLink(readLaunchUrl),
  subscribe: (listener) => attachLinkListener(listener),
};

const navTheme = {
  ...DarkTheme,
  colors: {
    ...DarkTheme.colors,
    primary: colors.accent,
    background: colors.background,
    card: colors.background,
    text: colors.text,
    border: colors.border,
    notification: colors.accent,
  },
};

function Tabs() {
  const { followedCount, liveWon, userLoading, dealsLoading, subsLoading, eventsLoading, dealsStale, dealsError } =
    useAppData();
  // Saved deals show at once, but the launch tab waits for the fresh list (see launchDataSettled).
  const dataReady = launchDataSettled({
    loading: userLoading || dealsLoading || subsLoading || eventsLoading,
    dealsStale,
    dealsFailed: dealsError !== null,
  });
  // Where the app opens, decided once at launch: see launchTab.
  const [launch] = useState(() => ({
    tab: launchTab(followedCount, liveWon, serverNow()),
    withData: dataReady,
  }));

  // If the splash gave up before the data arrived, move once to the right tab when it does, unless
  // the user already moved or a link opened a screen (see launchCorrection).
  const start = useRef<string | null>(null);
  const touched = useRef(false);
  const settled = useRef(launch.withData);
  useEffect(() => {
    if (settled.current) return;
    const check = () => {
      const root = navigationRef.getRootState();
      const snap = navSnapshot(root);
      if (!snap) return;
      if (start.current === null) {
        start.current = snap;
        // A link or push that opened the app (a deal over the tabs, or another tab) counts as a move.
        const tabs = root?.routes[0]?.state;
        const tab = tabs?.routes[tabs.index ?? 0]?.name ?? launch.tab;
        if ((root?.routes.length ?? 1) > 1 || tab !== launch.tab) touched.current = true;
      } else if (snap !== start.current) touched.current = true;
    };
    check();
    return navigationRef.addListener('state', check);
  }, [launch.tab]);
  useEffect(() => {
    if (settled.current || !dataReady) return;
    settled.current = true;
    const next = launchCorrection({
      initial: launch.tab,
      decidedWithData: launch.withData,
      dataReady,
      touched:
        touched.current ||
        (navigationRef.getRootState()?.routes.length ?? 1) > 1 ||
        (start.current !== null && navSnapshot(navigationRef.getRootState()) !== start.current),
      correct: launchTab(followedCount, liveWon, serverNow()),
    });
    if (next && navigationRef.isReady()) navigationRef.navigate('Tabs', { screen: next });
  }, [dataReady, followedCount, liveWon, launch]);

  const initial = launch.tab;
  return (
    <Tab.Navigator
      initialRouteName={initial}
      tabBar={(props) => <HudTabBar {...props} />}
      screenOptions={{ headerShown: false }}
    >
      <Tab.Screen name="Won" component={WonScreen} />
      <Tab.Screen name="Today" component={TodayScreen} />
      <Tab.Screen name="Teams" component={TeamsNavigator} options={{ popToTopOnBlur: true }} />
      <Tab.Screen name="You" component={YouScreen} />
    </Tab.Navigator>
  );
}

/**
 * Holds the native splash until fonts and the stored user, deals, events and follows are ready (2.5s
 * at most), so the first tab is chosen from real follows.
 */
function Gate() {
  const [fontsLoaded, fontError] = useFonts({ [fonts.display]: PressStart2P_400Regular });
  const { userLoading, dealsLoading, subsLoading, eventsLoading } = useAppData();
  const { config } = useAppConfig();
  const [timedOut, setTimedOut] = useState(false);
  useEffect(() => {
    const t = setTimeout(() => setTimedOut(true), 2500);
    return () => clearTimeout(t);
  }, []);

  const ready = timedOut || ((fontsLoaded || !!fontError) && !userLoading && !dealsLoading && !subsLoading && !eventsLoading);
  // Once ready, the launch rays take over from the native still (docs/ui.md Motion); the still is
  // hidden when their art has drawn. Under Reduce Motion the app shows straight away.
  const reduced = useReducedMotion();
  const [launchDone, setLaunchDone] = useState(false);
  const hideSplash = useCallback(() => {
    SplashScreen.hideAsync().catch(() => {});
  }, []);
  const finishLaunch = useCallback(() => setLaunchDone(true), []);
  const launching = ready && !reduced && !launchDone;
  useEffect(() => {
    if (ready && !launching) hideSplash();
  }, [ready, launching, hideSplash]);

  if (!ready) return null;
  const rays = launching ? <LaunchRays onReady={hideSplash} onDone={finishLaunch} /> : null;
  if (config.features.maintenance_mode === true) {
    return (
      <>
        <MaintenanceScreen />
        {rays}
      </>
    );
  }

  return (
    <>
      <NavigationContainer ref={navigationRef} linking={linking} theme={navTheme}>
        <Stack.Navigator
          screenOptions={{ headerShown: false, contentStyle: { backgroundColor: colors.background } }}
        >
          <Stack.Screen name="Tabs" component={Tabs} />
          {/* No iOS slides anywhere (docs/ui.md Motion): deals and alerts open as a hard cut, like PAUSE. */}
          <Stack.Screen name="Deal" component={DealScreen} options={{ presentation: 'fullScreenModal', animation: 'none' }} />
          <Stack.Screen name="Alerts" component={AlertsScreen} options={{ presentation: 'fullScreenModal', animation: 'none' }} />
          <Stack.Screen
            name="Pause"
            component={PauseScreen}
            options={{
              presentation: 'transparentModal',
              animation: 'none',
              contentStyle: { backgroundColor: colors.background },
            }}
          />
        </Stack.Navigator>
        {/* After the launch rays, so a win's cut-in isn't hidden under them. */}
        {launching ? null : <VictoryCutInHost />}
      </NavigationContainer>
      {rays}
    </>
  );
}

/** A push that arrives while the app is open refreshes the deals (and TODAY) behind it. */
function ForegroundRefresh() {
  const { refreshAll } = useAppData();
  useEffect(() => {
    try {
      const sub = Notifications.addNotificationReceivedListener(() => {
        refreshAll();
      });
      return () => sub.remove();
    } catch {
      return undefined;
    }
  }, [refreshAll]);
  return null;
}

export default function App() {
  return (
    <SafeAreaProvider>
      <StatusBar style="light" />
      <AppConfigProvider>
        <ToastProvider>
          <UserProvider>
            <AppDataProvider>
              <ForegroundRefresh />
              <ScreenWipeProvider>
                <Gate />
              </ScreenWipeProvider>
            </AppDataProvider>
          </UserProvider>
        </ToastProvider>
      </AppConfigProvider>
    </SafeAreaProvider>
  );
}
