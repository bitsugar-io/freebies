import React from 'react';
import { RefreshControl, ScrollView, StyleSheet, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { colors } from '../hooks/useTheme';
import { HudBar } from './pixel/HudBar';

/** The one ScrollView method screens need. */
export interface Scroller {
  scrollTo: (o: { x?: number; y?: number; animated?: boolean }) => void;
}

interface Props {
  title: string;
  status?: string;
  left?: React.ReactNode;
  onRefresh?: () => Promise<void> | void;
  /** For screens that scroll themselves to a section (YOU > About). */
  scrollRef?: React.RefObject<Scroller | null>;
  children: React.ReactNode;
}

/** Tab screen shell: HUD header, one scroll view, 16pt gutters owned here, pull to refresh. */
export function Screen({ title, status, left, onRefresh, scrollRef, children }: Props) {
  const insets = useSafeAreaInsets();
  const [refreshing, setRefreshing] = React.useState(false);
  const refresh = async () => {
    if (!onRefresh) return;
    setRefreshing(true);
    try {
      await onRefresh();
    } finally {
      setRefreshing(false);
    }
  };
  return (
    <View style={styles.root}>
      <HudBar title={title} status={status} left={left} />
      <ScrollView
        ref={scrollRef as React.Ref<any>}
        contentContainerStyle={[styles.content, { paddingBottom: insets.bottom + 24 }]}
        refreshControl={
          onRefresh ? (
            <RefreshControl refreshing={refreshing} onRefresh={refresh} tintColor={colors.accent} colors={[colors.accent]} />
          ) : undefined
        }
      >
        {children}
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: colors.background },
  content: { padding: 16, flexGrow: 1 },
});
