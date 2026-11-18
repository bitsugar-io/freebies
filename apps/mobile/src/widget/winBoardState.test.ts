import { describe, expect, test } from 'bun:test';
import { startsLabel, winBoardState } from './winBoardState';

const sprite = { width: 2, height: 1, pixels: ['ab'], palette: { a: '#000000', b: '#FFFFFF' } };
const win = (id: string, team: string, triggered: string, expires: string, available?: string) => ({
  id,
  triggeredAt: triggered,
  expiresAt: expires,
  availableAt: available,
  event: { teamId: team, teamName: `LA ${team}`, sprite },
});
const teams = [
  { id: 'LAR', name: 'Rams', colors: { primary: '#003594', secondary: '#FFA300' } },
  { id: 'LAL', name: 'Lakers', colors: { primary: '#552583', secondary: '#FDB927' } },
];
const now = Date.parse('2026-11-17T17:00:00Z'); // 9am PT
const s = (iso: string) => Date.parse(iso) / 1000;

describe('winBoardState', () => {
  test('no live wins: no card', () => {
    expect(winBoardState([], teams, now)).toBeNull();
  });

  test('one win: who won, the meal and its countdown; tap opens the deal', () => {
    const st = winBoardState([win('te1', 'LAR', '2026-11-17T15:00:00Z', '2026-11-18T07:59:59Z')], teams, now)!;
    expect(st.title).toBe('RAMS WON!');
    expect(st.meals).toHaveLength(1);
    expect(st.meals[0]).toMatchObject({ url: 'freebie://deal/te1', startsAt: 0, endsAt: s('2026-11-18T07:59:59Z') });
    expect(st.stripe).toEqual(['#003594', '#FFA300']);
    expect(st.url).toBe('freebie://deal/te1');
    expect(st.more).toBe(0);
  });

  test('several wins: the 3 that end soonest, each with its own deal link, then +N MORE', () => {
    const wins = [
      win('long', 'LAR', '2026-11-17T16:00:00Z', '2026-11-19T07:59:59Z'),
      win('soon', 'LAL', '2026-11-17T10:00:00Z', '2026-11-17T18:00:00Z'),
      win('mid', 'LAR', '2026-11-17T11:00:00Z', '2026-11-18T01:00:00Z'),
      win('later', 'LAL', '2026-11-17T12:00:00Z', '2026-11-18T07:59:59Z'),
      win('last', 'LAR', '2026-11-17T13:00:00Z', '2026-11-20T07:59:59Z'),
    ];
    const st = winBoardState(wins, teams, now)!;
    expect(st.title).toBe('5 WINS!');
    expect(st.meals.map((m) => m.url)).toEqual(['freebie://deal/soon', 'freebie://deal/mid', 'freebie://deal/later']);
    expect(st.more).toBe(2);
    expect(st.url).toBe('freebie://won');
    expect(st.staleAt).toBe(Date.parse('2026-11-17T18:00:00Z'));
  });

  test('a deal that has not started shows when it starts', () => {
    const st = winBoardState(
      [win('w', 'LAL', '2026-11-17T15:00:00Z', '2026-11-18T07:59:59Z', '2026-11-17T18:00:00Z')],
      teams,
      now
    )!;
    expect(st.meals[0].startsAt).toBe(s('2026-11-17T18:00:00Z'));
    expect(st.meals[0].startsLabel).toBe('10AM');
  });
});

describe('startsLabel', () => {
  test('today: the time; a later day: the weekday', () => {
    expect(startsLabel(Date.parse('2026-11-17T18:30:00Z'), now)).toBe('10:30AM');
    expect(startsLabel(Date.parse('2026-11-19T18:00:00Z'), now)).toBe('THU');
  });
});
