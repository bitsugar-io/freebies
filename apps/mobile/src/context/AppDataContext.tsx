import React, { createContext, useContext, useCallback, useEffect, useRef, useState, ReactNode } from 'react';
import { AppState, AppStateStatus } from 'react-native';
import { useEvents } from '../hooks/useEvents';
import { useSubscriptions } from '../hooks/useSubscriptions';
import { useActiveDeals } from '../hooks/useActiveDeals';
import { useUser } from '../hooks/useUser';
import { usePushNotifications } from '../hooks/usePushNotifications';
import { Event, ActiveDeal, League, api } from '../api/client';
import { useAppConfig } from './AppConfigContext';

interface AppDataContextType {
  // User
  user: ReturnType<typeof useUser>['user'];
  userLoading: boolean;
  userError: string | null;

  // Leagues
  leagues: League[];
  leaguesLoading: boolean;

  // Events
  events: Event[];
  eventsLoading: boolean;
  eventsError: string | null;
  refetchEvents: () => Promise<void>;

  // Subscriptions
  isSubscribed: (eventId: string) => boolean;
  toggleSubscription: (eventId: string) => Promise<void>;
  subscribedCount: number;
  subsLoading: boolean;

  // Deals
  undismissedDeals: ActiveDeal[];
  dismissedDeals: ActiveDeal[];
  dismissDeal: (triggeredEventId: string, type: 'got_it' | 'stop_reminding') => Promise<void>;
  undoDismissal: (triggeredEventId: string) => Promise<void>;
  dealsLoading: boolean;
  refetchDeals: () => Promise<void>;

  // Push
  expoPushToken: string | null;

  // Refresh
  refreshAll: () => Promise<void>;

  // Modal handlers (for deep linking)
  openDealModal: (eventId: string) => void;
  setModalHandlers: (handlers: {
    setSelectedEvent: (event: Event | null) => void;
    setSelectedDeal: (deal: ActiveDeal | null) => void;
    setModalVisible: (visible: boolean) => void;
  }) => void;
}

const AppDataContext = createContext<AppDataContextType | null>(null);

export function AppDataProvider({ children }: { children: ReactNode }) {
  const { user, isLoading: userLoading, error: userError } = useUser();
  const { events, isLoading: eventsLoading, error: eventsError, refetch: refetchEvents } = useEvents();
  const { config } = useAppConfig();

  // Leagues state
  const [leagues, setLeagues] = useState<League[]>([]);
  const [leaguesLoading, setLeaguesLoading] = useState(true);

  // Fetch leagues on mount
  useEffect(() => {
    const fetchLeagues = async () => {
      try {
        const data = await api.listLeagues();
        setLeagues(data);
      } catch (err) {
        console.error('Failed to fetch leagues:', err);
      } finally {
        setLeaguesLoading(false);
      }
    };
    fetchLeagues();
  }, []);
  const {
    isLoading: subsLoading,
    toggleSubscription: _toggleSubscription,
    isSubscribed,
    subscribedCount,
  } = useSubscriptions();
  const {
    undismissedDeals,
    dismissedDeals,
    dismissDeal,
    undoDismissal,
    isLoading: dealsLoading,
    refetch: refetchDeals,
  } = useActiveDeals();

  // Modal handlers ref for deep linking
  const modalHandlersRef = useRef<{
    setSelectedEvent: (event: Event | null) => void;
    setSelectedDeal: (deal: ActiveDeal | null) => void;
    setModalVisible: (visible: boolean) => void;
  } | null>(null);

  const setModalHandlers = useCallback((handlers: typeof modalHandlersRef.current) => {
    modalHandlersRef.current = handlers;
  }, []);

  const openDealModal = useCallback((eventId: string) => {
    if (!modalHandlersRef.current) return;

    const deal = undismissedDeals.find(d => d.eventId === eventId);
    if (deal) {
      modalHandlersRef.current.setSelectedEvent(deal.event);
      modalHandlersRef.current.setSelectedDeal(deal);
      modalHandlersRef.current.setModalVisible(true);
      return;
    }

    const event = events.find(e => e.id === eventId);
    if (event) {
      modalHandlersRef.current.setSelectedEvent(event);
      modalHandlersRef.current.setSelectedDeal(null);
      modalHandlersRef.current.setModalVisible(true);
    }
  }, [events, undismissedDeals]);

  const refreshAll = useCallback(async () => {
    await Promise.all([refetchEvents(), refetchDeals()]);
  }, [refetchEvents, refetchDeals]);

  // Wrap toggle subscription to refetch deals after (gated by feature flag)
  const toggleSubscription = useCallback(async (eventId: string) => {
    if (config.features.enable_subscriptions === false) return;
    await _toggleSubscription(eventId);
    refetchDeals();
  }, [_toggleSubscription, refetchDeals, config.features.enable_subscriptions]);

  // Refresh data when app comes to foreground
  const appState = useRef(AppState.currentState);
  useEffect(() => {
    const subscription = AppState.addEventListener('change', (nextAppState: AppStateStatus) => {
      if (appState.current.match(/inactive|background/) && nextAppState === 'active') {
        refreshAll();
      }
      appState.current = nextAppState;
    });
    return () => subscription.remove();
  }, [refreshAll]);

  // Handle notification tap
  const handleNotificationTap = useCallback((eventId: string) => {
    console.log('Opening deal from notification:', eventId);
    openDealModal(eventId);
  }, [openDealModal]);

  // Handle notification received in foreground
  const handleNotificationReceived = useCallback(() => {
    refreshAll();
  }, [refreshAll]);

  // Register for push notifications (only if enabled)
  const pushEnabled = config.features.enable_push_notifications !== false;
  const { expoPushToken } = usePushNotifications(
    pushEnabled ? (user?.id ?? null) : null,
    pushEnabled ? handleNotificationTap : undefined,
    pushEnabled ? handleNotificationReceived : undefined
  );

  const value: AppDataContextType = {
    user,
    userLoading,
    userError,
    leagues,
    leaguesLoading,
    events,
    eventsLoading,
    eventsError,
    refetchEvents,
    isSubscribed,
    toggleSubscription,
    subscribedCount,
    subsLoading,
    undismissedDeals,
    dismissedDeals,
    dismissDeal,
    undoDismissal,
    dealsLoading,
    refetchDeals,
    expoPushToken,
    refreshAll,
    openDealModal,
    setModalHandlers,
  };

  return (
    <AppDataContext.Provider value={value}>
      {children}
    </AppDataContext.Provider>
  );
}

export function useAppData() {
  const context = useContext(AppDataContext);
  if (!context) {
    throw new Error('useAppData must be used within AppDataProvider');
  }
  return context;
}
