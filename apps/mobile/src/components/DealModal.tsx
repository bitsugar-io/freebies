import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  Modal,
  TouchableOpacity,
  ScrollView,
} from 'react-native';
import { Event, ActiveDeal } from '../api/client';
import { useTheme } from '../hooks/useTheme';
import { formatExpiresAt, isExpiringSoon } from '../hooks/useActiveDeals';
import { SponsoredCard } from './SponsoredCard';
import { SwipeButton } from './SwipeButton';

interface DealModalProps {
  event: Event | null;
  visible: boolean;
  onClose: () => void;
  isSubscribed: boolean;
  onToggleSubscription: () => void;
  activeDeal?: ActiveDeal | null;
  onDismiss?: (type: 'got_it' | 'stop_reminding') => void;
}

export function DealModal({
  event,
  visible,
  onClose,
  isSubscribed,
  onToggleSubscription,
  activeDeal,
  onDismiss,
}: DealModalProps) {
  const { theme } = useTheme();
  const { colors } = theme;

  if (!event) return null;

  const hasExpiration = activeDeal?.expiresAt;
  const expiringSoon = hasExpiration ? isExpiringSoon(activeDeal.expiresAt, 6) : false;
  const expirationText = hasExpiration ? formatExpiresAt(activeDeal.expiresAt) : null;

  return (
    <Modal
      visible={visible}
      animationType="slide"
      presentationStyle="pageSheet"
      onRequestClose={onClose}
    >
      <View style={[styles.container, { backgroundColor: colors.background }]}>
        {/* Header */}
        <View style={[styles.header, { borderBottomColor: colors.border }]}>
          <TouchableOpacity onPress={onClose} style={styles.closeButton}>
            <Text style={[styles.closeText, { color: colors.textMuted }]}>Close</Text>
          </TouchableOpacity>
          <Text style={[styles.headerTitle, { color: colors.text }]}>Deal Details</Text>
          <View style={styles.closeButton} />
        </View>

        <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
          {/* Expiration Badge */}
          {expirationText && (
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
                {expiringSoon ? '⏰ Expiring soon: ' : '🕐 '}
                {expirationText}
              </Text>
            </View>
          )}

          {/* League, Team & Partner */}
          <View style={styles.badges}>
            <View style={[styles.badge, { backgroundColor: colors.primary }]}>
              <Text style={[styles.badgeText, { color: colors.primaryText }]}>
                {event.league?.toUpperCase() || '?'} • {event.teamId}
              </Text>
            </View>
            <View style={[styles.badge, { backgroundColor: colors.surfaceSecondary }]}>
              <Text style={[styles.badgeText, { color: colors.text }]}>
                {event.partnerName}
              </Text>
            </View>
          </View>

          {/* Offer Name */}
          <View style={styles.offerRow}>
            {event.icon && (
              <Text style={styles.offerIcon}>{event.icon}</Text>
            )}
            <Text style={[styles.offerName, { color: colors.text }]}>
              {event.offerName}
            </Text>
          </View>

          <Text style={[styles.teamName, { color: colors.textMuted }]}>
            {event.teamName}
          </Text>

          {/* Trigger Condition */}
          <View style={[styles.triggerBox, { backgroundColor: colors.successBackground }]}>
            <Text style={[styles.triggerLabel, { color: colors.success }]}>
              Triggered when:
            </Text>
            <Text style={[styles.triggerCondition, { color: colors.success }]}>
              {event.triggerCondition}
            </Text>
          </View>

          {/* Description */}
          <Text style={[styles.description, { color: colors.textSecondary }]}>
            {event.offerDescription}
          </Text>

          {/* Region */}
          {event.regionName && (
            <View style={[styles.regionBox, { backgroundColor: colors.infoBackground }]}>
              <Text style={[styles.regionText, { color: colors.info }]}>
                📍 Available in {event.regionName}
              </Text>
            </View>
          )}

          {/* Dismiss Buttons - only show if there's an active deal */}
          {activeDeal && onDismiss && !activeDeal.isDismissed && (
            <View style={styles.dismissActions}>
              <TouchableOpacity
                style={[styles.dismissButton, { backgroundColor: colors.accent }]}
                onPress={() => {
                  onDismiss('got_it');
                  onClose();
                }}
              >
                <Text style={styles.dismissButtonText}>Got It!</Text>
              </TouchableOpacity>
              <View style={styles.swipeButtonContainer}>
                <SwipeButton
                  label="Stop Reminding"
                  onSwipeComplete={() => {
                    onDismiss('stop_reminding');
                    onClose();
                  }}
                  backgroundColor={colors.surfaceSecondary}
                  textColor={colors.textMuted}
                />
              </View>
            </View>
          )}

          {/* Already Dismissed */}
          {activeDeal?.isDismissed && (
            <View style={[styles.dismissedBox, { backgroundColor: colors.successBackground }]}>
              <Text style={[styles.dismissedText, { color: colors.success }]}>
                {activeDeal.dismissalType === 'got_it'
                  ? '✓ You claimed this deal!'
                  : '✓ Reminders stopped for this deal'}
              </Text>
            </View>
          )}

          {/* Subscribe Button - only show when not viewing an active deal */}
          {!activeDeal && (
            <TouchableOpacity
              style={[
                styles.subscribeButton,
                { backgroundColor: isSubscribed ? colors.surfaceSecondary : colors.accent },
              ]}
              onPress={onToggleSubscription}
            >
              <Text
                style={[
                  styles.subscribeButtonText,
                  { color: isSubscribed ? colors.text : '#fff' },
                ]}
              >
                {isSubscribed ? '✓ Subscribed - Tap to Unsubscribe' : 'Subscribe to this Deal'}
              </Text>
            </TouchableOpacity>
          )}

          {/* Status */}
          {!event.isActive && (
            <View style={[styles.inactiveBox, { backgroundColor: colors.warningBackground }]}>
              <Text style={[styles.inactiveText, { color: colors.warning }]}>
                ⏳ This deal is not currently active
              </Text>
            </View>
          )}

          {/* Sponsored Affiliate Link */}
          <View style={styles.sponsoredSection}>
            <SponsoredCard event={event} />
          </View>
        </ScrollView>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    paddingVertical: 16,
    borderBottomWidth: 1,
  },
  closeButton: {
    width: 60,
  },
  closeText: {
    fontSize: 16,
  },
  headerTitle: {
    fontSize: 17,
    fontWeight: '600',
  },
  content: {
    flex: 1,
    padding: 20,
  },
  expirationBadge: {
    alignSelf: 'flex-start',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
    marginBottom: 16,
  },
  expirationText: {
    fontSize: 14,
    fontWeight: '600',
  },
  badges: {
    flexDirection: 'row',
    gap: 8,
    marginBottom: 16,
  },
  badge: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
  },
  badgeText: {
    fontSize: 14,
    fontWeight: '600',
  },
  offerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 4,
  },
  offerIcon: {
    fontSize: 32,
    marginRight: 10,
  },
  offerName: {
    fontSize: 28,
    fontWeight: 'bold',
    flex: 1,
  },
  teamName: {
    fontSize: 16,
    marginBottom: 20,
  },
  triggerBox: {
    padding: 16,
    borderRadius: 12,
    marginBottom: 20,
  },
  triggerLabel: {
    fontSize: 12,
    fontWeight: '600',
    marginBottom: 4,
  },
  triggerCondition: {
    fontSize: 18,
    fontWeight: '600',
  },
  description: {
    fontSize: 16,
    lineHeight: 24,
    marginBottom: 20,
  },
  regionBox: {
    padding: 12,
    borderRadius: 8,
    marginBottom: 20,
  },
  regionText: {
    fontSize: 14,
  },
  subscribeButton: {
    padding: 16,
    borderRadius: 12,
    alignItems: 'center',
    marginBottom: 16,
  },
  subscribeButtonText: {
    fontSize: 16,
    fontWeight: '600',
  },
  inactiveBox: {
    padding: 12,
    borderRadius: 8,
    marginBottom: 20,
  },
  inactiveText: {
    fontSize: 14,
    textAlign: 'center',
  },
  dismissActions: {
    flexDirection: 'row',
    gap: 12,
    marginBottom: 16,
  },
  dismissButton: {
    flex: 1,
    paddingVertical: 14,
    borderRadius: 12,
    alignItems: 'center',
  },
  swipeButtonContainer: {
    flex: 1,
  },
  dismissButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  dismissButtonTextSecondary: {
    fontSize: 16,
    fontWeight: '600',
  },
  dismissedBox: {
    padding: 12,
    borderRadius: 8,
    marginBottom: 16,
  },
  dismissedText: {
    fontSize: 14,
    textAlign: 'center',
    fontWeight: '500',
  },
  sponsoredSection: {
    marginTop: 8,
    paddingTop: 16,
    borderTopWidth: StyleSheet.hairlineWidth,
    borderTopColor: 'rgba(0,0,0,0.1)',
  },
});
