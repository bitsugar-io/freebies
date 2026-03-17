import React, { useState, useMemo } from 'react';
import { View, Text, FlatList, TouchableOpacity, StyleSheet } from 'react-native';
import { BlockProps } from './BlockRenderer';
import { useAppData } from '../../context/AppDataContext';
import { useTheme } from '../../hooks/useTheme';
import { useAppConfig } from '../../context/AppConfigContext';
import { Event } from '../../api/client';

const DEFAULT_TEAM_COLOR = '#666666';

interface TeamGroup {
  teamId: string;
  teamName: string;
  league: string;
  color: string;
  events: Event[];
  isFollowing: boolean;
  followedCount: number;
}

interface EventListProps extends BlockProps {
  selectedLeague?: string;
  onEventPress?: (event: Event) => void;
}

export function EventListBlock({ config, selectedLeague = 'all', onEventPress }: EventListProps) {
  const { theme } = useTheme();
  const { colors } = theme;
  const { events, isSubscribed, toggleSubscription } = useAppData();
  const { config: appConfig } = useAppConfig();
  const [expandedTeam, setExpandedTeam] = useState<string | null>(null);

  const teamGroups = useMemo(() => {
    const leagueFlags: Record<string, string> = {
      MLB: 'enable_mlb', NBA: 'enable_nba', NFL: 'enable_nfl', NHL: 'enable_nhl',
    };
    const enabledEvents = events.filter(event => {
      const flagKey = leagueFlags[event.league];
      return !flagKey || appConfig.features[flagKey] !== false;
    });

    const groups: Record<string, TeamGroup> = {};
    enabledEvents.forEach(event => {
      if (!groups[event.teamId]) {
        groups[event.teamId] = {
          teamId: event.teamId, teamName: event.teamName,
          league: event.league.toLowerCase(), color: event.teamColor || DEFAULT_TEAM_COLOR,
          events: [], isFollowing: false, followedCount: 0,
        };
      }
      groups[event.teamId].events.push(event);
      if (isSubscribed(event.id)) groups[event.teamId].followedCount++;
    });
    Object.values(groups).forEach(g => {
      g.isFollowing = g.events.length > 0 && g.events.every(e => isSubscribed(e.id));
    });
    return Object.values(groups);
  }, [events, isSubscribed, appConfig.features]);

  const filteredTeams = useMemo(() => {
    if (selectedLeague === 'all') return teamGroups;
    return teamGroups.filter(t => t.league === selectedLeague);
  }, [teamGroups, selectedLeague]);

  const handleFollowTeam = async (team: TeamGroup) => {
    for (const event of team.events) {
      if (team.isFollowing) {
        if (isSubscribed(event.id)) await toggleSubscription(event.id);
      } else {
        if (!isSubscribed(event.id)) await toggleSubscription(event.id);
      }
    }
  };

  const renderTeamCard = ({ item: team }: { item: TeamGroup }) => {
    const isExpanded = expandedTeam === team.teamId;
    return (
      <View style={[styles.teamCard, { backgroundColor: colors.surface }]}>
        <TouchableOpacity style={styles.teamHeader} onPress={() => setExpandedTeam(isExpanded ? null : team.teamId)} activeOpacity={0.7}>
          <View style={[styles.teamBadge, { backgroundColor: team.color }]}>
            <Text style={styles.teamBadgeText}>{team.teamId}</Text>
          </View>
          <View style={styles.teamInfo}>
            <Text style={[styles.teamName, { color: colors.text }]}>{team.teamName}</Text>
            <Text style={[styles.teamOffers, { color: colors.textMuted }]}>
              {team.events.length} offer{team.events.length !== 1 ? 's' : ''} available
              {team.followedCount > 0 && ` • ${team.followedCount} following`}
            </Text>
          </View>
          <TouchableOpacity
            style={[styles.followButton, { backgroundColor: team.isFollowing ? colors.surfaceSecondary : colors.accent }]}
            onPress={() => handleFollowTeam(team)}
          >
            <Text style={[styles.followButtonText, { color: team.isFollowing ? colors.text : '#fff' }]}>
              {team.isFollowing ? 'Following' : 'Follow'}
            </Text>
          </TouchableOpacity>
        </TouchableOpacity>

        {isExpanded && (
          <View style={[styles.eventsList, { borderTopColor: colors.border }]}>
            {team.events.map(event => (
              <TouchableOpacity key={event.id} style={[styles.eventRow, { borderBottomColor: colors.border }]} onPress={() => onEventPress?.(event)}>
                <View style={styles.eventInfo}>
                  <Text style={[styles.eventName, { color: colors.text }]}>{event.offerName}</Text>
                  <Text style={[styles.eventPartner, { color: colors.textMuted }]}>
                    {event.partnerName} • {event.triggerCondition}
                  </Text>
                </View>
                <TouchableOpacity
                  style={[styles.subscribeButton, { backgroundColor: isSubscribed(event.id) ? colors.success : colors.surfaceSecondary }]}
                  onPress={() => toggleSubscription(event.id)}
                >
                  <Text style={[styles.subscribeButtonText, { color: isSubscribed(event.id) ? '#fff' : colors.textMuted }]}>
                    {isSubscribed(event.id) ? '✓' : '+'}
                  </Text>
                </TouchableOpacity>
              </TouchableOpacity>
            ))}
          </View>
        )}

        <TouchableOpacity style={styles.expandHint} onPress={() => setExpandedTeam(isExpanded ? null : team.teamId)}>
          <Text style={[styles.expandHintText, { color: colors.textMuted }]}>
            {isExpanded ? '▲ Collapse' : '▼ See offers'}
          </Text>
        </TouchableOpacity>
      </View>
    );
  };

  return (
    <FlatList
      scrollEnabled={false}
      data={filteredTeams}
      renderItem={renderTeamCard}
      keyExtractor={(item) => item.teamId}
      contentContainerStyle={styles.listContent}
      ListEmptyComponent={
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyIcon}>🏟️</Text>
          <Text style={[styles.emptyTitle, { color: colors.text }]}>No Teams Found</Text>
          <Text style={[styles.emptyText, { color: colors.textMuted }]}>No teams available in this league yet.</Text>
        </View>
      }
    />
  );
}

const styles = StyleSheet.create({
  listContent: { padding: 16, paddingBottom: 100 },
  teamCard: { borderRadius: 12, marginBottom: 12, overflow: 'hidden' },
  teamHeader: { flexDirection: 'row', alignItems: 'center', padding: 16 },
  teamBadge: { width: 48, height: 48, borderRadius: 24, alignItems: 'center', justifyContent: 'center' },
  teamBadgeText: { color: '#fff', fontWeight: 'bold', fontSize: 14 },
  teamInfo: { flex: 1, marginLeft: 12 },
  teamName: { fontSize: 18, fontWeight: '600' },
  teamOffers: { fontSize: 13, marginTop: 2 },
  followButton: { paddingHorizontal: 16, paddingVertical: 8, borderRadius: 16 },
  followButtonText: { fontSize: 14, fontWeight: '600' },
  eventsList: { borderTopWidth: 1 },
  eventRow: { flexDirection: 'row', alignItems: 'center', paddingHorizontal: 16, paddingVertical: 12, borderBottomWidth: StyleSheet.hairlineWidth },
  eventInfo: { flex: 1 },
  eventName: { fontSize: 15, fontWeight: '500' },
  eventPartner: { fontSize: 12, marginTop: 2 },
  subscribeButton: { width: 32, height: 32, borderRadius: 16, alignItems: 'center', justifyContent: 'center' },
  subscribeButtonText: { fontSize: 16, fontWeight: '600' },
  expandHint: { paddingVertical: 10, alignItems: 'center' },
  expandHintText: { fontSize: 12 },
  emptyContainer: { padding: 32, alignItems: 'center', marginTop: 40 },
  emptyIcon: { fontSize: 48, marginBottom: 16 },
  emptyTitle: { fontSize: 20, fontWeight: '600', marginBottom: 8 },
  emptyText: { fontSize: 16, textAlign: 'center', lineHeight: 22 },
});
