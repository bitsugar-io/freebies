import React from 'react';
import { View, Text, FlatList, StyleSheet } from 'react-native';
import { BlockProps } from './BlockRenderer';
import { useAppData } from '../../context/AppDataContext';
import { useTheme } from '../../hooks/useTheme';
import { ActiveDealCard } from '../ActiveDealCard';

export function ActiveDealsBlock({ config }: BlockProps) {
  const { theme } = useTheme();
  const { colors } = theme;
  const { undismissedDeals, dismissDeal } = useAppData();

  const emptyTitle = (config.emptyTitle as string) ?? 'No Active Deals';
  const emptySubtitle = (config.emptySubtitle as string) ??
    'Subscribe to events in the Discover tab to get notified when deals drop!';

  return (
    <FlatList
      scrollEnabled={false}
      data={undismissedDeals}
      renderItem={({ item }) => (
        <ActiveDealCard
          deal={item}
          onPress={() => {}}
          onDismiss={(type) => dismissDeal(item.id, type)}
        />
      )}
      keyExtractor={(item) => item.id}
      contentContainerStyle={styles.listContent}
      ListEmptyComponent={
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyIcon}>🎁</Text>
          <Text style={[styles.emptyTitle, { color: colors.text }]}>{emptyTitle}</Text>
          <Text style={[styles.emptyText, { color: colors.textMuted }]}>{emptySubtitle}</Text>
        </View>
      }
    />
  );
}

const styles = StyleSheet.create({
  listContent: { padding: 16, paddingBottom: 100 },
  emptyContainer: { padding: 32, alignItems: 'center', marginTop: 40 },
  emptyIcon: { fontSize: 48, marginBottom: 16 },
  emptyTitle: { fontSize: 20, fontWeight: '600', marginBottom: 8 },
  emptyText: { fontSize: 16, textAlign: 'center', lineHeight: 22 },
});
