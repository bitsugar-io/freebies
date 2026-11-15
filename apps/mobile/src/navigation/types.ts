import type { NavigatorScreenParams } from '@react-navigation/native';

/** Teams is a small stack inside its tab: SELECT SPORT, then one sport's teams. */
export type TeamsStackParamList = {
  Sports: undefined;
  /** `sport` is the API's sport id, e.g. "hockey". */
  Sport: { sport: string };
};

export type TabParamList = {
  /** Home: deals to claim, on the line, and coming up. */
  Won: undefined;
  Teams: NavigatorScreenParams<TeamsStackParamList> | undefined;
  /** `anchor: 'about'` scrolls to the About section; `at` makes a repeat request count. */
  You: { anchor?: 'about'; at?: number } | undefined;
};

export type RootStackParamList = {
  Tabs: NavigatorScreenParams<TabParamList> | undefined;
  /** `id` is a triggered-event id (an unlocked deal) or an event id (a preview). */
  Deal: { id: string };
  /** Explainer shown before the system notification prompt. */
  Alerts: undefined;
  /** The NES pause menu, shown over everything (including the deal modal). */
  Pause: undefined;
};
