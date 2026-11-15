// Team tile color choices, free of native imports so bun tests can run them.

function channel(v: number): number {
  const c = v / 255;
  return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
}

/** WCAG relative luminance of #RRGGBB. */
export function relativeLuminance(hex: string): number {
  const n = parseInt(hex.slice(1), 16);
  return 0.2126 * channel(n >> 16) + 0.7152 * channel((n >> 8) & 255) + 0.0722 * channel(n & 255);
}

/** WCAG contrast ratio against the black screen (#000000). */
export function contrastOnBlack(hex: string): number {
  return (relativeLuminance(hex) + 0.05) / 0.05;
}

/** Effectively black: it would vanish into the screen. Navy and other dark-but-visible colors are not. */
export function isNearBlack(hex: string): boolean {
  return contrastOnBlack(hex) < 1.4;
}

/** Field and stripe colors for a tile: swap only when the primary is near black and the secondary is not. */
export function tilePair(primary: string, secondary: string): { field: string; stripe: string } {
  if (isNearBlack(primary) && !isNearBlack(secondary)) return { field: secondary, stripe: primary };
  return { field: primary, stripe: secondary };
}
