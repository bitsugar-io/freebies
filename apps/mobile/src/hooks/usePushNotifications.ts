import { useState, useEffect, useCallback } from 'react';
import { AppState, Platform } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import * as Device from 'expo-device';
import * as Notifications from 'expo-notifications';
import Constants from 'expo-constants';
import { api } from '../api/client';

// Show notifications while the app is in the foreground too.
Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: true,
    shouldSetBadge: true,
    shouldShowBanner: true,
    shouldShowList: true,
  }),
});

const PRIMED_KEY = 'freebie_alerts_primed';

export type AlertsStatus = 'granted' | 'denied' | 'undetermined';

async function readStatus(): Promise<AlertsStatus> {
  try {
    const { status, canAskAgain } = await Notifications.getPermissionsAsync();
    if (status === 'granted') return 'granted';
    // Never asked (or Android before its first ask) can still prompt; otherwise it's a hard no.
    return canAskAgain ? 'undetermined' : 'denied';
  } catch {
    return 'undetermined';
  }
}

async function fetchExpoToken(): Promise<string | null> {
  // iOS and Android need a physical device. Web can still get a token for testing.
  if (!Device.isDevice && Platform.OS !== 'web') {
    console.log('Push tokens need a physical device on iOS/Android');
    return null;
  }
  if (Platform.OS === 'android') {
    await Notifications.setNotificationChannelAsync('default', {
      name: 'default',
      importance: Notifications.AndroidImportance.MAX,
      vibrationPattern: [0, 250, 250, 250],
    });
  }
  const tokenData = await Notifications.getExpoPushTokenAsync({
    projectId: Constants.expoConfig?.extra?.eas?.projectId,
  });
  return tokenData.data;
}

/**
 * Deal alerts state and actions. Nothing here prompts on mount: the system dialog only appears
 * from `enableAlerts()`, which the app calls after the user follows something and taps TURN ON.
 * When permission was already granted, the token is re-registered silently.
 */
export function usePushAlerts(userId: string | null) {
  const [status, setStatus] = useState<AlertsStatus>('undetermined');
  const [expoPushToken, setExpoPushToken] = useState<string | null>(null);

  const register = useCallback(async () => {
    try {
      const token = await fetchExpoToken();
      if (token) {
        setExpoPushToken(token);
        if (userId) await api.updatePushToken(userId, token);
      }
    } catch (err) {
      console.log('Push registration error:', err instanceof Error ? err.message : err);
    }
  }, [userId]);

  const refreshStatus = useCallback(async () => {
    const next = await readStatus();
    setStatus(next);
    return next;
  }, []);

  // Check status on mount and whenever the app returns from system Settings.
  useEffect(() => {
    refreshStatus();
    const sub = AppState.addEventListener('change', (s) => s === 'active' && refreshStatus());
    return () => sub.remove();
  }, [refreshStatus]);

  // Already granted: keep the server's token fresh without asking anything.
  useEffect(() => {
    if (userId && status === 'granted') register();
  }, [userId, status, register]);

  /** Ask the system for permission (one-shot dialog) and register the token on a grant. */
  const enableAlerts = useCallback(async (): Promise<AlertsStatus> => {
    let next = await readStatus();
    if (next === 'undetermined') {
      try {
        await Notifications.requestPermissionsAsync();
      } catch (err) {
        console.log('Permission request failed:', err);
      }
      next = await readStatus();
    }
    setStatus(next);
    if (next === 'granted') await register();
    return next;
  }, [register]);

  return { alertsStatus: status, expoPushToken, enableAlerts, refreshAlertsStatus: refreshStatus };
}

/** Whether to show the explainer: permission never asked and we haven't shown it before. */
export async function shouldPrimeAlerts(): Promise<boolean> {
  if ((await readStatus()) !== 'undetermined') return false;
  return (await AsyncStorage.getItem(PRIMED_KEY).catch(() => null)) === null;
}

export function markAlertsPrimed(): void {
  AsyncStorage.setItem(PRIMED_KEY, '1').catch(() => {});
}

/**
 * Where a notification tap should go. Pushes carry `data.url` ("freebie://deal/<id>"), plus
 * `triggeredEventId` and `eventId`. Older pushes only have the ids, so handle every shape.
 */
export function urlFromNotification(notification: Notifications.Notification | null | undefined): string | null {
  const data = notification?.request?.content?.data as Record<string, unknown> | undefined;
  if (!data) return null;
  if (typeof data.url === 'string' && data.url) return data.url;
  const id = data.triggeredEventId ?? data.eventId;
  return typeof id === 'string' && id ? `freebie://deal/${encodeURIComponent(id)}` : null;
}

// Utility to send a local notification (for testing)
export async function sendLocalNotification(title: string, body: string, data?: Record<string, unknown>) {
  await Notifications.scheduleNotificationAsync({
    content: { title, body, sound: true, data },
    trigger: null, // Immediately
  });
}

// Store for active deals to use in test notifications
let cachedActiveDeals: Array<{
  id: string;
  eventId: string;
  expiresAt?: string;
  event: {
    offerName: string;
    teamName: string;
    triggerCondition: string;
  };
}> = [];

// Call this from useActiveDeals to keep test notifications in sync
export function setActiveDealsForNotifications(deals: typeof cachedActiveDeals) {
  cachedActiveDeals = deals;
}

/** Dev helper: fire a local notification shaped like the real ones. */
export async function sendRandomTestNotification(): Promise<{ title: string; body: string }> {
  const deal = cachedActiveDeals[Math.floor(Math.random() * cachedActiveDeals.length)];
  const notification = deal
    ? {
        title: `${deal.event.triggerCondition}! ${deal.event.offerName}`,
        body: `Claim at ${deal.event.teamName} partner`,
        data: { url: `freebie://deal/${deal.id}`, triggeredEventId: deal.id, eventId: deal.eventId },
      }
    : { title: 'No active deals', body: 'Run "task api:deals" to create test deals, then try again.', data: undefined };

  const { status } = await Notifications.getPermissionsAsync();
  if (status !== 'granted') {
    const { status: next } = await Notifications.requestPermissionsAsync();
    if (next !== 'granted') return notification;
  }
  await sendLocalNotification(notification.title, notification.body, notification.data);
  return notification;
}
