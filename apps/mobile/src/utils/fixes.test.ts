import { describe, expect, test } from 'bun:test';
import { HttpError, isNetworkError, shouldDiscardStoredUser } from '../api/errors';
import { clockOffset, estimateOffset, observeServerDate, resetClock, serverNow } from './serverClock';
import { parseCache, parseUserCache } from './cache';
import { createSequencer, once, singleFlight } from './sequence';
import { liveWonDeals } from './dealState';
import { dealScreenPhase, resolveDeal } from './dealResolve';
import { isForDate, pacificDate } from './today';

const NOW = Date.parse('2026-11-16T15:00:00Z');
const later = '2026-11-16T20:00:00Z';
const earlier = '2026-11-16T10:00:00Z';

describe('stored user (offline must not wipe it)', () => {
  test('only 401, 403 and 404 discard the stored user', () => {
    expect(shouldDiscardStoredUser(new HttpError(404, 'user not found'))).toBe(true);
    expect(shouldDiscardStoredUser(new HttpError(401, 'invalid token'))).toBe(true);
    expect(shouldDiscardStoredUser(new HttpError(403, 'forbidden'))).toBe(true);
  });
  test('offline, timeouts and server errors keep it', () => {
    expect(shouldDiscardStoredUser(new Error('Network request failed'))).toBe(false);
    expect(shouldDiscardStoredUser(new Error('Request timed out'))).toBe(false);
    expect(shouldDiscardStoredUser(new HttpError(500, 'boom'))).toBe(false);
    expect(shouldDiscardStoredUser(new HttpError(503, 'down'))).toBe(false);
  });
  test('a timed-out GOT IT is a network error (safe to retry; the server is idempotent)', () => {
    expect(isNetworkError(new Error('Request timed out'))).toBe(true);
    expect(isNetworkError(new HttpError(409, 'already dismissed'))).toBe(false);
  });
});

describe('server clock', () => {
  test('a phone 10 minutes slow gets a +10 minute offset', () => {
    const device = NOW - 600_000;
    const header = new Date(NOW).toUTCString();
    const off = estimateOffset(header, device - 100, device + 100)!;
    expect(Math.abs(off - 600_000)).toBeLessThan(1_000);
  });
  test('missing or bad header: device time, unchanged offset', () => {
    resetClock();
    observeServerDate(null, 1, 2);
    observeServerDate('not a date', 1, 2);
    expect(clockOffset()).toBe(0);
    expect(estimateOffset(undefined, 1, 2)).toBeNull();
  });
  test('sub-2s differences are noise', () => {
    const header = new Date(NOW).toUTCString();
    expect(estimateOffset(header, NOW - 200, NOW + 200)).toBe(0);
  });
  test('serverNow applies the offset', () => {
    resetClock();
    const before = Date.now();
    observeServerDate(new Date(before + 3_600_000).toUTCString(), before, before);
    expect(serverNow() - Date.now()).toBeGreaterThan(3_590_000);
    resetClock();
  });
});

describe('caches never block the fetch', () => {
  test('corrupt JSON parses to null instead of throwing', () => {
    expect(parseCache('{"deals": [', () => true)).toBeNull();
    expect(parseCache('null', () => true)).toBeNull();
    expect(parseCache(null, () => true)).toBeNull();
  });
  test('wrong shape or another user: null', () => {
    const raw = JSON.stringify({ userId: 'u1', deals: [] });
    expect(parseUserCache(raw, 'u2', (v) => Array.isArray(v.deals))).toBeNull();
    expect(parseUserCache(raw, 'u1', (v) => Array.isArray(v.teams))).toBeNull();
    expect(parseUserCache<{ userId: string; deals: unknown[] }>(raw, 'u1', (v) => Array.isArray(v.deals))?.deals).toEqual([]);
  });
});

describe('out-of-order responses', () => {
  test('an older response that lands last is ignored', async () => {
    const seq = createSequencer();
    let shown = 'none';
    const fetch = (value: string, ms: number) => {
      const fresh = seq.begin();
      return new Promise<void>((r) => setTimeout(r, ms)).then(() => {
        if (fresh()) shown = value;
      });
    };
    await Promise.all([fetch('old', 30), fetch('new', 5)]);
    expect(shown).toBe('new');
  });
  test('a refetch started before GOT IT cannot undo it', async () => {
    const seq = createSequencer();
    let dismissed = false;
    const fresh = seq.begin(); // refetch starts: server still says not dismissed
    dismissed = true; // GOT IT succeeds locally
    seq.invalidate();
    if (fresh()) dismissed = false; // the stale response arrives
    expect(dismissed).toBe(true);
  });
});

describe('double taps', () => {
  test('one GOT IT in flight at a time', async () => {
    let calls = 0;
    const gotIt = singleFlight(async () => {
      calls++;
      await new Promise((r) => setTimeout(r, 10));
    });
    await Promise.all([gotIt(), gotIt(), gotIt()]);
    expect(calls).toBe(1);
    await gotIt();
    expect(calls).toBe(2);
  });
  test('close runs once', () => {
    let closes = 0;
    const close = once(() => closes++);
    close();
    close();
    expect(closes).toBe(1);
  });
});

describe('liveWonDeals (badge, stat line, list, stripe and launch agree)', () => {
  const deals = [
    { id: 'a', expiresAt: later, isDismissed: false },
    { id: 'b', expiresAt: earlier, isDismissed: false },
    { id: 'c', expiresAt: later, isDismissed: true },
    { id: 'd', isDismissed: false },
    { id: 'e', expiresAt: later, availableAt: later, isDismissed: false },
  ];
  test('not expired, not dismissed; locked and no-expiry count', () => {
    expect(liveWonDeals(deals, NOW).map((d) => d.id)).toEqual(['a', 'd', 'e']);
  });
});

describe('deal screen', () => {
  const mine = [
    { id: 'te-old', eventId: 'ev', triggeredAt: '2026-11-10T00:00:00Z', expiresAt: later, isDismissed: true },
    { id: 'te-new', eventId: 'ev', triggeredAt: '2026-11-15T00:00:00Z', expiresAt: later, isDismissed: false },
  ];
  const live = [{ id: 'te-pub', eventId: 'ev2', triggeredAt: '2026-11-16T00:00:00Z', expiresAt: later }];
  const events = [{ id: 'ev' }, { id: 'ev2' }];
  const base = { mine, live, fetchedTriggered: null, events, fetchedEvent: null };

  test('a triggered id opens that exact win', () => {
    expect(resolveDeal('te-old', base)).toEqual({ kind: 'mine', deal: mine[0] });
  });
  test('an event id opens your newest open win of it', () => {
    expect(resolveDeal('ev', base)).toEqual({ kind: 'mine', deal: mine[1] });
  });
  test('a TODAY live deal for a team you do not follow opens the triggered deal, not a preview', () => {
    expect(resolveDeal('te-pub', base)).toEqual({ kind: 'public', deal: live[0] });
  });
  test('a fetched triggered deal (cold-start push after an unfollow) resolves', () => {
    const fetched = { id: 'te-x', eventId: 'ev', triggeredAt: '2026-11-16T00:00:00Z' };
    expect(resolveDeal('te-x', { ...base, fetchedTriggered: fetched })).toEqual({ kind: 'public', deal: fetched });
  });
  test('an event id with no wins is the preview', () => {
    expect(resolveDeal('ev2', base)).toEqual({ kind: 'event', event: events[1] });
  });

  const phase = (o: Partial<Parameters<typeof dealScreenPhase>[0]>) =>
    dealScreenPhase({ resolved: null, dealsKnown: true, refetchDone: true, lookupsDone: true, timedOut: false, ...o });

  test('cold-start push: no "not available" or public view before your deals load', () => {
    expect(phase({ dealsKnown: false, refetchDone: false })).toBe('loading');
    expect(phase({ resolved: 'public', dealsKnown: false, refetchDone: false })).toBe('loading');
    expect(phase({ resolved: 'public', refetchDone: false })).toBe('loading');
  });
  test('a win of yours shows at once', () => {
    expect(phase({ resolved: 'mine', dealsKnown: false, refetchDone: false, lookupsDone: false })).toBe('show');
  });
  test('a preview shows once deals are known, without waiting on the refresh', () => {
    expect(phase({ resolved: 'event', refetchDone: false })).toBe('show');
  });
  test('gone only after deals, the refresh and the lookups answered, or the wait timed out', () => {
    expect(phase({ lookupsDone: false })).toBe('loading');
    expect(phase({})).toBe('missing');
    expect(phase({ dealsKnown: false, refetchDone: false, lookupsDone: false, timedOut: true })).toBe('missing');
    expect(phase({ resolved: 'public', dealsKnown: false, timedOut: true })).toBe('show');
  });
});

describe('TODAY overnight', () => {
  test('the Pacific date rolls at midnight PT, not UTC', () => {
    expect(pacificDate(new Date('2026-11-17T07:59:00Z'))).toBe('2026-11-16');
    expect(pacificDate(new Date('2026-11-17T08:00:00Z'))).toBe('2026-11-17');
  });
  test("yesterday's TODAY is not shown as today", () => {
    expect(isForDate({ date: '2026-11-16' }, '2026-11-17')).toBe(false);
    expect(isForDate({ date: '2026-11-17' }, '2026-11-17')).toBe(true);
    expect(isForDate(null, '2026-11-17')).toBe(false);
  });
});
