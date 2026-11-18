import { describe, expect, test } from 'bun:test';
import { cutInDuration, cutInTimeline, winsToCelebrate } from './cutin';

describe('winsToCelebrate', () => {
  const live = [
    { id: 'a', triggeredAt: '2026-11-16T15:00:00Z' },
    { id: 'b', triggeredAt: '2026-11-17T15:00:00Z' },
    { id: 'c', triggeredAt: '2026-11-15T15:00:00Z' },
  ];
  test('only wins not celebrated yet, newest first', () => {
    expect(winsToCelebrate(live, new Set(['c'])).map((d) => d.id)).toEqual(['b', 'a']);
  });
  test('nothing new: nothing to play', () => {
    expect(winsToCelebrate(live, new Set(['a', 'b', 'c']))).toEqual([]);
  });
});

describe('cutInTimeline', () => {
  test('one win: flash, rays, a 4-step drop, then the banner types out to the full text', () => {
    const f = cutInTimeline(1, 'RAMS WON!');
    expect(f[0].kind).toBe('flash');
    expect(f[1].kind).toBe('rays');
    expect(f.filter((x) => x.kind === 'drop')).toHaveLength(4);
    const last = f[f.length - 1];
    expect(last.kind === 'type' && last.text).toBe('RAMS WON!');
  });
  test('several wins: one shelf frame per win, 8 at most', () => {
    const f = cutInTimeline(12, '12 WINS!');
    expect(f.filter((x) => x.kind === 'shelf' && x.hold < 200)).toHaveLength(8);
  });
  test('always under 2 seconds, even for a long team name or many wins', () => {
    expect(cutInDuration(cutInTimeline(1, 'CHARGERS WON!'))).toBeLessThan(2000);
    expect(cutInDuration(cutInTimeline(8, '8 WINS!'))).toBeLessThan(2000);
  });
});
