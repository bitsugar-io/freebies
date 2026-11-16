import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { colors } from '../hooks/useTheme';
import type { TeamsStackParamList } from '../navigation/types';
import { SportsScreen, SportTeamsScreen } from './TeamsScreen';

const Stack = createNativeStackNavigator<TeamsStackParamList>();

/** The Teams tab: SELECT SPORT, then a sport's teams. Native push, so swipe-back works. */
export function TeamsNavigator() {
  return (
    <Stack.Navigator screenOptions={{ headerShown: false, contentStyle: { backgroundColor: colors.background } }}>
      <Stack.Screen name="Sports" component={SportsScreen} />
      <Stack.Screen name="Sport" component={SportTeamsScreen} />
    </Stack.Navigator>
  );
}
