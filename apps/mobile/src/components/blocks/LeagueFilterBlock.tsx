import React, { useMemo } from 'react';
import { View, Text, ScrollView, TouchableOpacity, StyleSheet } from 'react-native';
import { BlockProps } from './BlockRenderer';
import { useAppData } from '../../context/AppDataContext';
import { useTheme } from '../../hooks/useTheme';
import { useAppConfig } from '../../context/AppConfigContext';

export function LeagueFilterBlock({ config, screenProps }: BlockProps) {
  const { theme } = useTheme();
  const { colors } = theme;
  const { leagues } = useAppData();
  const { config: appConfig } = useAppConfig();

  const selectedLeague = screenProps?.selectedLeague ?? 'all';
  const onSelectLeague = screenProps?.onSelectLeague;

  const leagueOptions = useMemo(() => {
    const all = { id: 'all', name: 'All', icon: '🌟', displayOrder: 0 };
    const enabledLeagues = leagues.filter(league => {
      const flagKey = `enable_${league.name.toLowerCase()}`;
      return appConfig.features[flagKey] !== false;
    });
    return [all, ...enabledLeagues];
  }, [leagues, appConfig.features]);

  return (
    <View style={[styles.container, { backgroundColor: colors.surface }]}>
      <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.scroll}>
        {leagueOptions.map(league => {
          const isActive = selectedLeague === league.id;
          return (
            <TouchableOpacity
              key={league.id}
              style={[styles.pill, { backgroundColor: isActive ? colors.accent : colors.surfaceSecondary }]}
              onPress={() => onSelectLeague?.(league.id)}
            >
              <Text style={styles.icon}>{league.icon}</Text>
              <Text style={[styles.label, { color: isActive ? '#fff' : colors.text }]}>{league.name}</Text>
            </TouchableOpacity>
          );
        })}
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { paddingVertical: 12 },
  scroll: { paddingHorizontal: 12 },
  pill: { flexDirection: 'row', alignItems: 'center', paddingHorizontal: 16, paddingVertical: 10, borderRadius: 20, marginRight: 8 },
  icon: { fontSize: 16, marginRight: 6 },
  label: { fontSize: 14, fontWeight: '600' },
});
