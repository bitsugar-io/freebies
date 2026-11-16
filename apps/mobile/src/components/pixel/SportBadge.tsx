import React, { useMemo } from 'react';
import { PixelArt } from './PixelArt';
import { SPORT_BADGES } from './art';
import { toGray } from '../PixelSprite';
import PixelSprite from '../PixelSprite';
import type { Sprite } from '../../api/client';

interface Props {
  /** Sport id from the API ("hockey"). Unknown sports render the generic fallback sprite. */
  sport?: string;
  /** Badge art from the API. Preferred; the bundled art is only the offline fallback. */
  badge?: Pick<Sprite, 'pixels' | 'palette' | 'width' | 'height' | 'name'> | null;
  /** Whole points per art pixel. Default fits the badge into 32pt. */
  scale?: number;
  /** Grey by brightness (off-season). No opacity. */
  gray?: boolean;
}

/**
 * A team's sport badge. Art and team data come from the API; the bundled badges here only cover
 * an offline cold start for the sports we know, and anything else gets the generic sprite.
 * Decorative: the team and league are always shown as text.
 */
export function SportBadge({ sport, badge, scale, gray }: Props) {
  const bundled = sport && Object.prototype.hasOwnProperty.call(SPORT_BADGES, sport)
    ? SPORT_BADGES[sport as keyof typeof SPORT_BADGES]
    : null;
  const art = badge ?? bundled;
  const palette = useMemo(() => (art && gray ? toGray(art.palette) : art?.palette), [art, gray]);
  if (!art) return <PixelSprite sprite={null} scale={scale ?? 2} gray={gray} />;
  const s = scale ?? Math.max(1, Math.floor(32 / Math.max(1, art.width)));
  return <PixelArt pixels={art.pixels} palette={palette} scale={s} />;
}
