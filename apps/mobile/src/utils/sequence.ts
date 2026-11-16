// Request ordering and tap guards. Pure (no React or native imports).

/**
 * Hands out request tickets so only the newest response is applied. `begin()` returns a check that
 * stays true while no newer request started and nothing invalidated it. `invalidate()` (called by a
 * local change such as GOT IT or a follow) makes every in-flight response stale, so a refetch that
 * started before the change can't undo it.
 */
export function createSequencer() {
  let current = 0;
  return {
    begin(): () => boolean {
      const ticket = ++current;
      return () => ticket === current;
    },
    invalidate(): void {
      current++;
    },
  };
}

/**
 * Wraps an async action so taps while it runs are ignored: one in flight at a time. Resolves to
 * undefined for an ignored call.
 */
export function singleFlight<A extends unknown[], R>(fn: (...args: A) => Promise<R>) {
  let running = false;
  const run = async (...args: A): Promise<R | undefined> => {
    if (running) return undefined;
    running = true;
    try {
      return await fn(...args);
    } finally {
      running = false;
    }
  };
  run.isRunning = () => running;
  return run;
}

/** A function that runs only its first call (closing a screen once). */
export function once(fn: () => void): () => void {
  let done = false;
  return () => {
    if (done) return;
    done = true;
    fn();
  };
}
