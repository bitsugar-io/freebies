import React, { useState, useMemo } from 'react';
import { StyleSheet, View } from 'react-native';
import { useAppData } from '../context/AppDataContext';
import { useTheme } from '../hooks/useTheme';
import { DealModal } from '../components/DealModal';
import { BlockRenderer } from '../components/blocks/BlockRenderer';
import { Event, ActiveDeal } from '../api/client';

export function DiscoverScreen() {
  const { theme } = useTheme();
  const { colors } = theme;
  const { refreshAll, isSubscribed, toggleSubscription, undismissedDeals } = useAppData();

  const [selectedLeague, setSelectedLeague] = useState('all');
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [selectedDeal, setSelectedDeal] = useState<ActiveDeal | null>(null);
  const [modalVisible, setModalVisible] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  const onRefresh = async () => {
    setRefreshing(true);
    await refreshAll();
    setRefreshing(false);
  };

  const handleEventPress = (event: Event) => {
    const deal = undismissedDeals.find(d => d.eventId === event.id);
    setSelectedEvent(event);
    setSelectedDeal(deal || null);
    setModalVisible(true);
  };

  const screenProps = useMemo(() => ({
    selectedLeague,
    onSelectLeague: setSelectedLeague,
    onEventPress: handleEventPress,
  }), [selectedLeague, undismissedDeals]);

  return (
    <View style={[styles.container, { backgroundColor: colors.background }]}>
      <BlockRenderer
        screenId="discover"
        refreshing={refreshing}
        onRefresh={onRefresh}
        screenProps={screenProps}
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
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1 },
});
