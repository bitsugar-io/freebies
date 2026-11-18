import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Linking, View } from 'react-native';
import Constants from 'expo-constants';
import { Sprite } from '../api/client';
import { useAppData } from '../context/AppDataContext';
import { Screen, Scroller } from '../components/Screen';
import { ServerBlocks } from '../components/ServerBlocks';
import { PixelBox } from '../components/pixel/PixelBox';
import { MenuRow, PixelMenu } from '../components/pixel/PixelMenu';
import { SportBadge } from '../components/pixel/SportBadge';
import { Text } from '../components/pixel/Text';
import { sendRandomTestNotification } from '../hooks/usePushNotifications';
import { RouteProp, useNavigation, useRoute } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { WatchRow } from '../components/DealRows';
import type { RootStackParamList, TabParamList } from '../navigation/types';
import { LICENSES } from '../licenses';
import { pauseUntilLabel, shortDatePacific } from '../utils/format';
import { useScreenWipe } from '../components/pixel/ScreenWipe';

/** Claimed deals shown before SHOW ALL. */
const TROPHY_PREVIEW = 10;

const DISCLAIMER =
  "Freebies: LA Sports Deals is not affiliated with, endorsed by or sponsored by any team, league, school or restaurant. Team and restaurant names are used only to describe the deals they offer. Deals are set by each restaurant and can change or end without notice. Check the restaurant's terms before you order.";

/** You: season card, alert settings, and about. Dark only, so there is no appearance setting. */
export function YouScreen() {
  const { user, events, isSubscribed, alertsStatus, enableAlerts, expoPushToken, refreshAll, stats, teams: apiTeams, claimedDeals, allDeals, alertsPaused, alertsPausedUntil, teamName } = useAppData();
  const nav = useNavigation<NativeStackNavigationProp<RootStackParamList>>();
  const wipe = useScreenWipe();
  const [showAllClaimed, setShowAllClaimed] = useState(false);
  const shownClaimed = showAllClaimed ? claimedDeals : claimedDeals.slice(0, TROPHY_PREVIEW);
  // Only claims whose window is still open can open the deal screen (to move one back).
  const activeIds = useMemo(() => new Set(allDeals.map((d) => d.id)), [allDeals]);
  const [showLicenses, setShowLicenses] = useState(false);

  // PAUSE > ABOUT lands here with an anchor: scroll to the About section once it has a position.
  const route = useRoute<RouteProp<TabParamList, 'You'>>();
  const scrollRef = useRef<Scroller>(null);
  const [aboutY, setAboutY] = useState<number | null>(null);
  const handled = useRef<number | undefined>(undefined);
  const anchor = route.params?.anchor;
  const anchorAt = route.params?.at;
  useEffect(() => {
    if (anchor !== 'about' || aboutY === null || handled.current === anchorAt) return;
    handled.current = anchorAt;
    scrollRef.current?.scrollTo({ y: Math.max(0, aboutY - 8), animated: false });
  }, [anchor, anchorAt, aboutY]);

  const followedTeams = useMemo(() => {
    const map = new Map<string, { id: string; name: string; sport?: string; badge?: Sprite }>();
    events.filter((e) => e.isActive !== false && isSubscribed(e.id)).forEach((e) => {
      if (!map.has(e.teamId)) {
        const team = apiTeams.find((x) => x.id === e.teamId);
        map.set(e.teamId, { id: e.teamId, name: teamName(e.teamId, e.teamName), sport: team?.sport, badge: team?.badge });
      }
    });
    return [...map.values()];
  }, [events, isSubscribed, apiTeams, teamName]);

  const onAlerts = () => {
    if (alertsStatus === 'undetermined') enableAlerts();
    else Linking.openSettings();
  };
  const version = Constants.expoConfig?.version ?? '';

  return (
    <Screen title="You" onRefresh={refreshAll} scrollRef={scrollRef}>
      <ServerBlocks keys={['you', 'profile']} />

      <PixelBox variant="panel" style={{ marginBottom: 16 }}>
        {stats ? (
          <Text
            variant="displayL"
            accessibilityLabel={`${stats.dealsClaimed} deals claimed`}
            style={{ marginBottom: 12 }}
          >
            <Text variant="displayL" color={stats.dealsClaimed > 0 ? 'accent' : 'text'}>
              {String(stats.dealsClaimed)}
            </Text>
            {' DEALS CLAIMED'}
          </Text>
        ) : null}
        <Text variant="caption" color="textSecondary">
          {followedTeams.length > 0
            ? `Following ${followedTeams.map((t) => t.name).join(', ')}`
            : 'Not following any teams yet.'}
        </Text>
        {followedTeams.length > 0 ? (
          <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 8, marginTop: 8 }}>
            {followedTeams.map((t) => (
              <SportBadge key={t.id} sport={t.sport} badge={t.badge} />
            ))}
          </View>
        ) : null}
      </PixelBox>

      <PixelMenu title="Trophy case">
        {claimedDeals.length > 0 ? (
          <>
            {shownClaimed.map((c) => (
              <WatchRow
                key={c.id}
                event={c.event}
                detail={`${teamName(c.event.teamId, c.event.teamName)} · CLAIMED ${shortDatePacific(c.claimedAt)}`}
                onPress={activeIds.has(c.id) ? () => wipe(() => nav.navigate('Deal', { id: c.id })) : undefined}
              />
            ))}
            {claimedDeals.length > TROPHY_PREVIEW ? (
              <MenuRow
                label={showAllClaimed ? 'Show fewer' : `Show all ${claimedDeals.length}`}
                onPress={() => setShowAllClaimed((v) => !v)}
                state={{ selected: showAllClaimed }}
              />
            ) : null}
          </>
        ) : (
          <Text variant="body" color="textSecondary" style={{ marginVertical: 8 }}>
            Claimed deals show up here so you can track your season.
          </Text>
        )}
      </PixelMenu>

      <PixelMenu title="Alerts">
        <MenuRow
          label="Deal alerts"
          value={alertsStatus === 'granted' ? 'On' : alertsStatus === 'denied' ? 'Off' : 'Turn on'}
          valueColor={alertsStatus === 'granted' ? 'success' : alertsStatus === 'denied' ? 'danger' : 'accent'}
          selected={alertsStatus === 'undetermined'}
          onPress={onAlerts}
          accessibilityHint={alertsStatus === 'undetermined' ? 'Asks to allow notifications' : 'Opens system settings'}
        />
        {alertsPaused && alertsPausedUntil ? (
          <MenuRow
            label="Paused"
            value={pauseUntilLabel(alertsPausedUntil)}
            valueColor="warning"
            onPress={() => nav.navigate('Pause')}
            accessibilityHint="Opens the pause menu to resume alerts"
          />
        ) : null}
        {alertsStatus === 'denied' ? (
          <Text variant="caption" color="textSecondary" style={{ marginBottom: 8 }}>
            Turn on in Settings to hear about deals.
          </Text>
        ) : null}
      </PixelMenu>

      <PixelMenu title="About" onLayout={(e) => setAboutY(e.nativeEvent.layout.y)}>
        <Text variant="body" color="textSecondary" style={{ marginVertical: 8 }}>
          {DISCLAIMER}
        </Text>
        <MenuRow
          label="Licenses"
          onPress={() => setShowLicenses((v) => !v)}
          state={{ selected: showLicenses }}
          value={showLicenses ? 'Hide' : undefined}
        />
        {showLicenses
          ? LICENSES.map((l) => (
              <PixelBox key={l.name} variant="dialog" style={{ marginBottom: 8 }}>
                <Text variant="bodyStrong" style={{ marginBottom: 4 }}>
                  {l.name}
                </Text>
                <Text variant="caption" color="textSecondary">
                  {l.text}
                </Text>
              </PixelBox>
            ))
          : null}
        {user ? <MenuRow label="Support code" value={user.id.replace(/-/g, '').slice(0, 8).toUpperCase()} /> : null}
        <MenuRow label="Version" value={version} />
      </PixelMenu>

      {__DEV__ ? (
        <PixelMenu title="Dev">
          <MenuRow label="Send test push" onPress={() => sendRandomTestNotification().catch(console.warn)} />
          {expoPushToken ? (
            <Text variant="caption" color="textSecondary" selectable>
              {expoPushToken}
            </Text>
          ) : null}
        </PixelMenu>
      ) : null}
    </Screen>
  );
}
