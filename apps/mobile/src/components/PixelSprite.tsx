import React, { memo, useMemo } from 'react';

import type { Sprite } from '../api/client';
import { PLATE } from './pixel/art';
import { PixelArt } from './pixel/PixelArt';

interface Props {
  /** Sprite from the API. Missing sprites render the bundled plate, never emoji. */
  sprite?: Pick<Sprite, 'pixels' | 'palette' | 'width' | 'height' | 'name'> | null;
  /** Whole points per art pixel (integer). Preferred over `size`. */
  scale?: number;
  /** Convenience: largest whole scale that fits this many points. Ignored when `scale` is set. */
  size?: number;
  /** Render in greys (expired or unavailable). Colors map by brightness, no opacity on art. */
  gray?: boolean;
  /** Omit when a text label sits next to the sprite; the sprite is then decorative. */
  accessibilityLabel?: string;
}

const GREYS = ['#7C7C7C', '#BCBCBC', '#FCFCFC'];

/** One color to a grey by brightness (black, #7C7C7C, #BCBCBC or white). */
export function grayHex(hex: string): string {
  const n = parseInt(hex.slice(1), 16);
  const lum = (0.299 * (n >> 16) + 0.587 * ((n >> 8) & 255) + 0.114 * (n & 255)) / 255;
  return lum < 0.1 ? '#000000' : GREYS[Math.min(2, Math.floor(lum * 3))];
}

export function toGray(palette: Record<string, string>): Record<string, string> {
  const out: Record<string, string> = {};
  for (const [k, hex] of Object.entries(palette)) out[k] = grayHex(hex);
  return out;
}

function PixelSprite({ sprite, scale, size = 48, gray, accessibilityLabel }: Props) {
  const art = sprite ?? PLATE;
  const palette = useMemo(() => (gray ? toGray(art.palette) : art.palette), [art, gray]);
  const s = scale ?? Math.max(1, Math.floor(size / Math.max(art.width, art.height)));
  return (
    <PixelArt
      pixels={art.pixels}
      palette={palette}
      scale={s}
      accessibilityLabel={accessibilityLabel}
    />
  );
}

export default memo(PixelSprite);
