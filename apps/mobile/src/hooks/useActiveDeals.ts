import { useState, useEffect, useCallback, useRef } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { api, ActiveDeal } from '../api/client';
import { useUser } from './useUser';
import { setActiveDealsForNotifications } from './usePushNotifications';

const CACHE_KEY = 'freebie_active_deals_cache';

interface DealsCache {
  userId: string;
  savedAt: number;
  deals: ActiveDeal[];
}

export function useActiveDeals() {
  const { user } = useUser();
  const [activeDeals, setActiveDeals] = useState<ActiveDeal[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  /** When the deals on screen were last fetched or cached. Null until something loads. */
  const [savedAt, setSavedAt] = useState<number | null>(null);
  /** True when the deals shown came from the cache because the last fetch failed. */
  const [isStale, setIsStale] = useState(false);

  const initialLoadDone = useRef(false);
  const userId = user?.id;

  const fetchActiveDeals = useCallback(async () => {
    if (!userId) {
      setActiveDeals([]);
      setIsLoading(false);
      return;
    }

    try {
      if (!initialLoadDone.current) {
        // Show the last known deals right away (a fan at a counter may have no signal).
        const raw = await AsyncStorage.getItem(CACHE_KEY).catch(() => null);
        if (raw) {
          const cache = JSON.parse(raw) as DealsCache;
          if (cache.userId === userId) {
            setActiveDeals(cache.deals);
            setSavedAt(cache.savedAt);
            setIsStale(true);
            setActiveDealsForNotifications(cache.deals);
          }
        }
        setIsLoading(false);
      }
      const deals = await api.listActiveDeals(userId);
      const now = Date.now();
      setActiveDeals(deals);
      setSavedAt(now);
      setIsStale(false);
      setError(null);
      setActiveDealsForNotifications(deals);
      AsyncStorage.setItem(CACHE_KEY, JSON.stringify({ userId, savedAt: now, deals } satisfies DealsCache)).catch(
        () => {}
      );
    } catch (err) {
      console.error('Failed to fetch active deals:', err);
      setError(err instanceof Error ? err.message : 'Failed to load active deals');
      setIsStale(true);
    } finally {
      setIsLoading(false);
      initialLoadDone.current = true;
    }
  }, [userId]);

  useEffect(() => {
    initialLoadDone.current = false;
    fetchActiveDeals();
  }, [fetchActiveDeals]);

  const mutateLocal = useCallback(
    (triggeredEventId: string, patch: Partial<ActiveDeal>) =>
      setActiveDeals((prev) => {
        const next = prev.map((d) => (d.id === triggeredEventId ? { ...d, ...patch } : d));
        if (userId && savedAt) {
          AsyncStorage.setItem(CACHE_KEY, JSON.stringify({ userId, savedAt, deals: next } satisfies DealsCache)).catch(
            () => {}
          );
        }
        return next;
      }),
    [userId, savedAt]
  );

  const dismissDeal = useCallback(
    async (triggeredEventId: string, type: 'got_it' | 'stop_reminding' = 'got_it') => {
      if (!userId) return;
      await api.createDismissal(userId, triggeredEventId, type);
      mutateLocal(triggeredEventId, { isDismissed: true, dismissalType: type });
    },
    [userId, mutateLocal]
  );

  const undoDismissal = useCallback(
    async (triggeredEventId: string) => {
      if (!userId) return;
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
