import React from 'react';
import { PixelArt } from './PixelArt';
import { LOGO_ICON } from './logoArt';

/** The app icon (Trophy Burger over victory rays) as 32x32 art, identical to the icon script's scene(). */
export function AppLogo({ scale = 1 }: { scale?: number }) {
  return <PixelArt pixels={LOGO_ICON.pixels} palette={LOGO_ICON.palette} scale={scale} />;
}
