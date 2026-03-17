import React, { useState } from 'react';
import { StyleSheet, View } from 'react-native';
import { useAppData } from '../context/AppDataContext';
import { useTheme } from '../hooks/useTheme';
import { BlockRenderer } from '../components/blocks/BlockRenderer';

export function ProfileScreen() {
  const { theme } = useTheme();
  const { colors } = theme;
  const { refreshAll } = useAppData();
  const [refreshing, setRefreshing] = useState(false);

  const onRefresh = async () => {
    setRefreshing(true);
    await refreshAll();
    setRefreshing(false);
  };

  return (
    <View style={[styles.container, { backgroundColor: colors.background }]}>
      <BlockRenderer
        screenId="profile"
        refreshing={refreshing}
        onRefresh={onRefresh}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1 },
});
