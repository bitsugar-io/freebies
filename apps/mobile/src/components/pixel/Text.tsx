import React from 'react';
import { Platform, Text as RNText, TextProps, TextStyle } from 'react-native';
import { colors, ColorToken, fonts, TextVariant, type, useLargeText } from '../../hooks/useTheme';

interface Props extends TextProps {
  variant?: TextVariant;
  color?: ColorToken;
  align?: TextStyle['textAlign'];
}

/**
 * The one text component. Display variants render uppercase in Press Start 2P and fall back to
 * the system bold face at accessibility text sizes. Body text is always the system font.
 */
export function Text({ variant = 'body', color = 'text', align, style, children, ...rest }: Props) {
  const t = type[variant];
  const large = useLargeText();
  const pixel = t.display && !large;
  let base: TextStyle = {
    color: colors[color],
    fontSize: t.fontSize,
    lineHeight: t.lineHeight,
    textAlign: align,
  };
  if (t.display && pixel) {
    base = {
      ...base,
      textTransform: 'uppercase',
      fontFamily: fonts.display,
      ...(Platform.OS === 'android' ? { includeFontPadding: false } : null),
    };
  } else if (t.display) {
    // Accessibility text sizes: system bold, slightly smaller since it is narrower than the pixel face.
    base = {
      ...base,
      textTransform: 'uppercase',
      fontWeight: '700',
      fontSize: Math.round(t.fontSize * 0.85),
      letterSpacing: 1,
    };
  } else if (variant === 'bodyStrong') {
    base = { ...base, fontWeight: '600' };
  }
  return (
    <RNText maxFontSizeMultiplier={t.maxScale} style={[base, style]} {...rest}>
      {children}
    </RNText>
  );
}
