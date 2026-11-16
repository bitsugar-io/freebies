import { useState, useEffect, useCallback, useRef } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { api, Team } from '../api/client';

const CACHE_KEY = 'freebie_teams_cache';

interface TeamsCache {
  savedAt: number;
  teams: Team[];
}

/**
 * The season board's teams. Shows the saved copy right away, then refreshes. `error` is set when
 * the latest fetch failed; the screen then falls back to teams derived from events.
 */
export function useTeams() {
  const [teams, setTeams] = useState<Team[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isStale, setIsStale] = useState(false);
  const loaded = useRef(false);

  const fetchTeams = useCallback(async () => {
    try {
      if (!loaded.current) {
        const raw = await AsyncStorage.getItem(CACHE_KEY).catch(() => null);
        if (raw) {
          const cache = JSON.parse(raw) as TeamsCache;
          if (Array.isArray(cache.teams) && cache.teams.length > 0) {
            setTeams(cache.teams);
            setIsStale(true);
            setIsLoading(false);
          }
        }
      }
      const data = await api.listTeams();
      setTeams(data);
      setError(null);
      setIsStale(false);
      loaded.current = true;
      AsyncStorage.setItem(CACHE_KEY, JSON.stringify({ savedAt: Date.now(), teams: data } satisfies TeamsCache)).catch(
        () => {}
      );
    } catch (err) {
      console.error('Failed to fetch teams:', err);
      setError(err instanceof Error ? err.message : 'Failed to load teams');
      setIsStale(true);
    } finally {
      setIsLoading(false);
      loaded.current = true;
    }
  }, []);

  useEffect(() => {
    fetchTeams();
  }, [fetchTeams]);

  return { teams, isLoading, error, isStale, refetch: fetchTeams };
}
