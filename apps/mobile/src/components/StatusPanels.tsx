import React, { useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import { colors } from '../hooks/useTheme';
import { clockTime } from '../utils/format';
import { DialogBox } from './pixel/DialogBox';
import { PixelBox } from './pixel/PixelBox';
import { PixelButton } from './pixel/PixelButton';
import { Text } from './pixel/Text';
import PixelSprite from './PixelSprite';
import { NO_SIGNAL } from './pixel/art';

/** Full error with nothing to fall back on: plain words, retry, and details for support. */
export function ErrorPanel({ message, onRetry }: { message?: string | null; onRetry: () => void }) {
  const [details, setDetails] = useState(false);
  return (
    <View style={styles.gap}>
      <View style={styles.sprite}>
        <PixelSprite sprite={NO_SIGNAL} scale={5} />
      </View>
      <DialogBox variant="alert" accessibilityLabel="Couldn't reach Freebies: LA Sports Deals. Check your connection and try again.">
        Couldn't reach Freebies: LA Sports Deals. Check your connection and try again.
      </DialogBox>
      <PixelButton label="Try again" onPress={onRetry} />
      {message ? (
        <Pressable
          onPress={() => setDetails((d) => !d)}
          accessibilityRole="button"
          accessibilityLabel={details ? 'Hide details' : 'Show details'}
          style={styles.link}
        >
          <Text variant="caption" color="textSecondary">
            {details ? `Error: ${message}` : 'Show details'}
          </Text>
        </Pressable>
      ) : null}
    </View>
  );
}

/** Strip shown above saved deals when the latest refresh failed. */
export function StaleBanner({ savedAt, onRetry }: { savedAt: number | null; onRetry: () => void }) {
  const when = savedAt ? ` Showing deals saved at ${clockTime(savedAt)}.` : '';
  return (
    <PixelBox variant="alert" padding={4} style={styles.banner}>
      <View style={styles.bannerRow}>
        <PixelSprite sprite={NO_SIGNAL} scale={2} />
        <View style={{ flex: 1 }}>
          <Text variant="caption">{`Couldn't refresh.${when}`}</Text>
        </View>
        <Pressable
          onPress={onRetry}
          accessibilityRole="button"
          accessibilityLabel="Retry"
          hitSlop={8}
          style={styles.retry}
        >
          <Text variant="displayM" color="accent">
            Retry
          </Text>
        </Pressable>
      </View>
    </PixelBox>
  );
}

/** Centered empty state: one icon or sprite, one line, one action. */
export function EmptyState({
  art,
  title,
  body,
  actionLabel,
  onAction,
}: {
  art: React.ReactNode;
  title: string;
  body: string;
  actionLabel?: string;
  onAction?: () => void;
}) {
  return (
    <View style={styles.empty}>
      {art}
      <Text variant="displayL" color="accent" align="center" accessibilityRole="header">
        {title}
      </Text>
      <Text variant="body" color="textSecondary" align="center">
        {body}
      </Text>
      {actionLabel && onAction ? <PixelButton label={actionLabel} onPress={onAction} style={{ alignSelf: 'stretch' }} /> : null}
    </View>
  );
}

const styles = StyleSheet.create({
  gap: { gap: 12 },
  link: { minHeight: 44, justifyContent: 'center', alignItems: 'center' },
  banner: { marginBottom: 12 },
  bannerRow: { flexDirection: 'row', alignItems: 'center', gap: 8, paddingLeft: 4 },
  sprite: { alignItems: 'center', paddingTop: 8 },
  retry: { minHeight: 44, minWidth: 44, alignItems: 'center', justifyContent: 'center', paddingHorizontal: 4 },
  empty: { alignItems: 'center', gap: 16, paddingVertical: 32, paddingHorizontal: 8 },
});
