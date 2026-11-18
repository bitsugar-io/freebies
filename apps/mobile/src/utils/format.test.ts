import { describe, expect, test } from 'bun:test';
import { deadlineLabel, formatCountdown } from './format';

describe('formatCountdown', () => {
  test('under a day: a ticking clock', () => {
    expect(formatCountdown((3 * 3600 + 28 * 60 + 11) * 1000)).toBe('03:28:11');
    expect(formatCountdown((23 * 3600 + 59 * 60 + 59) * 1000)).toBe('23:59:59');
  });
  test('a day or more: days and hours, never 152 hours', () => {
    expect(formatCountdown((152 * 3600 + 19 * 60) * 1000)).toBe('6D 08H');
    expect(formatCountdown(24 * 3600 * 1000)).toBe('1D 00H');
  });
  test('never negative', () => {
    expect(formatCountdown(-5000)).toBe('00:00:00');
  });
});

describe('deadlineLabel', () => {
  const MON_3PM = Date.parse('2026-11-16T23:00:00Z'); // Mon Nov 16, 3pm PT
  test('today: the time only', () => {
    expect(deadlineLabel('2026-11-17T07:59:59Z', MON_3PM)).toBe('11:59pm PT');
  });
  test('a later day: the day too, so a week-long deal never reads as tonight', () => {
    expect(deadlineLabel('2026-11-23T07:59:59Z', MON_3PM)).toBe('Sun Nov 22, 11:59pm PT');
  });
});

import { formatTimeLeft, spokenTimeLeft } from './timeLeft';

describe('formatTimeLeft', () => {
  const H = 3_600_000;
  test('between one and two days reads as a day and hours, not 31h', () => {
    expect(formatTimeLeft(31 * H + 47 * 60_000)).toBe('1 day 7h left');
    expect(spokenTimeLeft(31 * H)).toBe('1 day 7 hours left');
  });
  test('other ranges are unchanged', () => {
    expect(formatTimeLeft(7 * H + 47 * 60_000)).toBe('7h 47m left');
    expect(formatTimeLeft(50 * H)).toBe('2 days left');
    expect(formatTimeLeft(5 * 60_000)).toBe('5m left');
  });
});
