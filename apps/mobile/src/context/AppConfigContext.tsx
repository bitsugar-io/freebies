import React, { createContext, useContext, useState, useEffect, useCallback, useRef } from 'react';
import { AppState, AppStateStatus } from 'react-native';
import { api, AppConfig } from '../api/client';

// Fallback when /config can't be reached. The app renders Won, Teams and You itself;
// the server only adds banner and promo_card blocks (see ServerBlocks), keyed by screen name.
// The older keys (deals, discover, profile, deal_detail) are still served for builds that
// predate the redesign and are read here only for banners.
const DEFAULT_CONFIG: AppConfig = {
  features: {
    enable_mlb: true,
    enable_nba: true,
    enable_nfl: true,
    enable_nhl: true,
    enable_mls: true,
    show_affiliate_links: false,
    enable_push_notifications: true,
    enable_subscriptions: true,
    maintenance_mode: false,
  },
  screens: {
    tonight: [],
    deals_v2: [],
    teams: [],
    you: [],
    deal_v2: [],
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
      if (appState.current?.match(/inactive|background/) && nextState === 'active') {
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
