import React from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { colors } from '../../hooks/useTheme';
import { useVictory } from '../../hooks/useVictory';
import { navigationRef } from '../../navigation/ref';
import { PixelIcon } from './PixelIcon';
import { Text } from './Text';

interface Props {
  title: string;
  /** Short status on the right ("1 FOLLOWING"). Decorative detail, keep it to a few characters. */
  status?: string;
  /** Show a 44pt close button on the right instead of status. */
  onClose?: () => void;
  /** Left slot. */
  left?: React.ReactNode;
  /** Skip the top safe-area padding (modals that already sit below the status bar). */
  noInset?: boolean;
  /** Hide the MENU (pause menu) button. */
  noMenu?: boolean;
}

/**
 * Black header bar with the screen title in gold. Under it, a 4pt gold rule, or while a win is live,
 * a two-band stripe in that team's colors (from the API), like a jersey band. Doubles as the
 * screen's heading landmark.
 */
export function HudBar({ title, status, onClose, left, noInset, noMenu }: Props) {
  const insets = useSafeAreaInsets();
  const victory = useVictory();
  return (
    <View>
    <View style={[styles.bar, { paddingTop: (noInset ? 0 : insets.top) + 8 }]}>
      {left}
      <View style={{ flex: 1 }}>
        <Text variant="displayL" color="accent" numberOfLines={2} accessibilityRole="header">
          {title}
        </Text>
        {status ? (
          <Text variant="displayM" color="textSecondary" style={{ marginTop: 2 }}>
            {status}
          </Text>
        ) : null}
      </View>
      {!noMenu ? (
        <Pressable
          onPress={() => navigationRef.isReady() && navigationRef.navigate('Pause')}
          accessibilityRole="button"
          accessibilityLabel="Menu"
          accessibilityHint="Opens the menu: pause alerts, alert settings, about"
          hitSlop={4}
          style={({ pressed }) => [styles.start, pressed && { transform: [{ translateX: 2 }, { translateY: 2 }] }]}
        >
          <View style={styles.menuInner} accessibilityElementsHidden importantForAccessibility="no-hide-descendants">
            <PixelIcon name="menu" scale={2} color="text" />
            <Text variant="displayM" color="text">
              MENU
            </Text>
          </View>
        </Pressable>
      ) : null}
      {onClose ? (
        <Pressable
          onPress={onClose}
          hitSlop={8}
          accessibilityRole="button"
          accessibilityLabel="Close"
          style={styles.close}
        >
          <PixelIcon name="close" scale={2} color="text" />
        </Pressable>
      ) : null}
    </View>
    {victory ? (
      <View accessibilityElementsHidden importantForAccessibility="no-hide-descendants">
        <View style={[styles.band, { backgroundColor: victory.stripe[0] }]} />
        <View style={[styles.band, { backgroundColor: victory.stripe[1] }]} />
      </View>
    ) : (
      <View style={[styles.band, { backgroundColor: colors.accent }]} />
    )}
    </View>
  );
}

const styles = StyleSheet.create({
  bar: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    backgroundColor: colors.background,
    paddingHorizontal: 16,
    paddingBottom: 8,
    minHeight: 56,
  },
  menuInner: { flexDirection: 'row', alignItems: 'center', gap: 6 },
  start: {
    minWidth: 64,
    minHeight: 44,
    paddingHorizontal: 8,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.background,
    borderWidth: 3,
    borderColor: colors.border,
  },
  close: { width: 44, height: 44, alignItems: 'center', justifyContent: 'center' },
  band: { height: 4 },
});
