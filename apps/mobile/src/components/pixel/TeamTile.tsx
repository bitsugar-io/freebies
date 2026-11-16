import React from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import Svg, { Defs, Pattern, Rect } from 'react-native-svg';
import type { Sprite } from '../../api/client';
import { colors } from '../../hooks/useTheme';
import { grayHex } from '../PixelSprite';
import { tilePair } from '../../utils/tileColors';

import { PixelBox } from './PixelBox';
import { PixelIcon } from './PixelIcon';
import { SportBadge } from './SportBadge';
import { Text } from './Text';

interface Props {
  /** Short name, "Kings". */
  name: string;
  fullName?: string;
  /** Display league, "NHL". */
  league: string;
  sport: string;
  badge?: Sprite;
  /** Two-tone identity from the API (NES hex). Missing: the panel color. */
  teamColors?: { primary: string; secondary: string };
  /** Offers followed / offers available for this team. */
  followed: number;
  total: number;
  inSeason: boolean;
  /** "MAR" or "FEB 11" when off season. */
  returnLabel?: string | null;
  activeDeals: number;
  busy?: boolean;
  onPress: () => void;
}

/**
 * Team select cell. Top: the team's two colors (primary field with the sport badge, a 6pt
 * secondary stripe), then a black label strip with the name and status, so text never depends on
 * team colors. Off season: colors and badge in grey, a 2px checker over the field and an OFF tag,
 * so a grey or silver team (Kings) can never pass for off season. Followed: white border and a blinking gold
 * check (partial: "1/2"). Season and follow state sit on different channels: color vs border.
 */
export function TeamTile({
  name,
  fullName,
  league,
  sport,
  badge,
  teamColors,
  followed,
  total,
  inSeason,
  returnLabel,
  activeDeals,
  busy,
  onPress,
}: Props) {
  const all = total > 0 && followed === total;
  const some = followed > 0 && !all;
  const checked: boolean | 'mixed' = all ? true : some ? 'mixed' : false;
  const dealWord = activeDeals > 0 ? (activeDeals === 1 ? '1 DEAL' : `${activeDeals} DEALS`) : 'NO DEAL';
  const sub = inSeason ? (activeDeals > 0 ? dealWord : 'NO DEAL YET') : `BACK ${returnLabel ?? 'SOON'}`;
  const subColor = inSeason ? (activeDeals > 0 ? 'success' : 'textMuted') : 'textMuted';
  const status = all ? 'following all deals' : some ? `following ${followed} of ${total} deals` : 'not following';
  const season = inSeason ? '' : `, off season, back ${returnLabel ?? 'soon'}`;
  const deals = activeDeals === 0 ? ', no deal yet' : '';
  // A near-black primary would vanish into the black screen: swap so the lighter color is the field.
  // Navy and other dark-but-visible colors keep their order.
  const pair = teamColors ? tilePair(teamColors.primary, teamColors.secondary) : null;
  const primary0 = pair?.field ?? colors.surface;
  const secondary0 = pair?.stripe ?? colors.textMuted;
  const primary = inSeason ? primary0 : grayHex(primary0);
  const secondary = inSeason ? secondary0 : grayHex(secondary0);
  const border = all || some ? colors.border : colors.textMuted;
  return (
    <Pressable
      onPress={onPress}
      disabled={busy}
      accessibilityRole="checkbox"
      accessibilityLabel={`${fullName ?? name}, ${league}, ${status}${season}${deals}`}
      accessibilityHint={
        activeDeals === 0 ? 'Double tap to learn more' : all ? 'Double tap to unfollow' : 'Double tap to follow all deals'
      }
      accessibilityState={{ checked, busy: !!busy }}
      style={styles.pressable}
    >
      {({ pressed }) => (
        <PixelBox
          fill={colors.background}
          borderColor={border}
          padding={0}
          style={[pressed && { transform: [{ translateX: 2 }, { translateY: 2 }] }]}
        >
          <View style={[styles.field, { backgroundColor: primary }]}>
            {inSeason ? null : <Checker />}
            <SportBadge sport={sport} badge={badge} gray={!inSeason} />
          </View>
          <View style={{ height: 6, backgroundColor: secondary }} />
          <View style={styles.label}>
            {/* One fixed size for every name. Long names wrap to two centered lines in a fixed box. */}
            <View style={styles.nameBox}>
              <Text variant="displayM" color={!inSeason ? 'textMuted' : activeDeals > 0 ? 'text' : 'textSecondary'} align="center">
                {name}
              </Text>
            </View>
            <Text variant="caption" color={subColor} align="center" style={styles.status}>
              {sub}
            </Text>
          </View>
          {inSeason ? null : (
            <View style={styles.off} accessibilityElementsHidden importantForAccessibility="no-hide-descendants">
              <Text variant="displayM" color="text">
                OFF
              </Text>
            </View>
          )}
          {all || some ? (
            // Solid, never blinking: following is a state, and with several followed teams on a page
            // the blink budget made only some checks flicker, which read as "not followed".
            <View style={styles.badge}>
              {all ? (
                <PixelIcon name="check" scale={2} color="accent" />
              ) : (
                <Text variant="displayM" color="accent">
                  {followed}/{total}
                </Text>
              )}
            </View>
          ) : null}
        </PixelBox>
      )}
    </Pressable>
  );
}

/** 2px black checker over the field: an unmistakable off-season cue that doesn't rely on grey. */
function Checker() {
  return (
    <Svg style={StyleSheet.absoluteFill} width="100%" height="100%">
      <Defs>
        <Pattern id="off" patternUnits="userSpaceOnUse" width={4} height={4}>
          <Rect x={0} y={0} width={2} height={2} fill={colors.background} />
          <Rect x={2} y={2} width={2} height={2} fill={colors.background} />
        </Pattern>
      </Defs>
      <Rect x={0} y={0} width="100%" height="100%" fill="url(#off)" />
    </Svg>
  );
}

const styles = StyleSheet.create({
  pressable: { flex: 1 },
  field: { alignItems: 'center', justifyContent: 'center', paddingVertical: 10 },
  label: { paddingHorizontal: 6, paddingTop: 6, paddingBottom: 8, backgroundColor: colors.background },
  // Two lines of the name type, so every tile in the grid is the same height.
  nameBox: { minHeight: 40, justifyContent: 'center', alignSelf: 'stretch' },
  status: { marginTop: 4, alignSelf: 'stretch', fontWeight: '600' },
  off: { position: 'absolute', top: 3, left: 3, backgroundColor: colors.background, paddingHorizontal: 3 },
  // A black square so the gold check reads on any team color.
  badge: { position: 'absolute', top: 3, right: 3, backgroundColor: colors.background, padding: 3, minWidth: 24, alignItems: 'center' },
});
