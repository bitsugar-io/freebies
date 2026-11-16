import React, { memo, useMemo } from 'react';
import Svg, { Rect } from 'react-native-svg';

export interface Run {
  x: number;
  y: number;
  w: number;
  color: string;
}

/** Merge each row's consecutive same-color pixels into one rectangle. */
export function bitmapRuns(
  pixels: readonly string[],
  colorOf: (key: string) => string | undefined
): Run[] {
  const runs: Run[] = [];
  pixels.forEach((row, y) => {
    let x = 0;
    while (x < row.length) {
      const key = row[x];
      let end = x + 1;
      while (end < row.length && row[end] === key) end++;
      const color = key === '.' ? undefined : colorOf(key);
      if (color) runs.push({ x, y, w: end - x, color });
      x = end;
    }
  });
  return runs;
}

interface Props {
  pixels: readonly string[];
  /** Palette key -> color. */
  palette?: Record<string, string>;
  /** Single color for every non-'.' pixel (icons). */
  color?: string;
  /** Whole points per art pixel. Integer so pixels stay square and crisp. */
  scale: number;
  accessibilityLabel?: string;
}

/** Draws a bitmap as run-merged SVG rects at an integer scale. */
function PixelArtBase({ pixels, palette, color, scale, accessibilityLabel }: Props) {
  const runs = useMemo(
    () => bitmapRuns(pixels, (k) => (color ? color : palette?.[k])),
    [pixels, palette, color]
  );
  const width = pixels.reduce((m, r) => Math.max(m, r.length), 0);
  const s = Math.max(1, Math.round(scale));
  return (
    <Svg
      width={width * s}
      height={pixels.length * s}
      viewBox={`0 0 ${width} ${pixels.length}`}
      accessible={accessibilityLabel !== undefined}
      accessibilityRole={accessibilityLabel !== undefined ? 'image' : undefined}
      accessibilityLabel={accessibilityLabel}
    >
      {runs.map((r) => (
        // Slight overdraw hides hairline seams between rects on some devices.
        <Rect key={`${r.x}-${r.y}`} x={r.x} y={r.y} width={r.w + 0.02} height={1.02} fill={r.color} />
      ))}
    </Svg>
  );
}

export const PixelArt = memo(PixelArtBase);
