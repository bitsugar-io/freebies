import React, { useRef } from 'react';
import {
  View,
  Text,
  StyleSheet,
  Animated,
  PanResponder,
  Dimensions,
} from 'react-native';
import { useTheme } from '../hooks/useTheme';

interface SwipeButtonProps {
  onSwipeComplete: () => void;
  label: string;
  backgroundColor?: string;
  textColor?: string;
}

const SWIPE_THRESHOLD = 0.7; // 70% of button width to trigger

export function SwipeButton({
  onSwipeComplete,
  label,
  backgroundColor,
  textColor,
}: SwipeButtonProps) {
  const { theme } = useTheme();
  const { colors } = theme;

  const containerWidth = useRef(0);
  const translateX = useRef(new Animated.Value(0)).current;

  const panResponder = useRef(
    PanResponder.create({
      onStartShouldSetPanResponder: () => true,
      onMoveShouldSetPanResponder: () => true,
      onPanResponderMove: (_, gestureState) => {
        // Only allow moving right
        if (gestureState.dx > 0) {
          const maxTranslate = containerWidth.current - 56; // thumb width
          const clampedValue = Math.min(gestureState.dx, maxTranslate);
          translateX.setValue(clampedValue);
        }
      },
      onPanResponderRelease: (_, gestureState) => {
        const maxTranslate = containerWidth.current - 56;
        const threshold = maxTranslate * SWIPE_THRESHOLD;

        if (gestureState.dx >= threshold) {
          // Complete the swipe
          Animated.timing(translateX, {
            toValue: maxTranslate,
            duration: 100,
            useNativeDriver: true,
          }).start(() => {
            onSwipeComplete();
            // Reset after callback
            setTimeout(() => {
              translateX.setValue(0);
            }, 300);
          });
        } else {
          // Reset to start
          Animated.spring(translateX, {
            toValue: 0,
            useNativeDriver: true,
            tension: 50,
            friction: 8,
          }).start();
        }
      },
    })
  ).current;

  const bgColor = backgroundColor || colors.accent;
  const txtColor = textColor || '#fff';

  // Calculate opacity for hint text based on swipe progress
  const hintOpacity = translateX.interpolate({
    inputRange: [0, 100],
    outputRange: [1, 0],
    extrapolate: 'clamp',
  });

  return (
    <View
      style={[styles.container, { backgroundColor: bgColor }]}
      onLayout={(e) => {
        containerWidth.current = e.nativeEvent.layout.width;
      }}
    >
      <Animated.View style={{ opacity: hintOpacity }}>
        <Text style={[styles.hintText, { color: txtColor }]}>
          Slide to {label}
        </Text>
      </Animated.View>

      <Animated.View
        style={[
          styles.thumb,
          {
            backgroundColor: 'rgba(255,255,255,0.3)',
            transform: [{ translateX }],
          },
        ]}
        {...panResponder.panHandlers}
      >
        <Text style={[styles.thumbText, { color: txtColor }]}>→</Text>
      </Animated.View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    height: 56,
    borderRadius: 28,
    justifyContent: 'center',
    alignItems: 'center',
    overflow: 'hidden',
  },
  hintText: {
    fontSize: 16,
    fontWeight: '600',
  },
  thumb: {
    position: 'absolute',
    left: 4,
    width: 48,
    height: 48,
    borderRadius: 24,
    justifyContent: 'center',
    alignItems: 'center',
  },
  thumbText: {
    fontSize: 20,
    fontWeight: 'bold',
  },
});
