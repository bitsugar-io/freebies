import { useState, useEffect, useCallback } from 'react';
import { api, Subscription } from '../api/client';
import { useUser } from './useUser';

export function useSubscriptions() {
  const { user } = useUser();
  const [subscriptions, setSubscriptions] = useState<Set<string>>(new Set());
  const [subscriptionData, setSubscriptionData] = useState<Subscription[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Load subscriptions from API
  const fetchSubscriptions = useCallback(async () => {
    if (!user) return;

    try {
      setIsLoading(true);
      setError(null);
      const data = await api.listSubscriptions(user.id);
      setSubscriptionData(data);
      setSubscriptions(new Set(data.map((s) => s.eventId)));
    } catch (err) {
      console.error('Failed to load subscriptions:', err);
      setError(err instanceof Error ? err.message : 'Failed to load');
    } finally {
      setIsLoading(false);
    }
  }, [user]);

  useEffect(() => {
    if (user) {
      fetchSubscriptions();
    }
  }, [user, fetchSubscriptions]);

  const toggleSubscription = useCallback(
    async (eventId: string) => {
      if (!user) return;

      const wasSubscribed = subscriptions.has(eventId);

      // Optimistic update
      setSubscriptions((prev) => {
        const next = new Set(prev);
        if (wasSubscribed) {
          next.delete(eventId);
        } else {
          next.add(eventId);
        }
        return next;
      });

      try {
        if (wasSubscribed) {
          await api.deleteSubscription(user.id, eventId);
        } else {
          await api.createSubscription(user.id, eventId);
        }
      } catch (err) {
        console.error('Failed to toggle subscription:', err);
        // Revert on error
        setSubscriptions((prev) => {
          const next = new Set(prev);
          if (wasSubscribed) {
            next.add(eventId);
          } else {
            next.delete(eventId);
          }
          return next;
        });
        setError(err instanceof Error ? err.message : 'Failed to update');
      }
    },
    [user, subscriptions]
  );

  const isSubscribed = useCallback(
    (eventId: string) => subscriptions.has(eventId),
    [subscriptions]
  );

  return {
    subscriptions,
    subscriptionData,
    isLoading,
    error,
    toggleSubscription,
    isSubscribed,
    subscribedCount: subscriptions.size,
    refetch: fetchSubscriptions,
  };
}
