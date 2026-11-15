import React from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import { ActiveDeal, Event } from '../api/client';
import { colors } from '../hooks/useTheme';
import { useNow } from '../hooks/useNow';
import { Blink } from '../hooks/useBlink';
import { ExpiryBar, formatTimeLeft, isUrgent, spokenTimeLeft, timeLeftMs } from './pixel/DealBar';
import { PixelBox } from './pixel/PixelBox';
import { Text } from './pixel/Text';
import PixelSprite from './PixelSprite';
import { ifCondition } from '../utils/format';
import { isLocked, spokenUnlock, unlockTag } from '../utils/dealState';

const DAY_MS = 86_400_000;

/** Time left, first thing on the card. Orange with a blinking "ENDS SOON" under 6 hours. */
export function TimeLeft({ expiresAt }: { expiresAt?: string }) {
  const now = useNow(60_000);
  const left = timeLeftMs(expiresAt, now);
  if (left === null) return null;
  const urgent = isUrgent(left);
  if (left <= 0) {
    return (
      <Text variant="displayM" color="info">
        Expired
      </Text>
    );
  }
  return (
    <View style={styles.timeRow}>
      {urgent ? (
        <Blink priority={5}>
          <Text variant="displayM" color="warning">
            Ends soon
          </Text>
        </Blink>
      ) : null}
      <Text variant="displayM" color={urgent ? 'warning' : 'info'}>
        {formatTimeLeft(left)}
      </Text>
    </View>
  );
}

/** "NEW!" tag for deals unlocked in the last 24 hours. Blinks; solid under Reduce Motion. */
export function NewTag({ triggeredAt }: { triggeredAt: string }) {
  const now = useNow(60_000);
  const age = now - new Date(triggeredAt).getTime();
  if (!(age >= -60_000 && age < DAY_MS)) return null;
  return (
    <Blink priority={4}>
      <View style={styles.newTag} accessibilityElementsHidden importantForAccessibility="no-hide-descendants">
        <Text variant="displayM" color="onAccent">
          NEW!
        </Text>
      </View>
    </Blink>
  );
}

function isNew(triggeredAt: string) {
  return Date.now() - new Date(triggeredAt).getTime() < DAY_MS;
}

/** A won deal (locked until its code unlocks, if the API says so). The whole box is the one action: open the claim screen. */
export function UnlockedDealRow({ deal, onPress }: { deal: ActiveDeal; onPress: () => void }) {
  const { event } = deal;
  // Re-render each minute so a locked row flips to the normal state when the code unlocks.
  const now = useNow(60_000);
  const left = timeLeftMs(deal.expiresAt, now);
  const locked = isLocked(deal.availableAt, now) && !!deal.availableAt;
  const label = [
    !locked && isNew(deal.triggeredAt) ? 'New' : '',
    `${event.offerName} from ${event.partnerName}`,
    `Because ${event.triggerCondition}`,
    locked ? spokenUnlock(deal.availableAt!, now) : left !== null ? spokenTimeLeft(left) : '',
  ]
    .filter(Boolean)
    .join('. ');
  return (
    <Pressable onPress={onPress} accessibilityRole="button" accessibilityLabel={label} accessibilityHint={locked ? 'Opens the deal' : 'Opens how to claim'}>
      {({ pressed }) => (
        <PixelBox variant="panel" style={pressed ? styles.pressed : undefined}>
          <View style={styles.topRow}>
            {locked ? (
              // Locked: a solid status, never ENDS SOON or NEW, so nothing blinks over it.
              <Text variant="displayM" color="info">
                {unlockTag(deal.availableAt!, now)}
              </Text>
            ) : (
              <>
                <TimeLeft expiresAt={deal.expiresAt} />
                <NewTag triggeredAt={deal.triggeredAt} />
              </>
            )}
          </View>
          <View style={styles.main}>
            <PixelSprite sprite={event.sprite} scale={3} />
            <View style={{ flex: 1 }}>
              <Text variant="bodyStrong" style={styles.headline}>
                {event.offerName}
              </Text>
              <Text variant="caption" color="textSecondary">
                {event.partnerName}
              </Text>
            </View>
          </View>
          <Text variant="caption" color="success" style={{ marginBottom: 8 }}>
            {`Because ${event.triggerCondition}`}
          </Text>
          {locked ? null : <ExpiryBar triggeredAt={deal.triggeredAt} expiresAt={deal.expiresAt} />}
        </PixelBox>
      )}
    </Pressable>
  );
}

/** A deal you follow that hasn't unlocked: sprite, offer, and the condition in plain words. */
export function WatchRow({
  event,
  onPress,
  right,
  muted,
  detail,
}: {
  event: Event;
  /** Replaces the "If ..." condition line, e.g. "Dodgers · CLAIMED NOV 15". */
  detail?: string;
  onPress?: () => void;
  right?: React.ReactNode;
  /** Off-season team: grey sprite, muted text. */
  muted?: boolean;
}) {
  const body = (
    <View style={styles.watch}>
      <PixelSprite sprite={event.sprite} scale={2} gray={muted} />
      <View style={{ flex: 1 }}>
        <Text variant="bodyStrong" color={muted ? 'textMuted' : 'text'} style={{ fontSize: 16 }}>
          {event.offerName}
        </Text>
        <Text variant="caption" color={muted ? 'textMuted' : 'textSecondary'}>
          {detail ?? `${event.partnerName} · ${ifCondition(event.triggerCondition)}`}
        </Text>
      </View>
      {right}
    </View>
  );
  if (!onPress) return body;
  return (
    <Pressable
      onPress={onPress}
      accessibilityRole="button"
      accessibilityLabel={`${event.offerName} from ${event.partnerName}. ${detail ?? ifCondition(event.triggerCondition)}`}
    >
      {body}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  timeRow: { flexDirection: 'row', flexWrap: 'wrap', alignItems: 'center', columnGap: 8 },
  topRow: { flexDirection: 'row', flexWrap: 'wrap', alignItems: 'center', justifyContent: 'space-between', gap: 8 },
  newTag: { backgroundColor: colors.accent, paddingHorizontal: 6, paddingVertical: 2 },
  pressed: { transform: [{ translateX: 2 }, { translateY: 2 }] },
  main: { flexDirection: 'row', alignItems: 'center', gap: 12, marginVertical: 10 },
  headline: { fontSize: 20, lineHeight: 26 },
  watch: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
    minHeight: 56,
    paddingVertical: 8,
    borderBottomWidth: 2,
    borderBottomColor: colors.surface,
  },
});
