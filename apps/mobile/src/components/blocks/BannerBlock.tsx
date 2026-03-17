import React, { useState } from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { BlockProps } from './BlockRenderer';
import { useTheme } from '../../hooks/useTheme';

export function BannerBlock({ config }: BlockProps) {
  const [dismissed, setDismissed] = useState(false);
  const { theme } = useTheme();
  const { colors } = theme;

  if (dismissed) return null;

  const text = config.text as string;
  if (!text) return null;

  const backgroundColor = (config.backgroundColor as string) ?? colors.card;
  const textColor = (config.textColor as string) ?? colors.text;
  const dismissible = config.dismissible !== false;

  return (
    <View style={[styles.container, { backgroundColor }]}>
      <Text style={[styles.text, { color: textColor }]}>{text}</Text>
      {dismissible && (
        <TouchableOpacity onPress={() => setDismissed(true)} style={styles.dismiss}>
          <Text style={{ color: textColor, opacity: 0.7 }}>✕</Text>
        </TouchableOpacity>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row', alignItems: 'center', padding: 16,
    marginHorizontal: 16, marginTop: 8, marginBottom: 8, borderRadius: 12,
  },
  text: { flex: 1, fontSize: 15, fontWeight: '600' },
  dismiss: { paddingLeft: 12 },
});
