import React from 'react';
import { Text, TouchableOpacity, StyleSheet, Linking } from 'react-native';
import { BlockProps } from './BlockRenderer';

export function PromoCardBlock({ config }: BlockProps) {
  const title = config.title as string;
  const subtitle = config.subtitle as string;
  const url = config.url as string;
  const backgroundColor = (config.backgroundColor as string) ?? '#1a1a1a';
  const textColor = (config.textColor as string) ?? '#FFFFFF';

  if (!title || !url) return null;

  return (
    <TouchableOpacity style={[styles.container, { backgroundColor }]} onPress={() => Linking.openURL(url)}>
      <Text style={[styles.title, { color: textColor }]}>{title}</Text>
      {subtitle && <Text style={[styles.subtitle, { color: textColor, opacity: 0.7 }]}>{subtitle}</Text>}
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  container: { padding: 20, marginHorizontal: 16, marginTop: 8, marginBottom: 8, borderRadius: 12 },
  title: { fontSize: 17, fontWeight: '700' },
  subtitle: { fontSize: 14, marginTop: 4 },
});
