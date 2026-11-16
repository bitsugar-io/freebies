import React, { memo } from 'react';
import { StyleSheet, View } from 'react-native';
import Svg, { Polygon } from 'react-native-svg';
import { colors } from '../../hooks/useTheme';

const RAYS = 24;
const VIEW_W = 300;
const VIEW_H = 160;

/**
 * A hard-edged victory burst behind the YOU WON sprite, in the winning team's two colors (from the
 * API). Static: no motion, so nothing to turn off under Reduce Motion. Decorative only.
 */
export const VictoryRays = memo(function VictoryRays({
  pair,
  children,
}: {
  /** The team's colors; null falls back to gold and black. */
  pair: [string, string] | null;
  children: React.ReactNode;
}) {
  const [a, b] = pair ?? [colors.accent, colors.background];
  const cx = VIEW_W / 2;
  const cy = VIEW_H / 2;
  const r = VIEW_W; // long enough to reach every corner
  const wedges = Array.from({ length: RAYS }, (_, i) => {
    const t0 = (i / RAYS) * 2 * Math.PI;
    const t1 = ((i + 1) / RAYS) * 2 * Math.PI;
    const pts = `${cx},${cy} ${cx + r * Math.cos(t0)},${cy + r * Math.sin(t0)} ${cx + r * Math.cos(t1)},${cy + r * Math.sin(t1)}`;
    return <Polygon key={i} points={pts} fill={i % 2 === 0 ? a : b} />;
  });
  return (
    <View style={styles.box}>
      <View style={StyleSheet.absoluteFill} accessibilityElementsHidden importantForAccessibility="no-hide-descendants">
        <Svg width="100%" height="100%" viewBox={`0 0 ${VIEW_W} ${VIEW_H}`} preserveAspectRatio="xMidYMid slice">
          {wedges}
        </Svg>
      </View>
      {children}
    </View>
  );
});

const styles = StyleSheet.create({
  box: {
    height: 160,
    marginHorizontal: -16,
    marginBottom: 16,
    alignItems: 'center',
    justifyContent: 'center',
    overflow: 'hidden',
    borderBottomWidth: 4,
    borderBottomColor: colors.border,
  },
});
