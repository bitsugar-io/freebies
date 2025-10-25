import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useColorScheme } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';

const THEME_KEY = 'freebie_theme';

export type ThemeMode = 'light' | 'dark' | 'system';

export interface Theme {
  mode: ThemeMode;
  isDark: boolean;
  colors: {
    background: string;
    surface: string;
    surfaceSecondary: string;
    text: string;
    textSecondary: string;
    textMuted: string;
    border: string;
    primary: string;
    primaryText: string;
    accent: string;
    success: string;
    successBackground: string;
    warning: string;
    warningBackground: string;
    info: string;
    infoBackground: string;
  };
}

const lightColors: Theme['colors'] = {
  background: '#f5f5f5',
  surface: '#ffffff',
  surfaceSecondary: '#f0f0f0',
  text: '#1a1a2e',
  textSecondary: '#333333',
  textMuted: '#666666',
  border: '#eeeeee',
  primary: '#1a1a2e',
  primaryText: '#ffffff',
  accent: '#4CAF50',
  success: '#4CAF50',
  successBackground: '#e8f5e9',
  warning: '#f57c00',
  warningBackground: '#fff3e0',
  info: '#1976d2',
  infoBackground: '#e3f2fd',
};

const darkColors: Theme['colors'] = {
  background: '#121212',
  surface: '#1e1e1e',
  surfaceSecondary: '#2a2a2a',
  text: '#ffffff',
  textSecondary: '#e0e0e0',
  textMuted: '#9e9e9e',
  border: '#333333',
  primary: '#bb86fc',
  primaryText: '#000000',
  accent: '#03dac6',
  success: '#03dac6',
  successBackground: '#1a3a36',
  warning: '#ffb74d',
  warningBackground: '#3d2e1a',
  info: '#64b5f6',
  infoBackground: '#1a2a3d',
};

interface ThemeContextValue {
  theme: Theme;
  setThemeMode: (mode: ThemeMode) => void;
  toggleTheme: () => void;
}

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined);

export function ThemeProvider({ children }: { children: ReactNode }) {
  const systemColorScheme = useColorScheme();
  const [mode, setMode] = useState<ThemeMode>('system');
  const [isLoaded, setIsLoaded] = useState(false);

  // Load saved theme preference
  useEffect(() => {
    async function loadTheme() {
      try {
        const saved = await AsyncStorage.getItem(THEME_KEY);
        if (saved && ['light', 'dark', 'system'].includes(saved)) {
          setMode(saved as ThemeMode);
        }
      } catch (error) {
        console.error('Failed to load theme:', error);
      } finally {
        setIsLoaded(true);
      }
    }
    loadTheme();
  }, []);

  // Save theme preference
  const setThemeMode = async (newMode: ThemeMode) => {
    setMode(newMode);
    try {
      await AsyncStorage.setItem(THEME_KEY, newMode);
    } catch (error) {
      console.error('Failed to save theme:', error);
    }
  };

  const toggleTheme = () => {
    const nextMode = mode === 'light' ? 'dark' : mode === 'dark' ? 'system' : 'light';
    setThemeMode(nextMode);
  };

  const isDark = mode === 'system'
    ? systemColorScheme === 'dark'
    : mode === 'dark';

  const theme: Theme = {
    mode,
    isDark,
    colors: isDark ? darkColors : lightColors,
  };

  if (!isLoaded) {
    return null;
  }

  return (
    <ThemeContext.Provider value={{ theme, setThemeMode, toggleTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
}
