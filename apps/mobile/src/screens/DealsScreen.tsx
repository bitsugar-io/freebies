import React, { useState, useEffect } from 'react';
import { StyleSheet, View } from 'react-native';
import { useAppData } from '../context/AppDataContext';
import { useTheme } from '../hooks/useTheme';
import { DealModal } from '../components/DealModal';
import { BlockRenderer } from '../components/blocks/BlockRenderer';
import { Event, ActiveDeal } from '../api/client';

export function DealsScreen() {
  const { theme } = useTheme();
  const { colors } = theme;
  const {
    dismissDeal,
    refreshAll,
    isSubscribed,
    toggleSubscription,
    setModalHandlers,
    undismissedDeals,
  } = useAppData();

  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [selectedDeal, setSelectedDeal] = useState<ActiveDeal | null>(null);
  const [modalVisible, setModalVisible] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  useEffect(() => {
    setModalHandlers({ setSelectedEvent, setSelectedDeal, setModalVisible });
  }, [setModalHandlers]);

  const onRefresh = async () => {
    setRefreshing(true);
    await refreshAll();
    setRefreshing(false);
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
      <BlockRenderer
        screenId="deals"
        refreshing={refreshing}
        onRefresh={onRefresh}
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
  container: { flex: 1 },
});
