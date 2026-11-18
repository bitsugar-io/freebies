# App Store Screenshots

The 2.0 App Store screenshots, one folder per App Store Connect display slot. The shot list and
captions are in [APP-STORE.md](../../APP-STORE.md); how to capture them is in
[.claude/skills/ops/references/release-mobile.md](../../../../.claude/skills/ops/references/release-mobile.md).

| Folder      | Size      | App Store Connect slot   |
| ----------- | --------- | ------------------------ |
| `6.9-inch/` | 1320×2868 | iPhone 6.9" Display      |
| `6.5-inch/` | 1284×2778 | iPhone 6.5" Display      |
| `6.3-inch/` | 1206×2622 | iPhone 6.1"/6.3" Display |

The 6.9" set is the original, captured on the iPhone Pro Max simulator and framed with captions. Run
`./resize.sh` to remake the other two from it; never edit them by hand. Each slot rejects any other
size ("File dimensions are invalid"), so upload each folder to its own slot.

The README screenshots are separate, in `asset/screenshots/` at the repo root.
