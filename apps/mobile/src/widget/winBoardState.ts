// The Win Board Live Activity's content, built from the app's live wins. Pure (no native imports) so
// bun tests can run it; src/widget/winBoard.ts hands it to the native WinBoard module.
//
// Owner's pick (2026-11-17): no words, just the food and its countdown. One win shows the meal, who
// won and a big clock; several show the 3 that end soonest, each with its own clock, then
// "+N MORE". Each meal opens its own deal.

interface SpriteLike {
  width: number;
  height: number;
  pixels: string[];
  palette: Record<string, string>;
}

interface WinLike {
  id: string;
  triggeredAt: string;
  availableAt?: string;
  expiresAt?: string;
  event: { teamId: string; teamName: string; sprite?: SpriteLike };
}

interface TeamLike {
  id: string;
  name: string;
  colors?: { primary: string; secondary: string; accent?: string };
}

/** One meal on the card. Times are seconds since the epoch (what Swift's Date wants); 0 = none. */
export interface WinBoardMeal {
  url: string;
  sprite: { w: number; h: number; px: string[]; pal: Record<string, string> } | null;
  /** When it can be used, if that's still ahead ("STARTS 10AM"); 0 when it already can. */
  startsAt: number;
  /** "10AM" or "WED", shown under STARTS. */
  startsLabel: string;
  endsAt: number;
}

/** Mirrors WinBoardAttributes.ContentState in Swift (modules/win-board and targets/widget). */
export interface WinBoardState {
  title: string;
  count: number;
  /** Wins not on the card ("+N MORE"). */
  more: number;
  stripe: string[];
  meals: WinBoardMeal[];
  url: string;
  /** ms since epoch when the soonest deadline passes (iOS marks the card stale then). */
  staleAt: number | null;
}

const SHOWN = 3;
const GOLD = '#F8B800';
const PT = 'America/Los_Angeles';

/** "10AM", "10:30AM" today; "WED" on a later day. */
export function startsLabel(at: number, now: number): string {
  const day = (t: number) => new Date(t).toLocaleDateString('en-US', { timeZone: PT });
  if (day(at) !== day(now)) {
    return new Date(at).toLocaleDateString('en-US', { timeZone: PT, weekday: 'short' }).toUpperCase();
  }
  return new Date(at)
    .toLocaleTimeString('en-US', { timeZone: PT, hour: 'numeric', minute: '2-digit' })
    .replace(':00', '')
    .replace(' ', '')
    .toUpperCase();
}

/** One card for every live win (docs/ui.md: one Win Board, never one per win). */
export function winBoardState(live: readonly WinLike[], teams: readonly TeamLike[], now = Date.now()): WinBoardState | null {
  if (live.length === 0) return null;
  const sec = (iso?: string) => (iso ? Math.floor(new Date(iso).getTime() / 1000) : 0);
  const newest = [...live].sort((a, b) => new Date(b.triggeredAt).getTime() - new Date(a.triggeredAt).getTime());
  const colorOf = (w: WinLike) => teams.find((t) => t.id === w.event.teamId)?.colors;
  // Soonest first: a deal that hasn't started sorts by its start, the rest by their end.
  const key = (w: WinLike) => {
    const start = sec(w.availableAt) * 1000;
    return start > now ? start : sec(w.expiresAt) * 1000 || Number.MAX_SAFE_INTEGER;
  };
  const soonest = [...live].sort((a, b) => key(a) - key(b));
  const meals = soonest.slice(0, SHOWN).map((w): WinBoardMeal => {
    const s = w.event.sprite;
    const starts = sec(w.availableAt) * 1000 > now ? sec(w.availableAt) : 0;
    return {
      url: `freebie://deal/${w.id}`,
      sprite: s ? { w: s.width, h: s.height, px: s.pixels, pal: s.palette } : null,
      startsAt: starts,
      startsLabel: starts ? startsLabel(starts * 1000, now) : '',
      endsAt: sec(w.expiresAt),
    };
  });
  const first = newest[0];
  const single = live.length === 1;
  const c = colorOf(first);
  const ends = soonest.map((w) => sec(w.expiresAt)).filter(Boolean);
  const teamName = (teams.find((t) => t.id === first.event.teamId)?.name ?? first.event.teamName).toUpperCase();
  return {
    title: single ? `${teamName} WON!` : `${live.length} WINS!`,
    count: live.length,
    more: Math.max(0, live.length - SHOWN),
    stripe: single
      ? c
        ? [c.primary, c.secondary]
        : [GOLD, GOLD]
      : [...new Set(newest.map((w) => colorOf(w)?.primary ?? GOLD))].slice(0, 8),
    meals,
    url: single ? `freebie://deal/${first.id}` : 'freebie://won',
    staleAt: ends.length ? Math.min(...ends) * 1000 : null,
  };
}
