import { describe, expect, test } from 'bun:test';
import {
  followedTeamIds,
  countdownTarget,
  countdownTone,
  dealPhase,
  isLocked,
  ptStamp,
  unlockTag,
} from './dealState';
import { sportLabel, teamNameFrom } from './format';

// Sun Nov 15 2026, noon Pacific (UTC-8).
const NOW = Date.parse('2026-11-15T20:00:00Z');
const TUE = '2026-11-17T08:00:00Z'; // Tue Nov 17 12:00am PT
const END = '2026-11-24T08:00:00Z';

describe('locked state', () => {
  test('future availableAt is locked, past or missing is not', () => {
    expect(isLocked(TUE, NOW)).toBe(true);
    expect(isLocked('2026-11-15T19:00:00Z', NOW)).toBe(false);
    expect(isLocked(undefined, NOW)).toBe(false);
    expect(isLocked('not a date', NOW)).toBe(false);
  });

  test('phase: locked, then active when the clock passes availableAt, then expired', () => {
    const deal = { availableAt: TUE, expiresAt: END };
    expect(dealPhase(deal, NOW)).toBe('locked');
    expect(dealPhase(deal, Date.parse(TUE) - 1)).toBe('locked');
    expect(dealPhase(deal, Date.parse(TUE))).toBe('active');
    expect(dealPhase(deal, Date.parse(END))).toBe('expired');
  });

  test('a deal without availableAt is active as before', () => {
    expect(dealPhase({ expiresAt: END }, NOW)).toBe('active');
  });

  test('expired beats locked', () => {
    expect(dealPhase({ availableAt: TUE, expiresAt: '2026-11-15T10:00:00Z' }, NOW)).toBe('expired');
  });
});

describe('unlock labels', () => {
  test('uses the Pacific day name within 6 days', () => {
    expect(unlockTag(TUE, NOW)).toBe('CODE UNLOCKS TUE');
  });
  test('uses the date when more than 6 days away', () => {
    expect(unlockTag('2026-11-24T08:00:00Z', NOW)).toBe('CODE UNLOCKS NOV 24');
  });
  test('exactly 6 days still uses the day name', () => {
    expect(unlockTag('2026-11-21T20:00:00Z', NOW)).toBe('CODE UNLOCKS SAT');
  });
  test('same Pacific day reads TODAY', () => {
    expect(unlockTag('2026-11-16T05:00:00Z', NOW)).toBe('CODE UNLOCKS TODAY');
  });
  test('the day is the Pacific day, not the UTC day', () => {
    // 2026-11-18T03:00Z is still Tue Nov 17 at 7pm PT.
    expect(unlockTag('2026-11-18T03:00:00Z', NOW)).toBe('CODE UNLOCKS TUE');
  });
  test('stamp is Pacific wall clock', () => {
    expect(ptStamp(TUE)).toBe('Tue Nov 17 at 12:00am PT');
  });
});

describe('countdown', () => {
  test('counts to the unlock while locked, to the end otherwise', () => {
    expect(countdownTarget({ availableAt: TUE, expiresAt: END }, NOW)).toEqual({ to: TUE, kind: 'unlock' });
    expect(countdownTarget({ availableAt: TUE, expiresAt: END }, Date.parse(TUE))).toEqual({ to: END, kind: 'expiry' });
    expect(countdownTarget({}, NOW)).toBeNull();
  });
  test('timer color only under 6 hours', () => {
    expect(countdownTone(5 * 3_600_000)).toBe('warning');
    expect(countdownTone(6 * 3_600_000)).toBe('warning');
    expect(countdownTone(6 * 3_600_000 + 1)).toBe('text');
    expect(countdownTone(0)).toBe('text');
  });
});

describe('followed teams', () => {
  const events = [
    { id: 'ana-mcd', teamId: 'ANA', isActive: true },
    { id: 'ana-cfa', teamId: 'ANA', isActive: true },
    { id: 'lak-mcd', teamId: 'LAK', isActive: true },
    { id: 'lar-old', teamId: 'LAR', isActive: false },
  ];
  test('a team with two followed deals counts once', () => {
    const subs = new Set(['ana-mcd', 'ana-cfa']);
    expect(followedTeamIds(events, (id) => subs.has(id))).toEqual(['ANA']);
  });
  test('teams, not deals: Ducks (2 deals) + Kings (1) is 2', () => {
    const subs = new Set(['ana-mcd', 'ana-cfa', 'lak-mcd']);
    expect(followedTeamIds(events, (id) => subs.has(id))).toEqual(['ANA', 'LAK']);
  });
  test('retired deals and unknown ids never count', () => {
    const subs = new Set(['lar-old', 'gone']);
    expect(followedTeamIds(events, (id) => subs.has(id))).toEqual([]);
  });
});

describe('names from the API', () => {
  test('team name from the list, else the fallback', () => {
    expect(teamNameFrom([{ id: 'LAD', name: 'Dodgers' }], 'LAD', 'Los Angeles Dodgers')).toBe('Dodgers');
    expect(teamNameFrom([], 'LAD', 'Los Angeles Dodgers')).toBe('Los Angeles Dodgers');
  });
  test('sport label', () => {
    expect(sportLabel('hockey')).toBe('HOCKEY');
    expect(sportLabel('womens-soccer')).toBe('WOMENS SOCCER');
    expect(sportLabel('hockey', 'Ice Hockey')).toBe('ICE HOCKEY');
    expect(sportLabel('')).toBe('ALL TEAMS');
  });
});
