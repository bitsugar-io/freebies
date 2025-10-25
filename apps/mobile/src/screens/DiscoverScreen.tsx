import React, { useState, useMemo } from 'react';
import {
  StyleSheet,
  Text,
  View,
  FlatList,
  RefreshControl,
  TouchableOpacity,
  ScrollView,
} from 'react-native';
import { useAppData } from '../context/AppDataContext';
import { useTheme } from '../hooks/useTheme';
import { DealModal } from '../components/DealModal';
import { Event, ActiveDeal } from '../api/client';

// Default color if none provided
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

export function DiscoverScreen() {
  const { theme } = useTheme();
  const { colors } = theme;
  const {
    events,
    leagues,
    refreshAll,
    isSubscribed,
    toggleSubscription,
    undismissedDeals,
  } = useAppData();

  // Build league options from backend data with "All" prepended
  const leagueOptions = useMemo(() => {
    const all = { id: 'all', name: 'All', icon: '🌟', displayOrder: 0 };
    return [all, ...leagues];
  }, [leagues]);

  const [selectedLeague, setSelectedLeague] = useState('all');
  const [expandedTeam, setExpandedTeam] = useState<string | null>(null);
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [selectedDeal, setSelectedDeal] = useState<ActiveDeal | null>(null);
  const [modalVisible, setModalVisible] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  // Group events by team
  const teamGroups = useMemo(() => {
    const groups: Record<string, TeamGroup> = {};

    events.forEach(event => {
      if (!groups[event.teamId]) {
        groups[event.teamId] = {
          teamId: event.teamId,
          teamName: event.teamName,
          league: event.league.toLowerCase(),
          color: event.teamColor || DEFAULT_TEAM_COLOR,
          events: [],
          isFollowing: false,
          followedCount: 0,
        };
      }
      groups[event.teamId].events.push(event);
      if (isSubscribed(event.id)) {
        groups[event.teamId].followedCount++;
      }
    });

    // Calculate isFollowing (following all events for team)
    Object.values(groups).forEach(group => {
      group.isFollowing = group.events.length > 0 &&
        group.events.every(e => isSubscribed(e.id));
    });

    return Object.values(groups);
  }, [events, isSubscribed]);

  // Filter by league
  const filteredTeams = useMemo(() => {
    if (selectedLeague === 'all') return teamGroups;
    return teamGroups.filter(team => team.league === selectedLeague);
  }, [teamGroups, selectedLeague]);

  // Get league counts
  const leagueCounts = useMemo(() => {
    const counts: Record<string, number> = { all: teamGroups.length };
    teamGroups.forEach(team => {
      counts[team.league] = (counts[team.league] || 0) + 1;
    });
    return counts;
  }, [teamGroups]);

  const onRefresh = async () => {
    setRefreshing(true);
    await refreshAll();
    setRefreshing(false);
  };

  const handleFollowTeam = async (team: TeamGroup) => {
    // Toggle all events for this team
    for (const event of team.events) {
      if (team.isFollowing) {
        // Unfollow: only unsubscribe if currently subscribed
        if (isSubscribed(event.id)) {
          await toggleSubscription(event.id);
        }
      } else {
        // Follow: only subscribe if not already subscribed
        if (!isSubscribed(event.id)) {
          await toggleSubscription(event.id);
        }
      }
    }
  };

  const handleEventPress = (event: Event) => {
    const deal = undismissedDeals.find(d => d.eventId === event.id);
    setSelectedEvent(event);
    setSelectedDeal(deal || null);
    setModalVisible(true);
  };

  const renderTeamCard = ({ item: team }: { item: TeamGroup }) => {
    const isExpanded = expandedTeam === team.teamId;

    return (
      <View style={[styles.teamCard, { backgroundColor: colors.surface }]}>
        {/* Team Header */}
        <TouchableOpacity
          style={styles.teamHeader}
          onPress={() => setExpandedTeam(isExpanded ? null : team.teamId)}
          activeOpacity={0.7}
        >
          <View style={[styles.teamBadge, { backgroundColor: team.color }]}>
            <Text style={styles.teamBadgeText}>{team.teamId}</Text>
          </View>
          <View style={styles.teamInfo}>
            <Text style={[styles.teamName, { color: colors.text }]}>
              {team.teamName}
            </Text>
            <Text style={[styles.teamOffers, { color: colors.textMuted }]}>
              {team.events.length} offer{team.events.length !== 1 ? 's' : ''} available
              {team.followedCount > 0 && ` • ${team.followedCount} following`}
            </Text>
          </View>
          <TouchableOpacity
            style={[
              styles.followButton,
              { backgroundColor: team.isFollowing ? colors.surfaceSecondary : colors.accent },
            ]}
            onPress={() => handleFollowTeam(team)}
          >
            <Text style={[
              styles.followButtonText,
              { color: team.isFollowing ? colors.text : '#fff' },
            ]}>
              {team.isFollowing ? 'Following' : 'Follow'}
            </Text>
          </TouchableOpacity>
        </TouchableOpacity>

        {/* Expanded Events List */}
        {isExpanded && (
          <View style={[styles.eventsList, { borderTopColor: colors.border }]}>
            {team.events.map(event => (
              <TouchableOpacity
                key={event.id}
                style={[styles.eventRow, { borderBottomColor: colors.border }]}
                onPress={() => handleEventPress(event)}
              >
                <View style={styles.eventInfo}>
                  <Text style={[styles.eventName, { color: colors.text }]}>
                    {event.offerName}
                  </Text>
                  <Text style={[styles.eventPartner, { color: colors.textMuted }]}>
                    {event.partnerName} • {event.triggerCondition}
                  </Text>
                </View>
                <TouchableOpacity
                  style={[
                    styles.subscribeButton,
                    { backgroundColor: isSubscribed(event.id) ? colors.success : colors.surfaceSecondary },
                  ]}
                  onPress={() => toggleSubscription(event.id)}
                >
                  <Text style={[
                    styles.subscribeButtonText,
                    { color: isSubscribed(event.id) ? '#fff' : colors.textMuted },
                  ]}>
                    {isSubscribed(event.id) ? '✓' : '+'}
                  </Text>
                </TouchableOpacity>
              </TouchableOpacity>
            ))}
          </View>
        )}

        {/* Expand hint */}
        <TouchableOpacity
          style={styles.expandHint}
          onPress={() => setExpandedTeam(isExpanded ? null : team.teamId)}
        >
          <Text style={[styles.expandHintText, { color: colors.textMuted }]}>
            {isExpanded ? '▲ Collapse' : '▼ See offers'}
          </Text>
        </TouchableOpacity>
      </View>
    );
  };

  return (
    <View style={[styles.container, { backgroundColor: colors.background }]}>
      {/* League Filter Pills */}
      <View style={[styles.leagueContainer, { backgroundColor: colors.surface }]}>
        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          contentContainerStyle={styles.leagueScroll}
        >
          {leagueOptions.map(league => {
            const count = leagueCounts[league.id] || 0;
            const isActive = selectedLeague === league.id;

            // Hide leagues with no teams (except 'all')
            if (league.id !== 'all' && count === 0) {
              return null;
            }

            return (
              <TouchableOpacity
                key={league.id}
                style={[
                  styles.leaguePill,
                  { backgroundColor: isActive ? colors.accent : colors.surfaceSecondary },
                ]}
                onPress={() => setSelectedLeague(league.id)}
              >
                <Text style={styles.leagueIcon}>{league.icon}</Text>
                <Text style={[
                  styles.leagueLabel,
                  { color: isActive ? '#fff' : colors.text },
                ]}>
                  {league.name}
                </Text>
              </TouchableOpacity>
            );
          })}
        </ScrollView>
      </View>

      {/* Teams List */}
      <FlatList
        style={styles.list}
        data={filteredTeams}
        renderItem={renderTeamCard}
        keyExtractor={(item) => item.teamId}
        contentContainerStyle={styles.listContent}
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            tintColor={colors.accent}
          />
        }
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyIcon}>🏟️</Text>
            <Text style={[styles.emptyTitle, { color: colors.text }]}>
              No Teams Found
            </Text>
            <Text style={[styles.emptyText, { color: colors.textMuted }]}>
              No teams available in this league yet.
            </Text>
          </View>
        }
      />

      <DealModal
        event={selectedEvent}
        visible={modalVisible}
        onClose={() => {
          setModalVisible(false);
          setSelectedDeal(null);
        }}
        isSubscribed={selectedEvent ? isSubscribed(selectedEvent.id) : false}
        onToggleSubscription={() => {
          if (selectedEvent) {
            toggleSubscription(selectedEvent.id);
          }
        }}
        activeDeal={selectedDeal}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  leagueContainer: {
    paddingVertical: 12,
  },
  leagueScroll: {
    paddingHorizontal: 12,
  },
  leaguePill: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 20,
    marginRight: 8,
  },
  leagueIcon: {
    fontSize: 16,
    marginRight: 6,
  },
  leagueLabel: {
    fontSize: 14,
    fontWeight: '600',
  },
  list: {
    flex: 1,
  },
  listContent: {
    padding: 16,
    paddingBottom: 100,
  },
  teamCard: {
    borderRadius: 12,
    marginBottom: 12,
    overflow: 'hidden',
  },
  teamHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
  },
  teamBadge: {
    width: 48,
    height: 48,
    borderRadius: 24,
    alignItems: 'center',
    justifyContent: 'center',
  },
  teamBadgeText: {
    color: '#fff',
    fontWeight: 'bold',
    fontSize: 14,
  },
  teamInfo: {
    flex: 1,
    marginLeft: 12,
  },
  teamName: {
    fontSize: 18,
    fontWeight: '600',
  },
  teamOffers: {
    fontSize: 13,
    marginTop: 2,
  },
  followButton: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 16,
  },
  followButtonText: {
    fontSize: 14,
    fontWeight: '600',
  },
  eventsList: {
    borderTopWidth: 1,
  },
  eventRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: StyleSheet.hairlineWidth,
  },
  eventInfo: {
    flex: 1,
  },
  eventName: {
    fontSize: 15,
    fontWeight: '500',
  },
  eventPartner: {
    fontSize: 12,
    marginTop: 2,
  },
  subscribeButton: {
    width: 32,
    height: 32,
    borderRadius: 16,
    alignItems: 'center',
    justifyContent: 'center',
  },
  subscribeButtonText: {
    fontSize: 16,
    fontWeight: '600',
  },
  expandHint: {
    paddingVertical: 10,
    alignItems: 'center',
  },
  expandHintText: {
    fontSize: 12,
  },
  emptyContainer: {
    padding: 32,
    alignItems: 'center',
    marginTop: 40,
  },
  emptyIcon: {
    fontSize: 48,
    marginBottom: 16,
  },
  emptyTitle: {
    fontSize: 20,
    fontWeight: '600',
    marginBottom: 8,
  },
  emptyText: {
    fontSize: 16,
    textAlign: 'center',
    lineHeight: 22,
  },
});
