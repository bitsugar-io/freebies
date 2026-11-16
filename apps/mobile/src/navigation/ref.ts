import { createNavigationContainerRef } from '@react-navigation/native';
import type { RootStackParamList } from './types';

/** Lets code outside screens (data layer, push handling) navigate. */
export const navigationRef = createNavigationContainerRef<RootStackParamList>();
