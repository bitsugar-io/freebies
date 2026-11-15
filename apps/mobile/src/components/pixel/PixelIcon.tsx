import React from 'react';
import { colors, ColorToken } from '../../hooks/useTheme';
import { ICONS, IconName } from './art';
import { PixelArt } from './PixelArt';

interface Props {
  name: IconName;
  scale?: number;
  /** Token name or a raw color from a navigator (tab bar tint). */
  color?: ColorToken | string;
  accessibilityLabel?: string;
}

/** Single-color pixel icon. Decorative unless a label is given. */
export function PixelIcon({ name, scale = 2, color = 'text', accessibilityLabel }: Props) {
  const tint = color in colors ? colors[color as ColorToken] : color;
  return (
    <PixelArt
      pixels={ICONS[name]}
      color={tint}
      scale={scale}
      accessibilityLabel={accessibilityLabel}
    />
  );
}
