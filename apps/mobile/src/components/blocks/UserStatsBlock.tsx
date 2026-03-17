import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { BlockProps } from './BlockRenderer';
import { useAppData } from '../../context/AppDataContext';
import { useTheme } from '../../hooks/useTheme';
import { api } from '../../api/client';

export function UserStatsBlock({ config }: BlockProps) {
  const { theme } = useTheme();
  const { colors } = theme;
  const { user, undismissedDeals, subscribedCount } = useAppData();
  const [dealsClaimed, setDealsClaimed] = useState(0);

  useEffect(() => {
    if (user?.id) {
      api.getUserStats(user.id)
        .then(stats => setDealsClaimed(stats.dealsClaimed))
        .catch(err => console.error('Failed to fetch user stats:', err));
    }
  }, [user?.id, undismissedDeals]);

  return (
    <View style={[styles.section, { backgroundColor: colors.surface }]}>
      <Text style={[styles.sectionTitle, { color: colors.text }]}>Your Stats</Text>
      <View style={styles.statsGrid}>
        <View style={[styles.statCard, { backgroundColor: colors.surfaceSecondary }]}>
          <Text style={[styles.statNumber, { color: '#E91E63' }]}>{dealsClaimed}</Text>
          <Text style={[styles.statLabel, { color: colors.textMuted }]}>Claimed</Text>
        </View>
        <View style={[styles.statCard, { backgroundColor: colors.surfaceSecondary }]}>
          <Text style={[styles.statNumber, { color: colors.success }]}>{undismissedDeals.length}</Text>
          <Text style={[styles.statLabel, { color: colors.textMuted }]}>Active</Text>
        </View>
        <View style={[styles.statCard, { backgroundColor: colors.surfaceSecondary }]}>
          <Text style={[styles.statNumber, { color: colors.info }]}>{subscribedCount}</Text>
          <Text style={[styles.statLabel, { color: colors.textMuted }]}>Subscribed</Text>
        </View>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  section: { borderRadius: 12, padding: 16, marginBottom: 16 },
  sectionTitle: { fontSize: 18, fontWeight: '600', marginBottom: 16 },
  statsGrid: { flexDirection: 'row', gap: 12 },
  statCard: { flex: 1, padding: 16, borderRadius: 12, alignItems: 'center' },
  statNumber: { fontSize: 32, fontWeight: 'bold' },
  statLabel: { fontSize: 12, marginTop: 4 },
});
