import React from 'react';
import { StyleSheet, View } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useAppData } from '../context/AppDataContext';
import { DialogBox } from '../components/pixel/DialogBox';
import { MenuRow, PixelMenu } from '../components/pixel/PixelMenu';
import { PixelIcon } from '../components/pixel/PixelIcon';
import { colors } from '../hooks/useTheme';
import { useScreenWipe } from '../components/pixel/ScreenWipe';

/**
 * Explainer shown after the first follow and before the system notification prompt. The system
 * dialog only appears when the user taps TURN ON ALERTS.
 */
export function AlertsScreen() {
  const nav = useNavigation();
  const wipe = useScreenWipe();
  const insets = useSafeAreaInsets();
  const { enableAlerts } = useAppData();
  const close = () => wipe(() => nav.goBack());

  return (
    <View style={styles.scrim}>
      <View style={[styles.sheet, { paddingBottom: insets.bottom + 16 }]}>
        <View style={{ alignItems: 'center', marginBottom: 16 }}>
          <PixelIcon name="bell" scale={6} color="accent" />
        </View>
        <DialogBox
          text="Deals unlock after the game, and most are gone by the next morning. Want a ping when one drops?"
          more
        />
        <PixelMenu>
          <MenuRow
            label="Turn on alerts"
            selected
            onPress={async () => {
              await enableAlerts();
              close();
            }}
          />
          <MenuRow label="Not now" onPress={close} />
        </PixelMenu>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  scrim: { flex: 1, justifyContent: 'flex-end', backgroundColor: colors.background },
  sheet: { padding: 16 },
});
