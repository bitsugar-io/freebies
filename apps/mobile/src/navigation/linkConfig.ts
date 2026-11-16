import type { LinkingOptions } from '@react-navigation/native';
import type { RootStackParamList } from './types';

/**
 * Deep link routes, kept free of native imports so a plain unit test can run getStateFromPath on
 * every path. Push taps use freebie://deal/<triggeredEventId>; Expo Go uses exp://host/--/<path>.
 */
export const LINK_CONFIG: NonNullable<LinkingOptions<RootStackParamList>['config']> = {
  initialRouteName: 'Tabs',
  screens: {
    Deal: 'deal/:id',
    Pause: 'pause',
    Tabs: {
      screens: {
        // Old /tonight and /deals links still land on the home tab.
        Won: { path: 'won', alias: ['tonight', 'deals'] },
        Today: 'today',
        // /teams is SELECT SPORT; /teams/<sport> opens that sport's page.
        Teams: { path: 'teams', screens: { Sports: '', Sport: ':sport' } },
        You: 'you',
      },
    },
  },
};
