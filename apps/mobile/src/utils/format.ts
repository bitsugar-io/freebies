/**
 * The short display name for a team ("Dodgers"), from the API's team list. Until the list loads
 * (or for an unknown team) it falls back to the full name the event carries.
 */
export function teamNameFrom(teams: { id: string; name: string }[], teamId: string, fallback = ''): string {
  return teams.find((t) => t.id === teamId)?.name ?? (fallback || teamId);
}

/** A sport's display label: the API's `sportName` when it sends one, else the id ("womens-soccer" -> "WOMENS SOCCER"). */
export function sportLabel(sport: string, sportName?: string | null): string {
  if (sportName) return sportName.toUpperCase();
  if (!sport) return 'ALL TEAMS';
  return sport.replace(/[-_]+/g, ' ').toUpperCase();
}

/** "Dodgers pitchers record 7+ strikeouts" -> "If the Dodgers pitchers record 7+ strikeouts". */
export function ifCondition(condition: string): string {
  return `If ${condition.trim()}`;
}

/** Wall-clock time in Pacific time, e.g. "10:30am PT". Deals are an LA thing, so deadlines are too. */
export function clockTime(iso: number | string | Date): string {
  const t = new Date(iso)
    .toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', timeZone: 'America/Los_Angeles' })
    .toLowerCase()
    .replace(' ', '');
  return `${t} PT`;
}

/** "03:28:11" */
export function formatCountdown(ms: number): string {
  const total = Math.max(0, Math.floor(ms / 1000));
  const h = Math.floor(total / 3600);
  const m = Math.floor((total % 3600) / 60);
  const s = total % 60;
  const p = (n: number) => String(n).padStart(2, '0');
  return `${p(h)}:${p(m)}:${p(s)}`;
}

const PT = 'America/Los_Angeles';

/** Wall-clock parts of an instant in Pacific time. */
function ptParts(d: Date) {
  const parts = new Intl.DateTimeFormat('en-US', {
    timeZone: PT,
    hourCycle: 'h23',
    year: 'numeric',
    month: 'numeric',
    day: 'numeric',
    hour: 'numeric',
    minute: 'numeric',
    second: 'numeric',
  }).formatToParts(d);
  const n = (t: string) => Number(parts.find((p) => p.type === t)?.value ?? 0);
  return { y: n('year'), m: n('month'), d: n('day'), h: n('hour'), min: n('minute'), s: n('second') };
}

/** The next 8:00am Pacific after `from` (today if it is before 8am there, otherwise tomorrow). */
export function next8amPacific(from: Date = new Date()): Date {
  const p = ptParts(from);
  const day = p.h < 8 ? 0 : 1;
  // Pacific wall-clock 8:00 on that day, expressed as UTC, then corrected by the zone offset.
  const wall = Date.UTC(p.y, p.m - 1, p.d + day, 8, 0, 0);
  let guess = new Date(wall);
  for (let i = 0; i < 2; i++) {
    const g = ptParts(guess);
    const gWall = Date.UTC(g.y, g.m - 1, g.d, g.h, g.min, g.s);
    guess = new Date(guess.getTime() + (wall - gWall));
  }
  return guess;
}

/** "Sun 8:00am PT" for a pause end time. */
export function pauseUntilLabel(iso: string): string {
  const d = new Date(iso);
  if (d.getUTCFullYear() >= 9000) return 'until you resume';
  const day = d.toLocaleDateString('en-US', { weekday: 'short', timeZone: PT });
  return `${day} ${clockTime(d)}`;
}

/** "NOV 15": the Pacific calendar date of an instant, for the trophy case. */
export function shortDatePacific(iso: number | string | Date): string {
  const parts = new Intl.DateTimeFormat('en-US', { timeZone: PT, month: 'short', day: 'numeric' }).formatToParts(
    new Date(iso)
  );
  const get = (t: string) => parts.find((p) => p.type === t)?.value ?? '';
  return `${get('month')} ${get('day')}`.toUpperCase();
}
