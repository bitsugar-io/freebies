import React, { useCallback } from 'react';
import { StyleSheet, View } from 'react-native';
import { useFocusEffect, useNavigation } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import type { ActiveDeal, TodayGame } from '../api/client';
import { Screen } from '../components/Screen';
import { ErrorPanel, StaleBanner } from '../components/StatusPanels';
import { UnlockedDealRow, WatchRow } from '../components/DealRows';
import { PixelBox } from '../components/pixel/PixelBox';
import { SkeletonBox } from '../components/pixel/Skeleton';
import { timeLeftMs } from '../components/pixel/DealBar';
import { Text } from '../components/pixel/Text';
import { Blink } from '../hooks/useBlink';
import { colors } from '../hooks/useTheme';
import { useNow } from '../hooks/useNow';
import { useAppData } from '../context/AppDataContext';
import { useVictory } from '../hooks/useVictory';
import type { RootStackParamList } from '../navigation/types';

type Nav = NativeStackNavigationProp<RootStackParamList>;

/**
 * Today: every LA deal anyone can claim right now (not just your teams), then today's LA games with
 * a deal on the line. At-the-game deals are tagged; they're claimed at the venue that night, so
 * this is the only place they surface. All of it comes from GET /today.
 */
export function TodayScreen() {
  const nav = useNavigation<Nav>();
  const {
    today,
    todayLoading: isLoading,
    todayError: error,
    todaySavedAt: savedAt,
    refreshTodayIfStale,
    refreshAll: refresh,
  } = useAppData();
  // Refetch when the tab gains focus (at most once a minute); foreground, pushes, pull to refresh
  // and a new Pacific day refetch through the app data context.
  useFocusEffect(
    useCallback(() => {
      refreshTodayIfStale();
    }, [refreshTodayIfStale])
  );
  const now = useNow(60_000);
  const victory = useVictory();

  const live: ActiveDeal[] = (today?.live ?? [])
    .filter((d) => (timeLeftMs(d.expiresAt, now) ?? 1) > 0)
    .map((d) => ({ ...d, isDismissed: false }));
  const games = today?.games ?? [];
  const status = new Date()
    .toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })
    .replace(',', '');
  const failedEmpty = !!error && !today;
  /** A game's deal: the "if X, you get Y" preview, by event id. */
  const open = (eventId: string) => nav.navigate('Deal', { id: eventId });

  return (
    <Screen title="Today" status={victory ? `${victory.team} won!` : status} onRefresh={refresh}>
      {error && today && savedAt ? <StaleBanner savedAt={savedAt} onRetry={refresh} /> : null}
      {failedEmpty ? <ErrorPanel message={error} onRetry={refresh} /> : null}

      {isLoading && !today ? (
        <View style={{ gap: 12 }}>
          <SkeletonBox height={140} />
          <SkeletonBox height={140} />
        </View>
      ) : null}

      {today && live.length > 0 ? (
        <View style={styles.section}>
          <Text variant="displayM" color="info" accessibilityRole="header">
            Live now
          </Text>
          <Text variant="caption" color="textSecondary">
            Anyone can claim these, whoever you follow.
          </Text>
          {live.map((deal) => (
            <UnlockedDealRow key={deal.id} deal={deal} onPress={() => nav.navigate('Deal', { id: deal.id })} />
          ))}
        </View>
      ) : null}

      {today && games.length > 0 ? (
        <View style={styles.section}>
          <Text variant="displayM" color="info" accessibilityRole="header">
            Games today
          </Text>
          {games.map((g) => (
            <GameBox key={`${g.teamId}-${g.startsAt}`} game={g} onOpen={open} />
          ))}
        </View>
      ) : null}

      {today && live.length === 0 && games.length === 0 ? (
        <PixelBox variant="dialog" style={{ gap: 16 }}>
          <Text variant="displayL" color="accent" align="center" accessibilityRole="header">
            No games today
          </Text>
          <Text variant="body" color="textSecondary" align="center">
            On game days, every LA deal on the line shows up here, and anything you can claim right
            now.
          </Text>
        </PixelBox>
      ) : null}
    </Screen>
  );
}

/** One game: start time (or LIVE / FINAL), the matchup, then each deal riding on it. */
function GameBox({ game, onOpen }: { game: TodayGame; onOpen: (eventId: string) => void }) {
  const matchup = `${game.teamName} ${game.home ? 'vs' : 'at'} ${game.opponent}`;
  const when = startTime(game.startsAt);
  const spoken = game.state === 'in' ? 'Live now' : game.state === 'post' ? 'Final' : `Starts ${when}`;
  return (
    <PixelBox variant="panel">
      <View style={styles.gameTop} accessible accessibilityLabel={`${matchup}. ${spoken}`}>
        {game.state === 'in' ? (
          <Blink priority={4}>
            <View style={styles.liveTag}>
              <Text variant="displayM" color="text">
                Live
              </Text>
            </View>
          </Blink>
        ) : (
          <Text variant="displayM" color={game.state === 'post' ? 'textSecondary' : 'info'}>
            {game.state === 'post' ? 'Final' : when}
          </Text>
        )}
        <Text variant="bodyStrong" style={{ flex: 1 }}>
          {matchup}
        </Text>
      </View>
      {game.deals.map((e) => (
        <WatchRow key={e.id} event={e} onPress={() => onOpen(e.id)} />
      ))}
    </PixelBox>
  );
}

/** "7:30 PM" in Pacific time. */
function startTime(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  return d.toLocaleTimeString('en-US', { timeZone: 'America/Los_Angeles', hour: 'numeric', minute: '2-digit' });
}

const styles = StyleSheet.create({
  section: { gap: 12, marginBottom: 24 },
  gameTop: { flexDirection: 'row', flexWrap: 'wrap', alignItems: 'center', gap: 12, marginBottom: 4 },
  liveTag: { backgroundColor: colors.dangerDeep, paddingHorizontal: 6, paddingVertical: 2 },
});
