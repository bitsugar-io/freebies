import React, { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react';
import { AccessibilityInfo, StyleSheet, View } from 'react-native';
import { colors } from '../../hooks/useTheme';

const COLUMNS = 8;
const STEP_MS = 30;
// Columns covered at each of the 8 steps: black sweeps in from the left, then out to the right.
const COVER = [2, 4, 6, 8, 6, 4, 2, 0];

type RunWipe = (atFullCover: () => void) => void;
const WipeContext = createContext<RunWipe>((fn) => fn());

/** Runs `fn` at the midpoint of a stepped black wipe (8 steps, about 240ms). Instant under Reduce Motion. */
export function useScreenWipe(): RunWipe {
  return useContext(WipeContext);
}

/** Lightweight overlay for tab-switch wipes. It blocks touches only while a wipe is running. */
export function ScreenWipeProvider({ children }: { children: React.ReactNode }) {
  const [step, setStep] = useState(-1);
  const reduced = useRef(false);
  const timer = useRef<ReturnType<typeof setInterval> | null>(null);
  const pending = useRef<(() => void) | null>(null);

  useEffect(() => {
    AccessibilityInfo.isReduceMotionEnabled().then((v) => (reduced.current = v)).catch(() => {});
    const sub = AccessibilityInfo.addEventListener('reduceMotionChanged', (v) => (reduced.current = v));
    return () => {
      sub.remove();
      if (timer.current) clearInterval(timer.current);
    };
  }, []);

  const run = useCallback<RunWipe>((fn) => {
    if (reduced.current || timer.current) {
      fn();
      return;
    }
    pending.current = fn;
    let i = 0;
    setStep(0);
    timer.current = setInterval(() => {
      i += 1;
      if (i === 3 && pending.current) {
        pending.current();
        pending.current = null;
      }
      if (i >= COVER.length) {
        if (timer.current) clearInterval(timer.current);
        timer.current = null;
        setStep(-1);
        return;
      }
      setStep(i);
    }, STEP_MS);
  }, []);

  const covered = step >= 0 ? COVER[step] : 0;
  const fromLeft = step < 4;
  return (
    <WipeContext.Provider value={run}>
      {children}
      {step >= 0 ? (
        <View
          style={StyleSheet.absoluteFill}
          pointerEvents="auto"
          accessibilityElementsHidden
          importantForAccessibility="no-hide-descendants"
        >
          <View style={[styles.row, { justifyContent: fromLeft ? 'flex-start' : 'flex-end' }]}>
            {Array.from({ length: covered }, (_, i) => (
              <View key={i} style={styles.col} />
            ))}
          </View>
        </View>
      ) : null}
    </WipeContext.Provider>
  );
}

const styles = StyleSheet.create({
  row: { flex: 1, flexDirection: 'row' },
  col: { width: `${100 / COLUMNS}%`, backgroundColor: colors.background },
});
