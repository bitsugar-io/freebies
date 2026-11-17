import React, { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react';
import { AccessibilityInfo, Pressable, StyleSheet, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { PixelBox } from './PixelBox';
import { Text } from './Text';

interface ToastOptions {
  /** Label for an inline action, e.g. "UNDO". */
  actionLabel?: string;
  onAction?: () => void;
  durationMs?: number;
}

interface ToastItem extends ToastOptions {
  id: number;
  message: string;
}

const ToastContext = createContext<(message: string, options?: ToastOptions) => void>(() => {});

export function useToast() {
  return useContext(ToastContext);
}

// Clears the HUD tab bar (stat line, 48pt tabs and the 3pt border).
const TAB_BAR_CLEARANCE = 88;

/** Deep strip above the tab bar. Announced to screen readers. */
export function ToastProvider({ children }: { children: React.ReactNode }) {
  const insets = useSafeAreaInsets();
  const [toast, setToast] = useState<ToastItem | null>(null);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const counter = useRef(0);

  const dismiss = useCallback(() => {
    if (timer.current) clearTimeout(timer.current);
    setToast(null);
  }, []);

  const show = useCallback(
    (message: string, options: ToastOptions = {}) => {
      if (timer.current) clearTimeout(timer.current);
      counter.current += 1;
      setToast({ id: counter.current, message, ...options });
      AccessibilityInfo.announceForAccessibility(message);
      timer.current = setTimeout(() => setToast(null), options.durationMs ?? 4000);
    },
    []
  );

  useEffect(() => () => {
    if (timer.current) clearTimeout(timer.current);
  }, []);

  return (
    <ToastContext.Provider value={show}>
      {children}
      {toast ? (
        <View
          pointerEvents="box-none"
          style={[styles.wrap, { bottom: insets.bottom + TAB_BAR_CLEARANCE }]}
          accessibilityLiveRegion="polite"
        >
          <PixelBox variant="notice" padding={8}>
            <View style={styles.row}>
              <View style={{ flex: 1 }}>
                <Text variant="bodyStrong" color="onAccent" style={{ fontSize: 15, lineHeight: 20 }}>
                  {toast.message}
                </Text>
              </View>
              {toast.actionLabel ? (
                <Pressable
                  onPress={() => {
                    toast.onAction?.();
                    dismiss();
                  }}
                  accessibilityRole="button"
                  accessibilityLabel={toast.actionLabel}
                  hitSlop={8}
                  style={styles.action}
                >
                  <Text variant="displayM" color="onAccent" style={{ textDecorationLine: 'underline' }}>
                    {toast.actionLabel}
                  </Text>
                </Pressable>
              ) : null}
            </View>
          </PixelBox>
        </View>
      ) : null}
    </ToastContext.Provider>
  );
}

const styles = StyleSheet.create({
  wrap: { position: 'absolute', left: 16, right: 16 },
  row: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  action: { minHeight: 44, minWidth: 44, alignItems: 'center', justifyContent: 'center' },
});
