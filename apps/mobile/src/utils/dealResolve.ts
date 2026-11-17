// What the deal screen shows for a route id. Pure (no React or native imports).
//
// The `Deal` route id is a triggered-event id (a win, from WON, a push or a TODAY "Live now" row)
// or an event id (a deal you're considering, from TEAMS or a TODAY game). Order matters: a
// triggered id is matched first, so an event with several wins never opens the wrong one.

interface Timed {
  id: string;
  eventId: string;
  expiresAt?: string;
  availableAt?: string;
  triggeredAt: string;
}

export interface Owned extends Timed {
  isDismissed: boolean;
}

export type Resolved<D extends Owned, P extends Timed, E> =
  /** A win in your list: claim, silence, move back. */
  | { kind: 'mine'; deal: D }
  /** A triggered deal not in your list (a team you don't follow): open the partner, follow. */
  | { kind: 'public'; deal: P }
  /** An event id: the "if X, you get Y" preview. */
  | { kind: 'event'; event: E }
  | null;

export function resolveDeal<D extends Owned, P extends Timed, E extends { id: string }>(
  id: string,
  sources: {
    mine: readonly D[];
    /** Triggered deals known without a fetch: TODAY's live list. */
    live: readonly P[];
    /** GET /triggered/{id}, once fetched. */
    fetchedTriggered: P | null;
    events: readonly E[];
    /** GET /events/{id}, once fetched. */
    fetchedEvent: E | null;
  }
): Resolved<D, P, E> {
  const byId = sources.mine.find((d) => d.id === id);
  if (byId) return { kind: 'mine', deal: byId };
  const pub = sources.live.find((d) => d.id === id) ?? (sources.fetchedTriggered?.id === id ? sources.fetchedTriggered : null);
  if (pub) return { kind: 'public', deal: pub };
  // An event id: an old push with only the event id opens your newest open win of it.
  const ofEvent = sources.mine.filter((d) => d.eventId === id);
  const open = ofEvent.filter((d) => !d.isDismissed).sort((a, b) => b.triggeredAt.localeCompare(a.triggeredAt));
  if (open[0]) return { kind: 'mine', deal: open[0] };
  if (ofEvent[0]) return { kind: 'mine', deal: ofEvent[0] };
  const event = sources.events.find((e) => e.id === id) ?? (sources.fetchedEvent?.id === id ? sources.fetchedEvent : null);
  return event ? { kind: 'event', event } : null;
}

/**
 * Whether the screen may show what it resolved, keep the skeleton, or say the deal is gone. A deal
 * that may still be yours waits for your deals (cache or network) and the one refresh for this id,
 * so a cold-start push tap never flashes "not available" or the public view; `timedOut` ends the wait.
 */
export function dealScreenPhase(s: {
  resolved: 'mine' | 'public' | 'event' | null;
  /** Your deals are known (from cache or network). */
  dealsKnown: boolean;
  /** The one refresh of your deals for this id is done (a push can beat the list). */
  refetchDone: boolean;
  /** GET /triggered/{id} and GET /events/{id} have answered (found or not). */
  lookupsDone: boolean;
  timedOut: boolean;
}): 'show' | 'loading' | 'missing' {
  if (s.resolved === 'mine') return 'show';
  if (s.timedOut) return s.resolved ? 'show' : 'missing';
  // A preview shows once your deals are known; a triggered deal waits for the refresh too, since
  // it may turn out to be yours.
  if (s.resolved === 'event') return s.dealsKnown ? 'show' : 'loading';
  if (s.resolved === 'public') return s.dealsKnown && s.refetchDone ? 'show' : 'loading';
  return s.dealsKnown && s.refetchDone && s.lookupsDone ? 'missing' : 'loading';
}
