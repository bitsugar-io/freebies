import { useState, useEffect, useCallback, useRef } from 'react';
import { api, Subscription } from '../api/client';
import { useUser } from './useUser';

export function useSubscriptions() {
  const { user } = useUser();
  const [subscriptions, setSubscriptions] = useState<Set<string>>(new Set());
  const [subscriptionData, setSubscriptionData] = useState<Subscription[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  // Mirror of state so rapid taps see the latest optimistic value.
  const current = useRef<Set<string>>(new Set());

  const commit = useCallback((next: Set<string>) => {
    current.current = next;
    setSubscriptions(next);
  }, []);

  const fetchSubscriptions = useCallback(async () => {
    if (!user) return;

    try {
      setError(null);
      const data = await api.listSubscriptions(user.id);
      setSubscriptionData(data);
      commit(new Set(data.map((s) => s.eventId)));
    } catch (err) {
      console.error('Failed to load subscriptions:', err);
      setError(err instanceof Error ? err.message : 'Failed to load');
    } finally {
      setIsLoading(false);
    }
  }, [user, commit]);

  useEffect(() => {
    if (user) {
      fetchSubscriptions();
    }
  }, [user, fetchSubscriptions]);

  /**
   * Set many events on or off at once. Updates instantly, sends the requests in parallel and
   * reverts only the ones that failed. Resolves with the ids that failed.
   */
  const setMany = useCallback(
    async (eventIds: string[], on: boolean): Promise<string[]> => {
      if (!user) return eventIds;
      const changing = eventIds.filter((id) => current.current.has(id) !== on);
      if (changing.length === 0) return [];

      const optimistic = new Set(current.current);
      changing.forEach((id) => (on ? optimistic.add(id) : optimistic.delete(id)));
      commit(optimistic);

      const results = await Promise.allSettled(
        changing.map((id) =>
          on ? api.createSubscription(user.id, id) : api.deleteSubscription(user.id, id)
        )
      );
      const failed = changing.filter((_, i) => results[i].status === 'rejected');
      if (failed.length > 0) {
        const reverted = new Set(current.current);
        failed.forEach((id) => (on ? reverted.delete(id) : reverted.add(id)));
        commit(reverted);
        setError('Failed to update');
      }
      return failed;
    },
    [user, commit]
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
