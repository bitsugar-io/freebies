// Victory cut-in timing and selection, free of native imports so bun tests can run them.
// docs/ui.md Motion: stepped frames, under 2 seconds, never more than one cut-in at a time.

/** Live wins not yet celebrated, newest first. */
export function winsToCelebrate<T extends { id: string; triggeredAt: string }>(
  live: readonly T[],
  celebrated: ReadonlySet<string>
): T[] {
  return live
    .filter((d) => !celebrated.has(d.id))
    .sort((a, b) => new Date(b.triggeredAt).getTime() - new Date(a.triggeredAt).getTime());
}

/** One frame of the cut-in and how long it holds, in ms. */
export type CutInFrame =
  | { kind: 'flash'; hold: number }
  | { kind: 'rays'; hold: number }
  | { kind: 'drop'; step: number; hold: number }
  | { kind: 'type'; text: string; hold: number }
  | { kind: 'shelf'; shown: number; hold: number };

/** Frame unit: 4 frames at 60 fps, the stepped hold the rules allow. */
const F4 = 67;
const F8 = 133;

/**
 * The cut-in's frames. One win: a team-color flash, rays snap in, the sprite drops in 4 steps, the
 * banner types out, then a hold. Several wins: rays, the banner, then one sprite onto the shelf per
 * step (8 at most), then a hold. Both stay under 2 seconds.
 */
export function cutInTimeline(wins: number, banner: string): CutInFrame[] {
  if (wins <= 1) {
    const frames: CutInFrame[] = [
      { kind: 'flash', hold: F8 },
      { kind: 'rays', hold: F8 },
      ...[0, 1, 2, 3].map((step) => ({ kind: 'drop' as const, step, hold: F4 })),
    ];
    // Type in pairs of characters so a long name ("CHARGERS WON!") still fits the budget.
    for (let i = 2; i < banner.length; i += 2) frames.push({ kind: 'type', text: banner.slice(0, i), hold: F4 });
    frames.push({ kind: 'type', text: banner, hold: 500 });
    return frames;
  }
  const shown = Math.min(wins, 8);
  return [
    { kind: 'rays', hold: 300 },
    ...Array.from({ length: shown }, (_, i) => ({ kind: 'shelf' as const, shown: i + 1, hold: F8 - 13 })),
    { kind: 'shelf', shown, hold: 400 },
  ];
}

/** Total length in ms. */
export function cutInDuration(frames: readonly CutInFrame[]): number {
  return frames.reduce((t, f) => t + f.hold, 0);
}
