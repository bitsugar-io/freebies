// The Win Board Live Activity's content, built from the app's live wins. Pure (no native imports) so
// bun tests can run it; src/widget/winBoard.ts hands it to the native WinBoard module.

import { deadlineLabel } from '../utils/format';

interface SpriteLike {
  width: number;
  height: number;
  pixels: string[];
  palette: Record<string, string>;
}

interface WinLike {
  id: string;
  triggeredAt: string;
  expiresAt?: string;
  event: { teamId: string; teamName: string; offerName: string; partnerName: string; sprite?: SpriteLike };
}

interface TeamLike {
  id: string;
  name: string;
  colors?: { primary: string; secondary: string; accent?: string };
}

/** Mirrors WinBoardAttributes.ContentState in Swift (modules/win-board and targets/widget). */
export interface WinBoardState {
  title: string;
  /** System font, white: the offer, or "Next: <offer>" when there are several. */
  offer: string;
  /** System font, grey: the restaurant. */
  sub: string;
  /** Pixel font, lilac, one short line: "ENDS 11:59PM PT". */
  detail: string;
  count: number;
  stripe: string[];
  wins: { sprite: { w: number; h: number; px: string[]; pal: Record<string, string> } | null; primary: string; secondary: string }[];
  url: string;
  /** ms since epoch: when the next deadline passes (iOS marks the card stale then). */
  staleAt: number | null;
}

const GOLD = '#F8B800';

/** One card for every live win, newest first (docs/ui.md: one Win Board, never one per win). */
export function winBoardState(live: readonly WinLike[], teams: readonly TeamLike[], now = Date.now()): WinBoardState | null {
  if (live.length === 0) return null;
  const wins = [...live].sort((a, b) => new Date(b.triggeredAt).getTime() - new Date(a.triggeredAt).getTime());
  const teamOf = (id: string) => teams.find((t) => t.id === id);
  const pairOf = (w: WinLike): [string, string] => {
    const c = teamOf(w.event.teamId)?.colors;
    return c ? [c.primary, c.secondary] : [GOLD, GOLD];
  };
  const due = wins.filter((w) => w.expiresAt).sort((a, b) => new Date(a.expiresAt!).getTime() - new Date(b.expiresAt!).getTime());
  const next = due[0] ?? wins[0];
  const ends = next.expiresAt ? `ENDS ${deadlineLabel(next.expiresAt, now).toUpperCase()}` : '';
  const first = wins[0];
  const name = (teamOf(first.event.teamId)?.name ?? first.event.teamName).toUpperCase();
  const single = wins.length === 1;
  return {
    title: single ? `${name} WON!` : `${wins.length} WINS!`,
    offer: single ? first.event.offerName : `Next: ${next.event.offerName}`,
    sub: single ? first.event.partnerName : next.event.partnerName,
    detail: ends,
    count: wins.length,
    stripe: single ? pairOf(first) : [...new Set(wins.map((w) => pairOf(w)[0]))].slice(0, 8),
    wins: wins.slice(0, 4).map((w) => {
      const s = w.event.sprite;
      const [primary, secondary] = pairOf(w);
      return { sprite: s ? { w: s.width, h: s.height, px: s.pixels, pal: s.palette } : null, primary, secondary };
    }),
    url: single ? `freebie://deal/${first.id}` : 'freebie://won',
    staleAt: next.expiresAt ? new Date(next.expiresAt).getTime() : null,
  };
}
