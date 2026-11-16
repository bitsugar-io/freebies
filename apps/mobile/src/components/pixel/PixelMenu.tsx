import React from 'react';
import { AccessibilityRole, Pressable, StyleSheet, View, ViewProps } from 'react-native';
import { colors, ColorToken } from '../../hooks/useTheme';
import { useBlink } from '../../hooks/useBlink';
import { IconName } from './art';
import { PixelIcon } from './PixelIcon';
import { Text } from './Text';

export interface MenuRowProps {
  label: string;
  onPress?: () => void;
  /** The default action: shows a blinking ▶ (solid under Reduce Motion). */
  selected?: boolean;
  /** Right-aligned status or value, e.g. "ON". */
  value?: string;
  valueColor?: ColorToken;
  /** Blink priority for the ▶ (lower wins the blink budget). */
  blinkPriority?: number;
  /** A fixed icon in the cursor slot (e.g. a back arrow) when the cursor isn't showing. */
  leading?: IconName;
  tone?: 'default' | 'danger';
  disabled?: boolean;
  role?: AccessibilityRole;
  state?: { checked?: boolean | 'mixed'; selected?: boolean; disabled?: boolean };
  accessibilityLabel?: string;
  accessibilityHint?: string;
}

/** Full-width pause-menu row, 48pt minimum. The ▶ shows on the default row and while pressed. */
export function MenuRow({
  label,
  onPress,
  selected,
  value,
  valueColor = 'textSecondary',
  tone = 'default',
  leading,
  blinkPriority = 6,
  disabled,
  role = 'menuitem',
  state,
  accessibilityLabel,
  accessibilityHint,
}: MenuRowProps) {
  const blink = useBlink(blinkPriority, !!selected);
  const color: ColorToken = disabled ? 'textMuted' : tone === 'danger' ? 'danger' : selected ? 'accent' : 'text';
  const content = (showCursor: boolean) => (
    <View style={styles.row}>
      <View style={styles.cursor}>
        {showCursor ? (
          <PixelIcon name="cursor" scale={2} color={color} />
        ) : leading ? (
          <PixelIcon name={leading} scale={2} color={color} />
        ) : null}
      </View>
      <View style={{ flex: 1 }}>
        <Text variant="displayM" color={color}>
          {label}
        </Text>
      </View>
      {value ? (
        <Text variant="displayM" color={valueColor} style={styles.value}>
          {value}
        </Text>
      ) : null}
    </View>
  );
  if (!onPress) {
    return <View accessible accessibilityRole="text" accessibilityLabel={accessibilityLabel ?? (value ? `${label}, ${value}` : label)}>{content(false)}</View>;
  }
  return (
    <Pressable
      onPress={onPress}
      disabled={disabled}
      accessibilityRole={role}
      accessibilityLabel={accessibilityLabel ?? (value ? `${label}, ${value}` : label)}
      accessibilityHint={accessibilityHint}
      accessibilityState={{ disabled: !!disabled, ...state }}
    >
      {({ pressed }) => content(pressed || (!!selected && blink))}
    </Pressable>
  );
}

/** A group of rows with an optional pixel section label. */
export function PixelMenu({ title, children, onLayout }: { title?: string; children: React.ReactNode; onLayout?: ViewProps['onLayout'] }) {
  return (
    <View accessibilityRole="menu" onLayout={onLayout}>
      {title ? (
        <Text variant="displayM" color="textMuted" accessibilityRole="header" style={styles.title}>
          {title}
        </Text>
      ) : null}
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { minHeight: 48, flexDirection: 'row', alignItems: 'center', paddingRight: 4 },
  cursor: { width: 24, alignItems: 'flex-start' },
  value: { marginLeft: 8 },
  title: { marginTop: 8, marginBottom: 4, borderBottomWidth: 2, borderBottomColor: colors.surface, paddingBottom: 4 },
});
