import { useMemo } from 'react';
import { useAppData } from '../context/AppDataContext';
import { colors } from './useTheme';
import { isNearBlack } from '../utils/tileColors';

export interface Victory {
  /** Short team name from the API, e.g. "Rams". */
  team: string;
  /** The team's two colors from the API (`team.colors`), primary first. */
  stripe: [string, string];
}

/** A black or near-black team color (the Kings' too) would vanish against the black bar: use the team's accent, else white. */
function visible(c: string, accent?: string): string {
  if (!isNearBlack(c)) return c;
  return accent && !isNearBlack(accent) ? accent : colors.text;
}

/**
 * The most recent live win: the newest won deal that hasn't expired or been claimed, and its team's
 * colors from /teams. null when nothing is live or the team has no colors. Drives the team stripe
 * under the top bar and "RAMS WON!" under the title. Colors are server data, never hardcoded.
 */
export function useVictory(): Victory | null {
  const { liveWon, teams } = useAppData();
  return useMemo(() => {
    const live = [...liveWon].sort((a, b) => new Date(b.triggeredAt).getTime() - new Date(a.triggeredAt).getTime());
    for (const d of live) {
      const t = teams.find((x) => x.id === d.event.teamId);
      if (t?.colors) return { team: t.name, stripe: [visible(t.colors.primary, t.colors.accent), visible(t.colors.secondary, t.colors.accent)] };
    }
    return null;
  }, [liveWon, teams]);
}

/** Ray colors for a deal's team (its two API colors), for the YOU WON burst. */
export function useTeamColors(teamId: string | undefined): [string, string] | null {
  const { teams } = useAppData();
  return useMemo(() => {
    const t = teams.find((x) => x.id === teamId);
    return t?.colors ? [t.colors.primary, t.colors.secondary] : null;
  }, [teams, teamId]);
}
