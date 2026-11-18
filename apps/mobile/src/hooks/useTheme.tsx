// NES-style design tokens. Dark only: there is no light theme and no theme picker.
// This is the only file (besides sprite data) that may contain color literals.
import { useWindowDimensions } from 'react-native';

export const colors = {
  background: '#000000',
  /** Cards, text boxes (graphite). */
  surface: '#101010',
  /** Selected cells, card header strips, skeleton tone. */
  surfaceSecondary: '#2A2A2A',
  /** Hairline dividers between rows and under section titles. */
  divider: '#3A3A3A',
  text: '#FCFCFC',
  textSecondary: '#BCBCBC',
  /** Disabled and hint text (4.6:1 on surface, 4.8:1 on background). */
  textMuted: '#7C7C7C',
  /** Pixel borders (Dodger silver). */
  border: '#A5ACAF',
  primary: '#F8B800',
  /** Text and icons on the accent fill. */
  onAccent: '#000000',
  accent: '#F8B800',
  accentLight: '#F8D878',
  /** The hard drop under gold buttons. */
  accentShadow: '#8A6410',
  success: '#58D854',
  successDeep: '#007800',
  successBackground: '#007800',
  /** Countdowns, expiring soon. */
  warning: '#FCA044',
  warningBackground: '#000000',
  /** Dusk lilac: the information layer (times, upcoming, section labels); gold stays the action color. */
  info: '#D8B8F8',
  infoBackground: '#000000',
  /** Errors, live dot. Large text and icons only on surface (3.0:1). */
  danger: '#F83800',
  dangerDeep: '#A81000',
  ice: '#A4E4FC',
  skin: '#F0D0B0',
} as const;

export type ColorToken = keyof typeof colors;

export const spacing = { xs: 4, sm: 8, md: 12, lg: 16, xl: 24, xxl: 32 } as const;

/** Art pixel size in points. 1 tile = 8 art px. Integer only. */
export const PX = 3;

export const fonts = {
  /** Press Start 2P: display, HUD, scores, tab labels, menu rows. Short uppercase only. */
  display: 'PressStart2P_400Regular',
} as const;

export type TextVariant =
  | 'displayXl'
  | 'displayL'
  | 'displayM'
  | 'displayS'
  | 'body'
  | 'bodyStrong'
  | 'caption';

interface TypeStyle {
  display: boolean;
  fontSize: number;
  lineHeight: number;
  /** Cap for accessibility text scaling. */
  maxScale: number;
}

/** Type scale. Press Start 2P is never below 12pt. */
export const type: Record<TextVariant, TypeStyle> = {
  displayXl: { display: true, fontSize: 24, lineHeight: 34, maxScale: 1.2 },
  displayL: { display: true, fontSize: 16, lineHeight: 24, maxScale: 1.2 },
  displayM: { display: true, fontSize: 12, lineHeight: 20, maxScale: 1.3 },
  /** Tab labels under their icons. */
  displayS: { display: true, fontSize: 8, lineHeight: 12, maxScale: 1.4 },
  body: { display: false, fontSize: 17, lineHeight: 24, maxScale: 2 },
  bodyStrong: { display: false, fontSize: 17, lineHeight: 24, maxScale: 2 },
  caption: { display: false, fontSize: 13, lineHeight: 18, maxScale: 2 },
};

/** Above this system font scale, display text falls back to the system bold face. */
export const DISPLAY_FALLBACK_SCALE = 1.3;

export interface Theme {
  mode: 'dark';
  isDark: true;
  colors: typeof colors;
  spacing: typeof spacing;
  type: typeof type;
  fonts: typeof fonts;
}

const theme: Theme = { mode: 'dark', isDark: true, colors, spacing, type, fonts };

export function useTheme(): { theme: Theme } {
  return { theme };
}

/** True when the user's text size is large enough that pixel fonts should give way to system text. */
export function useLargeText(): boolean {
  return useWindowDimensions().fontScale > DISPLAY_FALLBACK_SCALE;
}
