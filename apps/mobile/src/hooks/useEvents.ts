import { useState, useEffect, useCallback, useRef } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { api, Event } from '../api/client';
import { parseCache } from '../utils/cache';
import { createSequencer } from '../utils/sequence';

const CACHE_KEY = 'freebie_events_cache';

interface EventsCache {
  savedAt: number;
  events: Event[];
}

/**
 * Every deal (event). Shows the saved copy first, so the follow count (follows x events) is known at
 * launch, then refreshes. `isLoading` is true only until the saved copy or the first fetch lands.
 */
export function useEvents() {
  const [events, setEvents] = useState<Event[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const restored = useRef(false);
  const seq = useRef(createSequencer()).current;

  const fetchEvents = useCallback(async () => {
    if (!restored.current) {
      restored.current = true;
      const raw = await AsyncStorage.getItem(CACHE_KEY).catch(() => null);
      const cache = parseCache<EventsCache>(raw, (v) => Array.isArray(v.events) && v.events.length > 0);
      if (cache) {
        setEvents(cache.events);
        setIsLoading(false);
      }
    }
    const fresh = seq.begin();
    try {
      const data = await api.listEvents();
      if (!fresh()) return;
      setError(null);
      setEvents(data);
      AsyncStorage.setItem(CACHE_KEY, JSON.stringify({ savedAt: Date.now(), events: data } satisfies EventsCache)).catch(
        () => {}
      );
    } catch (err) {
      if (!fresh()) return;
      console.error('Failed to fetch events:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch events');
    } finally {
      setIsLoading(false);
    }
  }, [seq]);

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
