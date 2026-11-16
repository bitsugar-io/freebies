import React, { useEffect, useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import { colors } from '../../hooks/useTheme';
import { useReducedMotion } from '../../hooks/useReducedMotion';
import { Text } from './Text';

interface Props {
  text: string;
  onDone: () => void;
  /** Total time before auto-dismiss. Capped at 2s. */
  durationMs?: number;
}

const FRAME_MS = 60;
const FRAMES = 4;

/**
 * Full-screen celebration banner ("DUCKS WIN!"). Drops in over 4 stepped frames, auto-dismisses
 * within 2s, and any tap dismisses it. Skipped entirely under Reduce Motion.
 */
export function CutIn({ text, onDone, durationMs = 1600 }: Props) {
  const reduced = useReducedMotion();
  const [frame, setFrame] = useState(0);

  useEffect(() => {
    if (reduced) {
      onDone();
      return;
    }
    const steps = setInterval(() => setFrame((f) => Math.min(FRAMES - 1, f + 1)), FRAME_MS);
    const end = setTimeout(onDone, Math.min(durationMs, 2000));
    return () => {
      clearInterval(steps);
      clearTimeout(end);
    };
  }, [reduced, durationMs, onDone]);

  if (reduced) return null;
  const drop = (FRAMES - 1 - frame) * -16;
  return (
    <Pressable
      style={styles.overlay}
      onPress={onDone}
      accessibilityRole="button"
      accessibilityLabel={`${text}. Tap to continue`}
    >
      <View style={{ transform: [{ translateY: drop }] }}>
        <Text variant="displayXl" color="accent" align="center" style={styles.text}>
          {text}
        </Text>
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  overlay: {
    ...StyleSheet.absoluteFill,
    backgroundColor: colors.background,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 24,
    zIndex: 10,
  },
  text: {
    textShadowColor: colors.dangerDeep,
    textShadowOffset: { width: 3, height: 3 },
    textShadowRadius: 0,
  },
});
