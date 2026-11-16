import React, { useEffect, useState } from 'react';
import { AccessibilityInfo, Pressable, StyleSheet, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { colors } from '../../hooks/useTheme';
import { Blink } from '../../hooks/useBlink';
import { AppLogo } from './AppLogo';
import { Text } from './Text';

const AUTO_CONTINUE_MS = 4000;

/**
 * First-launch title screen. Tap anywhere, or wait 4 seconds, to continue. The blinking prompt is
 * solid under Reduce Motion (the screen still shows). A full-screen button sits under it all.
 */
export function TitleScreen({ onDone }: { onDone: () => void }) {
  const insets = useSafeAreaInsets();
  // With a screen reader on, wait for a tap: 4 seconds is too short to hear the screen.
  const [screenReader, setScreenReader] = useState<boolean | null>(null);
  useEffect(() => {
    let alive = true;
    AccessibilityInfo.isScreenReaderEnabled()
      .then((on) => alive && setScreenReader(on))
      .catch(() => alive && setScreenReader(false));
    const sub = AccessibilityInfo.addEventListener('screenReaderChanged', setScreenReader);
    return () => {
      alive = false;
      sub.remove();
    };
  }, []);
  useEffect(() => {
    if (screenReader !== false) return undefined;
    const t = setTimeout(onDone, AUTO_CONTINUE_MS);
    return () => clearTimeout(t);
  }, [onDone, screenReader]);

  return (
    <Pressable
      onPress={onDone}
      accessibilityRole="button"
      accessibilityLabel="Freebies: LA Sports Deals. When LA wins, you win. Tap to start."
      style={[styles.root, { paddingTop: insets.top + 24, paddingBottom: insets.bottom + 24 }]}
    >
      <View style={{ gap: 12, alignItems: 'center' }}>
        <Text variant="displayXl" color="accent" align="center">
          Freebies
        </Text>
        <Text variant="displayM" color="text" align="center">
          LA Sports Deals
        </Text>
      </View>
      <AppLogo scale={6} />
      <View style={{ gap: 20, alignItems: 'center' }}>
        <Blink priority={0}>
          <Text variant="displayL" color="text" align="center">
            Tap to start
          </Text>
        </Blink>
        <Text variant="displayM" color="textSecondary" align="center">
          When LA wins, you win
        </Text>
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  root: {
    flex: 1,
    backgroundColor: colors.background,
    alignItems: 'center',
    justifyContent: 'space-around',
    paddingHorizontal: 24,
  },
});
