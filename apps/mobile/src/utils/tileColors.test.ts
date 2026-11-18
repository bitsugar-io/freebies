import { describe, expect, test } from 'bun:test';
import { contrastOnBlack, isNearBlack, tilePair } from './tileColors';

describe('tile colors', () => {
  test('black is near black, navy and other dark colors are not', () => {
    expect(isNearBlack('#000000')).toBe(true);
    expect(isNearBlack('#0000BC')).toBe(false); // Galaxy, Rams
    expect(isNearBlack('#A81000')).toBe(false);
    expect(contrastOnBlack('#0000BC')).toBeGreaterThan(1.4);
  });
  test('swaps for Kings, LAFC and Angel City, not for Galaxy or Angels', () => {
    expect(tilePair('#000000', '#BCBCBC')).toEqual({ field: '#BCBCBC', stripe: '#000000' });
    expect(tilePair('#000000', '#F8B800')).toEqual({ field: '#F8B800', stripe: '#000000' });
    expect(tilePair('#0000BC', '#F8B800')).toEqual({ field: '#0000BC', stripe: '#F8B800' });
    expect(tilePair('#F83800', '#0000BC')).toEqual({ field: '#F83800', stripe: '#0000BC' });
  });
  test('keeps both colors when both are black', () => {
    expect(tilePair('#000000', '#000000')).toEqual({ field: '#000000', stripe: '#000000' });
  });
  test('a black stripe uses the accent so the middle line shows (Kings, Ducks, LAFC, Angel City)', () => {
    expect(tilePair('#111111', '#A2AAAD', '#FFFFFF')).toEqual({ field: '#A2AAAD', stripe: '#FFFFFF' });
    expect(tilePair('#FC4C02', '#000000', '#B9975B')).toEqual({ field: '#FC4C02', stripe: '#B9975B' });
    expect(tilePair('#000000', '#C39E6D', '#E31837')).toEqual({ field: '#C39E6D', stripe: '#E31837' });
  });
  test('no accent: the stripe stays as it was; a visible stripe ignores the accent', () => {
    expect(tilePair('#000000', '#C39E6D')).toEqual({ field: '#C39E6D', stripe: '#000000' });
    expect(tilePair('#552583', '#FDB927', '#FFFFFF')).toEqual({ field: '#552583', stripe: '#FDB927' });
  });
});
