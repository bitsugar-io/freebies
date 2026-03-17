import React, { createContext, useContext, useState, useEffect, useCallback, useRef } from 'react';
import { AppState, AppStateStatus } from 'react-native';
import { api, AppConfig } from '../api/client';

const DEFAULT_CONFIG: AppConfig = {
  features: {
    enable_mlb: true,
    enable_nba: true,
    enable_nfl: true,
    enable_nhl: false,
    show_affiliate_links: false,
    maintenance_mode: false,
  },
  screens: {
    deals: [
      { type: 'active_deals', key: 'active-deals-list', config: { layout: 'list', emptyTitle: 'No Active Deals', emptySubtitle: 'Deals appear here when your teams trigger offers' } },
    ],
    discover: [
      { type: 'league_filter', key: 'league-filter-bar', config: {} },
      { type: 'event_list', key: 'event-list', config: { groupBy: 'team' } },
    ],
    profile: [
      { type: 'user_stats', key: 'user-stats', config: {} },
      { type: 'subscription_list', key: 'subscriptions', config: {} },
      { type: 'settings', key: 'settings', config: { showThemeToggle: true } },
    ],
  },
};

interface AppConfigContextType {
  config: AppConfig;
  isLoading: boolean;
  refreshConfig: () => Promise<void>;
}

const AppConfigContext = createContext<AppConfigContextType>({
  config: DEFAULT_CONFIG,
  isLoading: true,
  refreshConfig: async () => {},
});

export function AppConfigProvider({ children }: { children: React.ReactNode }) {
  const [config, setConfig] = useState<AppConfig>(DEFAULT_CONFIG);
  const [isLoading, setIsLoading] = useState(true);
  const appState = useRef(AppState.currentState);

  const refreshConfig = useCallback(async () => {
    try {
      const result = await api.getConfig();
      setConfig(result);
    } catch (err) {
      console.warn('Failed to fetch config, using defaults:', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    refreshConfig();
  }, [refreshConfig]);

  useEffect(() => {
    const sub = AppState.addEventListener('change', (nextState: AppStateStatus) => {
      if (appState.current.match(/inactive|background/) && nextState === 'active') {
        refreshConfig();
      }
      appState.current = nextState;
    });
    return () => sub.remove();
  }, [refreshConfig]);

  return (
    <AppConfigContext.Provider value={{ config, isLoading, refreshConfig }}>
      {children}
    </AppConfigContext.Provider>
  );
}

export function useAppConfig() {
  return useContext(AppConfigContext);
}

export { DEFAULT_CONFIG };
