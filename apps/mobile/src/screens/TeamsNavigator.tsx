import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { colors } from '../hooks/useTheme';
import type { TeamsStackParamList } from '../navigation/types';
import { SportsScreen, SportTeamsScreen } from './TeamsScreen';

const Stack = createNativeStackNavigator<TeamsStackParamList>();

/**
 * The Teams tab: SELECT SPORT, then a sport's teams. Hard cuts, no iOS slide (docs/ui.md Motion);
 * BACK and the TEAMS tab return to SELECT SPORT, and swipe-back still works.
 */
export function TeamsNavigator() {
  return (
    <Stack.Navigator screenOptions={{ headerShown: false, animation: 'none', contentStyle: { backgroundColor: colors.background } }}>
      <Stack.Screen name="Sports" component={SportsScreen} />
      <Stack.Screen name="Sport" component={SportTeamsScreen} />
    </Stack.Navigator>
  );
}
