import { useCallback, useEffect, useRef, useState } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { api, Today } from '../api/client';
import { parseCache } from '../utils/cache';
import { createSequencer } from '../utils/sequence';
import { serverNow } from '../utils/serverClock';
import { isForDate, pacificDate } from '../utils/today';

export { pacificDate } from '../utils/today';

const CACHE_KEY = 'freebie_today_cache';
/** Refetch on focus at most this often; pull to refresh, foreground and a new day always fetch. */
const FRESH_MS = 60_000;
/** How often to check whether the Pacific date rolled over. */
const DATE_CHECK_MS = 60_000;

interface TodayCache {
  savedAt: number;
  today: Today;
}

/**
 * GET /today: what anyone can claim now and today's games with deals on the line. Lives in the app
 * data context so foreground, pushes and refreshAll refetch it too. Shows the saved copy first (if
 * it is for today). When the Pacific date changes it refetches, and yesterday's copy is never shown
 * as today. `savedAt` is set while a failed fetch leaves the saved copy on screen.
 */
export function useToday() {
  const [stored, setStored] = useState<Today | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [savedAt, setSavedAt] = useState<number | null>(null);
  const [date, setDate] = useState(() => pacificDate(new Date(serverNow())));
  const lastFetch = useRef(0);
  const restored = useRef(false);
  const seq = useRef(createSequencer()).current;

  const fetchToday = useCallback(
    async (force: boolean) => {
      if (!restored.current) {
        restored.current = true;
        const raw = await AsyncStorage.getItem(CACHE_KEY).catch(() => null);
        const cache = parseCache<TodayCache>(raw, (v) => !!v.today && typeof v.today.date === 'string');
        if (cache && cache.today.date === pacificDate(new Date(serverNow()))) {
          setStored(cache.today);
          setSavedAt(cache.savedAt);
          setIsLoading(false);
        }
      }
      if (!force && Date.now() - lastFetch.current < FRESH_MS) return;
      lastFetch.current = Date.now();
      const fresh = seq.begin();
      try {
        const data = await api.getToday();
        if (!fresh()) return;
        setStored(data);
        setError(null);
        setSavedAt(null);
        AsyncStorage.setItem(CACHE_KEY, JSON.stringify({ savedAt: Date.now(), today: data } satisfies TodayCache)).catch(
          () => {}
        );
      } catch (err) {
        if (!fresh()) return;
        setError(err instanceof Error ? err.message : 'Failed to load today');
        lastFetch.current = 0;
      } finally {
        setIsLoading(false);
      }
    },
    [seq]
  );

  useEffect(() => {
    fetchToday(false);
  }, [fetchToday]);

  // Overnight: a new Pacific day means new games; refetch once when the date changes.
  useEffect(() => {
    const id = setInterval(() => {
      const d = pacificDate(new Date(serverNow()));
      setDate((prev) => (prev === d ? prev : d));
    }, DATE_CHECK_MS);
    return () => clearInterval(id);
  }, []);
  const firstDate = useRef(date);
  useEffect(() => {
    if (date === firstDate.current) return;
    firstDate.current = date;
    fetchToday(true);
  }, [date, fetchToday]);

  const today = isForDate(stored, date) ? stored : null;
  /** Fetch now (pull to refresh, foreground), catching up on the date first (timers sleep in the background). */
  const refresh = useCallback(() => {
    const d = pacificDate(new Date(serverNow()));
    firstDate.current = d;
    setDate(d);
    return fetchToday(true);
  }, [fetchToday]);
  /** Fetch unless fetched within the last minute (tab focus). */
  const refreshIfStale = useCallback(() => fetchToday(false), [fetchToday]);
  return { today, isLoading, error, savedAt, refresh, refreshIfStale };
}
