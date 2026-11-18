import { describe, expect, test } from 'bun:test';
import { winBoardState } from './winBoardState';

const sprite = { width: 2, height: 1, pixels: ['ab'], palette: { a: '#000000', b: '#FFFFFF' } };
const win = (id: string, team: string, offer: string, partner: string, triggered: string, expires: string) => ({
  id,
  triggeredAt: triggered,
  expiresAt: expires,
  event: { teamId: team, teamName: `LA ${team}`, offerName: offer, partnerName: partner, sprite },
});
const teams = [
  { id: 'LAR', name: 'Rams', colors: { primary: '#003594', secondary: '#FFA300' } },
  { id: 'LAL', name: 'Lakers', colors: { primary: '#552583', secondary: '#FDB927' } },
];
const now = Date.parse('2026-11-17T17:00:00Z'); // 9am PT

describe('winBoardState', () => {
  test('no live wins: no card', () => {
    expect(winBoardState([], teams, now)).toBeNull();
  });

  test('one win: the team and offer lead, the deal opens on tap', () => {
    const s = winBoardState([win('te1', 'LAR', 'Free Famous Star', "Carl's Jr.", '2026-11-17T15:00:00Z', '2026-11-18T07:59:59Z')], teams, now)!;
    expect(s.title).toBe('RAMS WON!');
    expect(s.offer).toBe('Free Famous Star');
    expect(s.sub).toBe("Carl's Jr.");
    expect(s.detail).toBe('ENDS 11:59PM PT');
    expect(s.stripe).toEqual(['#003594', '#FFA300']);
    expect(s.url).toBe('freebie://deal/te1');
    expect(s.wins[0].sprite?.px).toEqual(['ab']);
  });

  test('several wins: one card, the count leads, the soonest deadline is next, WON opens on tap', () => {
    const s = winBoardState(
      [
        win('a', 'LAR', 'Free Famous Star', "Carl's Jr.", '2026-11-17T15:00:00Z', '2026-11-19T07:59:59Z'),
        win('b', 'LAL', '2 free tacos', 'Jack in the Box', '2026-11-17T16:00:00Z', '2026-11-18T07:59:59Z'),
      ],
      teams,
      now
    )!;
    expect(s.title).toBe('2 WINS!');
    expect(s.count).toBe(2);
    expect(s.offer).toBe('Next: 2 free tacos');
    expect(s.sub).toBe('Jack in the Box');
    expect(s.detail).toBe('ENDS 11:59PM PT');
    expect(s.stripe).toEqual(['#552583', '#003594']); // newest first
    expect(s.url).toBe('freebie://won');
    expect(s.staleAt).toBe(Date.parse('2026-11-18T07:59:59Z'));
  });

  test('at most 4 sprites ride on the card', () => {
    const many = Array.from({ length: 8 }, (_, i) => win(`w${i}`, 'LAR', 'X', 'Y', `2026-11-17T1${i}:00:00Z`, '2026-11-18T07:59:59Z'));
    expect(winBoardState(many, teams, now)!.wins).toHaveLength(4);
  });
});
