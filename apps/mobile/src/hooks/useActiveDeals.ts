import { useState, useEffect, useCallback } from 'react';
import { api, ActiveDeal } from '../api/client';
import { useUser } from './useUser';
import { setActiveDealsForNotifications } from './usePushNotifications';

export function useActiveDeals() {
  const { user } = useUser();
  const [activeDeals, setActiveDeals] = useState<ActiveDeal[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchActiveDeals = useCallback(async () => {
    if (!user?.id) {
      setActiveDeals([]);
      setIsLoading(false);
      return;
    }

    try {
      setIsLoading(true);
      const deals = await api.listActiveDeals(user.id);
      setActiveDeals(deals);
      // Update the notification system with current deals
      setActiveDealsForNotifications(deals);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch active deals:', err);
      setError(err instanceof Error ? err.message : 'Failed to load active deals');
    } finally {
      setIsLoading(false);
    }
  }, [user?.id]);

  useEffect(() => {
    fetchActiveDeals();
  }, [fetchActiveDeals]);

  const dismissDeal = useCallback(async (
    triggeredEventId: string,
    type: 'got_it' | 'stop_reminding' = 'got_it'
  ) => {
    if (!user?.id) return;

    try {
      await api.createDismissal(user.id, triggeredEventId, type);
      // Update local state
      setActiveDeals(prev =>
        prev.map(deal =>
          deal.id === triggeredEventId
            ? { ...deal, isDismissed: true, dismissalType: type }
            : deal
        )
      );
    } catch (err) {
      console.error('Failed to dismiss deal:', err);
      throw err;
    }
  }, [user?.id]);

  const undoDismissal = useCallback(async (triggeredEventId: string) => {
    if (!user?.id) return;

    try {
      await api.deleteDismissal(user.id, triggeredEventId);
      // Update local state
      setActiveDeals(prev =>
        prev.map(deal =>
          deal.id === triggeredEventId
            ? { ...deal, isDismissed: false, dismissalType: undefined }
            : deal
        )
      );
    } catch (err) {
      console.error('Failed to undo dismissal:', err);
      throw err;
    }
  }, [user?.id]);

  // Filter to show only non-dismissed deals (or all, depending on needs)
  const undismissedDeals = activeDeals.filter(deal => !deal.isDismissed);
  const dismissedDeals = activeDeals.filter(deal => deal.isDismissed);

  return {
    activeDeals,
    undismissedDeals,
    dismissedDeals,
    isLoading,
    error,
    refetch: fetchActiveDeals,
    dismissDeal,
    undoDismissal,
  };
}

// Utility to format expiration time
export function formatExpiresAt(expiresAt: string | undefined): string {
  if (!expiresAt) return 'No expiration';

  const expires = new Date(expiresAt);
  const now = new Date();
  const diff = expires.getTime() - now.getTime();

  if (diff < 0) return 'Expired';

  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

  if (hours > 24) {
    const days = Math.floor(hours / 24);
    return `${days} day${days > 1 ? 's' : ''} left`;
  }

  if (hours > 0) {
    return `${hours}h ${minutes}m left`;
  }

  return `${minutes}m left`;
}

export function isExpiringSoon(expiresAt: string | undefined, hoursThreshold: number = 6): boolean {
  if (!expiresAt) return false;

  const expires = new Date(expiresAt);
  const now = new Date();
  const diff = expires.getTime() - now.getTime();
  const hoursLeft = diff / (1000 * 60 * 60);

  return hoursLeft > 0 && hoursLeft <= hoursThreshold;
}
