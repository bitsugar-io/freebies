import React from 'react';
import { Linking, Pressable } from 'react-native';
import { Event } from '../api/client';
import { useAppConfig } from '../context/AppConfigContext';
import { PixelBox } from './pixel/PixelBox';
import { Text } from './pixel/Text';

/** Affiliate "team gear" link. Hidden unless the show_affiliate_links flag is on. */
export function SponsoredCard({ event }: { event: Event }) {
  const { config } = useAppConfig();
  const url = event.affiliateUrl;
  if (config.features.show_affiliate_links !== true || !url) return null;

  return (
    <Pressable
      onPress={() => Linking.openURL(url)}
      accessibilityRole="link"
      accessibilityLabel={`Sponsored: ${event.teamName} gear. ${event.affiliateTagline ?? ''}`}
      style={{ marginTop: 16 }}
    >
      <PixelBox variant="panel">
        <Text variant="caption" color="textMuted">
          Sponsored
        </Text>
        <Text variant="bodyStrong">{`${event.teamName} Gear`}</Text>
        {event.affiliateTagline ? (
          <Text variant="caption" color="textSecondary">
            {event.affiliateTagline}
          </Text>
        ) : null}
      </PixelBox>
    </Pressable>
  );
}
