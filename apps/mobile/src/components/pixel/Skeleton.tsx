import React from 'react';
import { StyleProp, View, ViewStyle } from 'react-native';
import { colors } from '../../hooks/useTheme';
import { useBlinkClock } from '../../hooks/useBlink';
import { PixelBox } from './PixelBox';

/** Loading placeholder: a panel outline with bars. Blinks between two tones in steps (static under Reduce Motion). */
export function SkeletonBox({ height = 96, style }: { height?: number; style?: StyleProp<ViewStyle> }) {
  const on = useBlinkClock();
  const bar = on ? colors.surface : colors.surfaceSecondary;
  return (
    <PixelBox
      variant="off"
      style={[{ minHeight: height }, style]}
      accessible
      accessibilityLabel="Loading"
      accessibilityRole="progressbar"
    >
      <View style={{ height: 14, width: '60%', backgroundColor: bar }} />
      <View style={{ height: 10, width: '85%', backgroundColor: bar, marginTop: 10 }} />
      <View style={{ height: 10, width: '40%', backgroundColor: bar, marginTop: 8 }} />
    </PixelBox>
  );
}
