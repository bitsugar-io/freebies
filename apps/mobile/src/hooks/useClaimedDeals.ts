import { useState, useEffect, useCallback, useRef } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { api, ClaimedDeal } from '../api/client';
import { useUser } from './useUser';
import { parseUserCache } from '../utils/cache';

const CACHE_KEY = 'freebie_claimed_deals_cache';

interface ClaimedCache {
  userId: string;
  savedAt: number;
  claimed: ClaimedDeal[];
}

/** Every deal the user claimed, newest first, including ones whose window has closed. */
export function useClaimedDeals() {
  const { user } = useUser();
  const userId = user?.id;
  const [claimed, setClaimed] = useState<ClaimedDeal[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const initialLoadDone = useRef(false);

  const fetchClaimed = useCallback(async () => {
    if (!userId) {
      setClaimed([]);
      setIsLoading(false);
      return;
    }
    try {
      if (!initialLoadDone.current) {
        // Parsed on its own: a corrupt cache never blocks the fetch below.
        const raw = await AsyncStorage.getItem(CACHE_KEY).catch(() => null);
        const cache = parseUserCache<ClaimedCache>(raw, userId, (v) => Array.isArray(v.claimed));
        if (cache) setClaimed(cache.claimed);
        setIsLoading(false);
      }
      const list = await api.listClaimed(userId);
      setClaimed(list);
      AsyncStorage.setItem(
        CACHE_KEY,
        JSON.stringify({ userId, savedAt: Date.now(), claimed: list } satisfies ClaimedCache)
      ).catch(() => {});
    } catch (err) {
      // Keep showing the cached list; the trophy case is history, not live data.
      console.error('Failed to fetch claimed deals:', err);
    } finally {
      setIsLoading(false);
      initialLoadDone.current = true;
    }
  }, [userId]);

  useEffect(() => {
    initialLoadDone.current = false;
    fetchClaimed();
  }, [fetchClaimed]);

  return { claimed, isLoading, refetch: fetchClaimed };
}
