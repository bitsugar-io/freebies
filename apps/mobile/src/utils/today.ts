// TODAY's date rules. Pure (no React or native imports).

/** Today's date in Pacific time, YYYY-MM-DD (matches the API's `date`). */
export function pacificDate(d = new Date()): string {
  return d.toLocaleDateString('en-CA', { timeZone: 'America/Los_Angeles' });
}

/** True when a /today response is for the given Pacific date (yesterday's games never show today). */
export function isForDate(today: { date: string } | null, date: string): boolean {
  return !!today && today.date === date;
}
