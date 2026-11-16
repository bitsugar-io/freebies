import React, { useEffect, useRef, useSyncExternalStore } from 'react';
import { AccessibilityInfo, StyleProp, View, ViewStyle } from 'react-native';
import { NavigationContext } from '@react-navigation/native';

/**
 * The one blink clock for the whole app: 1 Hz (500ms on, 500ms off), stepped, no easing. A single
 * interval serves every subscriber and runs only while something is subscribed. Under Reduce
 * Motion the interval never starts and everything is solid.
 *
 * Budget: at most MAX_BLINKERS things blink at once. Each blinker registers with a priority
 * (lower wins; ties go to whoever registered first) while its screen is focused. The rest stay
 * solid. Priorities in use: 0 pause/title, 1 PAUSED, 2 active tab arrow, 3 unlocked badge,
 * 4 NEW tag, 5 ENDS SOON, 6 menu cursor, 7 countdown colon.
 *
 * Rules (docs/ui.md, Motion): blink only selection cursors, tags and status words; never body
 * text; never faster than 1 Hz.
 */
const HALF_PERIOD_MS = 500;
export const MAX_BLINKERS = 3;

let phase = true;
let reduced = false;
let timer: ReturnType<typeof setInterval> | null = null;
let reduceSub: { remove: () => void } | null = null;
let nextId = 1;
const listeners = new Set<() => void>();
const slots: { id: number; priority: number }[] = [];
let allowed = new Set<number>();

function emit() {
  listeners.forEach((l) => l());
}

function syncTimer() {
  if (!reduced && listeners.size > 0 && !timer) {
    timer = setInterval(() => {
      phase = !phase;
      emit();
    }, HALF_PERIOD_MS);
  } else if ((reduced || listeners.size === 0) && timer) {
    clearInterval(timer);
    timer = null;
    phase = true;
  }
}

function recompute() {
  const ranked = [...slots].sort((a, b) => a.priority - b.priority || a.id - b.id);
  allowed = new Set(ranked.slice(0, MAX_BLINKERS).map((s) => s.id));
  emit();
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  if (listeners.size === 1) {
    AccessibilityInfo.isReduceMotionEnabled()
      .then((v) => {
        reduced = v;
        syncTimer();
        emit();
      })
      .catch(() => {});
    reduceSub = AccessibilityInfo.addEventListener('reduceMotionChanged', (v) => {
      reduced = v;
      syncTimer();
      emit();
    });
  }
  syncTimer();
  return () => {
    listeners.delete(listener);
    if (listeners.size === 0) {
      reduceSub?.remove();
      reduceSub = null;
    }
    syncTimer();
  };
}

/** True while the screen this component is on is focused (always true outside a navigator). */
function useScreenFocused(): boolean {
  const nav = React.useContext(NavigationContext);
  const [focused, setFocused] = React.useState(() => (nav ? nav.isFocused() : true));
  useEffect(() => {
    if (!nav) return undefined;
    setFocused(nav.isFocused());
    const a = nav.addListener('focus', () => setFocused(true));
    const b = nav.addListener('blur', () => setFocused(false));
    return () => {
      a();
      b();
    };
  }, [nav]);
  return focused;
}

/** The raw clock, outside the budget: loading skeleton tones and the ▼ bob. True under Reduce Motion. */
export function useBlinkClock(): boolean {
  return useSyncExternalStore(subscribe, () => (reduced ? true : phase), () => true);
}

/**
 * True during the visible half of a blink. Solid (true) when Reduce Motion is on, when `active`
 * is false, or when the blink budget is spent on higher-priority things.
 */
export function useBlink(priority = 6, active = true): boolean {
  const idRef = useRef(0);
  if (idRef.current === 0) idRef.current = nextId++;
  const id = idRef.current;
  const focused = useScreenFocused();
  const live = active && focused;
  useEffect(() => {
    if (!live) return undefined;
    slots.push({ id, priority });
    recompute();
    return () => {
      const i = slots.findIndex((s) => s.id === id);
      if (i >= 0) slots.splice(i, 1);
      recompute();
    };
  }, [id, priority, live]);
  return useSyncExternalStore(
    subscribe,
    () => (reduced || !live || !allowed.has(id) ? true : phase),
    () => true
  );
}

interface BlinkProps {
  children: React.ReactNode;
  priority?: number;
  /** Set false to show the children solid. */
  active?: boolean;
  style?: StyleProp<ViewStyle>;
}

/**
 * Hard on/off visibility for its children on the shared clock. The layout box stays put so
 * nothing shifts. It is a visibility switch, not a fade: there is no in-between.
 */
export function Blink({ children, priority = 6, active = true, style }: BlinkProps) {
  const on = useBlink(priority, active);
  return <View style={[style, !on && { opacity: 0 }]}>{children}</View>;
}
