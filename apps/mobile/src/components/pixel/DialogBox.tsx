import React from 'react';
import { StyleProp, View, ViewStyle } from 'react-native';
import { useBlinkClock } from '../../hooks/useBlink';
import { PixelBox, BoxVariant } from './PixelBox';
import { PixelIcon } from './PixelIcon';
import { Text } from './Text';

interface Props {
  /** Pass children for rich text, or `text` for a plain string. */
  children?: React.ReactNode;
  text?: string;
  variant?: BoxVariant;
  /** Show the bobbing ▼ (more text follows). */
  more?: boolean;
  /** Full sentence for screen readers when children are nested styled text. */
  accessibilityLabel?: string;
  style?: StyleProp<ViewStyle>;
}

/** RPG-style text box: black fill, white pixel border. Use `<Text color="accent">` for key words. */
export function DialogBox({ children, text, variant = 'dialog', more, accessibilityLabel, style }: Props) {
  const up = useBlinkClock();
  return (
    <PixelBox variant={variant} style={style} accessible accessibilityLabel={accessibilityLabel ?? text}>
      <Text variant="body">{children ?? text}</Text>
      {more ? (
        <View style={{ alignItems: 'flex-end', marginTop: 4, transform: [{ translateY: up ? 0 : 2 }] }}>
          <PixelIcon name="more" scale={2} color="accent" />
        </View>
      ) : null}
    </PixelBox>
  );
}
