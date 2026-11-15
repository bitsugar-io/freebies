import React, { useRef } from 'react';
import { Animated, PanResponder, Pressable, StyleSheet } from 'react-native';
import { colors, PX } from '../../hooks/useTheme';
import { PixelIcon } from './PixelIcon';
import { Text } from './Text';

interface Props {
  /** The text on the track, e.g. "SLIDE TO STOP REMINDERS". */
  label: string;
  /** Fires when slid past 70%. */
  onComplete: () => void;
  /** Fires on a tap anywhere on the track (and on a screen-reader double tap): ask first. */
  onTap?: () => void;
  accessibilityLabel: string;
}

const THUMB = 48;
const THRESHOLD = 0.7;

/**
 * Deliberate-gesture control for "stop reminders". Slide the thumb across, or tap anywhere on the
 * track to be asked first (VoiceOver gets a plain button that does the same). The callbacks go through a ref so the
 * gesture always sees the latest handlers.
 */
export function PixelSlide({ label, onComplete, onTap, accessibilityLabel }: Props) {
  const width = useRef(0);
  const x = useRef(new Animated.Value(0)).current;
  const handlers = useRef({ onComplete, onTap });
  handlers.current = { onComplete, onTap };

  const pan = useRef(
    PanResponder.create({
      onStartShouldSetPanResponder: () => true,
      onMoveShouldSetPanResponder: () => true,
      onPanResponderMove: (_, g) => {
        const max = width.current - THUMB;
        if (g.dx > 0) x.setValue(Math.min(g.dx, max));
      },
      onPanResponderRelease: (_, g) => {
        const max = width.current - THUMB;
        if (g.dx >= max * THRESHOLD) {
          // Stepped: snap to the end in one frame, no easing.
          x.setValue(max);
          handlers.current.onComplete();
          setTimeout(() => x.setValue(0), 300);
        } else {
          x.setValue(0);
          if (Math.abs(g.dx) < 6 && Math.abs(g.dy) < 6) handlers.current.onTap?.();
        }
      },
    })
  ).current;

  return (
    <Pressable
      style={styles.track}
      onLayout={(e) => (width.current = e.nativeEvent.layout.width)}
      onPress={() => handlers.current.onTap?.()}
      accessibilityRole="button"
      accessibilityLabel={accessibilityLabel}
      accessibilityHint="Asks you to confirm"
    >
      <Text variant="displayM" color="textSecondary" align="center" style={styles.hint} numberOfLines={2}>
        {label}
      </Text>
      <Animated.View style={[styles.thumb, { transform: [{ translateX: x }] }]} {...pan.panHandlers}>
        <PixelIcon name="cursor" scale={2} color="onAccent" />
      </Animated.View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  track: {
    height: THUMB + 2 * PX,
    borderWidth: PX,
    borderColor: colors.textSecondary,
    backgroundColor: colors.background,
    justifyContent: 'center',
  },
  hint: { paddingLeft: THUMB, paddingRight: 8 },
  thumb: {
    position: 'absolute',
    left: 0,
    width: THUMB,
    height: THUMB,
    backgroundColor: colors.accent,
    alignItems: 'center',
    justifyContent: 'center',
  },
});
