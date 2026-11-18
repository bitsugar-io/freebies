import { Platform } from 'react-native';
import { requireOptionalNativeModule } from 'expo-modules-core';
import type { WinBoardState } from './winBoardState';

interface WinBoardModule {
  isSupported(): boolean;
  sync(json: string, staleAt: number): Promise<void>;
}

// Missing in Expo Go and on Android: the Win Board only exists in a real iOS build.
const native = Platform.OS === 'ios' ? requireOptionalNativeModule<WinBoardModule>('WinBoard') : null;

/**
 * Shows the Win Board Live Activity for the current live wins, updates it when they change, and
 * ends it when there are none (claimed or expired). `state` comes from winBoardState().
 */
export function syncWinBoard(state: WinBoardState | null): void {
  if (!native) return;
  try {
    if (state && !native.isSupported()) return;
    native.sync(JSON.stringify(state), state?.staleAt ?? 0).catch(() => {});
  } catch {
    // The Win Board is extra; never let it break the app.
  }
}
