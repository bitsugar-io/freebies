import React, { useState } from 'react';
import { StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { useTheme } from '../hooks/useTheme';

interface ErrorCardProps {
  message: string;
  onRetry?: () => void;
}

export function ErrorCard({ message, onRetry }: ErrorCardProps) {
  const { theme } = useTheme();
  const { colors } = theme;
  const [expanded, setExpanded] = useState(false);

  return (
    <View style={styles.wrapper}>
      <TouchableOpacity
        style={[styles.card, { backgroundColor: colors.surface, borderColor: colors.border }]}
        onPress={() => setExpanded(!expanded)}
        activeOpacity={0.7}
      >
        <View style={styles.header}>
          <Text style={styles.icon}>🫠</Text>
          <View style={styles.headerText}>
            <Text style={[styles.title, { color: colors.text }]}>
              Something went wrong
            </Text>
            <Text style={[styles.subtitle, { color: colors.textMuted }]}>
              Tap to see details
            </Text>
          </View>
          <Text style={[styles.chevron, { color: colors.textMuted }]}>
            {expanded ? '▲' : '▼'}
          </Text>
        </View>

        {expanded && (
          <View style={[styles.details, { borderTopColor: colors.border }]}>
            <Text style={[styles.errorMessage, { color: colors.textMuted }]}>
              {message}
            </Text>
          </View>
        )}
      </TouchableOpacity>

      {onRetry && (
        <TouchableOpacity
          style={[styles.retryButton, { backgroundColor: colors.accent }]}
          onPress={onRetry}
          activeOpacity={0.7}
        >
          <Text style={styles.retryText}>Try Again</Text>
        </TouchableOpacity>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: {
    padding: 16,
    alignItems: 'center',
    marginTop: 40,
  },
  card: {
    width: '100%',
    borderRadius: 12,
    borderWidth: 1,
    overflow: 'hidden',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
  },
  icon: {
    fontSize: 32,
    marginRight: 12,
  },
  headerText: {
    flex: 1,
  },
  title: {
    fontSize: 16,
    fontWeight: '600',
  },
  subtitle: {
    fontSize: 13,
    marginTop: 2,
  },
  chevron: {
    fontSize: 12,
    marginLeft: 8,
  },
  details: {
    padding: 16,
    paddingTop: 12,
    borderTopWidth: 1,
  },
  errorMessage: {
    fontSize: 13,
    fontFamily: 'monospace',
    lineHeight: 18,
  },
  retryButton: {
    marginTop: 12,
    paddingHorizontal: 24,
    paddingVertical: 10,
    borderRadius: 8,
  },
  retryText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
});
