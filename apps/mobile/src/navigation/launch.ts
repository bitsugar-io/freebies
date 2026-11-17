/**
 * Pure navigation rules, free of native imports so plain unit tests cover them.
 */

import { liveWonDeals } from '../utils/dealState';

/** The tab the app opens on (owner rule, 2026-11-15). Links and push taps still open their own screen. */
export type LaunchTab = 'Teams' | 'Won' | 'Today';

/**
 * No teams followed yet: TEAMS, to pick some. A live win (see liveWonDeals: not expired, not
 * claimed or silenced): WON. Otherwise: TODAY. `now` is server-corrected time.
 */
export function launchTab(
  followedCount: number,
  deals: readonly { expiresAt?: string | null; isDismissed?: boolean }[],
  now: number
): LaunchTab {
  if (followedCount === 0) return 'Teams';
  return liveWonDeals(deals, now).length > 0 ? 'Won' : 'Today';
}

/**
 * The launch tab may be picked before the follows and deals load (the splash gives up at 2.5s).
 * When they do load, move once to the right tab, but only if the user is still sitting where the
 * app put them: never after they changed tabs or a link or push opened a screen.
 */
export function launchCorrection(s: {
  /** The tab the app opened on. */
  initial: LaunchTab;
  /** True when that choice was made from complete data (nothing to correct). */
  decidedWithData: boolean;
  /** True once the user, follows and deals are known. */
  dataReady: boolean;
  /** True once the user navigated or a link opened a screen since launch. */
  touched: boolean;
  /** The tab that is right now that the data is in. */
  correct: LaunchTab;
}): LaunchTab | null {
  if (s.decidedWithData || !s.dataReady || s.touched) return null;
  return s.correct === s.initial ? null : s.correct;
}

/**
 * True once the launch tab can be trusted: everything loaded and the deals list is fresh from the
 * server, or its fetch failed (offline: the saved list is the best there is). Saved deals alone are
 * not enough: yesterday's list has no win from last night, and opening on TODAY then would hide it.
 */
export function launchDataSettled(s: { loading: boolean; dealsStale: boolean; dealsFailed: boolean }): boolean {
  return !s.loading && (!s.dealsStale || s.dealsFailed);
}

/** Minimal navigation state shape (React Navigation's, without the import). */
interface NavState {
  index?: number;
  routes: { name: string; state?: NavState }[];
}

/**
 * Where the user is, at the level a launch correction cares about: the screens over the tabs (a
 * deal opened by a link or push), the focused tab and whether its stack went deeper. Nested state
 * a navigator fills in on its own never changes it, so only a real move (a tap or a link) does.
 */
export function navSnapshot(root: NavState | undefined): string {
  if (!root) return '';
  const tabs = root.routes[0]?.state;
  const tab = tabs ? tabs.routes[tabs.index ?? 0] : undefined;
  return [root.routes.map((r) => r.name).join('>'), tab?.name ?? '', tab?.state?.index ?? 0].join('|');
}

/**
 * Where a push tap goes, from the notification's data. Pushes carry `url` ("freebie://deal/<id>"),
 * plus `triggeredEventId` and `eventId`; older pushes only have the ids, so handle every shape.
 */
export function urlFromPushData(data: Record<string, unknown> | null | undefined): string | null {
  if (!data) return null;
  if (typeof data.url === 'string' && data.url) return data.url;
  const id = data.triggeredEventId ?? data.eventId;
  return typeof id === 'string' && id ? `freebie://deal/${encodeURIComponent(id)}` : null;
}
