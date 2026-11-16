import { useState, useEffect, useCallback, useRef } from 'react';
import { api, Event } from '../api/client';

export function useEvents() {
  const [events, setEvents] = useState<Event[]>([]);
  // True only until the first fetch settles; refetches keep showing what's already loaded.
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const loaded = useRef(false);

  const fetchEvents = useCallback(async () => {
    try {
      setError(null);
      const data = await api.listEvents();
      setEvents(data);
      loaded.current = true;
    } catch (err) {
      console.error('Failed to fetch events:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch events');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchEvents();
  }, [fetchEvents]);

  return {
    events,
    isLoading,
    error,
    refetch: fetchEvents,
  };
}
