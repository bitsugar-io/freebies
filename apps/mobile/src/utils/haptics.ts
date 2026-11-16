import * as Haptics from 'expo-haptics';

// Haptics are a nicety: never let a missing engine (web, simulator) surface as an error.
export const haptics = {
  /** Toggling a team or deal. */
  selection: () => Haptics.selectionAsync().catch(() => {}),
  /** A deal unlocked or was claimed. */
  success: () => Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => {}),
  /** Dismissing a deal. */
  light: () => Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light).catch(() => {}),
};
