import { Platform } from 'react-native';
import { ExtensionStorage } from '@bacons/apple-targets';

/** App Group shared with the lock-screen widget (targets/widget). Must match app.json. */
export const APP_GROUP = 'group.io.bitsugar.freebies';

// A no-op in Expo Go and on Android: the native module only exists in a real iOS build.
const storage = Platform.OS === 'ios' ? new ExtensionStorage(APP_GROUP) : null;

/**
 * Hands the widget what it needs to fetch its own lines from GET /users/{id}/widget: the API base,
 * the user id and the user's token. The widget refreshes from the API on its own, so a win reaches
 * the lock screen without the app being opened.
 */
export function shareWithWidget(apiBase: string, userId: string, token: string | null): void {
  if (!storage) return;
  try {
    storage.set('apiBase', apiBase);
    storage.set('userId', userId);
    if (token) storage.set('token', token);
    ExtensionStorage.reloadWidget();
  } catch {
    // The widget is extra; never let it break the app.
  }
}

/** Asks iOS to redraw the widget now (a follow, claim or new win changed its lines). */
export function refreshWidget(): void {
  if (!storage) return;
  try {
    ExtensionStorage.reloadWidget();
  } catch {
    // ignore
  }
}
