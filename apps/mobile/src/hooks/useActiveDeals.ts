import { useState, useEffect, useCallback, useRef } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { api, ActiveDeal } from '../api/client';
import { isNetworkError } from '../api/errors';
import { useUser } from './useUser';
import { setActiveDealsForNotifications } from './usePushNotifications';
import { parseUserCache } from '../utils/cache';
import { createSequencer } from '../utils/sequence';

const CACHE_KEY = 'freebie_active_deals_cache';

interface DealsCache {
  userId: string;
  savedAt: number;
  deals: ActiveDeal[];
}

/**
 * The user's won deals. `isLoading` means "we don't know the user's deals yet": it stays true until
 * the stored user's cached deals are read (a cache hit) or the network answers, so the launch tab
 * sees a win. Out-of-order responses are dropped, and a local change (GOT IT) makes any fetch already
 * in flight stale so it can't undo the change.
 */
export function useActiveDeals() {
  const { user, isLoading: userLoading } = useUser();
  const [activeDeals, setActiveDeals] = useState<ActiveDeal[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  /** When the deals on screen were last fetched or cached. Null until something loads. */
  const [savedAt, setSavedAt] = useState<number | null>(null);
  /** True when the deals shown came from the cache because the last fetch failed. */
  const [isStale, setIsStale] = useState(false);

  const initialLoadDone = useRef(false);
  const seq = useRef(createSequencer()).current;
  const userId = user?.id;
  // Mirror for writing the cache from mutations without a stale closure.
  const savedAtRef = useRef<number | null>(null);

  const fetchActiveDeals = useCallback(async () => {
    if (!userId) {
      seq.invalidate();
      setActiveDeals([]);
      return;
    }

    if (!initialLoadDone.current) {
      initialLoadDone.current = true;
      // Show the last known deals right away (a fan at a counter may have no signal). Parsed on its
      // own: a corrupt cache never blocks the fetch below.
      const raw = await AsyncStorage.getItem(CACHE_KEY).catch(() => null);
      const cache = parseUserCache<DealsCache>(raw, userId, (v) => Array.isArray(v.deals));
      if (cache) {
        setActiveDeals(cache.deals);
        setSavedAt(cache.savedAt);
        savedAtRef.current = cache.savedAt;
        setIsStale(true);
        setActiveDealsForNotifications(cache.deals);
        setIsLoading(false);
      }
    }

    const fresh = seq.begin();
    try {
      const deals = await api.listActiveDeals(userId);
      if (!fresh()) return;
      const now = Date.now();
      setActiveDeals(deals);
      setSavedAt(now);
      savedAtRef.current = now;
      setIsStale(false);
      setError(null);
      setActiveDealsForNotifications(deals);
      AsyncStorage.setItem(CACHE_KEY, JSON.stringify({ userId, savedAt: now, deals } satisfies DealsCache)).catch(
        () => {}
      );
    } catch (err) {
      if (!fresh()) return;
      console.error('Failed to fetch active deals:', err);
      setError(err instanceof Error ? err.message : 'Failed to load active deals');
      setIsStale(true);
    } finally {
      // Known now, from the network or as well as we can (a failure keeps the cache).
      setIsLoading(false);
    }
  }, [userId, seq]);

  useEffect(() => {
    initialLoadDone.current = false;
    fetchActiveDeals();
  }, [fetchActiveDeals]);

  // No user and none coming (creating one failed): there is nothing to wait for.
  useEffect(() => {
    if (!userId && !userLoading) setIsLoading(false);
    if (!userId && userLoading) setIsLoading(true);
  }, [userId, userLoading]);

  const mutateLocal = useCallback(
    (triggeredEventId: string, patch: Partial<ActiveDeal>) => {
      // A fetch that started before this change would bring back the old state.
      seq.invalidate();
      setActiveDeals((prev) => {
        const next = prev.map((d) => (d.id === triggeredEventId ? { ...d, ...patch } : d));
        const at = savedAtRef.current;
        if (userId && at) {
          AsyncStorage.setItem(CACHE_KEY, JSON.stringify({ userId, savedAt: at, deals: next } satisfies DealsCache)).catch(
            () => {}
          );
        }
        return next;
      });
    },
    [userId, seq]
  );

  const dismissDeal = useCallback(
    async (triggeredEventId: string, type: 'got_it' | 'stop_reminding' = 'got_it') => {
      if (!userId) throw new Error('No user');
      let dismissal;
      try {
        dismissal = await api.createDismissal(userId, triggeredEventId, type);
      } catch (err) {
        // The first request may have reached the server and only the answer was lost (a timeout on
        // a weak signal). The server is idempotent, so one retry settles it either way.
        if (!isNetworkError(err)) throw err;
        dismissal = await api.createDismissal(userId, triggeredEventId, type);
      }
      // The server never downgrades a claim, so use the type it kept.
      const kept = dismissal?.type ?? type;
      mutateLocal(triggeredEventId, { isDismissed: true, dismissalType: kept });
    },
    [userId, mutateLocal]
  );

  const undoDismissal = useCallback(
    async (triggeredEventId: string) => {
      if (!userId) throw new Error('No user');
      await api.deleteDismissal(userId, triggeredEventId);
      mutateLocal(triggeredEventId, { isDismissed: false, dismissalType: undefined });
    },
    [userId, mutateLocal]
  );

  const undismissedDeals = activeDeals.filter((deal) => !deal.isDismissed);
  const dismissedDeals = activeDeals.filter((deal) => deal.isDismissed);

  return {
    activeDeals,
    undismissedDeals,
    dismissedDeals,
    isLoading,
    error,
    savedAt,
    isStale,
    refetch: fetchActiveDeals,
    dismissDeal,
    undoDismissal,
  };
}
