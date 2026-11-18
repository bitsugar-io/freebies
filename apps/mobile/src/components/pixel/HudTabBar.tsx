import React from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import type { BottomTabBarProps } from '@react-navigation/bottom-tabs';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useAppData } from '../../context/AppDataContext';
import { colors } from '../../hooks/useTheme';
import { Blink } from '../../hooks/useBlink';
import { PixelIcon } from './PixelIcon';
import type { IconName } from './art';
import { useScreenWipe } from './ScreenWipe';
import { Text } from './Text';

const LABELS: Record<string, string> = { Won: 'Won', Today: 'Today', Teams: 'Teams', You: 'You' };
const TAB_ICONS: Record<string, IconName> = { Won: 'trophy', Today: 'calendar', Teams: 'jersey', You: 'you' };

/** First screen of each tab's nested stack; the tab bar always returns there. */
const TAB_ROOTS: Record<string, string> = { Teams: 'Sports' };

/**
 * The game HUD tab bar: black, a 3pt lilac top rule, then four tabs, each a pixel icon over a small
 * Press Start 2P label (owner pick 2026-11-17, idea 4). The active tab's icon and label are gold. The
 * win count is a solid gold badge on the trophy; nothing here blinks except PAUSED, which shows only
 * while alerts are paused.
 */
export function HudTabBar({ state, navigation }: BottomTabBarProps) {
  const insets = useSafeAreaInsets();
  const wipe = useScreenWipe();
  const { liveWon, followedCount, alertsPaused } = useAppData();
  // Same list as the WON screen (liveWonDeals): expired and dismissed wins never count.
  const won = liveWon.length;

  return (
    <View style={[styles.bar, { paddingBottom: insets.bottom }]}>
      {alertsPaused ? (
        <View style={styles.paused} accessible accessibilityLabel={`Alerts paused. Following ${followedCount} teams.`}>
          <Blink priority={1}>
            <Text variant="displayS" color="warning">
              ALERTS PAUSED
            </Text>
          </Blink>
        </View>
      ) : null}
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
              // Already on this tab: back to its first screen, behind the same wipe as switching tabs.
              const depth = (route.state as { index?: number } | undefined)?.index ?? 0;
              if (root && depth > 0) wipe(go);
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
              <View style={styles.iconBox} accessibilityElementsHidden importantForAccessibility="no-hide-descendants">
                <PixelIcon name={TAB_ICONS[route.name] ?? 'more'} scale={2} color={focused ? 'accent' : 'textMuted'} />
                {badge ? (
                  <View style={styles.badge}>
                    <Text variant="displayS" color="onAccent">
                      {String(badge)}
                    </Text>
                  </View>
                ) : null}
              </View>
              <Text variant="displayS" color={focused ? 'accent' : 'textMuted'}>
                {name}
              </Text>
            </Pressable>
          );
        })}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  bar: { backgroundColor: colors.background, borderTopWidth: 3, borderTopColor: colors.info },
  paused: { alignItems: 'center', paddingTop: 6 },
  // Equal columns, so the tabs sit evenly.
  tabs: { flexDirection: 'row', paddingHorizontal: 4 },
  tab: { flex: 1, minHeight: 56, alignItems: 'center', justifyContent: 'center', gap: 6, paddingTop: 8, paddingBottom: 4 },
  iconBox: { width: 22, height: 22 },
  badge: { position: 'absolute', top: -7, right: -14, backgroundColor: colors.accent, paddingHorizontal: 3, minWidth: 14, alignItems: 'center' },
});
