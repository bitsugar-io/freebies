// Deal state logic with no React or native imports, so unit tests can run it with bun.

const DAY_MS = 86_400_000;
const PT = 'America/Los_Angeles';

/** Where a deal is in its life. A locked deal is won, but its code isn't usable yet. */
export type DealPhase = 'locked' | 'active' | 'expired';

interface Timed {
  availableAt?: string | null;
  expiresAt?: string | null;
}

function ms(iso?: string | null): number | null {
  if (!iso) return null;
  const t = new Date(iso).getTime();
  return Number.isNaN(t) ? null : t;
}

/** True while the code unlocks later than `now`. */
export function isLocked(availableAt: string | null | undefined, now: number): boolean {
  const t = ms(availableAt);
  return t !== null && t > now;
}

/** Expired wins over locked: a window that already closed never reads as "unlocks soon". */
export function dealPhase(deal: Timed, now: number): DealPhase {
  const end = ms(deal.expiresAt);
  if (end !== null && end <= now) return 'expired';
  return isLocked(deal.availableAt, now) ? 'locked' : 'active';
}

/** Pacific calendar day number of an instant, for same-day checks. */
function ptDay(t: number): string {
  return new Date(t).toLocaleDateString('en-US', { timeZone: PT, year: 'numeric', month: 'numeric', day: 'numeric' });
}

/** "TUE", or "NOV 24" when it is more than 6 days away, or "TODAY". Pacific time. */
export function unlockWhen(availableAt: string, now: number): string {
  const t = new Date(availableAt).getTime();
  if (ptDay(t) === ptDay(now)) return 'TODAY';
  if (t - now > 6 * DAY_MS) {
    const d = new Date(t);
    const month = d.toLocaleDateString('en-US', { timeZone: PT, month: 'short' });
    const day = d.toLocaleDateString('en-US', { timeZone: PT, day: 'numeric' });
    return `${month} ${day}`.toUpperCase();
  }
  return new Date(t).toLocaleDateString('en-US', { timeZone: PT, weekday: 'short' }).toUpperCase();
}

/** "CODE UNLOCKS TUE": the WON row's status while locked. */
export function unlockTag(availableAt: string, now: number): string {
  return `CODE UNLOCKS ${unlockWhen(availableAt, now)}`;
}

/** "Tue Nov 17 at 12:00am PT" */
export function ptStamp(iso: string): string {
  const d = new Date(iso);
  const date = d.toLocaleDateString('en-US', { timeZone: PT, weekday: 'short', month: 'short', day: 'numeric' }).replace(',', '');
  const time = d
    .toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', timeZone: PT })
    .toLowerCase()
    .replace(' ', '');
  return `${date} at ${time} PT`;
}

/** Spoken form for the row label: "Code unlocks Tuesday". */
export function spokenUnlock(availableAt: string, now: number): string {
  const when = unlockWhen(availableAt, now);
  if (when === 'TODAY') return 'Code unlocks today';
  const t = new Date(availableAt).getTime();
  if (t - now > 6 * DAY_MS) {
    return `Code unlocks ${new Date(t).toLocaleDateString('en-US', { timeZone: PT, month: 'long', day: 'numeric' })}`;
  }
  return `Code unlocks ${new Date(t).toLocaleDateString('en-US', { timeZone: PT, weekday: 'long' })}`;
}

/** The countdown on the deal screen counts to the unlock while locked, otherwise to the end. */
export function countdownTarget(deal: Timed, now: number): { to: string; kind: 'unlock' | 'expiry' } | null {
  if (isLocked(deal.availableAt, now) && deal.availableAt) return { to: deal.availableAt, kind: 'unlock' };
  return deal.expiresAt ? { to: deal.expiresAt, kind: 'expiry' } : null;
}

/** Timer color only under 6 hours; otherwise plain text color. */
export function countdownTone(leftMs: number): 'warning' | 'text' {
  return leftMs > 0 && leftMs <= 6 * 3_600_000 ? 'warning' : 'text';
}

/** Count of followed events that are still active. Retired events never count. */
export function countActiveFollowed(
  events: { id: string; isActive?: boolean }[],
  isSubscribed: (eventId: string) => boolean
): number {
  return events.filter((e) => e.isActive !== false && isSubscribed(e.id)).length;
}
