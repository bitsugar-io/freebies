/**
 * Holds links that arrive while no NavigationContainer can take them (splash,
 * maintenance, between mounts), so none is dropped and none is replayed.
 *
 * Why this exists: React Navigation only listens for links while its container is mounted, and it
 * asks getInitialURL on every mount. The OS launch URL (and the last push response) never change
 * for the life of the process, so a remount re-applied the previous link, and a link that landed
 * before the mount was lost.
 */
export type LinkListener = (url: string) => void;

let listener: LinkListener | null = null;
let pending: string | null = null;
let launchRead = false;
const handled = new Set<string>();

/** A link arrived. Hand it to the live container, or keep the newest until one mounts. */
export function deliverLink(url: string): void {
  if (listener) listener(url);
  else pending = url;
}

/** The container started listening. Anything that arrived before it is flushed now. */
export function attachLinkListener(next: LinkListener): () => void {
  listener = next;
  if (pending) {
    const url = pending;
    pending = null;
    next(url);
  }
  return () => {
    if (listener === next) listener = null;
  };
}

/**
 * The URL a freshly mounted container should open on: a link that arrived first, else the launch
 * URL, which counts only once per process.
 */
export async function initialLink(readLaunch: () => Promise<string | null>): Promise<string | null> {
  if (pending) {
    const url = pending;
    pending = null;
    launchRead = true; // newer than the launch URL, which must not replay later
    return url;
  }
  if (launchRead) return null;
  launchRead = true;
  const url = await readLaunch();
  // A link may have arrived while the launch URL was being read; it is newer, so it wins.
  if (pending) {
    const newer = pending;
    pending = null;
    return newer;
  }
  return url;
}

/** True the first time a notification id is seen (cold start and listener can both report one). */
export function firstSighting(notificationId: string | undefined): boolean {
  if (!notificationId) return true;
  if (handled.has(notificationId)) return false;
  handled.add(notificationId);
  return true;
}

/** Test hook. */
export function resetLinks(): void {
  listener = null;
  pending = null;
  launchRead = false;
  handled.clear();
}
