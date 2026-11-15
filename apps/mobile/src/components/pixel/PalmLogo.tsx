import React from 'react';
import { PixelArt } from './PixelArt';
import { LOGO_PALM } from './logoArt';

/** The app icon (pixel sunset palm) as 32x32 art, identical to the icon script's scene(). */
export function PalmLogo({ scale = 1 }: { scale?: number }) {
  return <PixelArt pixels={LOGO_PALM.pixels} palette={LOGO_PALM.palette} scale={scale} />;
}
