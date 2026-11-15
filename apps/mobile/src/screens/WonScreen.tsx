import React, { useMemo } from 'react';
import { View } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import type { BottomTabNavigationProp } from '@react-navigation/bottom-tabs';
import { useAppData } from '../context/AppDataContext';
import { Screen } from '../components/Screen';
import { ServerBlocks } from '../components/ServerBlocks';
import { ErrorPanel, StaleBanner } from '../components/StatusPanels';
import { UnlockedDealRow } from '../components/DealRows';
import { PalmLogo } from '../components/pixel/PalmLogo';
import { SkeletonBox } from '../components/pixel/Skeleton';
import { PixelBox } from '../components/pixel/PixelBox';
import { PixelButton } from '../components/pixel/PixelButton';
import { timeLeftMs } from '../components/pixel/DealBar';
import { Text } from '../components/pixel/Text';
import { useNow } from '../hooks/useNow';
import type { RootStackParamList, TabParamList } from '../navigation/types';

type Nav = BottomTabNavigationProp<TabParamList, 'Won'> & NativeStackNavigationProp<RootStackParamList>;

/**
 * Won, the home screen: only deals you have won and can still claim (unlocked, not expired, not
 * claimed). Followed deals that haven't unlocked live in Teams, never here.
 */
export function WonScreen() {
  const nav = useNavigation<Nav>();
  const { undismissedDeals, eventsLoading, dealsLoading, dealsError, dealsSavedAt, userError, refreshAll, retry } =
    useAppData();
  const now = useNow(60_000);

  const won = useMemo(
    () => undismissedDeals.filter((d) => (timeLeftMs(d.expiresAt, now) ?? 1) > 0),
    [undismissedDeals, now]
  );

  const today = new Date()
    .toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })
    .replace(',', '');
  const failedEmpty = (!!userError || !!dealsError) && !dealsSavedAt && undismissedDeals.length === 0;
  const loading = (eventsLoading || dealsLoading) && !failedEmpty;

  return (
    <Screen title="Freebies" status={today} left={<PalmLogo scale={1} />} onRefresh={refreshAll}>
      <ServerBlocks keys={['tonight', 'won']} />
      {dealsError && dealsSavedAt ? <StaleBanner savedAt={dealsSavedAt} onRetry={retry} /> : null}
      {failedEmpty ? <ErrorPanel message={userError ?? dealsError} onRetry={retry} /> : null}

      {loading ? (
        <View style={{ gap: 12 }}>
          <SkeletonBox height={140} />
          <SkeletonBox height={140} />
        </View>
      ) : null}

      {!loading && !failedEmpty && won.length > 0 ? (
        <View style={{ gap: 12 }}>
          {won.map((deal) => (
            <UnlockedDealRow key={deal.id} deal={deal} onPress={() => nav.navigate('Deal', { id: deal.id })} />
          ))}
        </View>
      ) : null}

      {!loading && !failedEmpty && won.length === 0 ? (
        <PixelBox variant="dialog" style={{ gap: 16 }}>
          <Text variant="displayL" color="accent" align="center" accessibilityRole="header">
            Nothing won yet
          </Text>
          <Text variant="body" color="textSecondary" align="center">
            When your teams pull it off, your free food shows up here.
          </Text>
          <PixelButton label="Pick teams" onPress={() => nav.navigate('Teams')} />
        </PixelBox>
      ) : null}
    </Screen>
  );
}
