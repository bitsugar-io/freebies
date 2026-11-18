import React, { useEffect, useRef, useState } from 'react';
import { Image, Pressable, StyleSheet, useWindowDimensions } from 'react-native';
import { colors } from '../../hooks/useTheme';

// Launch animation (owner's pick 2026-11-17, 'Spinning Rays'). iOS shows assets/launch.png as a
// still; this takes over from it with the same picture and turns the rays behind the icon for
// 1.5s, then the app shows. Art comes from `icons.py launch`, so it matches the still exactly.

/** The still is a 440x956pt picture drawn with aspect fill; the overlay copies that scale. */
const STILL_W = 440;
const STILL_H = 956;
const RAYS_PT = 1056;
const ICON_PT = 240;
/** Stepped like the rest of the app's motion: a hold of 4 frames (1/15s) per step. */
const STEP_MS = (4 * 1000) / 60;
const DEG_PER_STEP = 2;
const DURATION_MS = 1500;

const RAYS = require('../../../assets/launch-rays.png');
const ICON = require('../../../assets/launch-icon.png');

interface Props {
  /** Called once both images have drawn: hide the native still now, so there's no gap. */
  onReady: () => void;
  /** Called when the animation ends or is tapped away. */
  onDone: () => void;
}

export function LaunchRays({ onReady, onDone }: Props) {
  const { width, height } = useWindowDimensions();
  const scale = Math.max(width / STILL_W, height / STILL_H);
  const [step, setStep] = useState(0);
  const loaded = useRef(0);
  const done = useRef(false);

  const finish = () => {
    if (done.current) return;
    done.current = true;
    onDone();
  };

  const onLoad = () => {
    loaded.current += 1;
    if (loaded.current === 2) onReady();
  };

  useEffect(() => {
    // If an image never reports loading, don't hold the still or the app.
    const giveUp = setTimeout(onReady, 1000);
    return () => clearTimeout(giveUp);
  }, [onReady]);

  useEffect(() => {
    const tick = setInterval(() => setStep((s) => s + 1), STEP_MS);
    const end = setTimeout(finish, DURATION_MS);
    return () => {
      clearInterval(tick);
      clearTimeout(end);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const rays = RAYS_PT * scale;
  const icon = ICON_PT * scale;
  return (
    <Pressable
      style={styles.fill}
      onPress={finish}
      accessibilityRole='image'
      accessibilityLabel='Freebies: LA Sports Deals'
    >
      <Image
        source={RAYS}
        onLoad={onLoad}
        fadeDuration={0}
        style={{
          position: 'absolute',
          width: rays,
          height: rays,
          left: (width - rays) / 2,
          top: (height - rays) / 2,
          transform: [{ rotate: `${step * DEG_PER_STEP}deg` }],
        }}
      />
      <Image
        source={ICON}
        onLoad={onLoad}
        fadeDuration={0}
        style={{
          position: 'absolute',
          width: icon,
          height: icon,
          left: (width - icon) / 2,
          top: (height - icon) / 2,
        }}
      />
    </Pressable>
  );
}

const styles = StyleSheet.create({
  fill: {
    ...StyleSheet.absoluteFill,
    backgroundColor: colors.background,
    overflow: 'hidden',
    zIndex: 100,
  },
});
