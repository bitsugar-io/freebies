import React, { useCallback, useEffect, useState } from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Linking, Platform, AppState } from 'react-native';
import * as Notifications from 'expo-notifications';
import { BlockProps } from './BlockRenderer';
import { useAppData } from '../../context/AppDataContext';
import { useTheme, ThemeMode } from '../../hooks/useTheme';
import { sendRandomTestNotification } from '../../hooks/usePushNotifications';

type PermissionStatus = 'granted' | 'denied' | 'undetermined' | 'loading';

export function SettingsBlock({ config }: BlockProps) {
  const { theme, setThemeMode } = useTheme();
  const { colors, mode } = theme;
  const { user, expoPushToken } = useAppData();
  const [permissionStatus, setPermissionStatus] = useState<PermissionStatus>('loading');

  const checkPermission = useCallback(() => {
    Notifications.getPermissionsAsync().then(({ status }) => {
      setPermissionStatus(status);
    });
  }, []);

  useEffect(() => {
    checkPermission();
    const sub = AppState.addEventListener('change', (state) => {
      if (state === 'active') checkPermission();
    });
    return () => sub.remove();
  }, [checkPermission]);

  const showThemeToggle = config.showThemeToggle !== false;

  const themeOptions: { label: string; value: ThemeMode; icon: string }[] = [
    { label: 'Light', value: 'light', icon: '☀️' },
    { label: 'Dark', value: 'dark', icon: '🌙' },
    { label: 'System', value: 'system', icon: '⚙️' },
  ];

  return (
    <>
      {showThemeToggle && (
        <View style={[styles.section, { backgroundColor: colors.surface }]}>
          <Text style={[styles.sectionTitle, { color: colors.text }]}>Appearance</Text>
          <View style={styles.themeOptions}>
            {themeOptions.map(option => (
              <TouchableOpacity
                key={option.value}
                style={[styles.themeOption, { backgroundColor: colors.surfaceSecondary }, mode === option.value && { backgroundColor: colors.accent }]}
                onPress={() => setThemeMode(option.value)}
              >
                <Text style={styles.themeIcon}>{option.icon}</Text>
                <Text style={[styles.themeLabel, { color: mode === option.value ? '#fff' : colors.text }]}>{option.label}</Text>
              </TouchableOpacity>
            ))}
          </View>
        </View>
      )}

      <View style={[styles.section, { backgroundColor: colors.surface }]}>
        <Text style={[styles.sectionTitle, { color: colors.text }]}>Notifications</Text>
        <TouchableOpacity
          style={styles.row}
          disabled={permissionStatus === 'granted' || permissionStatus === 'loading'}
          onPress={() => {
            if (permissionStatus === 'denied') {
              Linking.openSettings();
            } else if (permissionStatus === 'undetermined') {
              Notifications.requestPermissionsAsync().then(({ status }) => {
                setPermissionStatus(status);
              });
            }
          }}
        >
          <View style={styles.rowText}>
            <Text style={[styles.rowTitle, { color: colors.text }]}>Push Notifications</Text>
            <Text style={[styles.rowSubtitle, { color: colors.textMuted }]}>
              {permissionStatus === 'loading' ? 'Checking...' :
               permissionStatus === 'granted' ? 'Enabled' :
               permissionStatus === 'denied' ? 'Disabled in Settings — tap to open' :
               'Tap to enable'}
            </Text>
          </View>
          <View style={[styles.statusBadge, { backgroundColor: permissionStatus === 'granted' ? colors.successBackground : colors.warningBackground }]}>
            <Text style={[styles.statusText, { color: permissionStatus === 'granted' ? colors.success : colors.warning }]}>
              {permissionStatus === 'granted' ? 'ON' : 'OFF'}
            </Text>
          </View>
        </TouchableOpacity>
        {__DEV__ && (
          <TouchableOpacity style={[styles.button, { backgroundColor: colors.surfaceSecondary }]} onPress={() => sendRandomTestNotification()}>
            <Text style={[styles.buttonText, { color: colors.text }]}>🔔 Send Test Notification</Text>
          </TouchableOpacity>
        )}
      </View>

      <View style={[styles.section, { backgroundColor: colors.surface }]}>
        <Text style={[styles.sectionTitle, { color: colors.text }]}>Account</Text>
        <View style={styles.row}>
          <View style={styles.rowText}>
            <Text style={[styles.rowTitle, { color: colors.text }]}>User ID</Text>
            <Text style={[styles.rowSubtitle, { color: colors.textMuted }]} numberOfLines={1}>{user?.id || 'Not logged in'}</Text>
          </View>
        </View>
        {__DEV__ && expoPushToken && (
          <View style={styles.row}>
            <View style={styles.rowText}>
              <Text style={[styles.rowTitle, { color: colors.text }]}>Push Token</Text>
              <Text style={[styles.rowSubtitle, { color: colors.textMuted }]} numberOfLines={1}>{expoPushToken.slice(0, 30)}...</Text>
            </View>
          </View>
        )}
      </View>

      <View style={[styles.section, { backgroundColor: colors.surface }]}>
        <Text style={[styles.sectionTitle, { color: colors.text }]}>About</Text>
        <View style={styles.row}>
          <Text style={[styles.rowTitle, { color: colors.text }]}>Version</Text>
          <Text style={[styles.rowValue, { color: colors.textMuted }]}>1.0.0</Text>
        </View>
      </View>

      <View style={styles.footer}>
        <Text style={[styles.footerText, { color: colors.textMuted }]}>Freebies - Never miss a free offer</Text>
      </View>
    </>
  );
}

const styles = StyleSheet.create({
  section: { borderRadius: 12, padding: 16, marginBottom: 16 },
  sectionTitle: { fontSize: 18, fontWeight: '600', marginBottom: 16 },
  themeOptions: { flexDirection: 'row', gap: 8 },
  themeOption: { flex: 1, padding: 12, borderRadius: 12, alignItems: 'center' },
  themeIcon: { fontSize: 24, marginBottom: 4 },
  themeLabel: { fontSize: 12, fontWeight: '500' },
  row: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', paddingVertical: 12, borderBottomWidth: StyleSheet.hairlineWidth, borderBottomColor: 'rgba(0,0,0,0.1)' },
  rowText: { flex: 1, marginRight: 12 },
  rowTitle: { fontSize: 16 },
  rowSubtitle: { fontSize: 12, marginTop: 2 },
  rowValue: { fontSize: 16 },
  statusBadge: { paddingHorizontal: 10, paddingVertical: 4, borderRadius: 12 },
  statusText: { fontSize: 12, fontWeight: '600' },
  button: { padding: 14, borderRadius: 12, alignItems: 'center', marginTop: 12 },
  buttonText: { fontSize: 16, fontWeight: '500' },
  footer: { alignItems: 'center', paddingVertical: 24 },
  footerText: { fontSize: 14 },
});
