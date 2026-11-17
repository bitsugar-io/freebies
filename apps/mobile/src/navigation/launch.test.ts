import { describe, expect, test } from 'bun:test';
import { getStateFromPath } from '@react-navigation/core';
import { LINK_CONFIG } from './linkConfig';
import { launchCorrection, launchDataSettled, launchTab, navSnapshot, urlFromPushData } from './launch';

const NOW = Date.parse('2026-11-16T15:00:00Z');
const later = '2026-11-17T07:59:59Z';
const earlier = '2026-11-16T07:59:59Z';

describe('launchTab', () => {
  test('first open, no teams followed: TEAMS', () => {
    expect(launchTab(0, [], NOW)).toBe('Teams');
  });
  test('no teams followed wins over an old live deal: TEAMS', () => {
    expect(launchTab(0, [{ expiresAt: later }], NOW)).toBe('Teams');
  });
  test('teams followed, nothing won: TODAY', () => {
    expect(launchTab(2, [], NOW)).toBe('Today');
  });
  test('teams followed, a win you can still claim: WON', () => {
    expect(launchTab(2, [{ expiresAt: later }], NOW)).toBe('Won');
  });
  test('teams followed, only expired wins: TODAY', () => {
    expect(launchTab(2, [{ expiresAt: earlier }], NOW)).toBe('Today');
  });
  test('a win with no expiry counts as live: WON', () => {
    expect(launchTab(1, [{}], NOW)).toBe('Won');
  });
});

/** The leaf route names and params a URL opens. */
function opens(url: string) {
  const path = url.replace(/^freebie:\/\//, '');
  let state: any = getStateFromPath(path, LINK_CONFIG);
  const names: string[] = [];
  let params: unknown;
  while (state) {
    const route = state.routes[state.index ?? state.routes.length - 1];
    names.push(route.name);
    params = route.params;
    state = route.state;
  }
  return { names, params };
}

describe('push taps', () => {
  // The exact data the worker sends (services/api/internal/worker/service.go): trigger and reminder.
  const serverPush = { triggeredEventId: 'te-123', eventId: 'event-nfl-lar-carlsjr', url: 'freebie://deal/te-123' };

  test('a "you won" or reminder push opens YOU WON for that win', () => {
    const url = urlFromPushData(serverPush)!;
    expect(url).toBe('freebie://deal/te-123');
    expect(opens(url)).toEqual({ names: ['Deal'], params: { id: 'te-123' } });
  });
  test('an older push with only ids still opens the deal', () => {
    expect(urlFromPushData({ triggeredEventId: 'te-9' })).toBe('freebie://deal/te-9');
    expect(urlFromPushData({ eventId: 'event-x' })).toBe('freebie://deal/event-x');
  });
  test('a push with no usable data opens nothing (the app opens normally)', () => {
    expect(urlFromPushData(undefined)).toBeNull();
    expect(urlFromPushData({})).toBeNull();
    expect(urlFromPushData({ url: '' })).toBeNull();
  });
});

describe('launchTab uses the shared live-win rule', () => {
  test('a claimed or silenced win does not open WON', () => {
    expect(launchTab(1, [{ expiresAt: later, isDismissed: true }], NOW)).toBe('Today');
  });
});

describe('launchCorrection (data arrived after the splash gave up)', () => {
  const base = { initial: 'Teams' as const, decidedWithData: false, dataReady: true, touched: false, correct: 'Won' as const };
  test('a follower opened on TEAMS before follows loaded moves once to WON', () => {
    expect(launchCorrection(base)).toBe('Won');
  });
  test('a winner opened on TODAY before deals loaded moves to WON', () => {
    expect(launchCorrection({ ...base, initial: 'Today' })).toBe('Won');
  });
  test('never moves a user who navigated or opened a link', () => {
    expect(launchCorrection({ ...base, touched: true })).toBeNull();
  });
  test('no move when the first choice used complete data, data is still loading, or it was right', () => {
    expect(launchCorrection({ ...base, decidedWithData: true })).toBeNull();
    expect(launchCorrection({ ...base, dataReady: false })).toBeNull();
    expect(launchCorrection({ ...base, correct: 'Teams' })).toBeNull();
  });
});

describe('navSnapshot (did the user move since launch?)', () => {
  const tabs = (index: number, teamsState?: { index: number; routes: { name: string }[] }) => ({
    index,
    routes: [{ name: 'Won' }, { name: 'Today' }, { name: 'Teams', state: teamsState }, { name: 'You' }],
  });
  test('a nested stack filling in its first screen is not a move', () => {
    const before = navSnapshot({ routes: [{ name: 'Tabs', state: tabs(2) }] });
    const after = navSnapshot({ routes: [{ name: 'Tabs', state: tabs(2, { index: 0, routes: [{ name: 'Sports' }] }) }] });
    expect(after).toBe(before);
  });
  test('a tab change, a pushed sport or a deal opened by a push is a move', () => {
    const start = navSnapshot({ routes: [{ name: 'Tabs', state: tabs(2) }] });
    expect(navSnapshot({ routes: [{ name: 'Tabs', state: tabs(1) }] })).not.toBe(start);
    expect(
      navSnapshot({ routes: [{ name: 'Tabs', state: tabs(2, { index: 1, routes: [{ name: 'Sports' }, { name: 'Sport' }] }) }] })
    ).not.toBe(start);
    expect(navSnapshot({ routes: [{ name: 'Tabs', state: tabs(2) }, { name: 'Deal' }] })).not.toBe(start);
  });
});

describe('launchDataSettled', () => {
  test('saved deals only (fresh list still coming): not settled, so a new win can still move to WON', () => {
    expect(launchDataSettled({ loading: false, dealsStale: true, dealsFailed: false })).toBe(false);
  });
  test('fresh deals from the server: settled', () => {
    expect(launchDataSettled({ loading: false, dealsStale: false, dealsFailed: false })).toBe(true);
  });
  test('offline (fetch failed): the saved list is final', () => {
    expect(launchDataSettled({ loading: false, dealsStale: true, dealsFailed: true })).toBe(true);
  });
  test('still loading: not settled', () => {
    expect(launchDataSettled({ loading: true, dealsStale: false, dealsFailed: false })).toBe(false);
  });
});
