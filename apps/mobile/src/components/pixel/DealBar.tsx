import React from 'react';
import { StyleSheet, View } from 'react-native';
import { colors } from '../../hooks/useTheme';
import { useNow } from '../../hooks/useNow';
import { serverNow } from '../../utils/serverClock';

interface BarProps {
  /** 0..1 */
  value: number;
  cells?: number;
  tone?: 'success' | 'warning';
  accessibilityLabel?: string;
}

/** HP-style bar of discrete cells. Also serves as period lights with `cells={3}`. */
export function DealBar({ value, cells = 12, tone = 'success', accessibilityLabel }: BarProps) {
  const lit = Math.max(0, Math.min(cells, Math.ceil(value * cells - 1e-6)));
  return (
    <View
      accessible
      accessibilityRole="progressbar"
      accessibilityLabel={accessibilityLabel}
      accessibilityValue={{ min: 0, max: cells, now: lit }}
      style={styles.bar}
    >
      {Array.from({ length: cells }, (_, i) => (
        <View
          key={i}
          style={[
            styles.cell,
            i < lit ? { backgroundColor: colors[tone] } : styles.off,
          ]}
        />
      ))}
    </View>
  );
}

const MS_HOUR = 3_600_000;

export function timeLeftMs(expiresAt: string | undefined, now = serverNow()): number | null {
  if (!expiresAt) return null;
  return new Date(expiresAt).getTime() - now;
}

/** "3h 28m left", "12m left", "2 days left". */
export function formatTimeLeft(ms: number): string {
  if (ms <= 0) return 'Expired';
  const hours = Math.floor(ms / MS_HOUR);
  const minutes = Math.floor((ms % MS_HOUR) / 60_000);
  if (hours >= 48) return `${Math.floor(hours / 24)} days left`;
  if (hours > 0) return `${hours}h ${minutes}m left`;
  return `${Math.max(1, minutes)}m left`;
}

/** Spoken form: "3 hours 28 minutes left". */
export function spokenTimeLeft(ms: number): string {
  if (ms <= 0) return 'Expired';
  const hours = Math.floor(ms / MS_HOUR);
  const minutes = Math.floor((ms % MS_HOUR) / 60_000);
  const parts: string[] = [];
  if (hours >= 48) return `${Math.floor(hours / 24)} days left`;
  if (hours > 0) parts.push(`${hours} ${hours === 1 ? 'hour' : 'hours'}`);
  parts.push(`${minutes} ${minutes === 1 ? 'minute' : 'minutes'}`);
  return `${parts.join(' ')} left`;
}

export function isUrgent(ms: number | null, hours = 6): boolean {
  return ms !== null && ms > 0 && ms <= hours * MS_HOUR;
}

interface ExpiryProps {
  triggeredAt: string;
  expiresAt?: string;
}

/** Expiry mode: cells = time left / total window. Green, then timer orange under 6 hours. */
export function ExpiryBar({ triggeredAt, expiresAt }: ExpiryProps) {
  const now = useNow(60_000);
  const left = timeLeftMs(expiresAt, now);
  if (left === null || !expiresAt) return null;
  const total = new Date(expiresAt).getTime() - new Date(triggeredAt).getTime();
  const value = total > 0 ? Math.max(0, left) / total : 0;
  return (
    <DealBar
      value={value}
      tone={isUrgent(left) ? 'warning' : 'success'}
      accessibilityLabel={spokenTimeLeft(left)}
    />
  );
}

const styles = StyleSheet.create({
  bar: {
    flexDirection: 'row',
    gap: 2,
    borderWidth: 2,
    borderColor: colors.border,
    backgroundColor: colors.background,
    padding: 2,
  },
  cell: { flex: 1, height: 10 },
  off: { backgroundColor: colors.background, borderWidth: 1, borderColor: colors.textMuted },
});
