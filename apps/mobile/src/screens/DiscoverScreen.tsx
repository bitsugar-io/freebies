import React, { useState } from 'react';
import { StyleSheet, View, ScrollView, RefreshControl } from 'react-native';
import { useAppData } from '../context/AppDataContext';
import { useTheme } from '../hooks/useTheme';
import { useAppConfig } from '../context/AppConfigContext';
import { DealModal } from '../components/DealModal';
import { LeagueFilterBlock } from '../components/blocks/LeagueFilterBlock';
import { EventListBlock } from '../components/blocks/EventListBlock';
import { PromoCardBlock } from '../components/blocks/PromoCardBlock';
import { BannerBlock } from '../components/blocks/BannerBlock';
import { Event, ActiveDeal } from '../api/client';

export function DiscoverScreen() {
  const { theme } = useTheme();
  const { colors } = theme;
  const { refreshAll, isSubscribed, toggleSubscription, undismissedDeals } = useAppData();
  const { config } = useAppConfig();

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

  // Render blocks from config, passing shared state to blocks that need it
  const blocks = config?.screens['discover'] ?? [];

  const renderBlock = (block: { type: string; key: string; config: Record<string, any> }) => {
    switch (block.type) {
      case 'league_filter':
        return (
          <LeagueFilterBlock
            key={block.key}
            config={block.config}
            selectedLeague={selectedLeague}
            onSelectLeague={setSelectedLeague}
          />
        );
      case 'event_list':
        return (
          <EventListBlock
            key={block.key}
            config={block.config}
            selectedLeague={selectedLeague}
            onEventPress={handleEventPress}
          />
        );
      case 'promo_card':
        return <PromoCardBlock key={block.key} config={block.config} />;
      case 'banner':
        return <BannerBlock key={block.key} config={block.config} />;
      default:
        return null;
    }
  };

  return (
    <View style={[styles.container, { backgroundColor: colors.background }]}>
      <ScrollView
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} tintColor={colors.accent} />
        }
      >
        {blocks.map(renderBlock)}
      </ScrollView>

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
