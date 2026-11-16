import React, { useCallback, useEffect, useRef, useState } from 'react';
import { Linking, ScrollView, StyleSheet, Text as RNText, View } from 'react-native';
import { useNavigation, useRoute, RouteProp } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { ActiveDeal, api, Event } from '../api/client';
import { useAppData } from '../context/AppDataContext';
import { ServerBlocks } from '../components/ServerBlocks';
import { SponsoredCard } from '../components/SponsoredCard';
import { AtTheGameTag, venueLine } from '../components/DealRows';
import PixelSprite from '../components/PixelSprite';
import { CutIn } from '../components/pixel/CutIn';
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
import { useReducedMotion } from '../hooks/useReducedMotion';
import { clockTime, formatCountdown, ifCondition } from '../utils/format';
import { countdownTarget, countdownTone, isLocked, ptStamp } from '../utils/dealState';
import { haptics } from '../utils/haptics';
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

/**
 * Deal detail. Route `Deal` with `id` = triggered-event id (an unlocked deal) or event id
 * (a preview of a deal you're considering). Reachable from a push tap or freebie://deal/<id>.
 */
export function DealScreen() {
  const nav = useNavigation<NativeStackNavigationProp<RootStackParamList>>();
  const { id } = useRoute<RouteProp<RootStackParamList, 'Deal'>>().params;
  const insets = useSafeAreaInsets();
  const showToast = useToast();
  const reduced = useReducedMotion();
  const {
    allDeals,
    events,
    dealsLoading,
    eventsLoading,
    refetchDeals,
    dismissDeal,
    undoDismissal,
    isSubscribed,
    setSubscribed,
    teamName,
  } = useAppData();

  const deal: ActiveDeal | undefined =
    allDeals.find((d) => d.id === id) ??
    allDeals.find((d) => d.eventId === id && !d.isDismissed) ??
    allDeals.find((d) => d.eventId === id);

  // If a push arrives before the list knows the deal, refresh once, then fall back to the event.
  const [settled, setSettled] = useState(false);
  const [fetched, setFetched] = useState<Event | null>(null);
  const refetched = useRef(false);
  useEffect(() => {
    if (deal || dealsLoading || refetched.current) return;
    refetched.current = true;
    refetchDeals().finally(() => setSettled(true));
  }, [deal, dealsLoading, refetchDeals]);

  const knownEvent = deal?.event ?? events.find((e) => e.id === id);
  useEffect(() => {
    if (deal || knownEvent || !settled || eventsLoading) return;
    let alive = true;
    api.getEvent(id).then((e) => alive && setFetched(e)).catch(() => {});
    return () => {
      alive = false;
    };
  }, [deal, knownEvent, settled, eventsLoading, id]);
  const event = knownEvent ?? fetched;

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

  // Cut-in: once per unlocked deal, then the sprite rises in stepped frames.
  const [cutIn, setCutIn] = useState(false);
  const [rise, setRise] = useState(reduced ? RISE_FRAMES - 1 : 0);
  const checkedSeen = useRef(false);
  useEffect(() => {
    if ((kind !== 'unlocked' && kind !== 'locked') || !deal || checkedSeen.current) return;
    checkedSeen.current = true;
    loadSeen().then((seen) => {
      if (seen.includes(deal.id)) return;
      AsyncStorage.setItem(SEEN_KEY, JSON.stringify([...seen, deal.id].slice(-50))).catch(() => {});
      haptics.success();
      setCutIn(true);
    });
  }, [kind, deal]);
  useEffect(() => {
    if (cutIn || reduced || rise >= RISE_FRAMES - 1) return;
    const t = setTimeout(() => setRise((r) => r + 1), 110);
    return () => clearTimeout(t);
  }, [cutIn, reduced, rise]);
  const endCutIn = useCallback(() => setCutIn(false), []);

  const [confirmStop, setConfirmStop] = useState(false);
  const close = () => (nav.canGoBack() ? nav.goBack() : nav.navigate('Tabs'));
  const failed = () => showToast("Couldn't save that. Try again.");

  const gotIt = async () => {
    if (!deal) return;
    try {
      await dismissDeal(deal.id, 'got_it');
      haptics.light();
      showToast(`Claimed: ${deal.event.offerName}`);
      close();
    } catch {
      failed();
    }
  };
  const stopReminders = async () => {
    if (!deal) return;
    try {
      await dismissDeal(deal.id, 'stop_reminding');
      showToast(`Reminders off for ${deal.event.offerName}.`);
      close();
    } catch {
      failed();
    }
  };
  const moveBack = async () => {
    if (!deal) return;
    try {
      await undoDismissal(deal.id);
      showToast(`${deal.event.offerName} is back in Won.`);
    } catch {
      failed();
    }
  };

  // Loading: a skeleton of this layout, never the preview view for a deal that is really unlocked.
  if (!event) {
    const missing = settled && !eventsLoading && !dealsLoading;
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

        <View
          style={[styles.sprite, { transform: [{ translateY: (RISE_FRAMES - 1 - rise) * 4 }] }]}
          accessible
          accessibilityRole="image"
          accessibilityLabel={event.offerName}
        >
          <PixelSprite sprite={event.sprite} scale={6} gray={kind === 'expired'} />
        </View>

        <Text variant="bodyStrong" align="center" accessibilityRole="header" style={styles.offerName}>
          {event.offerName}
        </Text>

        {kind === 'unlocked' || kind === 'locked' || kind === 'quiet' ? (
          <>
            <DialogBox
              accessibilityLabel={`You won ${event.offerName} from ${event.partnerName}${event.partnerName.endsWith('.') ? '' : '.'} Because ${event.triggerCondition}.${locked && deal?.availableAt ? ` Your code unlocks ${ptStamp(deal.availableAt)}.` : ''}`}
            >
              You won {key(event.offerName)} from {event.partnerName}{event.partnerName.endsWith('.') ? '' : '.'} Because {event.triggerCondition.replace(/\.$/, '')}.
              {locked && deal?.availableAt ? ` Your code unlocks ${ptStamp(deal.availableAt)}.` : ''}
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
                    : `left · ends ${clockTime(target.to)}`}
                </Text>
                {target.kind === 'expiry' ? (
                  <View style={{ alignSelf: 'stretch', marginTop: 12 }}>
                    <ExpiryBar triggeredAt={deal.triggeredAt} expiresAt={deal.expiresAt} />
                  </View>
                ) : null}
              </View>
            ) : null}

            <VenueNote event={event} />
            <Text variant="displayM" color="textMuted" accessibilityRole="header" style={styles.label}>
              How to claim
            </Text>
            <Text variant="body" color="textSecondary">
              {event.offerDescription}
            </Text>

            {/* Locked: no OPEN and no GOT IT until the code works. */}
            {locked && kind !== 'quiet' ? null : (
              <PixelMenu>
                {!locked && event.offerUrl ? (
                  <MenuRow label={`Open ${event.partnerName}`} selected onPress={() => openUrl(event.offerUrl)} />
                ) : null}
                {!locked ? <MenuRow label="Got it, claimed" selected={!event.offerUrl} onPress={gotIt} /> : null}
                {kind === 'quiet' ? (
                  <MenuRow label="Turn reminders on" selected={locked} onPress={moveBack} />
                ) : null}
              </PixelMenu>
            )}

            {kind === 'unlocked' || kind === 'locked' ? (
              confirmStop ? (
                <View style={{ gap: 8, marginTop: 8 }}>
                  <DialogBox variant="alert" text={`Stop reminders for ${event.offerName}? You can still claim it until it ends.`} />
                  <PixelMenu>
                    <MenuRow label="Yes, stop reminders" tone="danger" onPress={stopReminders} />
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
                <MenuRow label="Move back to won" onPress={moveBack} />
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
              text={`This one ended${deal?.expiresAt ? ` at ${clockTime(deal.expiresAt)}` : ''}. Keep following ${team} to win the next one.`}
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
  sprite: { alignItems: 'center', marginVertical: 16 },
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
