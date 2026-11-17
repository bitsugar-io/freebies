import { useState, useEffect, useCallback, useRef } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { api, Subscription } from '../api/client';
import { useUser } from './useUser';
import { parseUserCache } from '../utils/cache';
import { createSequencer } from '../utils/sequence';

const CACHE_KEY = 'freebie_subscriptions_cache';

/** The followed event ids, saved per user so a returning user's follows are known at launch. */
interface SubsCache {
  userId: string;
  savedAt: number;
  eventIds: string[];
}

export function useSubscriptions() {
  const { user, isLoading: userLoading } = useUser();
  const userId = user?.id;
  const [subscriptions, setSubscriptions] = useState<Set<string>>(new Set());
  const [subscriptionData, setSubscriptionData] = useState<Subscription[]>([]);
  /** True until the follows are known (saved copy or network). */
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  // Mirror of state so rapid taps see the latest optimistic value.
  const current = useRef<Set<string>>(new Set());
  const seq = useRef(createSequencer()).current;
  const restored = useRef(false);

  const commit = useCallback(
    (next: Set<string>) => {
      current.current = next;
      setSubscriptions(next);
      if (userId) {
        AsyncStorage.setItem(
          CACHE_KEY,
          JSON.stringify({ userId, savedAt: Date.now(), eventIds: [...next] } satisfies SubsCache)
        ).catch(() => {});
      }
    },
    [userId]
  );

  const fetchSubscriptions = useCallback(async () => {
    if (!userId) return;

    if (!restored.current) {
      restored.current = true;
      const raw = await AsyncStorage.getItem(CACHE_KEY).catch(() => null);
      const cache = parseUserCache<SubsCache>(raw, userId, (v) => Array.isArray(v.eventIds));
      if (cache && current.current.size === 0) {
        const ids = new Set(cache.eventIds);
        current.current = ids;
        setSubscriptions(ids);
        setIsLoading(false);
      }
    }

    const fresh = seq.begin();
    try {
      setError(null);
      const data = await api.listSubscriptions(userId);
      // An older response, or one that started before a follow or unfollow, would undo it.
      if (!fresh()) return;
      setSubscriptionData(data);
      commit(new Set(data.map((s) => s.eventId)));
    } catch (err) {
      if (!fresh()) return;
      console.error('Failed to load subscriptions:', err);
      setError(err instanceof Error ? err.message : 'Failed to load');
    } finally {
      setIsLoading(false);
    }
  }, [userId, commit, seq]);

  useEffect(() => {
    restored.current = false;
    current.current = new Set();
    setSubscriptions(new Set());
    setSubscriptionData([]);
    if (userId) fetchSubscriptions();
  }, [userId, fetchSubscriptions]);

  // No user and none coming (creating one failed): nothing to wait for.
  useEffect(() => {
    if (!userId) setIsLoading(userLoading);
  }, [userId, userLoading]);

  /**
   * Set many events on or off at once. Updates instantly, sends the requests in parallel and
   * reverts only the ones that failed. Resolves with the ids that failed.
   */
  const setMany = useCallback(
    async (eventIds: string[], on: boolean): Promise<string[]> => {
      if (!userId) return eventIds;
      const changing = eventIds.filter((id) => current.current.has(id) !== on);
      if (changing.length === 0) return [];

      seq.invalidate();
      const optimistic = new Set(current.current);
      changing.forEach((id) => (on ? optimistic.add(id) : optimistic.delete(id)));
      commit(optimistic);

      const results = await Promise.allSettled(
        changing.map((id) => (on ? api.createSubscription(userId, id) : api.deleteSubscription(userId, id)))
      );
      // A list fetched while these were in flight may predate them.
      seq.invalidate();
      const failed = changing.filter((_, i) => results[i].status === 'rejected');
      if (failed.length > 0) {
        const reverted = new Set(current.current);
        failed.forEach((id) => (on ? reverted.delete(id) : reverted.add(id)));
        commit(reverted);
        setError('Failed to update');
      }
      return failed;
    },
    [userId, commit, seq]
  );

  const isSubscribed = useCallback((eventId: string) => subscriptions.has(eventId), [subscriptions]);

  return {
    subscriptions,
    subscriptionData,
    isLoading,
    error,
    setMany,
    isSubscribed,
    refetch: fetchSubscriptions,
  };
}
