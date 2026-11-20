#!/usr/bin/env bash
# Makes the 6.5", 6.3" and 6.1" Face ID App Store screenshot sets from the 6.9" originals (APP-STORE.md).
# Usage: apps/mobile/assets/screenshots/resize.sh   (needs ffmpeg)
set -euo pipefail
cd "$(dirname "$0")"
mkdir -p 6.5-inch 6.3-inch 6.1-inch-faceid
for f in 6.9-inch/0*.png; do
  n=$(basename "$f")
  # 6.5": 1284x2778. The aspect differs slightly, so fit the height and pad the sides with black.
  ffmpeg -loglevel error -y -i "$f" -vf "scale=-1:2778:flags=lanczos,pad=1284:2778:(ow-iw)/2:0:black,format=rgb24" "6.5-inch/$n"
  # 6.1"/6.3": 1206x2622, the same aspect as 6.9".
  ffmpeg -loglevel error -y -i "$f" -vf "scale=1206:2622:flags=lanczos,format=rgb24" "6.3-inch/$n"
  # 6.1" Face ID (iPhone with Face ID, medium): 1170x2532. Fit the height and pad, like 6.5".
  ffmpeg -loglevel error -y -i "$f" -vf "scale=-1:2532:flags=lanczos,pad=1170:2532:(ow-iw)/2:0:black,format=rgb24" "6.1-inch-faceid/$n"
done
echo "wrote 6.5-inch/, 6.3-inch/ and 6.1-inch-faceid/"
