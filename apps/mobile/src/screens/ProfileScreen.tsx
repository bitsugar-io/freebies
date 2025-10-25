import React, { useMemo, useEffect, useState } from 'react';
import {
  StyleSheet,
  Text,
  View,
  ScrollView,
  TouchableOpacity,
} from 'react-native';
import { useAppData } from '../context/AppDataContext';
import { useTheme, ThemeMode } from '../hooks/useTheme';
import { sendRandomTestNotification } from '../hooks/usePushNotifications';
import { api } from '../api/client';

export function ProfileScreen() {
  const { theme, setThemeMode } = useTheme();
  const { colors, mode } = theme;
  const {
    user,
    subscribedCount,
    undismissedDeals,
    expoPushToken,
    events,
    isSubscribed,
    toggleSubscription,
  } = useAppData();

  const [dealsClaimed, setDealsClaimed] = useState(0);

  // Fetch user stats
  useEffect(() => {
    if (user?.id) {
      api.getUserStats(user.id)
        .then(stats => setDealsClaimed(stats.dealsClaimed))
        .catch(err => console.error('Failed to fetch user stats:', err));
    }
  }, [user?.id, undismissedDeals]); // Refetch when deals change

  // Get subscribed events
  const subscribedEvents = useMemo(() => {
    return events.filter(e => isSubscribed(e.id));
  }, [events, isSubscribed]);

  const themeOptions: { label: string; value: ThemeMode; icon: string }[] = [
    { label: 'Light', value: 'light', icon: '☀️' },
    { label: 'Dark', value: 'dark', icon: '🌙' },
    { label: 'System', value: 'system', icon: '⚙️' },
  ];

  return (
    <ScrollView
      style={[styles.container, { backgroundColor: colors.background }]}
      contentContainerStyle={styles.content}
    >
      {/* Stats Section */}
      <View style={[styles.section, { backgroundColor: colors.surface }]}>
        <Text style={[styles.sectionTitle, { color: colors.text }]}>Your Stats</Text>
        <View style={styles.statsGrid}>
          <View style={[styles.statCard, { backgroundColor: colors.surfaceSecondary }]}>
            <Text style={[styles.statNumber, { color: '#E91E63' }]}>
              {dealsClaimed}
            </Text>
            <Text style={[styles.statLabel, { color: colors.textMuted }]}>
              Claimed
            </Text>
          </View>
          <View style={[styles.statCard, { backgroundColor: colors.surfaceSecondary }]}>
            <Text style={[styles.statNumber, { color: colors.success }]}>
              {undismissedDeals.length}
            </Text>
            <Text style={[styles.statLabel, { color: colors.textMuted }]}>
              Active
            </Text>
          </View>
          <View style={[styles.statCard, { backgroundColor: colors.surfaceSecondary }]}>
            <Text style={[styles.statNumber, { color: colors.info }]}>
              {subscribedCount}
            </Text>
            <Text style={[styles.statLabel, { color: colors.textMuted }]}>
              Subscribed
            </Text>
          </View>
        </View>
      </View>

      {/* Subscriptions Section */}
      {subscribedEvents.length > 0 && (
        <View style={[styles.section, { backgroundColor: colors.surface }]}>
          <Text style={[styles.sectionTitle, { color: colors.text }]}>Your Subscriptions</Text>
          {subscribedEvents.map(event => (
            <View
              key={event.id}
              style={[styles.subscriptionRow, { borderBottomColor: colors.border }]}
            >
              <View style={[styles.teamBadge, { backgroundColor: event.teamColor || colors.accent }]}>
                <Text style={styles.teamBadgeText}>{event.teamId}</Text>
              </View>
              <View style={styles.subscriptionInfo}>
                <Text style={[styles.subscriptionName, { color: colors.text }]} numberOfLines={1}>
                  {event.offerName}
                </Text>
                <Text style={[styles.subscriptionTeam, { color: colors.textMuted }]}>
                  {event.teamName} • {event.partnerName}
                </Text>
              </View>
              <TouchableOpacity
                style={[styles.unsubscribeButton, { backgroundColor: colors.surfaceSecondary }]}
                onPress={() => toggleSubscription(event.id)}
              >
                <Text style={[styles.unsubscribeText, { color: colors.warning }]}>
                  Remove
                </Text>
              </TouchableOpacity>
            </View>
          ))}
        </View>
      )}

      {/* Appearance Section */}
      <View style={[styles.section, { backgroundColor: colors.surface }]}>
        <Text style={[styles.sectionTitle, { color: colors.text }]}>Appearance</Text>
        <View style={styles.themeOptions}>
          {themeOptions.map(option => (
            <TouchableOpacity
              key={option.value}
              style={[
                styles.themeOption,
                { backgroundColor: colors.surfaceSecondary },
                mode === option.value && { backgroundColor: colors.accent },
              ]}
              onPress={() => setThemeMode(option.value)}
            >
              <Text style={styles.themeIcon}>{option.icon}</Text>
              <Text style={[
                styles.themeLabel,
                { color: mode === option.value ? '#fff' : colors.text },
              ]}>
                {option.label}
              </Text>
            </TouchableOpacity>
          ))}
        </View>
      </View>

      {/* Notifications Section */}
      <View style={[styles.section, { backgroundColor: colors.surface }]}>
        <Text style={[styles.sectionTitle, { color: colors.text }]}>Notifications</Text>

        <View style={styles.row}>
          <View style={styles.rowText}>
            <Text style={[styles.rowTitle, { color: colors.text }]}>Push Notifications</Text>
            <Text style={[styles.rowSubtitle, { color: colors.textMuted }]}>
              {expoPushToken ? 'Enabled' : 'Not available'}
            </Text>
          </View>
          <View style={[
            styles.statusBadge,
            { backgroundColor: expoPushToken ? colors.successBackground : colors.warningBackground },
          ]}>
            <Text style={[
              styles.statusText,
              { color: expoPushToken ? colors.success : colors.warning },
            ]}>
              {expoPushToken ? 'ON' : 'OFF'}
            </Text>
          </View>
        </View>

        {__DEV__ && (
          <TouchableOpacity
            style={[styles.button, { backgroundColor: colors.surfaceSecondary }]}
            onPress={() => sendRandomTestNotification()}
          >
            <Text style={[styles.buttonText, { color: colors.text }]}>
              🔔 Send Test Notification
            </Text>
          </TouchableOpacity>
        )}
      </View>

      {/* Account Section */}
      <View style={[styles.section, { backgroundColor: colors.surface }]}>
        <Text style={[styles.sectionTitle, { color: colors.text }]}>Account</Text>

        <View style={styles.row}>
          <View style={styles.rowText}>
            <Text style={[styles.rowTitle, { color: colors.text }]}>User ID</Text>
            <Text style={[styles.rowSubtitle, { color: colors.textMuted }]} numberOfLines={1}>
              {user?.id || 'Not logged in'}
            </Text>
          </View>
        </View>

        {expoPushToken && (
          <View style={styles.row}>
            <View style={styles.rowText}>
              <Text style={[styles.rowTitle, { color: colors.text }]}>Push Token</Text>
              <Text style={[styles.rowSubtitle, { color: colors.textMuted }]} numberOfLines={1}>
                {expoPushToken.slice(0, 30)}...
              </Text>
            </View>
          </View>
        )}
      </View>

      {/* About Section */}
      <View style={[styles.section, { backgroundColor: colors.surface }]}>
        <Text style={[styles.sectionTitle, { color: colors.text }]}>About</Text>

        <View style={styles.row}>
          <Text style={[styles.rowTitle, { color: colors.text }]}>Version</Text>
          <Text style={[styles.rowValue, { color: colors.textMuted }]}>1.0.0</Text>
        </View>

        <View style={styles.row}>
          <Text style={[styles.rowTitle, { color: colors.text }]}>Build</Text>
          <Text style={[styles.rowValue, { color: colors.textMuted }]}>Development</Text>
        </View>
      </View>

      <View style={styles.footer}>
        <Text style={[styles.footerText, { color: colors.textMuted }]}>
          Freebie - Never miss a free offer
        </Text>
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  content: {
    padding: 16,
    paddingBottom: 100,
  },
  section: {
    borderRadius: 12,
    padding: 16,
    marginBottom: 16,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    marginBottom: 16,
  },
  statsGrid: {
    flexDirection: 'row',
    gap: 12,
  },
  statCard: {
    flex: 1,
    padding: 16,
    borderRadius: 12,
    alignItems: 'center',
  },
  statNumber: {
    fontSize: 32,
    fontWeight: 'bold',
  },
  statLabel: {
    fontSize: 12,
    marginTop: 4,
  },
  themeOptions: {
    flexDirection: 'row',
    gap: 8,
  },
  themeOption: {
    flex: 1,
    padding: 12,
    borderRadius: 12,
    alignItems: 'center',
  },
  themeIcon: {
    fontSize: 24,
    marginBottom: 4,
  },
  themeLabel: {
    fontSize: 12,
    fontWeight: '500',
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 12,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: 'rgba(0,0,0,0.1)',
  },
  rowText: {
    flex: 1,
    marginRight: 12,
  },
  rowTitle: {
    fontSize: 16,
  },
  rowSubtitle: {
    fontSize: 12,
    marginTop: 2,
  },
  rowValue: {
    fontSize: 16,
  },
  statusBadge: {
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  statusText: {
    fontSize: 12,
    fontWeight: '600',
  },
  button: {
    padding: 14,
    borderRadius: 12,
    alignItems: 'center',
    marginTop: 12,
  },
  buttonText: {
    fontSize: 16,
    fontWeight: '500',
  },
  footer: {
    alignItems: 'center',
    paddingVertical: 24,
  },
  footerText: {
    fontSize: 14,
  },
  subscriptionRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 12,
    borderBottomWidth: StyleSheet.hairlineWidth,
  },
  teamBadge: {
    width: 40,
    height: 40,
    borderRadius: 20,
    alignItems: 'center',
    justifyContent: 'center',
  },
  teamBadgeText: {
    color: '#fff',
    fontWeight: 'bold',
    fontSize: 12,
  },
  subscriptionInfo: {
    flex: 1,
    marginLeft: 12,
    marginRight: 8,
  },
  subscriptionName: {
    fontSize: 15,
    fontWeight: '500',
  },
  subscriptionTeam: {
    fontSize: 12,
    marginTop: 2,
  },
  unsubscribeButton: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 12,
  },
  unsubscribeText: {
    fontSize: 13,
    fontWeight: '500',
  },
});
