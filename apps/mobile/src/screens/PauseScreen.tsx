import React, { useEffect, useRef, useState } from 'react';
import { AccessibilityInfo, Linking, StyleSheet, View, findNodeHandle } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useAppData } from '../context/AppDataContext';
import { PixelBox } from '../components/pixel/PixelBox';
import { MenuRow, PixelMenu } from '../components/pixel/PixelMenu';
import { Text } from '../components/pixel/Text';
import { useToast } from '../components/pixel/Toast';
import { colors } from '../hooks/useTheme';
import { next8amPacific, pauseUntilLabel } from '../utils/format';
import type { RootStackParamList } from '../navigation/types';

const FOREVER = '9999-12-31T00:00:00Z';
type Page = 'main' | 'pause';

/**
 * The pause menu: a black full-screen overlay with a bordered box. It is a transparent modal route
 * with no animation, so it appears instantly (Reduce Motion safe) and the back button closes it.
 */
export function PauseScreen() {
  const nav = useNavigation<NativeStackNavigationProp<RootStackParamList>>();
  const insets = useSafeAreaInsets();
  const showToast = useToast();
  const { alertsPaused, alertsPausedUntil, alertsStatus, enableAlerts, pauseAlerts } = useAppData();
  const [page, setPage] = useState<Page>('main');
  const titleRef = useRef<React.ElementRef<typeof View>>(null);
  const close = () => (nav.canGoBack() ? nav.goBack() : nav.navigate('Tabs'));

  // Move screen reader focus to the title when the menu opens or changes page.
  useEffect(() => {
    const t = setTimeout(() => {
      const node = titleRef.current ? findNodeHandle(titleRef.current) : null;
      if (node) AccessibilityInfo.setAccessibilityFocus(node);
    }, 100);
    return () => clearTimeout(t);
  }, [page]);

  const pauseFor = async (until: Date | string, label: string) => {
    const iso = typeof until === 'string' ? until : until.toISOString();
    close();
    if (await pauseAlerts(iso)) showToast(`Alerts paused ${label}.`);
  };
  const resume = async () => {
    close();
    if (await pauseAlerts(null)) showToast('Alerts are back on.');
  };
  const alertSettings = () => {
    if (alertsStatus === 'undetermined') {
      nav.replace('Alerts');
    } else {
      close();
      Linking.openSettings();
    }
  };
  // Scroll YOU down to its About section (the nonce makes a repeat request count).
  const about = () => nav.navigate('Tabs', { screen: 'You', params: { anchor: 'about', at: Date.now() } });

  return (
    <View
      style={[styles.root, { paddingTop: insets.top + 16, paddingBottom: insets.bottom + 16 }]}
      accessibilityViewIsModal
      onAccessibilityEscape={close}
    >
      <PixelBox variant="dialog" padding={12}>
        <View ref={titleRef} accessible accessibilityRole="header" accessibilityLabel={page === 'main' ? 'Pause' : 'Pause alerts'}>
          <Text variant="displayL" color="accent" style={{ marginBottom: 8 }}>
            {page === 'main' ? 'Pause' : 'Pause alerts'}
          </Text>
        </View>
        {page === 'main' ? (
          <PixelMenu>
            {alertsPaused ? (
              <MenuRow label="Resume alerts" selected blinkPriority={0} onPress={resume} value={undefined} accessibilityHint={alertsPausedUntil ? `Paused until ${pauseUntilLabel(alertsPausedUntil)}` : undefined} />
            ) : (
              <MenuRow label="Pause alerts" selected blinkPriority={0} onPress={() => setPage('pause')} />
            )}
            <MenuRow label="Alert settings" onPress={alertSettings} />
            <MenuRow label="About" onPress={about} />
            <MenuRow label="Back" onPress={close} accessibilityLabel="Back to the game" />
          </PixelMenu>
        ) : (
          <PixelMenu>
            <MenuRow
              label="1 hour"
              selected
              blinkPriority={0}
              onPress={() => pauseFor(new Date(Date.now() + 3_600_000), 'for 1 hour')}
            />
            <MenuRow
              label="Until tomorrow"
              onPress={() => {
                const t = next8amPacific();
                pauseFor(t, `until ${pauseUntilLabel(t.toISOString())}`);
              }}
              accessibilityHint="Until 8am Pacific time"
            />
            <MenuRow label="Until I resume" onPress={() => pauseFor(FOREVER, 'until you resume')} />
            <MenuRow label="Back" onPress={() => setPage('main')} />
          </PixelMenu>
        )}
      </PixelBox>
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: colors.background, justifyContent: 'center', paddingHorizontal: 16 },
});
