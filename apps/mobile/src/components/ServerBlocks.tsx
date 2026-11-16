import React, { useState } from 'react';
import { Linking, Pressable, StyleSheet, View } from 'react-native';
import { useAppConfig } from '../context/AppConfigContext';
import { ColorToken } from '../hooks/useTheme';
import { PixelBox, BoxVariant } from './pixel/PixelBox';
import { PixelIcon } from './pixel/PixelIcon';
import { Text } from './pixel/Text';

/** Server blocks pick a palette tone by name. Unknown tones fall back to panel; hex is ignored. */
const TONES: Record<string, { variant: BoxVariant; text: ColorToken }> = {
  coin: { variant: 'selected', text: 'accent' },
  clear: { variant: 'panel', text: 'success' },
  sky: { variant: 'panel', text: 'info' },
  timer: { variant: 'panel', text: 'warning' },
  hit: { variant: 'alert', text: 'danger' },
  panel: { variant: 'panel', text: 'text' },
};

function toneOf(config: Record<string, any>) {
  return TONES[String(config.tone ?? 'panel')] ?? TONES.panel;
}

function Banner({ config }: { config: Record<string, any> }) {
  const [dismissed, setDismissed] = useState(false);
  const text = config.text as string | undefined;
  if (dismissed || !text) return null;
  const tone = toneOf(config);
  return (
    <PixelBox variant={tone.variant} padding={4} style={styles.block}>
      <View style={styles.bannerRow}>
        <View style={{ flex: 1, paddingLeft: 8 }}>
          <Text variant="bodyStrong" color={tone.text}>
            {text}
          </Text>
        </View>
        {config.dismissible !== false ? (
          <Pressable
            onPress={() => setDismissed(true)}
            accessibilityRole="button"
            accessibilityLabel="Dismiss message"
            style={styles.dismiss}
          >
            <PixelIcon name="close" scale={2} color="textSecondary" />
          </Pressable>
        ) : null}
      </View>
    </PixelBox>
  );
}

function Promo({ config }: { config: Record<string, any> }) {
  const title = config.title as string | undefined;
  const url = config.url as string | undefined;
  if (!title || !url || !/^https?:\/\//i.test(url)) return null;
  const tone = toneOf(config);
  const subtitle = config.subtitle as string | undefined;
  return (
    <Pressable
      onPress={() => Linking.openURL(url)}
      accessibilityRole="link"
      accessibilityLabel={subtitle ? `${title}. ${subtitle}` : title}
      style={styles.block}
    >
      <PixelBox variant={tone.variant}>
        <Text variant="bodyStrong" color={tone.text}>
          {title}
        </Text>
        {subtitle ? (
          <Text variant="caption" color="textSecondary">
            {subtitle}
          </Text>
        ) : null}
      </PixelBox>
    </Pressable>
  );
}

/**
 * Renders the server-driven `banner` and `promo_card` blocks for a screen. Every other block type
 * is served only to older app builds and is ignored here. `keys` lists the /config screen keys to
 * read (the new key first, then the pre-redesign one).
 */
export function ServerBlocks({ keys }: { keys: string[] }) {
  const { config } = useAppConfig();
  const blocks = keys.flatMap((k) => config.screens[k] ?? []);
  return (
    <>
      {blocks.map((b) => {
        if (b.type === 'banner') return <Banner key={`${b.type}-${b.key}`} config={b.config} />;
        if (b.type === 'promo_card') return <Promo key={`${b.type}-${b.key}`} config={b.config} />;
        return null;
      })}
    </>
  );
}

const styles = StyleSheet.create({
  block: { marginBottom: 12 },
  bannerRow: { flexDirection: 'row', alignItems: 'center' },
  dismiss: { width: 44, height: 44, alignItems: 'center', justifyContent: 'center' },
});
