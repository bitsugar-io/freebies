import { useMemo } from 'react';
import { useAppData } from '../context/AppDataContext';
import { colors } from './useTheme';

export interface Victory {
  /** Short team name from the API, e.g. "Rams". */
  team: string;
  /** The team's two colors from the API (`team.colors`), primary first. */
  stripe: [string, string];
}

/** A black team color would vanish against the black bar; draw it as white instead. */
function visible(c: string): string {
  return c.toUpperCase() === colors.background ? colors.text : c;
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
      if (t?.colors) return { team: t.name, stripe: [visible(t.colors.primary), visible(t.colors.secondary)] };
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
