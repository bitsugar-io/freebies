import React, { useMemo, useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import { useNavigation, useRoute, RouteProp } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { Event, Team as ApiTeam } from '../api/client';
import { PixelButton } from '../components/pixel/PixelButton';
import { MenuRow } from '../components/pixel/PixelMenu';
import { SportBadge } from '../components/pixel/SportBadge';
import { useBlink } from '../hooks/useBlink';
import { useAppData } from '../context/AppDataContext';
import { Screen } from '../components/Screen';
import { ServerBlocks } from '../components/ServerBlocks';
import { ErrorPanel } from '../components/StatusPanels';
import { useToast } from '../components/pixel/Toast';
import PixelSprite from '../components/PixelSprite';
import { PixelBox } from '../components/pixel/PixelBox';
import { PixelIcon } from '../components/pixel/PixelIcon';
import { PixelMenu } from '../components/pixel/PixelMenu';
import { TeamTile } from '../components/pixel/TeamTile';
import { Text } from '../components/pixel/Text';
import { colors, useLargeText } from '../hooks/useTheme';
import { haptics } from '../utils/haptics';
import { ifCondition, sportLabel } from '../utils/format';
import type { RootStackParamList, TeamsStackParamList } from '../navigation/types';

/** A team's offers, keyed by the event's teamId (which is the API team id). */
interface TeamEvents {
  teamId: string;
  name: string;
  league: string;
  events: Event[];
}

const NO_EVENTS: Event[] = [];

function chunk<T>(items: T[], size: number): T[][] {
  const rows: T[][] = [];
  for (let i = 0; i < items.length; i += size) rows.push(items.slice(i, i + size));
  return rows;
}

/** One offer under a team: tap the text for details, tap the box to follow just this deal. */
function OfferRow({ event, on, onToggle, onOpen }: { event: Event; on: boolean; onToggle: () => void; onOpen: () => void }) {
  return (
    <View style={styles.offer}>
      <Pressable
        onPress={onOpen}
        style={styles.offerMain}
        accessibilityRole="button"
        accessibilityLabel={`${event.offerName} from ${event.partnerName}. ${ifCondition(event.triggerCondition)}`}
        accessibilityHint="Opens deal details"
      >
        <PixelSprite sprite={event.sprite} scale={2} />
        <View style={{ flex: 1 }}>
          <Text variant="bodyStrong" style={{ fontSize: 16 }}>
            {event.offerName}
          </Text>
          <Text variant="caption" color="textSecondary">
            {`${event.partnerName} · ${ifCondition(event.triggerCondition)}`}
          </Text>
        </View>
      </Pressable>
      <Pressable
        onPress={onToggle}
        accessibilityRole="checkbox"
        accessibilityLabel={`Follow ${event.offerName}`}
        accessibilityState={{ checked: on }}
        style={styles.check}
      >
        <View style={[styles.box, on && styles.boxOn]}>{on ? <PixelIcon name="check" scale={2} color="onAccent" /> : null}</View>
      </Pressable>
    </View>
  );
}

/** Everything the Teams screens share: the season board, offers by team, and follow counts. */
function useBoard() {
  const { events, eventsLoading, eventsError, isSubscribed, setSubscribed, refreshAll, retry, teams: apiTeams, teamsLoading, teamsError, refetchTeams, teamName } =
    useAppData();

  // Active offers grouped by team id.
  const teams = useMemo<TeamEvents[]>(() => {
    const map = new Map<string, TeamEvents>();
    events
      .filter((e) => e.isActive !== false)
      .forEach((e) => {
        const t = map.get(e.teamId) ?? {
          teamId: e.teamId,
          name: teamName(e.teamId, e.teamName),
          league: e.league,
          events: [],
        };
        t.events.push(e);
        map.set(e.teamId, t);
      });
    return [...map.values()];
  }, [events, teamName]);

  // The season board: the server's list, or (when /teams failed with nothing saved) the teams
  // that have offers, shown as in season under one generic sport.
  const board = useMemo<ApiTeam[]>(() => {
    if (apiTeams.length > 0) return apiTeams;
    return teams.map((t) => ({
      id: t.teamId,
      name: t.name,
      fullName: t.name,
      league: t.league,
      sport: '',
      leagueName: t.league.toUpperCase(),
      inSeason: true,
      returnLabel: null,
      activeDeals: t.events.length,
    }));
  }, [apiTeams, teams]);

  const eventsOf = (teamId: string) => teams.find((t) => t.teamId === teamId)?.events ?? NO_EVENTS;
  const followedOf = (list: Event[]) => list.filter((e) => isSubscribed(e.id)).length;
  return {
    teams, board, eventsOf, followedOf, isSubscribed, setSubscribed, refreshAll, retry, refetchTeams,
    eventsLoading, eventsError, teamsLoading, teamsError, usingFallback: apiTeams.length === 0, hasApiTeams: apiTeams.length > 0,
  };
}

interface SportGroup {
  sport: string;
  label: string;
  teams: ApiTeam[];
}

/** Sports come from the API's team list, in server order. A new sport appears with no app change. */
function groupBySport(board: ApiTeam[]): SportGroup[] {
  const map = new Map<string, SportGroup>();
  board.forEach((t) => {
    const g = map.get(t.sport) ?? { sport: t.sport, label: sportLabel(t.sport, t.sportName), teams: [] };
    g.teams.push(t);
    map.set(t.sport, g);
  });
  return [...map.values()];
}

function BoardNote({ fallback, onRetry }: { fallback: boolean; onRetry: () => void }) {
  return (
    <PixelBox variant="alert" padding={4} style={{ marginBottom: 12 }}>
      <View style={styles.note}>
        <PixelIcon name="warn" scale={2} color="danger" />
        <View style={{ flex: 1 }}>
          <Text variant="caption">
            {fallback ? "Couldn't load the season board. Showing teams with deals." : "Couldn't refresh the season board. Showing the saved list."}
          </Text>
        </View>
        <Pressable onPress={onRetry} accessibilityRole="button" accessibilityLabel="Retry" hitSlop={8} style={styles.retry}>
          <Text variant="displayM" color="accent">
            Retry
          </Text>
        </Pressable>
      </View>
    </PixelBox>
  );
}

/** One ▶ row per sport: badge, name, followed teams out of total, and "BACK MAR" when all are out of season. */
function SportRow({ group, followedTeams, selected, onPress }: { group: SportGroup; followedTeams: number; selected: boolean; onPress: () => void }) {
  const blink = useBlink(6, selected);
  const total = group.teams.length;
  const allOff = group.teams.every((t) => !t.inSeason);
  const back = allOff ? group.teams.find((t) => t.returnLabel)?.returnLabel : null;
  const badge = group.teams.find((t) => t.badge)?.badge;
  return (
    <Pressable
      onPress={onPress}
      accessibilityRole="menuitem"
      accessibilityLabel={`${group.label.toLowerCase()}, ${followedTeams} of ${total} teams followed${allOff ? `, off season${back ? `, back ${back}` : ''}` : ''}`}
      style={styles.sportRow}
    >
      {({ pressed }) => (
        <>
          <View style={styles.cursor} accessibilityElementsHidden importantForAccessibility="no-hide-descendants">
            {pressed || (selected && blink) ? <PixelIcon name="cursor" scale={2} color="accent" /> : null}
          </View>
          <SportBadge sport={group.sport} badge={badge} gray={allOff} />
          <View style={{ flex: 1 }}>
            {/* The ▶ alone marks focus; accent text would read as "picked". */}
            <Text variant="displayM" color={allOff ? 'textMuted' : 'text'}>
              {group.label}
            </Text>
            {allOff ? (
              <Text variant="displayM" color="textMuted">
                {`BACK ${back ?? 'SOON'}`}
              </Text>
            ) : null}
          </View>
          <Text variant="displayM" color={followedTeams > 0 ? 'accent' : 'textSecondary'}>
            {`${followedTeams}/${total}`}
          </Text>
        </>
      )}
    </Pressable>
  );
}

/** Teams, screen 1: SELECT SPORT. */
export function SportsScreen() {
  const nav = useNavigation<NativeStackNavigationProp<TeamsStackParamList>>();
  const b = useBoard();
  const groups = useMemo(() => groupBySport(b.board), [b.board]);
  const pickedTeams = b.board.filter((t) => b.followedOf(b.eventsOf(t.id)) > 0).length;
  const loading = b.eventsLoading || (b.teamsLoading && !b.hasApiTeams);
  const refresh = async () => {
    await b.refreshAll();
  };
  return (
    <Screen title="Select sport" status={`${pickedTeams} FOLLOWING`} onRefresh={refresh}>
      <ServerBlocks keys={['teams', 'discover']} />
      {b.eventsError && b.teams.length === 0 && !b.hasApiTeams ? (
        <ErrorPanel message={b.eventsError} onRetry={b.retry} />
      ) : loading ? (
        <View style={{ gap: 8 }}>
          {[0, 1, 2, 3].map((i) => (
            <PixelBox key={i} variant="off" style={{ minHeight: 56 }} accessible accessibilityLabel="Loading" />
          ))}
        </View>
      ) : (
        <>
          {b.teamsError ? <BoardNote fallback={b.usingFallback} onRetry={b.refetchTeams} /> : null}
          <Text variant="caption" color="textSecondary" style={{ marginBottom: 8 }}>
            Pick a sport, then follow its teams.
          </Text>
          <View accessibilityRole="menu">
            {groups.map((g, i) => (
              <SportRow
                key={g.sport || 'all'}
                group={g}
                selected={i === 0}
                followedTeams={g.teams.filter((t) => b.followedOf(b.eventsOf(t.id)) > 0).length}
                onPress={() => nav.navigate('Sport', { sport: g.sport })}
              />
            ))}
          </View>
        </>
      )}
    </Screen>
  );
}

/**
 * Teams, screen 2: one sport's teams. One tap follows or unfollows every deal for a team. Below
 * the tiles, each team expands to its individual deals with a check per deal.
 */
export function SportTeamsScreen() {
  const nav = useNavigation<NativeStackNavigationProp<TeamsStackParamList & RootStackParamList>>();
  const { sport } = useRoute<RouteProp<TeamsStackParamList, 'Sport'>>().params;
  const showToast = useToast();
  const b = useBoard();
  const [expanded, setExpanded] = useState<string | null>(null);
  const [busy, setBusy] = useState<Set<string>>(new Set());
  // 3 columns (about 100pt tiles), 2 at large text.
  const columns = useLargeText() ? 2 : 3;

  const list = useMemo(() => b.board.filter((t) => t.sport === sport), [b.board, sport]);
  const label = sportLabel(sport, list.find((t) => t.sportName)?.sportName);
  const loading = b.eventsLoading || (b.teamsLoading && !b.hasApiTeams);
  // Teams with deals first (stable), so the ones you can follow lead.
  const sorted = useMemo(
    () => [...list].sort((a, c) => Number(c.activeDeals > 0) - Number(a.activeDeals > 0)),
    [list]
  );
  const noneHaveDeals = list.length > 0 && list.every((t) => t.activeDeals === 0);
  const withEvents = list.filter((t) => b.eventsOf(t.id).length > 0);

  const toggleTeam = async (t: ApiTeam) => {
    const evs = b.eventsOf(t.id);
    if (t.activeDeals === 0 || evs.length === 0) {
      showToast(`No deal for the ${t.name} yet. We'll add them when a sponsor runs one.`);
      return;
    }
    haptics.selection();
    const all = b.followedOf(evs) === evs.length;
    setBusy((x) => new Set(x).add(t.id));
    try {
      await b.setSubscribed(evs.map((e) => e.id), !all);
    } finally {
      setBusy((x) => {
        const next = new Set(x);
        next.delete(t.id);
        return next;
      });
    }
  };

  return (
    <Screen title={label} status={list.length > 0 ? `${list.filter((t) => b.followedOf(b.eventsOf(t.id)) > 0).length}/${list.length} FOLLOWING` : undefined} onRefresh={b.refreshAll}>
      <PixelMenu>
        <MenuRow label="Back" onPress={() => nav.goBack()} accessibilityLabel="Back to select sport" leading="back" />
      </PixelMenu>
      {b.teamsError ? <BoardNote fallback={b.usingFallback} onRetry={b.refetchTeams} /> : null}
      {loading ? (
        <View style={{ gap: 8 }}>
          {[0, 1].map((i) => (
            <View key={i} style={styles.gridRow}>
              {Array.from({ length: columns }, (_, k) => (
                <PixelBox key={k} variant="off" padding={0} style={{ flex: 1, minHeight: 120 }} accessible accessibilityLabel="Loading" />
              ))}
            </View>
          ))}
        </View>
      ) : b.teamsError && list.length === 0 && !b.hasApiTeams ? (
        <ErrorPanel message={b.teamsError} onRetry={b.refetchTeams} />
      ) : list.length === 0 ? (
        <PixelBox variant="dialog" style={{ gap: 16 }}>
          <Text variant="displayL" color="accent" accessibilityRole="header">
            {`No teams for ${label}`}
          </Text>
          <Text variant="body" color="textSecondary">
            We don't track that sport yet.
          </Text>
          <PixelButton label="Back" onPress={() => nav.goBack()} />
        </PixelBox>
      ) : (
      <>
      <Text variant="caption" color="textSecondary" style={{ marginBottom: 12 }}>
        {noneHaveDeals
          ? `No ${label.toLowerCase()} deals yet. We'll add them when a sponsor runs one.`
          : 'Tap a team to follow all its deals.'}
      </Text>
      <View style={{ gap: 8 }}>
        {chunk(sorted, columns).map((row, i) => (
          <View key={i} style={styles.gridRow}>
            {row.map((t) => {
              const evs = b.eventsOf(t.id);
              return (
                <TeamTile
                  key={t.id}
                  name={t.name}
                  fullName={t.fullName}
                  league={t.leagueName ?? t.league.toUpperCase()}
                  sport={t.sport}
                  badge={t.badge}
                  teamColors={t.colors}
                  followed={b.followedOf(evs)}
                  total={evs.length}
                  inSeason={t.inSeason}
                  returnLabel={t.returnLabel}
                  activeDeals={t.activeDeals}
                  busy={busy.has(t.id)}
                  onPress={() => toggleTeam(t)}
                />
              );
            })}
            {row.length < columns
              ? Array.from({ length: columns - row.length }, (_, k) => <View key={`s${k}`} style={{ flex: 1 }} />)
              : null}
          </View>
        ))}
      </View>

      {withEvents.length > 0 ? (
        <Text variant="displayM" color="info" accessibilityRole="header" style={{ marginTop: 24, marginBottom: 8 }}>
          Deals by team
        </Text>
      ) : null}
      {withEvents.map((t) => {
        const evs = b.eventsOf(t.id);
        const open = expanded === t.id;
        const n = b.followedOf(evs);
        return (
          <View key={t.id}>
            <Pressable
              onPress={() => setExpanded(open ? null : t.id)}
              accessibilityRole="button"
              accessibilityLabel={`${t.name}, ${n} of ${evs.length} deals followed`}
              accessibilityState={{ expanded: open }}
              style={styles.teamRow}
            >
              <View style={{ flex: 1 }}>
                <Text variant="displayM" color={n > 0 ? 'accent' : 'text'}>
                  {t.name}
                </Text>
                <Text variant="caption" color="textSecondary">
                  {`${n} of ${evs.length} following`}
                </Text>
              </View>
              <View style={{ transform: [{ rotate: open ? '0deg' : '-90deg' }] }}>
                <PixelIcon name="more" scale={2} color="textSecondary" />
              </View>
            </Pressable>
            {open
              ? evs.map((e) => (
                  <OfferRow
                    key={e.id}
                    event={e}
                    on={b.isSubscribed(e.id)}
                    onToggle={() => {
                      haptics.selection();
                      b.setSubscribed([e.id], !b.isSubscribed(e.id));
                    }}
                    onOpen={() => nav.navigate('Deal', { id: e.id })}
                  />
                ))
              : null}
          </View>
        );
      })}
      </>
      )}
    </Screen>
  );
}

const styles = StyleSheet.create({
  sportRow: { minHeight: 56, flexDirection: 'row', alignItems: 'center', gap: 12, borderBottomWidth: 2, borderBottomColor: colors.divider, paddingVertical: 8 },
  cursor: { width: 20, alignItems: 'flex-start' },
  gridRow: { flexDirection: 'row', gap: 8 },
  section: { marginTop: 8, marginBottom: 8 },
  note: { flexDirection: 'row', alignItems: 'center', gap: 8, paddingLeft: 4 },
  retry: { minHeight: 44, minWidth: 44, alignItems: 'center', justifyContent: 'center', paddingHorizontal: 4 },
  teamRow: {
    minHeight: 56,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    borderBottomWidth: 2,
    borderBottomColor: colors.divider,
    paddingVertical: 8,
  },
  offer: { flexDirection: 'row', alignItems: 'center', minHeight: 56, paddingLeft: 12 },
  offerMain: { flex: 1, flexDirection: 'row', alignItems: 'center', gap: 12, paddingVertical: 8 },
  check: { width: 48, height: 48, alignItems: 'center', justifyContent: 'center' },
  box: {
    width: 28,
    height: 28,
    borderWidth: 3,
    borderColor: colors.border,
    backgroundColor: colors.background,
    alignItems: 'center',
    justifyContent: 'center',
  },
  boxOn: { backgroundColor: colors.accent },
});
