import React from 'react';
import { Pressable, StyleProp, View, ViewStyle } from 'react-native';
import { colors, PX } from '../../hooks/useTheme';
import { Text } from './Text';

interface Props {
  label: string;
  onPress: () => void;
  disabled?: boolean;
  accessibilityLabel?: string;
  accessibilityHint?: string;
  style?: StyleProp<ViewStyle>;
}

/** Primary action: accent fill, black label. Press moves it down-right by one art pixel, instantly. */
export function PixelButton({
  label,
  onPress,
  disabled,
  accessibilityLabel,
  accessibilityHint,
  style,
}: Props) {
  return (
    <Pressable
      onPress={onPress}
      disabled={disabled}
      accessibilityRole="button"
      accessibilityLabel={accessibilityLabel ?? label}
      accessibilityHint={accessibilityHint}
      accessibilityState={{ disabled: !!disabled }}
      style={style}
    >
      {({ pressed }) => (
        <View style={{ paddingRight: PX, paddingBottom: PX }}>
          {/* The drop: a dark-gold offset the button covers while pressed. */}
          <View
            style={{
              position: 'absolute',
              left: PX,
              top: PX,
              right: 0,
              bottom: 0,
              backgroundColor: colors.accentShadow,
            }}
          />
          <View style={{ transform: pressed ? [{ translateX: PX }, { translateY: PX }] : [] }}>
            <View
              style={{
                backgroundColor: disabled ? colors.textMuted : colors.accent,
                minHeight: 48,
                justifyContent: 'center',
                alignItems: 'center',
                paddingHorizontal: 12,
                borderWidth: PX,
                borderColor: colors.border,
              }}
            >
              <Text variant="displayM" color="onAccent" align="center">
                {label}
              </Text>
            </View>
          </View>
        </View>
      )}
    </Pressable>
  );
}
