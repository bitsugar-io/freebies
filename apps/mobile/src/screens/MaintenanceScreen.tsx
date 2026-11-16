import React, { useEffect } from 'react';
import { ScrollView, StyleSheet, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useAppConfig } from '../context/AppConfigContext';
import { useAppData } from '../context/AppDataContext';
import { UnlockedDealRow } from '../components/DealRows';
import { DialogBox } from '../components/pixel/DialogBox';
import { PalmLogo } from '../components/pixel/PalmLogo';
import { Text } from '../components/pixel/Text';
import { colors } from '../hooks/useTheme';

/** Shown when the server sets maintenance_mode. Saved deals stay readable. Re-checks every minute. */
export function MaintenanceScreen() {
  const insets = useSafeAreaInsets();
  const { refreshConfig } = useAppConfig();
  const { undismissedDeals } = useAppData();

  useEffect(() => {
    const id = setInterval(refreshConfig, 60_000);
    return () => clearInterval(id);
  }, [refreshConfig]);

  return (
    <ScrollView
      style={{ backgroundColor: colors.background }}
      contentContainerStyle={[styles.content, { paddingTop: insets.top + 32, paddingBottom: insets.bottom + 24 }]}
    >
      <View style={{ alignItems: 'center', gap: 16 }}>
        <PalmLogo scale={6} />
        <Text variant="displayL" color="accent" align="center" accessibilityRole="header">
          Paused
        </Text>
      </View>
      <DialogBox text="Freebies: LA Sports Deals is down for a quick fix. Your deals are safe. We'll check again in a minute." />
      {undismissedDeals.length > 0 ? (
        <View style={{ gap: 12 }}>
          <Text variant="displayM" color="textMuted" accessibilityRole="header">
            Your deals
          </Text>
          {undismissedDeals.map((d) => (
            <UnlockedDealRow key={d.id} deal={d} onPress={() => {}} />
          ))}
        </View>
      ) : null}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  content: { padding: 16, gap: 24, flexGrow: 1 },
});
