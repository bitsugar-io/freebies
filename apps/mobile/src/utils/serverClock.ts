// Device clock correction from the API's Date header. Pure (no React or native imports).
//
// Expiry, "ends soon" and the lock checks compare server times to now. A phone whose clock is off
// would show a live deal as ended, or an ended one as live, so "now" is the device time plus the
// estimated server-minus-device offset. With no usable header the offset stays 0 (device time).

/** Offsets under this are noise: the Date header has 1 second resolution. */
const MIN_OFFSET_MS = 2_000;

let offsetMs = 0;

/**
 * The server-minus-device offset from one response, or null when the header is missing or bad.
 * The server stamped the header somewhere between send and receive; the midpoint is the best guess,
 * and +500ms recenters the header's truncated second.
 */
export function estimateOffset(dateHeader: string | null | undefined, sentAt: number, receivedAt: number): number | null {
  if (!dateHeader) return null;
  const server = Date.parse(dateHeader);
  if (Number.isNaN(server) || receivedAt < sentAt) return null;
  const offset = server + 500 - (sentAt + receivedAt) / 2;
  return Math.abs(offset) < MIN_OFFSET_MS ? 0 : Math.round(offset);
}

/** Record a response's Date header. Ignored when missing or unparseable. */
export function observeServerDate(dateHeader: string | null | undefined, sentAt: number, receivedAt: number): void {
  const est = estimateOffset(dateHeader, sentAt, receivedAt);
  if (est !== null) offsetMs = est;
}

/** Server-corrected now, in ms. */
export function serverNow(): number {
  return Date.now() + offsetMs;
}

/** The current offset (tests and diagnostics). */
export function clockOffset(): number {
  return offsetMs;
}

/** Test hook. */
export function resetClock(): void {
  offsetMs = 0;
}
