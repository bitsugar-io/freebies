import React from 'react';
import { StyleProp, StyleSheet, View, ViewProps, ViewStyle } from 'react-native';
import { colors, PX } from '../../hooks/useTheme';

export type BoxVariant = 'panel' | 'dialog' | 'selected' | 'partial' | 'off' | 'alert' | 'notice';

const VARIANTS: Record<BoxVariant, { fill: string; border: string }> = {
  panel: { fill: colors.surface, border: colors.border },
  dialog: { fill: colors.background, border: colors.border },
  selected: { fill: colors.surfaceSecondary, border: colors.accent },
  partial: { fill: colors.surface, border: colors.accent },
  off: { fill: colors.background, border: colors.textMuted },
  alert: { fill: colors.background, border: colors.danger },
  /** Toasts: a solid lilac box with a white border, so it stands out on the black screen. */
  notice: { fill: colors.info, border: colors.text },
};

interface Props extends ViewProps {
  variant?: BoxVariant;
  /** Inner padding in points, in addition to the border. */
  padding?: number;
  /** Override the variant's fill / border with theme colors (e.g. season fill, follow border). */
  fill?: string;
  borderColor?: string;
  /** Border thickness in points. Keep a whole multiple of PX. */
  border?: number;
  style?: StyleProp<ViewStyle>;
}

/**
 * Panel with a pixel border and notched corners. Built from bars, never borderRadius:
 * the fill is two overlapping rects (so the corner squares stay empty) with border bars on top.
 */
export function PixelBox({
  variant = 'panel',
  padding = 12,
  fill,
  borderColor,
  border = PX,
  style,
  children,
  ...rest
}: Props) {
  const v = { ...VARIANTS[variant], ...(fill ? { fill } : null), ...(borderColor ? { border: borderColor } : null) };
  const b = border;
  return (
    <View style={[{ padding: padding + b }, style]} {...rest}>
      <View style={StyleSheet.absoluteFill} pointerEvents="none" importantForAccessibility="no-hide-descendants">
        <View style={[styles.abs, { left: 0, right: 0, top: b, bottom: b, backgroundColor: v.fill }]} />
        <View style={[styles.abs, { left: b, right: b, top: 0, bottom: 0, backgroundColor: v.fill }]} />
        <View style={[styles.abs, { left: b, right: b, top: 0, height: b, backgroundColor: v.border }]} />
        <View style={[styles.abs, { left: b, right: b, bottom: 0, height: b, backgroundColor: v.border }]} />
        <View style={[styles.abs, { left: 0, width: b, top: b, bottom: b, backgroundColor: v.border }]} />
        <View style={[styles.abs, { right: 0, width: b, top: b, bottom: b, backgroundColor: v.border }]} />
      </View>
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  abs: { position: 'absolute' },
});
