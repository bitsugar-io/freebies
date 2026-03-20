import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
} from 'react-native';
import { ActiveDeal } from '../api/client';
import { useTheme } from '../hooks/useTheme';
import { formatExpiresAt, isExpiringSoon } from '../hooks/useActiveDeals';
interface ActiveDealCardProps {
  deal: ActiveDeal;
  onPress: () => void;
}

export function ActiveDealCard({ deal, onPress }: ActiveDealCardProps) {
  const { theme } = useTheme();
  const { colors } = theme;

  const expiringSoon = isExpiringSoon(deal.expiresAt, 6);
  const expirationText = formatExpiresAt(deal.expiresAt);

  return (
    <TouchableOpacity
      style={[styles.card, { backgroundColor: colors.surface, borderColor: colors.border }]}
      onPress={onPress}
      activeOpacity={0.7}
    >
      {/* Expiration Badge */}
      <View
        style={[
          styles.expirationBadge,
          {
            backgroundColor: expiringSoon ? colors.warningBackground : colors.infoBackground,
          },
        ]}
      >
        <Text
          style={[
            styles.expirationText,
            { color: expiringSoon ? colors.warning : colors.info },
          ]}
        >
          {expiringSoon ? '⏰ ' : '🕐 '}
          {expirationText}
        </Text>
      </View>

      {/* League, Team & Partner Row */}
      <View style={styles.tagsRow}>
        <View style={[styles.tag, { backgroundColor: colors.primary }]}>
          <Text style={[styles.tagText, { color: colors.primaryText }]}>
            {deal.event.league?.toUpperCase() || '?'} • {deal.event.teamId}
          </Text>
        </View>
        <View style={[styles.tag, { backgroundColor: colors.surfaceSecondary }]}>
          <Text style={[styles.tagText, { color: colors.text }]}>
            {deal.event.partnerName}
          </Text>
        </View>
      </View>

      {/* Offer Name */}
      <View style={styles.offerRow}>
        {deal.event.icon && (
          <Text style={styles.offerIcon}>{deal.event.icon}</Text>
        )}
        <Text style={[styles.offerName, { color: colors.text }]}>
          {deal.event.offerName}
        </Text>
      </View>

      {/* Team Name */}
      <Text style={[styles.teamName, { color: colors.textMuted }]}>
        {deal.event.teamName}
      </Text>

      {/* Trigger Condition */}
      <View style={[styles.triggerBox, { backgroundColor: colors.successBackground }]}>
        <Text style={[styles.triggerText, { color: colors.success }]}>
          {deal.event.triggerCondition}
        </Text>
      </View>

      {/* Redeem Button */}
      <TouchableOpacity
        style={[styles.redeemButton, { backgroundColor: colors.accent }]}
        onPress={onPress}
      >
        <Text style={styles.redeemButtonText}>Redeem</Text>
      </TouchableOpacity>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  card: {
    borderRadius: 16,
    padding: 16,
    marginBottom: 12,
    borderWidth: 1,
  },
  expirationBadge: {
    alignSelf: 'flex-start',
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
    marginBottom: 12,
  },
  expirationText: {
    fontSize: 12,
    fontWeight: '600',
  },
  tagsRow: {
    flexDirection: 'row',
    gap: 8,
    marginBottom: 8,
  },
  tag: {
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  tagText: {
    fontSize: 12,
    fontWeight: '600',
  },
  offerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 4,
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
  teamName: {
    fontSize: 14,
    marginBottom: 12,
  },
  triggerBox: {
    padding: 12,
    borderRadius: 8,
    marginBottom: 12,
  },
  triggerText: {
    fontSize: 14,
    fontWeight: '500',
  },
  redeemButton: {
    height: 56,
    borderRadius: 28,
    alignItems: 'center',
    justifyContent: 'center',
  },
  redeemButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
});
