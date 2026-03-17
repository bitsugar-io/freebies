import React, { useMemo } from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { BlockProps } from './BlockRenderer';
import { useAppData } from '../../context/AppDataContext';
import { useTheme } from '../../hooks/useTheme';

export function SubscriptionListBlock({ config }: BlockProps) {
  const { theme } = useTheme();
  const { colors } = theme;
  const { events, isSubscribed, toggleSubscription } = useAppData();

  const subscribedEvents = useMemo(() => events.filter(e => isSubscribed(e.id)), [events, isSubscribed]);

  if (subscribedEvents.length === 0) return null;

  return (
    <View style={[styles.section, { backgroundColor: colors.surface }]}>
      <Text style={[styles.sectionTitle, { color: colors.text }]}>Your Subscriptions</Text>
      {subscribedEvents.map(event => (
        <View key={event.id} style={[styles.row, { borderBottomColor: colors.border }]}>
          <View style={[styles.teamBadge, { backgroundColor: event.teamColor || colors.accent }]}>
            <Text style={styles.teamBadgeText}>{event.teamId}</Text>
          </View>
          <View style={styles.info}>
            <Text style={[styles.name, { color: colors.text }]} numberOfLines={1}>{event.offerName}</Text>
            <Text style={[styles.team, { color: colors.textMuted }]}>{event.teamName} • {event.partnerName}</Text>
          </View>
          <TouchableOpacity style={[styles.removeButton, { backgroundColor: colors.surfaceSecondary }]} onPress={() => toggleSubscription(event.id)}>
            <Text style={[styles.removeText, { color: colors.warning }]}>Remove</Text>
          </TouchableOpacity>
        </View>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  section: { borderRadius: 12, padding: 16, marginBottom: 16 },
  sectionTitle: { fontSize: 18, fontWeight: '600', marginBottom: 16 },
  row: { flexDirection: 'row', alignItems: 'center', paddingVertical: 12, borderBottomWidth: StyleSheet.hairlineWidth },
  teamBadge: { width: 40, height: 40, borderRadius: 20, alignItems: 'center', justifyContent: 'center' },
  teamBadgeText: { color: '#fff', fontWeight: 'bold', fontSize: 12 },
  info: { flex: 1, marginLeft: 12, marginRight: 8 },
  name: { fontSize: 15, fontWeight: '500' },
  team: { fontSize: 12, marginTop: 2 },
  removeButton: { paddingHorizontal: 12, paddingVertical: 6, borderRadius: 12 },
  removeText: { fontSize: 13, fontWeight: '500' },
});
