import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Linking, ScrollView, StyleSheet, Text as RNText, View } from 'react-native';
import { useNavigation, useRoute, RouteProp } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { ActiveDeal, api, Event, TodayDeal } from '../api/client';
import { useAppData } from '../context/AppDataContext';
import { ServerBlocks } from '../components/ServerBlocks';
import { SponsoredCard } from '../components/SponsoredCard';
import { AtTheGameTag, venueLine } from '../components/DealRows';
import PixelSprite from '../components/PixelSprite';
import { CutIn } from '../components/pixel/CutIn';
import { VictoryRays } from '../components/pixel/VictoryRays';
import { ExpiryBar, spokenTimeLeft, timeLeftMs } from '../components/pixel/DealBar';
import { DialogBox } from '../components/pixel/DialogBox';
import { HudBar } from '../components/pixel/HudBar';
import { MenuRow, PixelMenu } from '../components/pixel/PixelMenu';
import { PixelSlide } from '../components/pixel/PixelSlide';
import { SkeletonBox } from '../components/pixel/Skeleton';
import { Text } from '../components/pixel/Text';
import { useToast } from '../components/pixel/Toast';
import { colors } from '../hooks/useTheme';
import { useNow } from '../hooks/useNow';
import { useBlink } from '../hooks/useBlink';
import { useTeamColors } from '../hooks/useVictory';
import { useReducedMotion } from '../hooks/useReducedMotion';
import { deadlineLabel, formatCountdown, ifCondition } from '../utils/format';
import { countdownTarget, countdownTone, isLocked, ptStamp } from '../utils/dealState';
import { haptics } from '../utils/haptics';
import { dealScreenPhase, resolveDeal } from '../utils/dealResolve';
import { once, singleFlight } from '../utils/sequence';
import type { RootStackParamList } from '../navigation/types';

const SEEN_KEY = 'freebie_cutin_seen';
const RISE_FRAMES = 4;

/** Deal ids whose "unlocked" cut-in already played, so it only plays once per deal. */
/** "03:28:11" with the colons blinking on the shared clock (solid under Reduce Motion). */
function CountdownDigits({ value }: { value: string }) {
  const on = useBlink(7);
  return (
    <>
      {value.split(':').map((part, i) => (
        <React.Fragment key={i}>
          {i > 0 ? <RNText style={on ? undefined : { color: colors.background }}>:</RNText> : null}
          {part}
        </React.Fragment>
      ))}
    </>
  );
}

async function loadSeen(): Promise<string[]> {
  try {
    return JSON.parse((await AsyncStorage.getItem(SEEN_KEY)) ?? '[]');
  } catch {
    return [];
  }
}

function cutInText(event: Event, team: string, locked: boolean): string {
  return /\bwins?\b/i.test(event.triggerCondition) ? `${team} win!` : locked ? 'Deal won!' : 'Deal unlocked!';
}

/** "Code unlocks in 2 days" or "3 hours 28 minutes left". */
function spokenCountdown(kind: 'unlock' | 'expiry', ms: number): string {
  const spoken = spokenTimeLeft(ms);
  return kind === 'unlock' ? `Code unlocks in ${spoken.replace(/ left$/, '')}` : spoken;
}

function openUrl(url?: string) {
  if (url && /^https?:\/\//i.test(url)) Linking.openURL(url);
}

type Kind = 'unlocked' | 'locked' | 'quiet' | 'claimed' | 'expired' | 'preview';

/** How long a deal screen waits for your deals and the lookups before it decides. */
const WAIT_MS = 8_000;

/**
 * Deal detail. Route `Deal` with `id` = triggered-event id (a win: yours, or a live one from TODAY)
 * or event id (a preview of a deal you're considering). Reachable from a push tap or
 * freebie://deal/<id>. Keyed by id, so a second push while it's open starts fresh.
 */
export function DealScreen() {
  const { id } = useRoute<RouteProp<RootStackParamList, 'Deal'>>().params;
  return <DealContent key={id} id={id} />;
}

function DealContent({ id }: { id: string }) {
  const nav = useNavigation<NativeStackNavigationProp<RootStackParamList>>();
  const insets = useSafeAreaInsets();
  const showToast = useToast();
  const reduced = useReducedMotion();
  const {
    allDeals,
    events,
    today,
    dealsLoading,
    refetchDeals,
    dismissDeal,
    undoDismissal,
    isSubscribed,
    setSubscribed,
    teamName,
  } = useAppData();

  // Public lookups for a deal that isn't in your list: the triggered deal (a live one from TODAY, or
  // a push for a team you've since unfollowed), else the event. Started at once, not after the user.
  const [fetchedTriggered, setFetchedTriggered] = useState<TodayDeal | null>(null);
  const [fetchedEvent, setFetchedEvent] = useState<Event | null>(null);
  const [lookupsDone, setLookupsDone] = useState(false);
  const mineAtOpen = useRef(allDeals.some((d) => d.id === id));
  const eventAtOpen = useRef(events.some((e) => e.id === id));
  useEffect(() => {
    if (mineAtOpen.current) {
      setLookupsDone(true);
      return;
    }
    let alive = true;
    (async () => {
      if (!eventAtOpen.current) {
        try {
          const t = await api.getTriggered(id);
          if (alive) setFetchedTriggered(t);
          return;
        } catch {
          // Not a triggered id (or offline): try it as an event id.
        }
        try {
          const e = await api.getEvent(id);
          if (alive) setFetchedEvent(e);
        } catch {
          // Unknown or offline.
        }
      }
    })().finally(() => alive && setLookupsDone(true));
    return () => {
      alive = false;
    };
  }, [id]);

  const resolved = resolveDeal(id, {
    mine: allDeals,
    live: today?.live ?? [],
    fetchedTriggered,
    events,
    fetchedEvent,
  });

  // Not yours yet: refresh your deals once (a push can arrive before the list knows the win).
  const [refetchDone, setRefetchDone] = useState(false);
  const refetched = useRef(false);
  useEffect(() => {
    if (resolved?.kind === 'mine' || dealsLoading || refetched.current) return;
    refetched.current = true;
    refetchDeals().finally(() => setRefetchDone(true));
  }, [resolved?.kind, dealsLoading, refetchDeals]);

  const [timedOut, setTimedOut] = useState(false);
  useEffect(() => {
    const t = setTimeout(() => setTimedOut(true), WAIT_MS);
    return () => clearTimeout(t);
  }, []);

  const phase = dealScreenPhase({
    resolved: resolved?.kind ?? null,
    dealsKnown: !dealsLoading,
    refetchDone,
    lookupsDone,
    timedOut,
  });
  const ready = phase === 'show' && resolved ? resolved : null;
  /** A win in your list: claim, silence, move back. */
  const mine: ActiveDeal | undefined = ready?.kind === 'mine' ? ready.deal : undefined;
  /** A live deal that isn't in your list: open the partner, follow the team. */
  const pub: ActiveDeal | undefined = ready?.kind === 'public' ? { ...ready.deal, isDismissed: false } : undefined;
  const deal = mine ?? pub;
  const owned = !!mine;
  const event: Event | null = ready ? (ready.kind === 'event' ? ready.event : ready.deal.event) : null;
  // The YOU WON burst is always in the colors of this deal's team, whoever else won.
  const rayColors = useTeamColors(event?.teamId);

  const now = useNow(1000);
  const minute = useNow(60_000);
  const left = timeLeftMs(deal?.expiresAt, now);
  // Locked until availableAt: the deal is won, but the code can't be used yet. useNow flips it.
  const locked = !!deal?.availableAt && isLocked(deal.availableAt, now);
  const target = deal ? countdownTarget(deal, now) : null;
  const targetLeft = target ? timeLeftMs(target.to, now) : null;
  const kind: Kind | null = !event
    ? null
    : !deal
      ? 'preview'
      : deal.isDismissed
        ? deal.dismissalType === 'got_it'
          ? 'claimed'
          : 'quiet'
        : left !== null && left <= 0
          ? 'expired'
          : locked
            ? 'locked'
            : 'unlocked';

  // Cut-in: once per unlocked deal of yours, then the sprite rises in stepped frames.
  const [cutIn, setCutIn] = useState(false);
  const [rise, setRise] = useState(reduced ? RISE_FRAMES - 1 : 0);
  const checkedSeen = useRef(false);
  useEffect(() => {
    if ((kind !== 'unlocked' && kind !== 'locked') || !mine || checkedSeen.current) return;
    checkedSeen.current = true;
    loadSeen().then((seen) => {
      if (seen.includes(mine.id)) return;
      AsyncStorage.setItem(SEEN_KEY, JSON.stringify([...seen, mine.id].slice(-50))).catch(() => {});
      haptics.success();
      setCutIn(true);
    });
  }, [kind, mine]);
  useEffect(() => {
    if (cutIn || reduced || rise >= RISE_FRAMES - 1) return;
    const t = setTimeout(() => setRise((r) => r + 1), 110);
    return () => clearTimeout(t);
  }, [cutIn, reduced, rise]);
  const endCutIn = useCallback(() => setCutIn(false), []);

  const [confirmStop, setConfirmStop] = useState(false);
  // Close once, however many taps or completions land.
  const close = useMemo(() => once(() => (nav.canGoBack() ? nav.goBack() : nav.navigate('Tabs'))), [nav]);
  const failed = () => showToast("Couldn't save that. Try again.");

  // One action in flight at a time: a double tap on GOT IT, or a slide plus a tap, sends one request.
  const [busy, setBusy] = useState(false);
  const act = useRef(
    singleFlight(async (fn: () => Promise<void>) => {
      setBusy(true);
      try {
        await fn();
      } finally {
        setBusy(false);
      }
    })
  ).current;

  const gotIt = () =>
    act(async () => {
      if (!mine) return;
      try {
        await dismissDeal(mine.id, 'got_it');
        haptics.light();
        showToast(`Claimed: ${mine.event.offerName}`);
        close();
      } catch {
        failed();
      }
    });
  const stopReminders = () =>
    act(async () => {
      if (!mine) return;
      try {
        await dismissDeal(mine.id, 'stop_reminding');
        showToast(`Reminders off for ${mine.event.offerName}.`);
        close();
      } catch {
        failed();
      }
    });
  const moveBack = () =>
    act(async () => {
      if (!mine) return;
      try {
        await undoDismissal(mine.id);
        showToast(`${mine.event.offerName} is back in Won.`);
      } catch {
        failed();
      }
    });

  // Loading: a skeleton of this layout, never the preview view for a deal that is really unlocked.
  if (!event) {
    const missing = phase === 'missing';
    return (
      <View style={styles.root}>
        <HudBar title="Deal" onClose={close} noMenu />
        <View style={styles.content}>
          {missing ? (
            <>
              <DialogBox variant="alert" text="That deal isn't available anymore. It may have ended." />
              <PixelMenu>
                <MenuRow label="Close" selected onPress={close} />
              </PixelMenu>
            </>
          ) : (
            <View style={{ gap: 12 }}>
              <SkeletonBox height={96} />
              <SkeletonBox height={120} />
            </View>
          )}
        </View>
      </View>
    );
  }

  const team = teamName(event.teamId, event.teamName);
  const following = isSubscribed(event.id);
  // Follow the whole team (all its active deals), as the TEAMS screen does.
  const teamEvents = events.filter((e) => e.teamId === event.teamId && e.isActive !== false).map((e) => e.id);
  const teamIds = teamEvents.length > 0 ? teamEvents : [event.id];
  const followingTeam = teamIds.some((eid) => isSubscribed(eid));
  const key = (text: string) => (
    <Text variant="bodyStrong" color="accent">
      {text}
    </Text>
  );

  return (
    <View style={styles.root}>
      <HudBar title="Deal" onClose={close} noMenu />
      <ScrollView contentContainerStyle={[styles.content, { paddingBottom: insets.bottom + 24 }]}>
        <ServerBlocks keys={['deal_v2', 'deal_detail']} />

        <VictoryRays pair={kind === 'expired' ? [colors.textMuted, colors.surfaceSecondary] : rayColors}>
          <View
            style={{ transform: [{ translateY: (RISE_FRAMES - 1 - rise) * 4 }] }}
            accessible
            accessibilityRole="image"
            accessibilityLabel={event.offerName}
          >
            <PixelSprite sprite={event.sprite} scale={6} gray={kind === 'expired'} />
          </View>
        </VictoryRays>

        <Text variant="bodyStrong" align="center" accessibilityRole="header" style={styles.offerName}>
          {event.offerName}
        </Text>

        {kind === 'unlocked' || kind === 'locked' || kind === 'quiet' ? (
          <>
            <DialogBox
              accessibilityLabel={`${owned ? 'You won' : 'Live now:'} ${event.offerName} from ${event.partnerName}${event.partnerName.endsWith('.') ? '' : '.'} Because ${event.triggerCondition}.${locked && deal?.availableAt ? ` ${owned ? 'Your code' : 'The code'} unlocks ${ptStamp(deal.availableAt)}.` : ''}`}
            >
              {owned ? 'You won' : 'Live now:'} {key(event.offerName)} from {event.partnerName}{event.partnerName.endsWith('.') ? '' : '.'} Because {event.triggerCondition.replace(/\.$/, '')}.
              {locked && deal?.availableAt ? ` ${owned ? 'Your code' : 'The code'} unlocks ${ptStamp(deal.availableAt)}.` : ''}
            </DialogBox>
            {target && targetLeft !== null && deal ? (
              <View style={styles.countdown}>
                <Text
                  variant="displayXl"
                  color={countdownTone(targetLeft)}
                  align="center"
                  accessibilityLabel={spokenCountdown(target.kind, timeLeftMs(target.to, minute) ?? 0)}
                >
                  <CountdownDigits value={formatCountdown(targetLeft)} />
                </Text>
                <Text variant="caption" color="textSecondary" align="center">
                  {target.kind === 'unlock'
                    ? `until your code unlocks${deal.expiresAt ? ` · good until ${ptStamp(deal.expiresAt)}` : ''}`
                    : `left · ends ${deadlineLabel(target.to, minute)}`}
                </Text>
                {target.kind === 'expiry' ? (
                  <View style={{ alignSelf: 'stretch', marginTop: 12 }}>
                    <ExpiryBar triggeredAt={deal.triggeredAt} expiresAt={deal.expiresAt} />
                  </View>
                ) : null}
              </View>
            ) : null}

            <VenueNote event={event} />
            <Text variant="displayM" color="info" accessibilityRole="header" style={styles.label}>
              How to claim
            </Text>
            <Text variant="body" color="textSecondary">
              {event.offerDescription}
            </Text>

            {/* Locked: no OPEN and no GOT IT until the code works. */}
            {owned ? (
              locked && kind !== 'quiet' ? null : (
                <PixelMenu>
                  {!locked && event.offerUrl ? (
                    <MenuRow label={`Open ${event.partnerName}`} selected onPress={() => openUrl(event.offerUrl)} />
                  ) : null}
                  {!locked ? (
                    <MenuRow label="Got it, claimed" selected={!event.offerUrl} disabled={busy} onPress={gotIt} />
                  ) : null}
                  {kind === 'quiet' ? (
                    <MenuRow label="Turn reminders on" selected={locked} disabled={busy} onPress={moveBack} />
                  ) : null}
                </PixelMenu>
              )
            ) : (
              // Not in your list (a team you don't follow): open the partner, or follow the team so
              // the next one is yours. Following adds this live one to WON too.
              <PixelMenu>
                {!locked && event.offerUrl ? (
                  <MenuRow label={`Open ${event.partnerName}`} selected onPress={() => openUrl(event.offerUrl)} />
                ) : null}
                <MenuRow
                  label={`Follow ${team}`}
                  value={followingTeam ? 'On' : 'Off'}
                  valueColor={followingTeam ? 'success' : 'textMuted'}
                  selected={locked || !event.offerUrl}
                  role="switch"
                  state={{ checked: followingTeam }}
                  accessibilityLabel={`Follow ${team}`}
                  onPress={() => {
                    haptics.selection();
                    setSubscribed(teamIds, !followingTeam);
                  }}
                />
              </PixelMenu>
            )}

            {owned && (kind === 'unlocked' || kind === 'locked') ? (
              confirmStop ? (
                <View style={{ gap: 8, marginTop: 8 }}>
                  <DialogBox variant="alert" text={`Stop reminders for ${event.offerName}? You can still claim it until it ends.`} />
                  <PixelMenu>
                    <MenuRow label="Yes, stop reminders" tone="danger" disabled={busy} onPress={stopReminders} />
                    <MenuRow label="Keep reminders" selected onPress={() => setConfirmStop(false)} />
                  </PixelMenu>
                </View>
              ) : (
                <View style={{ marginTop: 8 }}>
                  <PixelSlide
                    label="SLIDE TO STOP REMINDERS"
                    accessibilityLabel={`Stop reminders for ${event.offerName}`}
                    onComplete={stopReminders}
                    onTap={() => setConfirmStop(true)}
                  />
                </View>
              )
            ) : null}
          </>
        ) : null}

        {kind === 'claimed' ? (
          <>
            <DialogBox text={`You claimed ${event.offerName} from ${event.partnerName}.`} />
            <PixelMenu>
              {deal?.expiresAt && left !== null && left > 0 ? (
                <MenuRow label="Move back to won" disabled={busy} onPress={moveBack} />
              ) : null}
              <MenuRow label="Close" selected onPress={close} />
            </PixelMenu>
          </>
        ) : null}

        {kind === 'expired' ? (
          <>
            <Text variant="displayL" color="danger" align="center" style={{ marginBottom: 12 }}>
              Expired
            </Text>
            <DialogBox
              text={`This one ended${deal?.expiresAt ? ` ${deadlineLabel(deal.expiresAt, minute)}` : ''}. ${followingTeam ? 'Keep following' : 'Follow'} ${team} to win the next one.`}
            />
            <PixelMenu>
              <MenuRow label="Close" selected onPress={close} />
            </PixelMenu>
          </>
        ) : null}

        {kind === 'preview' ? (
          <>
            <DialogBox
              accessibilityLabel={`${ifCondition(event.triggerCondition)}, you get ${event.offerName} from ${event.partnerName}.`}
            >
              {ifCondition(event.triggerCondition)}, you get {key(event.offerName)} from {event.partnerName}.
            </DialogBox>
            <VenueNote event={event} />
            <Text variant="body" color="textSecondary" style={{ marginTop: 12 }}>
              {event.offerDescription}
            </Text>
            <PixelMenu>
              <MenuRow
                label="Following"
                value={following ? 'On' : 'Off'}
                valueColor={following ? 'success' : 'textMuted'}
                selected={!following}
                role="switch"
                state={{ checked: following }}
                accessibilityLabel={`Following ${event.offerName}`}
                onPress={() => {
                  haptics.selection();
                  setSubscribed([event.id], !following);
                }}
              />
              {event.offerUrl ? <MenuRow label={`Open ${event.partnerName}`} onPress={() => openUrl(event.offerUrl)} /> : null}
            </PixelMenu>
          </>
        ) : null}

        <SponsoredCard event={event} />
      </ScrollView>
      {cutIn ? <CutIn text={cutInText(event, team, locked)} onDone={endCutIn} /> : null}
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: colors.background },
  content: { padding: 16, flexGrow: 1 },
  countdown: { alignItems: 'center', marginTop: 16 },
  label: { marginTop: 20, marginBottom: 8 },
  offerName: { fontSize: 24, lineHeight: 30, marginBottom: 16 },
});

/** At-the-game deals: the tag plus where, in plain words, so nobody expects to claim it from home. */
function VenueNote({ event }: { event: Event }) {
  if (!event.inVenue) return null;
  return (
    <View style={{ flexDirection: 'row', flexWrap: 'wrap', alignItems: 'center', gap: 8, marginTop: 12 }}>
      <AtTheGameTag event={event} />
      <Text variant="caption" color="textSecondary">
        {venueLine(event)}
      </Text>
    </View>
  );
}
