import React, { useState, useEffect } from 'react';
import {
  StyleSheet,
  Text,
  View,
  FlatList,
  RefreshControl,
} from 'react-native';
import { useAppData } from '../context/AppDataContext';
import { useTheme } from '../hooks/useTheme';
import { ActiveDealCard } from '../components/ActiveDealCard';
import { DealModal } from '../components/DealModal';
import { Event, ActiveDeal } from '../api/client';

export function DealsScreen() {
  const { theme } = useTheme();
  const { colors } = theme;
  const {
    undismissedDeals,
    dismissDeal,
    dealsLoading,
    refreshAll,
    isSubscribed,
    toggleSubscription,
    setModalHandlers,
  } = useAppData();

  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [selectedDeal, setSelectedDeal] = useState<ActiveDeal | null>(null);
  const [modalVisible, setModalVisible] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  // Register modal handlers for deep linking
  useEffect(() => {
    setModalHandlers({ setSelectedEvent, setSelectedDeal, setModalVisible });
  }, [setModalHandlers]);

  const onRefresh = async () => {
    setRefreshing(true);
    await refreshAll();
    setRefreshing(false);
  };

  const handleDealPress = (deal: ActiveDeal) => {
    setSelectedEvent(deal.event);
    setSelectedDeal(deal);
    setModalVisible(true);
  };

  const handleDismiss = async (type: 'got_it' | 'stop_reminding') => {
    if (selectedDeal) {
      try {
        await dismissDeal(selectedDeal.id, type);
      } catch (err) {
        console.error('Failed to dismiss deal:', err);
      }
    }
  };

  return (
    <View style={[styles.container, { backgroundColor: colors.background }]}>
      <FlatList
        style={styles.list}
        data={undismissedDeals}
        renderItem={({ item }) => (
          <ActiveDealCard
            deal={item}
            onPress={() => handleDealPress(item)}
            onDismiss={(type) => dismissDeal(item.id, type)}
          />
        )}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.listContent}
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            tintColor={colors.accent}
          />
        }
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={[styles.emptyIcon]}>🎁</Text>
            <Text style={[styles.emptyTitle, { color: colors.text }]}>
              No Active Deals
            </Text>
            <Text style={[styles.emptyText, { color: colors.textMuted }]}>
              Subscribe to events in the Discover tab to get notified when deals drop!
            </Text>
          </View>
        }
      />

      <DealModal
        event={selectedEvent}
        visible={modalVisible}
        onClose={() => {
          setModalVisible(false);
          setSelectedDeal(null);
        }}
        isSubscribed={selectedEvent ? isSubscribed(selectedEvent.id) : false}
        onToggleSubscription={() => {
          if (selectedEvent) {
            toggleSubscription(selectedEvent.id);
          }
        }}
        activeDeal={selectedDeal}
        onDismiss={handleDismiss}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  list: {
    flex: 1,
  },
  listContent: {
    padding: 16,
    paddingBottom: 100,
  },
  emptyContainer: {
    padding: 32,
    alignItems: 'center',
    marginTop: 40,
  },
  emptyIcon: {
    fontSize: 48,
    marginBottom: 16,
  },
  emptyTitle: {
    fontSize: 20,
    fontWeight: '600',
    marginBottom: 8,
  },
  emptyText: {
    fontSize: 16,
    textAlign: 'center',
    lineHeight: 22,
  },
});
