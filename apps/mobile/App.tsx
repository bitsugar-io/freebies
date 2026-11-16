import React, { useCallback, useEffect, useState } from 'react';
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
import { TitleScreen } from './src/components/pixel/TitleScreen';
import { PauseScreen } from './src/screens/PauseScreen';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { ToastProvider } from './src/components/pixel/Toast';
import { UserProvider } from './src/hooks/useUser';
import { colors, fonts } from './src/hooks/useTheme';
import { urlFromNotification } from './src/hooks/usePushNotifications';
import { AppDataProvider, useAppData } from './src/context/AppDataContext';
import { AppConfigProvider, useAppConfig } from './src/context/AppConfigContext';
import { navigationRef } from './src/navigation/ref';
import { LINK_CONFIG } from './src/navigation/linkConfig';
import { attachLinkListener, deliverLink, firstSighting, initialLink } from './src/navigation/pendingLink';
import type { RootStackParamList, TabParamList } from './src/navigation/types';

SplashScreen.preventAutoHideAsync().catch(() => {});

const Tab = createBottomTabNavigator<TabParamList>();
const Stack = createNativeStackNavigator<RootStackParamList>();

const TITLE_SEEN_KEY = 'freebie_title_seen';

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
  return (
    <Tab.Navigator
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

/** Holds the native splash until fonts and the stored user and deals are ready (2.5s at most). */
function Gate() {
  const [fontsLoaded, fontError] = useFonts({ [fonts.display]: PressStart2P_400Regular });
  const { userLoading, dealsLoading } = useAppData();
  const { config } = useAppConfig();
  const [timedOut, setTimedOut] = useState(false);
  // First launch only: 'check' until the stored flag is read, then 'show' or 'done'.
  const [title, setTitle] = useState<'check' | 'show' | 'done'>('check');

  useEffect(() => {
    AsyncStorage.getItem(TITLE_SEEN_KEY)
      .then((v) => setTitle(v ? 'done' : 'show'))
      .catch(() => setTitle('done'));
  }, []);
  // Dev only: exp://.../--/title shows the title screen again.
  useEffect(() => {
    if (!__DEV__) return undefined;
    const sub = Linking.addEventListener('url', ({ url }) => {
      if (/\/title\/?$/.test(url)) setTitle('show');
    });
    return () => sub.remove();
  }, []);
  const finishTitle = useCallback(() => {
    AsyncStorage.setItem(TITLE_SEEN_KEY, '1').catch(() => {});
    setTitle('done');
  }, []);

  useEffect(() => {
    const t = setTimeout(() => setTimedOut(true), 2500);
    return () => clearTimeout(t);
  }, []);

  const ready = timedOut || ((fontsLoaded || !!fontError) && !userLoading && !dealsLoading && title !== 'check');
  useEffect(() => {
    if (ready) SplashScreen.hideAsync().catch(() => {});
  }, [ready]);

  if (!ready) return null;
  if (config.features.maintenance_mode === true) return <MaintenanceScreen />;
  if (title === 'show') return <TitleScreen onDone={finishTitle} />;

  return (
    <NavigationContainer ref={navigationRef} linking={linking} theme={navTheme}>
      <Stack.Navigator
        screenOptions={{ headerShown: false, contentStyle: { backgroundColor: colors.background } }}
      >
        <Stack.Screen name="Tabs" component={Tabs} />
        <Stack.Screen name="Deal" component={DealScreen} options={{ presentation: 'modal' }} />
        <Stack.Screen name="Alerts" component={AlertsScreen} options={{ presentation: 'modal' }} />
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
    </NavigationContainer>
  );
}

/** A push that arrives while the app is open refreshes the deals behind it. */
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
