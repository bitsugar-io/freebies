import React from 'react';
import { View } from 'react-native';
import { PixelArt } from './PixelArt';
import { LOGO_ICON } from './logoArt';

/** iOS app-icon corner radius as a fraction of the side. */
const ICON_RADIUS = 0.2237;

/**
 * The app icon (Trophy Burger over victory rays) as 32x32 art, identical to the icon script's
 * scene(). `rounded` clips it to the iOS icon shape where it stands for the app icon (the
 * maintenance screen); the one place pixel UI uses a radius (docs/ui.md, Logo art).
 */
export function AppLogo({ scale = 1, rounded = false }: { scale?: number; rounded?: boolean }) {
  const art = <PixelArt pixels={LOGO_ICON.pixels} palette={LOGO_ICON.palette} scale={scale} />;
  if (!rounded) return art;
  const side = LOGO_ICON.width * scale;
  return <View style={{ width: side, height: side, borderRadius: side * ICON_RADIUS, overflow: 'hidden' }}>{art}</View>;
}
