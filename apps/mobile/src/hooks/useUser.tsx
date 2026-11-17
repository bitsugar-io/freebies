import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from 'react';
import { Platform } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import * as Application from 'expo-application';
import { api, User } from '../api/client';
import { shouldDiscardStoredUser } from '../api/errors';
import { parseCache } from '../utils/cache';

const USER_KEY = 'freebie_user';

interface UserContextValue {
  user: User | null;
  isLoading: boolean;
  error: string | null;
  refreshUser: () => Promise<void>;
}

const UserContext = createContext<UserContextValue | undefined>(undefined);

async function getStableDeviceId(): Promise<string> {
  try {
    if (Platform.OS === 'ios') {
      const id = await Application.getIosIdForVendorAsync();
      if (id) return `ios_${id}`;
    }
    if (Platform.OS === 'android') {
      const id = await Application.getAndroidId();
      if (id) return `android_${id}`;
    }
  } catch {
    // Platform ID unavailable (e.g., simulator)
  }
  return 'device_' + Math.random().toString(36).substring(2, 15);
}

function getPlatform(): string {
  if (Platform.OS === 'ios') return 'ios';
  if (Platform.OS === 'android') return 'android';
  return 'web';
}

export function UserProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const initUser = async () => {
    try {
      setIsLoading(true);
      setError(null);

      // Load auth token from secure storage first
      await api.loadAuthToken();

      // The stored user is used right away, before the server check, so saved deals and follows
      // (cached per user id) show offline and at launch.
      const savedUser = parseCache<User>(await AsyncStorage.getItem(USER_KEY).catch(() => null), (v) => typeof v.id === 'string');
      if (savedUser) {
        setUser(savedUser);
        setIsLoading(false);
        try {
          const serverUser = await api.getUser(savedUser.id);
          setUser(serverUser);
          AsyncStorage.setItem(USER_KEY, JSON.stringify(serverUser)).catch(() => {});
          return;
        } catch (err) {
          // Offline, a timeout or a server error: keep the stored user. Only a server that says the
          // user is gone (404) or the token is bad (401/403) starts a new one.
          if (!shouldDiscardStoredUser(err)) {
            console.warn('Could not verify the stored user; keeping it:', err);
            return;
          }
          await AsyncStorage.removeItem(USER_KEY);
          await api.clearAuthToken();
          setUser(null);
          setIsLoading(true);
        }
      }

      // Create new anonymous user (stores token automatically)
      const deviceId = await getStableDeviceId();
      const platform = getPlatform();
      const newUser = await api.createUser(deviceId, platform);

      // Save user to storage (token is saved separately in SecureStore)
      await AsyncStorage.setItem(USER_KEY, JSON.stringify(newUser));
      setUser(newUser);
    } catch (err) {
      console.error('Failed to initialize user:', err);
      setError(err instanceof Error ? err.message : 'Failed to initialize');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    initUser();
  }, []);

  const refreshUser = async () => {
    await initUser();
  };

  return (
    <UserContext.Provider value={{ user, isLoading, error, refreshUser }}>
      {children}
    </UserContext.Provider>
  );
}

export function useUser() {
  const context = useContext(UserContext);
  if (!context) {
    throw new Error('useUser must be used within a UserProvider');
  }
  return context;
}
