import React, { useEffect, useMemo, useRef, useState } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { AccessibilityInfo, Pressable, StyleSheet, View } from 'react-native';
import Svg, { Polygon } from 'react-native-svg';
import { ActiveDeal } from '../../api/client';
import { useAppData } from '../../context/AppDataContext';
import { colors } from '../../hooks/useTheme';
import { navigationRef } from '../../navigation/ref';
import { CutInFrame, cutInTimeline, winsToCelebrate } from '../../utils/cutin';
import PixelSprite from '../PixelSprite';
import { Text } from './Text';

const CELEBRATED_KEY = 'freebie_celebrated_wins';
const KEEP = 60;

/** Full-screen hard-edged rays in two colors (the YOU WON burst, screen-sized). */
function Rays({ a, b }: { a: string; b: string }) {
  const W = 300;
  const H = 650;
  const cx = W / 2;
  const cy = H * 0.42;
  const r = H;
  const n = 24;
  return (
    <Svg style={StyleSheet.absoluteFill} viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="xMidYMid slice">
      {Array.from({ length: n }, (_, i) => {
        const t0 = (i / n) * 2 * Math.PI;
        const t1 = ((i + 1) / n) * 2 * Math.PI;
        const pts = `${cx},${cy} ${cx + r * Math.cos(t0)},${cy + r * Math.sin(t0)} ${cx + r * Math.cos(t1)},${cy + r * Math.sin(t1)}`;
        return <Polygon key={i} points={pts} fill={i % 2 === 0 ? a : b} />;
      })}
    </Svg>
  );
}

/**
 * The victory cut-in (docs/ui.md Motion): the first time the app has fresh data with a win you
 * haven't seen celebrated. One win: a team-color flash, rays, the food sprite dropping in, "RAMS
 * WON!" typing out, then the deal opens. Several: one "N WINS!" with the sprites stacking on a
 * shelf, then WON. Under 2 seconds, a tap skips it, and Reduce Motion skips it entirely.
 */
export function VictoryCutInHost() {
  const { liveWon, teams, dealsLoading, dealsStale } = useAppData();
  const [celebrated, setCelebrated] = useState<Set<string> | null>(null);
  const [playing, setPlaying] = useState<ActiveDeal[] | null>(null);

  useEffect(() => {
    AsyncStorage.getItem(CELEBRATED_KEY)
      .then((raw) => {
        const ids = raw ? (JSON.parse(raw) as unknown) : [];
        setCelebrated(new Set(Array.isArray(ids) ? ids.filter((x) => typeof x === 'string') : []));
      })
      .catch(() => setCelebrated(new Set()));
  }, []);

  // Fresh deals only: a saved list from yesterday must not replay or hide a win.
  useEffect(() => {
    if (!celebrated || playing || dealsLoading || dealsStale) return;
    const fresh = winsToCelebrate(liveWon, celebrated);
    if (fresh.length === 0) return;
    let cancelled = false;
    AccessibilityInfo.isReduceMotionEnabled()
      .catch(() => false)
      .then((reduced) => {
        if (cancelled) return;
        if (reduced) remember(fresh);
        else setPlaying(fresh);
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [celebrated, liveWon, dealsLoading, dealsStale, playing]);

  const remember = (wins: ActiveDeal[]) => {
    const next = new Set(celebrated ?? []);
    wins.forEach((w) => next.add(w.id));
    setCelebrated(next);
    AsyncStorage.setItem(CELEBRATED_KEY, JSON.stringify([...next].slice(-KEEP))).catch(() => {});
  };

  const done = () => {
    const wins = playing ?? [];
    remember(wins);
    setPlaying(null);
    if (!navigationRef.isReady() || wins.length === 0) return;
    if (wins.length === 1) navigationRef.navigate('Deal', { id: wins[0].id });
    else navigationRef.navigate('Tabs', { screen: 'Won' });
  };

  if (!playing) return null;
  const first = playing[0];
  const team = teams.find((t) => t.id === first.event.teamId);
  return (
    <VictoryCutIn
      wins={playing}
      teamName={(team?.name ?? first.event.teamName).toUpperCase()}
      pair={team?.colors ? [team.colors.primary, team.colors.secondary] : [colors.accent, colors.background]}
      onDone={done}
    />
  );
}

function VictoryCutIn({
  wins,
  teamName,
  pair,
  onDone,
}: {
  wins: ActiveDeal[];
  teamName: string;
  pair: [string, string];
  onDone: () => void;
}) {
  const banner = wins.length === 1 ? `${teamName} WON!` : `${wins.length} WINS!`;
  const frames = useMemo<CutInFrame[]>(() => cutInTimeline(wins.length, banner), [wins.length, banner]);
  const [i, setI] = useState(0);
  const finished = useRef(false);
  const finish = () => {
    if (finished.current) return;
    finished.current = true;
    onDone();
  };

  useEffect(() => {
    AccessibilityInfo.announceForAccessibility(
      wins.length === 1 ? `${banner} ${wins[0].event.offerName}` : `${banner} Open WON to see them.`
    );
  }, [banner, wins]);

  useEffect(() => {
    if (i >= frames.length) {
      finish();
      return;
    }
    const t = setTimeout(() => setI((n) => n + 1), frames[i].hold);
    return () => clearTimeout(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [i, frames]);

  const f = frames[Math.min(i, frames.length - 1)];
  const many = wins.length > 1;
  const rayPair: [string, string] = many ? [colors.background, colors.surfaceSecondary] : pair;
  const dropTop = f.kind === 'drop' ? 60 + f.step * 45 : 195;
  const showSprite = !many && (f.kind === 'drop' || f.kind === 'type');
  const text = f.kind === 'type' ? f.text : many && f.kind === 'shelf' ? banner : '';

  return (
    <Pressable
      style={[StyleSheet.absoluteFill, styles.root]}
      onPress={finish}
      accessibilityRole="button"
      accessibilityLabel={`${banner} Tap to continue.`}
    >
      {f.kind === 'flash' ? (
        <View style={[StyleSheet.absoluteFill, { backgroundColor: pair[0] }]} />
      ) : (
        <Rays a={rayPair[0]} b={rayPair[1]} />
      )}
      {showSprite ? (
        <View style={[styles.sprite, { top: `${dropTop / 6.5}%` }]}>
          <PixelSprite sprite={wins[0].event.sprite} scale={7} />
        </View>
      ) : null}
      {text ? (
        <View style={[styles.banner, many ? styles.bannerHigh : null]}>
          <Text variant="displayXl" color={many ? 'accent' : 'text'} style={styles.bannerText}>
            {text}
          </Text>
        </View>
      ) : null}
      {many && f.kind === 'shelf' ? (
        <View style={styles.shelf}>
          {wins.slice(0, f.shown).map((w) => (
            <View key={w.id} style={styles.slot}>
              <PixelSprite sprite={w.event.sprite} scale={3} />
            </View>
          ))}
        </View>
      ) : null}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  root: { zIndex: 50, backgroundColor: colors.background },
  sprite: { position: 'absolute', left: 0, right: 0, alignItems: 'center' },
  banner: { position: 'absolute', left: 0, right: 0, top: '62%', alignItems: 'center', paddingHorizontal: 16 },
  bannerHigh: { top: '28%' },
  bannerText: { textAlign: 'center', textShadowColor: colors.background, textShadowOffset: { width: 3, height: 3 }, textShadowRadius: 0 },
  shelf: {
    position: 'absolute',
    top: '42%',
    left: 16,
    right: 16,
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'center',
    gap: 12,
  },
  slot: { width: 64, alignItems: 'center' },
});
