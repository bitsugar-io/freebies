import { useCallback, useRef, useState } from 'react';
import { useFocusEffect } from '@react-navigation/native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { api, Today } from '../api/client';

const CACHE_KEY = 'freebie_today_cache';
/** Refetch on focus at most this often; pull to refresh always fetches. */
const FRESH_MS = 60_000;

interface TodayCache {
  savedAt: number;
  today: Today;
}

/**
 * GET /today: what anyone can claim now and today's games with deals on the line. Shows the saved
 * copy first (if it is for today), refetches when the tab gains focus. `savedAt` is set while a
 * failed fetch leaves the saved copy on screen.
 */
export function useToday() {
  const [today, setToday] = useState<Today | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [savedAt, setSavedAt] = useState<number | null>(null);
  const lastFetch = useRef(0);
  const restored = useRef(false);

  const fetchToday = useCallback(async (force: boolean) => {
    if (!restored.current) {
      restored.current = true;
      const raw = await AsyncStorage.getItem(CACHE_KEY).catch(() => null);
      if (raw) {
        try {
          const cache = JSON.parse(raw) as TodayCache;
          if (cache.today?.date === pacificDate()) {
            setToday(cache.today);
            setSavedAt(cache.savedAt);
            setIsLoading(false);
          }
        } catch {
          // A corrupt cache is ignored.
        }
      }
    }
    if (!force && Date.now() - lastFetch.current < FRESH_MS) return;
    lastFetch.current = Date.now();
    try {
      const data = await api.getToday();
      setToday(data);
      setError(null);
      setSavedAt(null);
      AsyncStorage.setItem(CACHE_KEY, JSON.stringify({ savedAt: Date.now(), today: data } satisfies TodayCache)).catch(
        () => {}
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load today');
      lastFetch.current = 0;
    } finally {
      setIsLoading(false);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      fetchToday(false);
    }, [fetchToday])
  );

  const refresh = useCallback(() => fetchToday(true), [fetchToday]);
  return { today, isLoading, error, savedAt, refresh };
}

/** Today's date in Pacific time, YYYY-MM-DD (matches the API's `date`). */
export function pacificDate(d = new Date()): string {
  return d.toLocaleDateString('en-CA', { timeZone: 'America/Los_Angeles' });
}
