// Original pixel art as data. Sprites use palette keys, '.' is transparent.
// Sport badges are 16x16 ports of the procedural badges in docs/pivot/la-jumbotron-vs-8bit.html.
// The palm logo lives in logoArt.ts. All colors are NES master-palette values.

export interface Bitmap {
  width: number;
  height: number;
  pixels: string[];
  palette: Record<string, string>;
}


export const SPORT_BADGES: Record<'basketball' | 'baseball' | 'soccer' | 'football' | 'hockey', Bitmap> = {
  basketball: {
    "width": 16,
    "height": 16,
    "pixels": [
      "................",
      "....KKKKKKKK....",
      "...KKffsfffKK...",
      "..KsfffsffffsK..",
      ".KKfshfsfffsfKK.",
      ".KffhhhsffssffK.",
      ".KfffhfsffsfffK.",
      ".KssssssssssssK.",
      ".KfffsfsffsfffK.",
      ".KfffsfsffsfffK.",
      ".KffssfsffssffK.",
      ".KKfsffsfffsfKK.",
      "..KsfffsffffsK..",
      "...KKffsfffKK...",
      "....KKKKKKKK....",
      "................"
    ],
    "palette": {
      "K": "#000000",
      "f": "#FCA044",
      "h": "#FCA044",
      "s": "#000000"
    }
  },
  baseball: {
    "width": 16,
    "height": 16,
    "pixels": [
      "................",
      "....KKKKKKKK....",
      "...KKffffffKK...",
      "..KrffffffffrK..",
      ".KKrrhfffffrrKK.",
      ".KffhhhffffrffK.",
      ".KffrhfffffrffK.",
      ".KffrrffffrrffK.",
      ".KffrrffffrrffK.",
      ".KffrffffffrffK.",
      ".KffrffffffrffK.",
      ".KKrrffffffrrKK.",
      "..KrffffffffrK..",
      "...KKffffffKK...",
      "....KKKKKKKK....",
      "................"
    ],
    "palette": {
      "K": "#000000",
      "f": "#FCFCFC",
      "h": "#FCFCFC",
      "r": "#F83800"
    }
  },
  soccer: {
    "width": 16,
    "height": 16,
    "pixels": [
      "................",
      "....KKKKKKKK....",
      "...KKffffffKK...",
      "..KsffffffffsK..",
      ".KKsshfffffssKK.",
      ".KsshhhfffffssK.",
      ".KfffhfssfffffK.",
      ".KffffssssffffK.",
      ".KffffssssffffK.",
      ".KfffffssfffffK.",
      ".KssffffffffssK.",
      ".KKssffffffssKK.",
      "..KssffffffssK..",
      "...KKffssffKK...",
      "....KKKKKKKK....",
      "................"
    ],
    "palette": {
      "K": "#000000",
      "f": "#FCFCFC",
      "h": "#FCFCFC",
      "s": "#000000"
    }
  },
  football: {
    "width": 16,
    "height": 16,
    "pixels": [
      "................",
      "................",
      "................",
      "....KKKKKKKK....",
      "..KKKhfffffKKK..",
      ".KKhhhhhfffffKK.",
      ".KfffhwfwfwfffK.",
      "KKfffwwwwwwfffKK",
      "KKffffwfwfwfffKK",
      ".KffffffffffffK.",
      ".KKffffffffffKK.",
      "..KKKffffffKKK..",
      "....KKKKKKKK....",
      "................",
      "................",
      "................"
    ],
    "palette": {
      "K": "#000000",
      "f": "#A81000",
      "h": "#FCA044",
      "w": "#FCFCFC"
    }
  },
  hockey: {
    "width": 16,
    "height": 16,
    "pixels": [
      "............KK..",
      "...........KbK..",
      "...........KbK..",
      "..........KbK...",
      "..........KbK...",
      ".........KbK....",
      ".........KbK....",
      "........KbK.....",
      "........KtK.....",
      ".......KtK......",
      "..KKKKKKtK......",
      ".KbbbbbbbK......",
      ".KKKKKKKK.......",
      "..........KKKK..",
      ".........KggggK.",
      "..........KKKK.."
    ],
    "palette": {
      "K": "#000000",
      "b": "#F8B800",
      "t": "#FCFCFC",
      "g": "#7C7C7C"
    }
  },
};

export type Sport = keyof typeof SPORT_BADGES;

/** Single-color UI icons. 'X' = filled. Tinted at render time. */
export const ICONS = {
  moon: [
    "....XX.....",
    "..XXX......",
    ".XXX.......",
    ".XXX.......",
    "XXXX.......",
    "XXXX.......",
    "XXXXX......",
    ".XXXXX.....",
    ".XXXXXXX.X.",
    "..XXXXXXX..",
    "....XXX...."
  ],
  bag: [
    "....XXXX....",
    "...X....X...",
    "...X....X...",
    ".XXXXXXXXXX.",
    ".X........X.",
    ".X........X.",
    ".X........X.",
    ".X........X.",
    ".X........X.",
    ".XXXXXXXXXX."
  ],
  teams: [
    "XXXXX.XXXXX",
    "X...X.X...X",
    "X...X.X...X",
    "X...X.X...X",
    "XXXXX.XXXXX",
    "...........",
    "XXXXX.XXXXX",
    "X...X.XXXXX",
    "X...X.XXXXX",
    "X...X.XXXXX",
    "XXXXX.XXXXX"
  ],
  you: [
    "...XXXXX...",
    "..XXXXXXX..",
    "..XXXXXXX..",
    "..XXXXXXX..",
    "...XXXXX...",
    "...........",
    ".XXXXXXXXX.",
    "XXXXXXXXXXX",
    "XXXXXXXXXXX",
    "XXXXXXXXXXX"
  ],
  bell: [
    ".....X.....",
    "....XXX....",
    "...XXXXX...",
    "...XXXXX...",
    "..XXXXXXX..",
    "..XXXXXXX..",
    ".XXXXXXXXX.",
    "XXXXXXXXXXX",
    "...........",
    "....XXX...."
  ],
  menu: [
    "XXXXXXX",
    "XXXXXXX",
    ".......",
    "XXXXXXX",
    "XXXXXXX",
    ".......",
    "XXXXXXX",
    "XXXXXXX"
  ],
  close: [
    "XX.....XX",
    "XXX...XXX",
    ".XXX.XXX.",
    "..XXXXX..",
    "...XXX...",
    "..XXXXX..",
    ".XXX.XXX.",
    "XXX...XXX",
    "XX.....XX"
  ],
  check: [
    "........X",
    ".......XX",
    "X.....XX.",
    "XX...XX..",
    ".XX.XX...",
    "..XXX....",
    "...X....."
  ],
  warn: [
    ".....X.....",
    "....XXX....",
    "...XX.XX...",
    "...XX.XX...",
    "..XXX.XXX..",
    "..XXXXXXX..",
    ".XXXX.XXXX.",
    ".XXXXXXXXX.",
    "XXXXXXXXXXX"
  ],
  trophy: [
    "XXXXXXXXXXX",
    "X.XXXXXXX.X",
    "X.XXXXXXX.X",
    ".XXXXXXXXX.",
    "..XXXXXXX..",
    "...XXXXX...",
    "....XXX....",
    ".....X.....",
    ".....X.....",
    "...XXXXX...",
    "..XXXXXXX.."
  ],
  calendar: [
    ".X.......X.",
    "XXXXXXXXXXX",
    "XXXXXXXXXXX",
    "X.........X",
    "X.X.X.X.X.X",
    "X.........X",
    "X.X.X.XXX.X",
    "X.....XXX.X",
    "X.........X",
    "XXXXXXXXXXX"
  ],
  jersey: [
    "..XX...XX..",
    "XXXX...XXXX",
    "XXXXX.XXXXX",
    "XXXXXXXXXXX",
    ".XXXXXXXXX.",
    "..XXXXXXX..",
    "..XXXXXXX..",
    "..XXXXXXX..",
    "..XXXXXXX..",
    "..XXXXXXX..",
    "..XXXXXXX.."
  ],
  cursor: [
    "X....",
    "XX...",
    "XXX..",
    "XXXX.",
    "XXX..",
    "XX...",
    "X...."
  ],
  back: [
    "....X",
    "...XX",
    "..XXX",
    ".XXXX",
    "..XXX",
    "...XX",
    "....X"
  ],
  more: [
    "XXXXXXX",
    ".XXXXX.",
    "..XXX..",
    "...X..."
  ],
} satisfies Record<string, string[]>;

export type IconName = keyof typeof ICONS;

/** Fallback when a deal has no sprite: a plate. Never emoji. */
export const PLATE: Bitmap = {
  width: 16,
  height: 16,
  pixels: [
    "................",
    "................",
    "....KKKKKKKK....",
    "..KKwwwwwwwwKK..",
    ".KwwwwwwwwwwwwK.",
    ".KwwwwwwwwwwwwK.",
    "KwwwwKKKKKKwwwwK",
    "KwwwKwwwwwwKwwwK",
    "KwwwKwwwwwwKwwwK",
    "KwwwwKKKKKKwwwwK",
    ".KwwwwwwwwwwwwK.",
    ".KwwwwwwwwwwwwK.",
    "..KKwwwwwwwwKK..",
    "....KKKKKKKK....",
    "................",
    "................"
  ],
  palette: { K: '#000000', w: '#FCFCFC' },
};

/**
 * "No signal": an 8-bit TV full of static, for when the API can't be reached. Bundled (not from the
 * API) because it shows exactly when the API is down; the offline fallback the server-driven rule
 * allows (docs/architecture.md).
 */
export const NO_SIGNAL: Bitmap & { name: string } = {
  name: 'No signal',
  width: 16,
  height: 16,
  pixels: [
    "................",
    "....a......a....",
    ".....a....a.....",
    "......a..a......",
    ".......aa.......",
    ".gggggggggggggg.",
    "gKKKKKKKKKgggggg",
    "gKwlKwlwKKggoggg",
    "gKlKwwlKwKgggggg",
    "gKwwlKwlwKggoggg",
    "gKKlwlwwlKgggggg",
    "gKwlwKlwKKggoggg",
    "gKKKKKKKKKgggggg",
    ".gggggggggggggg.",
    "..gg........gg..",
    "................"
  ],
  palette: { K: '#000000', a: '#BCBCBC', g: '#7C7C7C', w: '#FCFCFC', l: '#BCBCBC', o: '#F8B800' },
};
