import React, { createContext, useContext, useCallback, useEffect, useMemo, useRef, useState, ReactNode } from 'react';
import { AppState, AppStateStatus } from 'react-native';
import { useEvents } from '../hooks/useEvents';
import { useTeams } from '../hooks/useTeams';
import { useSubscriptions } from '../hooks/useSubscriptions';
import { useClaimedDeals } from '../hooks/useClaimedDeals';
import { useActiveDeals } from '../hooks/useActiveDeals';
import { useUser } from '../hooks/useUser';
import { useToday } from '../hooks/useToday';
import { useNow } from '../hooks/useNow';
import { serverNow } from '../utils/serverClock';
import { AlertsStatus, markAlertsPrimed, shouldPrimeAlerts, usePushAlerts } from '../hooks/usePushNotifications';
import { api, Event, ActiveDeal, ClaimedDeal, Team, Today, UserStats } from '../api/client';
import { navigationRef } from '../navigation/ref';
import { useToast } from '../components/pixel/Toast';
import { teamNameFrom } from '../utils/format';
import { followedTeamIds, liveWonDeals } from '../utils/dealState';
import { useAppConfig } from './AppConfigContext';

interface AppDataContextType {
  // User
  user: ReturnType<typeof useUser>['user'];
  userLoading: boolean;
  userError: string | null;

  // Events
  events: Event[];
  eventsLoading: boolean;
  eventsError: string | null;
  refetchEvents: () => Promise<void>;

  // Teams (season board from the API)
  teams: Team[];
  teamsLoading: boolean;
  teamsError: string | null;
  refetchTeams: () => Promise<void>;

  // Subscriptions
  isSubscribed: (eventId: string) => boolean;
  /** Follow or unfollow offers in one go. Instant, parallel, with a toast if any fail. */
  setSubscribed: (eventIds: string[], on: boolean) => Promise<boolean>;
  /** How many teams you follow (a team with several followed deals counts once). */
  followedCount: number;
  /** The ids of the teams you follow, sorted. */
  followedTeams: string[];
  /** Short team name from the API's team list ("Dodgers"); falls back to the event's full name. */
  teamName: (teamId: string, fallback?: string) => string;
  subsLoading: boolean;

  // Deals
  /** Every deal the server returned: unlocked and dismissed. */
  allDeals: ActiveDeal[];
  undismissedDeals: ActiveDeal[];
  /** Won, not claimed or silenced, not expired (liveWonDeals): the badge, stat line and WON list. */
  liveWon: ActiveDeal[];
  dismissedDeals: ActiveDeal[];
  dismissDeal: (triggeredEventId: string, type: 'got_it' | 'stop_reminding') => Promise<void>;
  undoDismissal: (triggeredEventId: string) => Promise<void>;
  /** Every deal claimed (GOT IT), newest first, kept after the window closes. */
  claimedDeals: ClaimedDeal[];
  dealsLoading: boolean;
  dealsError: string | null;
  /** True when the deals on screen are the saved copy, not a fresh fetch. */
  dealsStale: boolean;
  dealsSavedAt: number | null;
  refetchDeals: () => Promise<void>;

  // Today (GET /today), refreshed on foreground, pushes, pull to refresh and a new Pacific day
  today: Today | null;
  todayLoading: boolean;
  todayError: string | null;
  todaySavedAt: number | null;
  /** Refetch unless fetched in the last minute (tab focus). */
  refreshTodayIfStale: () => Promise<void>;

  // Alerts
  alertsStatus: AlertsStatus;
  expoPushToken: string | null;
  enableAlerts: () => Promise<AlertsStatus>;

  // Stats (null until loaded, or when the request fails)
  stats: UserStats | null;

  // Alerts pause
  /** RFC3339 time alerts are paused until, or null. */
  alertsPausedUntil: string | null;
  /** True while paused (the stored time is still in the future). */
  alertsPaused: boolean;
  /** Pause until a time, or resume with null. Optimistic; reverts with a toast on failure. */
  pauseAlerts: (until: string | null) => Promise<boolean>;

  // Refresh
  refreshAll: () => Promise<void>;
  /** Retry whatever failed: user first, then data. */
  retry: () => Promise<void>;
}

const AppDataContext = createContext<AppDataContextType | null>(null);

export function AppDataProvider({ children }: { children: ReactNode }) {
  const { user, isLoading: userLoading, error: userError, refreshUser } = useUser();
  const { events, isLoading: eventsLoading, error: eventsError, refetch: refetchEvents } = useEvents();
  const { teams, isLoading: teamsLoading, error: teamsError, refetch: refetchTeams } = useTeams();
  const { config } = useAppConfig();
  const showToast = useToast();

  const {
    isLoading: subsLoading,
    setMany,
    isSubscribed,
    refetch: refetchSubs,
  } = useSubscriptions();
  const {
    activeDeals,
    undismissedDeals,
    dismissedDeals,
    dismissDeal: dismissActive,
    undoDismissal: undoActive,
    isLoading: dealsLoading,
    error: dealsError,
    isStale: dealsStale,
    savedAt: dealsSavedAt,
    refetch: refetchDeals,
  } = useActiveDeals();

  const {
    today,
    isLoading: todayLoading,
    error: todayError,
    savedAt: todaySavedAt,
    refresh: refetchToday,
    refreshIfStale: refreshTodayIfStale,
  } = useToday();

  // One shared "live won deals" list, re-checked each minute as deals expire.
  const minute = useNow(60_000);
  const liveWon = useMemo(() => liveWonDeals(undismissedDeals, minute), [undismissedDeals, minute]);

  const followedTeams = useMemo(() => followedTeamIds(events, isSubscribed), [events, isSubscribed]);
  const followedCount = followedTeams.length;
  const teamName = useCallback((teamId: string, fallback?: string) => teamNameFrom(teams, teamId, fallback), [teams]);

  const { claimed: claimedDeals, refetch: refetchClaimed } = useClaimedDeals();
  // Claiming or moving a deal back changes the trophy case, so refresh it after each.
  const dismissDeal = useCallback(
    async (triggeredEventId: string, type: 'got_it' | 'stop_reminding') => {
      await dismissActive(triggeredEventId, type);
      if (type === 'got_it') refetchClaimed();
    },
    [dismissActive, refetchClaimed]
  );
  const undoDismissal = useCallback(
    async (triggeredEventId: string) => {
      await undoActive(triggeredEventId);
      refetchClaimed();
    },
    [undoActive, refetchClaimed]
  );

  const pushEnabled = config.features.enable_push_notifications !== false;
  const { alertsStatus, expoPushToken, enableAlerts } = usePushAlerts(pushEnabled ? (user?.id ?? null) : null);

  // Season stats for the HUD and the You screen. A failed request hides the number, never shows 0.
  const [stats, setStats] = useState<UserStats | null>(null);
  const userId = user?.id;
  const refetchStats = useCallback(async () => {
    if (!userId) return;
    try {
      setStats(await api.getUserStats(userId));
    } catch {
      setStats(null);
    }
  }, [userId]);
  useEffect(() => {
    refetchStats();
  }, [refetchStats, dealsSavedAt, claimedDeals.length]);

  // Alerts pause, seeded from the user and kept optimistically.
  const [alertsPausedUntil, setAlertsPausedUntil] = useState<string | null>(null);
  useEffect(() => {
    setAlertsPausedUntil(user?.alertsPausedUntil ?? null);
  }, [user?.alertsPausedUntil, userId]);
  const alertsPaused = !!alertsPausedUntil && new Date(alertsPausedUntil).getTime() > serverNow();
  const pauseAlerts = useCallback(
    async (until: string | null): Promise<boolean> => {
      if (!userId) return false;
      const previous = alertsPausedUntil;
      setAlertsPausedUntil(until);
      try {
        const res = await api.setAlertsPause(userId, until);
        setAlertsPausedUntil(res.pausedUntil ?? null);
        return true;
      } catch {
        setAlertsPausedUntil(previous);
        showToast(until ? "Couldn't pause alerts. Try again." : "Couldn't resume alerts. Try again.");
        return false;
      }
    },
    [userId, alertsPausedUntil, showToast]
  );

  const refreshAll = useCallback(async () => {
    await Promise.all([
      refetchEvents(),
      refetchDeals(),
      refetchSubs(),
      refetchStats(),
      refetchTeams(),
      refetchClaimed(),
      refetchToday(),
    ]);
  }, [refetchEvents, refetchDeals, refetchSubs, refetchStats, refetchTeams, refetchClaimed, refetchToday]);

  const retry = useCallback(async () => {
    if (userError || !user) await refreshUser();
    await refreshAll();
  }, [userError, user, refreshUser, refreshAll]);

  const setSubscribed = useCallback(
    async (eventIds: string[], on: boolean): Promise<boolean> => {
      if (config.features.enable_subscriptions === false) return false;
      const wasEmpty = followedCount === 0;
      const failed = await setMany(eventIds, on);
      if (failed.length > 0) {
        const names = events.filter((e) => failed.includes(e.id));
        const team = names[0] ? teamName(names[0].teamId, names[0].teamName) : 'that';
        showToast(`Couldn't ${on ? 'follow' : 'unfollow'} ${team} deals. Try again.`);
      }
      // One refetch for the whole batch: newly followed deals may already be unlocked.
      refetchDeals();
      // First follow is the value moment: explain alerts before the system prompt.
      if (on && wasEmpty && failed.length < eventIds.length && navigationRef.isReady()) {
        if (await shouldPrimeAlerts()) {
          markAlertsPrimed();
          navigationRef.navigate('Alerts');
        }
      }
      return failed.length === 0;
    },
    [config.features.enable_subscriptions, followedCount, setMany, events, teamName, showToast, refetchDeals]
  );

  // Refresh data when app comes to foreground
  const appState = useRef(AppState.currentState);
  useEffect(() => {
    const subscription = AppState.addEventListener('change', (nextAppState: AppStateStatus) => {
      if (appState.current?.match(/inactive|background/) && nextAppState === 'active') {
        refreshAll();
      }
      appState.current = nextAppState;
    });
    return () => subscription.remove();
  }, [refreshAll]);

  const value: AppDataContextType = {
    user,
    userLoading,
    userError,
    events,
    eventsLoading,
    eventsError,
    refetchEvents,
    teams,
    teamsLoading,
    teamsError,
    refetchTeams,
    isSubscribed,
    setSubscribed,
    followedCount,
    followedTeams,
    teamName,
    subsLoading,
    allDeals: activeDeals,
    undismissedDeals,
    liveWon,
    dismissedDeals,
    dismissDeal,
    undoDismissal,
    claimedDeals,
    dealsLoading,
    dealsError,
    dealsStale,
    dealsSavedAt,
    refetchDeals,
    today,
    todayLoading,
    todayError,
    todaySavedAt,
    refreshTodayIfStale,
    alertsStatus,
    expoPushToken,
    enableAlerts,
    stats,
    alertsPausedUntil,
    alertsPaused,
    pauseAlerts,
    refreshAll,
    retry,
  };

  return <AppDataContext.Provider value={value}>{children}</AppDataContext.Provider>;
}

export function useAppData() {
  const context = useContext(AppDataContext);
  if (!context) {
    throw new Error('useAppData must be used within AppDataProvider');
  }
  return context;
}
