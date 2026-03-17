import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Linking,
} from 'react-native';
import { Event } from '../api/client';
import { useTheme } from '../hooks/useTheme';
import { useAppConfig } from '../context/AppConfigContext';

interface SponsoredCardProps {
  event: Event;
}

export function SponsoredCard({ event }: SponsoredCardProps) {
  const { theme } = useTheme();
  const { colors } = theme;
  const { config } = useAppConfig();

  // Only render if affiliate links are enabled and event has affiliate data
  if (config.features.show_affiliate_links === false || !event.affiliateUrl) {
    return null;
  }

  const handlePress = () => {
    if (event.affiliateUrl) {
      Linking.openURL(event.affiliateUrl);
    }
  };

  return (
    <TouchableOpacity
      style={[styles.card, { backgroundColor: colors.surface, borderColor: colors.border }]}
      onPress={handlePress}
      activeOpacity={0.7}
    >
      <View style={styles.header}>
        <Text style={[styles.sponsoredLabel, { color: colors.textMuted }]}>
          SPONSORED
        </Text>
      </View>
      <View style={styles.content}>
        <Text style={[styles.icon]}>🛍️</Text>
        <View style={styles.textContainer}>
          <Text style={[styles.title, { color: colors.text }]}>
            {event.teamName} Gear
          </Text>
          {event.affiliateTagline && (
            <Text style={[styles.tagline, { color: colors.textMuted }]}>
              {event.affiliateTagline}
            </Text>
          )}
        </View>
        <Text style={[styles.arrow, { color: colors.textMuted }]}>→</Text>
      </View>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  card: {
    borderRadius: 12,
    padding: 12,
    marginBottom: 16,
    borderWidth: 1,
  },
  header: {
    marginBottom: 8,
  },
  sponsoredLabel: {
    fontSize: 10,
    fontWeight: '600',
    letterSpacing: 1,
  },
  content: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  icon: {
    fontSize: 24,
    marginRight: 12,
  },
  textContainer: {
    flex: 1,
  },
  title: {
    fontSize: 16,
    fontWeight: '600',
  },
  tagline: {
    fontSize: 13,
    marginTop: 2,
  },
  arrow: {
    fontSize: 18,
    marginLeft: 8,
  },
});
