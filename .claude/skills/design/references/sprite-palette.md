# NES master palette

The NES has no single canonical RGB palette. This is the widely used 55-color set from Lospec (see
`docs/pivot/nes-style-research.md` section 1.2 for sources and caveats). `sprites.py` refers to
colors by index: add an index to its `NES` map only from this table.

| Row | x0     | x1     | x2     | x3     | x4     | x5     | x6     | x7     | x8     | x9     | xA     | xB     | xC     | xD     |
| --- | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ | ------ |
| 0x  | 7C7C7C | 0000FC | 0000BC | 4428BC | 940084 | A80020 | A81000 | 881400 | 503000 | 007800 | 006800 | 005800 | 004058 | 000000 |
| 1x  | BCBCBC | 0078F8 | 0058F8 | 6844FC | D800CC | E40058 | F83800 | E45C10 | AC7C00 | 00B800 | 00A800 | 00A844 | 008888 | 000000 |
| 2x  | F8F8F8 | 3CBCFC | 6888FC | 9878F8 | F878F8 | F85898 | F87858 | FCA044 | F8B800 | B8F818 | 58D854 | 58F898 | 00E8D8 | 787878 |
| 3x  | FCFCFC | A4E4FC | B8B8F8 | D8B8F8 | F8B8F8 | F8A4C0 | F0D0B0 | FCE0A8 | F8D878 | D8F878 | B8F8B8 | B8F8D8 | 00FCFC | F8D8F8 |

Use `0F` (#000000) for black. Avoid `0D`.

## Indices used by existing food sprites

| Index | Hex     | Typical use                       |
| ----- | ------- | --------------------------------- |
| 0F    | #000000 | outline                           |
| 30    | #FCFCFC | egg white, rice, plate, straw     |
| 10    | #BCBCBC | foil, plate shade, lid            |
| 17    | #E45C10 | bun, pizza crust                  |
| 27    | #FCA044 | bun highlight, nuggets            |
| 07    | #881400 | patty, meat                       |
| 06    | #A81000 | sausage, dark red                 |
| 16    | #F83800 | tomato, cup, pepperoni            |
| 26    | #F87858 | ribbon highlight                  |
| 28    | #F8B800 | cheese, shell, fries, fried crust |
| 38    | #F8D878 | shine on yellow                   |
| 18    | #AC7C00 | griddle-cake dimples              |
| 1A    | #00A800 | lettuce, pickles, greens          |
| 36    | #F0D0B0 | tortilla, skin                    |
| 21    | #3CBCFC | gift box                          |
| 31    | #A4E4FC | gift highlight                    |
