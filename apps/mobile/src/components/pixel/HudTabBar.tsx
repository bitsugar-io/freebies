import React from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import type { BottomTabBarProps } from '@react-navigation/bottom-tabs';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useAppData } from '../../context/AppDataContext';
import { colors } from '../../hooks/useTheme';
import { Blink } from '../../hooks/useBlink';
import { PixelIcon } from './PixelIcon';
import { useScreenWipe } from './ScreenWipe';
import { Text } from './Text';

const LABELS: Record<string, string> = { Won: 'Won', Today: 'Today', Teams: 'Teams', You: 'You' };

/** A thin pixel bar between stat-line items. */
function Sep() {
  return <View style={styles.sep} accessibilityElementsHidden importantForAccessibility="no-hide-descendants" />;
}

/** First screen of each tab's nested stack; the tab bar always returns there. */
const TAB_ROOTS: Record<string, string> = { Teams: 'Sports' };

/**
 * The game HUD tab bar: black, 3pt white top border, a stat line, then three text tabs. The active
 * tab is accent with a blinking ▶. Blinkers here: PAUSED, the ▶ and the won count.
 */
export function HudTabBar({ state, navigation }: BottomTabBarProps) {
  const insets = useSafeAreaInsets();
  const wipe = useScreenWipe();
  const { undismissedDeals, followedCount, alertsPaused } = useAppData();
  const won = undismissedDeals.length;

  const statLabel = [
    `${won} won`,
    `following ${followedCount}`,
    alertsPaused ? 'alerts paused' : null,
  ]
    .filter(Boolean)
    .join(', ');

  return (
    <View style={[styles.bar, { paddingBottom: insets.bottom }]}>
      <View style={styles.stats} accessible accessibilityLabel={statLabel}>
        <Text variant="displayM" color="textSecondary">{`WON ${won}`}</Text>
        <Sep />
        <Text variant="displayM" color="textSecondary">{`FOLLOWING ${followedCount}`}</Text>
        {alertsPaused ? (
          <>
            <Sep />
          <Blink priority={1}>
            <Text variant="displayM" color="warning">
              PAUSED
            </Text>
          </Blink>
          </>
        ) : null}
      </View>
      <View style={styles.tabs} accessibilityRole="tablist">
        {state.routes.map((route, index) => {
          const focused = state.index === index;
          const badge = route.name === 'Won' && won > 0 ? won : 0;
          const name = LABELS[route.name] ?? route.name;
          const onPress = () => {
            const event = navigation.emit({ type: 'tabPress', target: route.key, canPreventDefault: true });
            if (event.defaultPrevented) return;
            // Tabs with a nested stack always open on their first screen, whether you come from
            // another tab or tap the tab you're already on.
            const root = TAB_ROOTS[route.name];
            const go = () =>
              (navigation.navigate as (name: string, params?: object) => void)(
                route.name,
                root ? { screen: root } : undefined,
              );
            if (focused) {
              if (root) go();
              return;
            }
            wipe(go);
          };
          return (
            <Pressable
              key={route.key}
              onPress={onPress}
              onLongPress={() => navigation.emit({ type: 'tabLongPress', target: route.key })}
              accessibilityRole="tab"
              accessibilityLabel={badge ? `${name}, ${badge} won` : name}
              accessibilityState={{ selected: focused }}
              style={styles.tab}
            >
              <View style={styles.cursor} accessibilityElementsHidden importantForAccessibility="no-hide-descendants">
                {focused ? (
                  <Blink priority={2}>
                    <PixelIcon name="cursor" scale={2} color="accent" />
                  </Blink>
                ) : null}
              </View>
              <Text variant="displayM" color={focused ? 'accent' : 'textSecondary'}>
                {name}
              </Text>
              {badge ? (
                <Blink priority={3} style={styles.badge}>
                  <Text variant="displayM" color="onAccent">
                    {String(badge)}
                  </Text>
                </Blink>
              ) : null}
            </Pressable>
          );
        })}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  bar: { backgroundColor: colors.background, borderTopWidth: 3, borderTopColor: colors.border },
  stats: { flexDirection: 'row', flexWrap: 'wrap', alignItems: 'center', gap: 8, paddingHorizontal: 12, paddingTop: 8, paddingBottom: 2 },
  sep: { width: 2, height: 12, marginHorizontal: 4, backgroundColor: colors.textMuted },
  // Equal columns, so the tabs sit evenly.
  tabs: { flexDirection: 'row', paddingHorizontal: 4 },
  tab: { flex: 1, minHeight: 48, flexDirection: 'row', alignItems: 'center', justifyContent: 'center', paddingHorizontal: 4 },
  cursor: { width: 12, alignItems: 'flex-start' },
  badge: { marginLeft: 6, backgroundColor: colors.accent, paddingHorizontal: 4, minWidth: 20, alignItems: 'center' },
});
