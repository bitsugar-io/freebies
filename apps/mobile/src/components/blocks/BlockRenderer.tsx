import React from 'react';
import { ScrollView, RefreshControl } from 'react-native';
import { ScreenBlock } from '../../api/client';
import { useAppConfig } from '../../context/AppConfigContext';
import { BannerBlock } from './BannerBlock';
import { ActiveDealsBlock } from './ActiveDealsBlock';
import { LeagueFilterBlock } from './LeagueFilterBlock';
import { EventListBlock } from './EventListBlock';
import { PromoCardBlock } from './PromoCardBlock';
import { UserStatsBlock } from './UserStatsBlock';
import { SubscriptionListBlock } from './SubscriptionListBlock';
import { SettingsBlock } from './SettingsBlock';

export interface BlockProps {
  config: Record<string, any>;
  screenProps?: Record<string, any>;
}

const BLOCK_REGISTRY: Record<string, React.ComponentType<BlockProps>> = {
  banner: BannerBlock,
  active_deals: ActiveDealsBlock,
  league_filter: LeagueFilterBlock,
  event_list: EventListBlock,
  promo_card: PromoCardBlock,
  user_stats: UserStatsBlock,
  subscription_list: SubscriptionListBlock,
  settings: SettingsBlock,
};

interface BlockRendererProps {
  screenId: string;
  refreshing?: boolean;
  onRefresh?: () => void;
  screenProps?: Record<string, any>;
  scrollEnabled?: boolean;
}

export function BlockRenderer({ screenId, refreshing, onRefresh, screenProps, scrollEnabled = true }: BlockRendererProps) {
  const { config } = useAppConfig();
  const blocks = config?.screens[screenId] ?? [];

  return (
    <ScrollView
      scrollEnabled={scrollEnabled}
      refreshControl={
        onRefresh ? (
          <RefreshControl refreshing={refreshing ?? false} onRefresh={onRefresh} />
        ) : undefined
      }
    >
      {blocks.map((block) => {
        const Component = BLOCK_REGISTRY[block.type];
        if (!Component) return null;
        return <Component key={block.key} config={block.config} screenProps={screenProps} />;
      })}
    </ScrollView>
  );
}
