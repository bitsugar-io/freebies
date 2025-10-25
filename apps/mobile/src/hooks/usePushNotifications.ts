import { useState, useEffect, useRef } from 'react';
import { Platform } from 'react-native';
import * as Device from 'expo-device';
import * as Notifications from 'expo-notifications';
import Constants from 'expo-constants';
import { api } from '../api/client';

// Configure how notifications are handled when app is foregrounded
Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: true,
    shouldSetBadge: true,
    shouldShowBanner: true,
    shouldShowList: true,
  }),
});

export interface PushNotificationState {
  expoPushToken: string | null;
  notification: Notifications.Notification | null;
  error: string | null;
}

export function usePushNotifications(
  userId: string | null,
  onNotificationTap?: (eventId: string) => void,
  onNotificationReceived?: () => void
) {
  const [expoPushToken, setExpoPushToken] = useState<string | null>(null);
  const [notification, setNotification] = useState<Notifications.Notification | null>(null);
  const [error, setError] = useState<string | null>(null);

  const notificationListener = useRef<Notifications.Subscription | null>(null);
  const responseListener = useRef<Notifications.Subscription | null>(null);

  useEffect(() => {
    // Register for push notifications
    registerForPushNotificationsAsync()
      .then((token) => {
        if (token) {
          setExpoPushToken(token);
          // Send token to API if we have a user
          if (userId) {
            api.updatePushToken(userId, token).catch(console.error);
          }
        }
      })
      .catch((err) => {
        console.log('Push notification setup error:', err.message);
        setError(err.message);
      });

    // Listen for incoming notifications (only if supported)
    try {
      notificationListener.current = Notifications.addNotificationReceivedListener(
        (notification) => {
          setNotification(notification);
          // Refresh data when notification received in foreground
          if (onNotificationReceived) {
            onNotificationReceived();
          }
        }
      );

      // Listen for notification taps - this is the deep link handler
      responseListener.current = Notifications.addNotificationResponseReceivedListener(
        (response) => {
          console.log('Notification tapped:', response);
          const data = response.notification.request.content.data;
          if (data?.eventId && onNotificationTap) {
            onNotificationTap(data.eventId as string);
          }
        }
      );
    } catch (e) {
      console.log('Notification listeners not supported in Expo Go');
    }

    return () => {
      // Clean up subscriptions safely
      if (notificationListener.current?.remove) {
        notificationListener.current.remove();
      }
      if (responseListener.current?.remove) {
        responseListener.current.remove();
      }
    };
  }, [userId, onNotificationTap, onNotificationReceived]);

  // Update token on server when userId becomes available
  useEffect(() => {
    if (userId && expoPushToken) {
      api.updatePushToken(userId, expoPushToken).catch(console.error);
    }
  }, [userId, expoPushToken]);

  return {
    expoPushToken,
    notification,
    error,
  };
}

async function registerForPushNotificationsAsync(): Promise<string | null> {
  let token: string | null = null;

  // Set up Android notification channel
  if (Platform.OS === 'android') {
    await Notifications.setNotificationChannelAsync('default', {
      name: 'default',
      importance: Notifications.AndroidImportance.MAX,
      vibrationPattern: [0, 250, 250, 250],
      lightColor: '#FF231F7C',
    });
  }

  // Check if running on a device (required for push notifications on iOS)
  if (!Device.isDevice) {
    // For web or simulator, we can still get a token for testing
    if (Platform.OS === 'web') {
      // Web push - will prompt for permission
      const { status: existingStatus } = await Notifications.getPermissionsAsync();
      let finalStatus = existingStatus;

      if (existingStatus !== 'granted') {
        const { status } = await Notifications.requestPermissionsAsync();
        finalStatus = status;
      }

      if (finalStatus !== 'granted') {
        console.log('Push notification permission denied');
        return null;
      }

      // Get Expo push token for web
      const tokenData = await Notifications.getExpoPushTokenAsync({
        projectId: Constants.expoConfig?.extra?.eas?.projectId,
      });
      token = tokenData.data;
    } else {
      console.log('Must use physical device for push notifications on iOS/Android');
      return null;
    }
  } else {
    // Physical device
    const { status: existingStatus } = await Notifications.getPermissionsAsync();
    let finalStatus = existingStatus;

    if (existingStatus !== 'granted') {
      const { status } = await Notifications.requestPermissionsAsync();
      finalStatus = status;
    }

    if (finalStatus !== 'granted') {
      console.log('Push notification permission denied');
      return null;
    }

    // Get Expo push token
    const tokenData = await Notifications.getExpoPushTokenAsync({
      projectId: Constants.expoConfig?.extra?.eas?.projectId,
    });
    token = tokenData.data;
  }

  return token;
}

// Utility to send a local notification (for testing)
export async function sendLocalNotification(title: string, body: string, data?: Record<string, unknown>) {
  await Notifications.scheduleNotificationAsync({
    content: {
      title,
      body,
      sound: true,
      data,
    },
    trigger: null, // Immediately
  });
}

// Test notification types
interface TestNotification {
  title: string;
  body: string;
  eventId?: string;
  triggeredEventId?: string;
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

function formatTimeRemaining(expiresAt: string): string {
  const expires = new Date(expiresAt);
  const now = new Date();
  const diff = expires.getTime() - now.getTime();

  if (diff < 0) return 'Expired!';

  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

  if (hours > 0) return `${hours}h ${minutes}m left`;
  return `${minutes}m left`;
}

export async function sendRandomTestNotification(): Promise<TestNotification> {
  let notification: TestNotification;

  // Only send test notifications if there are active deals
  if (cachedActiveDeals.length > 0) {
    const deal = cachedActiveDeals[Math.floor(Math.random() * cachedActiveDeals.length)];
    const timeRemaining = deal.expiresAt ? formatTimeRemaining(deal.expiresAt) : '';
    const isExpiringSoon = deal.expiresAt && new Date(deal.expiresAt).getTime() - Date.now() < 6 * 60 * 60 * 1000;

    if (isExpiringSoon) {
      // Reminder notification
      notification = {
        title: `⏰ Expiring Soon: ${deal.event.offerName}`,
        body: `Only ${timeRemaining}! Don't miss your free deal from ${deal.event.teamName}.`,
        eventId: deal.eventId,
        triggeredEventId: deal.id,
      };
    } else {
      // New deal notification
      notification = {
        title: `🎉 ${deal.event.offerName}!`,
        body: `${deal.event.triggerCondition} just triggered! ${timeRemaining ? `Expires in ${timeRemaining}.` : ''}`,
        eventId: deal.eventId,
        triggeredEventId: deal.id,
      };
    }
  } else {
    // No active deals - show info notification
    notification = {
      title: '📱 No Active Deals',
      body: 'Run "task backend:deals" to create test deals, then try again.',
    };
  }

  try {
    // Check/request permissions first
    const { status } = await Notifications.getPermissionsAsync();
    if (status !== 'granted') {
      const { status: newStatus } = await Notifications.requestPermissionsAsync();
      if (newStatus !== 'granted') {
        alert('Please enable notifications in settings');
        return notification;
      }
    }

    await sendLocalNotification(notification.title, notification.body, {
      type: 'freebie',
      eventId: notification.eventId,
      triggeredEventId: notification.triggeredEventId,
    });
    console.log('Notification sent:', notification.title, 'eventId:', notification.eventId);
  } catch (error) {
    console.error('Failed to send notification:', error);
    // Fallback to alert for Expo Go
    alert(`${notification.title}\n\n${notification.body}`);
  }

  return notification;
}
