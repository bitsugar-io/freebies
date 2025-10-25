import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from 'react';
import { Platform } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { api, User } from '../api/client';

const USER_KEY = 'freebie_user';

interface UserContextValue {
  user: User | null;
  isLoading: boolean;
  error: string | null;
  refreshUser: () => Promise<void>;
}

const UserContext = createContext<UserContextValue | undefined>(undefined);

function generateDeviceId(): string {
  // Generate a random device ID for anonymous users
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

      // Check for existing user in storage
      const stored = await AsyncStorage.getItem(USER_KEY);
      if (stored) {
        const savedUser = JSON.parse(stored) as User;
        // Verify user still exists on server (requires auth token)
        try {
          const serverUser = await api.getUser(savedUser.id);
          setUser(serverUser);
          return;
        } catch {
          // User doesn't exist on server or token invalid, create new one
          await AsyncStorage.removeItem(USER_KEY);
          await api.clearAuthToken();
        }
      }

      // Create new anonymous user (stores token automatically)
      const deviceId = generateDeviceId();
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
