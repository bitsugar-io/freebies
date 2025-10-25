import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Switch,
} from 'react-native';
import { Event } from '../api/client';
import { useTheme } from '../hooks/useTheme';

interface EventCardProps {
  event: Event;
  isSubscribed: boolean;
  onToggleSubscription: () => void;
}

export function EventCard({
  event,
  isSubscribed,
  onToggleSubscription,
}: EventCardProps) {
  const { theme } = useTheme();
  const { colors } = theme;

  return (
    <View
      style={[
        styles.card,
        { backgroundColor: colors.surface },
        !event.isActive && styles.cardInactive,
      ]}
    >
      <View style={styles.header}>
        <View style={[styles.teamBadge, { backgroundColor: colors.primary }]}>
          <Text style={[styles.teamId, { color: colors.primaryText }]}>
            {event.league?.toUpperCase() || '?'} • {event.teamId}
          </Text>
        </View>
        <View style={styles.headerText}>
          <Text style={[styles.partnerName, { color: colors.textMuted }]}>
            {event.partnerName}
          </Text>
          <Text style={[styles.teamName, { color: colors.textSecondary }]}>
            {event.teamName}
          </Text>
        </View>
        {event.isActive && (
          <Switch
            value={isSubscribed}
            onValueChange={onToggleSubscription}
            trackColor={{ false: colors.surfaceSecondary, true: colors.success }}
            thumbColor={isSubscribed ? '#fff' : '#f4f3f4'}
          />
        )}
      </View>

      <View style={styles.offerRow}>
        {event.icon && (
          <Text style={styles.offerIcon}>{event.icon}</Text>
        )}
        <Text style={[styles.offerName, { color: colors.text }]}>
          {event.offerName}
        </Text>
      </View>

      <View
        style={[
          styles.triggerContainer,
          { backgroundColor: colors.surfaceSecondary },
        ]}
      >
        <Text style={[styles.triggerLabel, { color: colors.textMuted }]}>
          When:
        </Text>
        <Text style={[styles.triggerCondition, { color: colors.accent }]}>
          {event.triggerCondition}
        </Text>
      </View>

      <Text style={[styles.description, { color: colors.textMuted }]}>
        {event.offerDescription}
      </Text>

      {event.regionName && (
        <View
          style={[
            styles.regionBadge,
            { backgroundColor: colors.infoBackground },
          ]}
        >
          <Text style={[styles.regionText, { color: colors.info }]}>
            {event.regionName}
          </Text>
        </View>
      )}

      {!event.isActive && (
        <View
          style={[
            styles.inactiveBadge,
            { backgroundColor: colors.warningBackground },
          ]}
        >
          <Text style={[styles.inactiveText, { color: colors.warning }]}>
            Coming Soon
          </Text>
        </View>
      )}

      {isSubscribed && event.isActive && (
        <TouchableOpacity
          style={[
            styles.subscribedBadge,
            { backgroundColor: colors.successBackground },
          ]}
        >
          <Text style={[styles.subscribedText, { color: colors.success }]}>
            Subscribed
          </Text>
        </TouchableOpacity>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    borderRadius: 12,
    padding: 16,
    marginBottom: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  cardInactive: {
    opacity: 0.6,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 12,
  },
  teamBadge: {
    width: 48,
    height: 48,
    borderRadius: 24,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 12,
  },
  teamId: {
    fontWeight: 'bold',
    fontSize: 14,
  },
  headerText: {
    flex: 1,
  },
  partnerName: {
    fontSize: 12,
    textTransform: 'uppercase',
    letterSpacing: 1,
  },
  teamName: {
    fontSize: 14,
    fontWeight: '500',
  },
  offerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  offerIcon: {
    fontSize: 24,
    marginRight: 8,
  },
  offerName: {
    fontSize: 20,
    fontWeight: 'bold',
    flex: 1,
  },
  triggerContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
    padding: 8,
    borderRadius: 8,
  },
  triggerLabel: {
    fontSize: 12,
    marginRight: 4,
  },
  triggerCondition: {
    fontSize: 14,
    fontWeight: '600',
  },
  description: {
    fontSize: 14,
    lineHeight: 20,
  },
  regionBadge: {
    marginTop: 12,
    alignSelf: 'flex-start',
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  regionText: {
    fontSize: 12,
  },
  inactiveBadge: {
    marginTop: 12,
    alignSelf: 'flex-start',
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  inactiveText: {
    fontSize: 12,
    fontWeight: '600',
  },
  subscribedBadge: {
    marginTop: 12,
    alignSelf: 'flex-start',
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  subscribedText: {
    fontSize: 12,
    fontWeight: '600',
  },
});
