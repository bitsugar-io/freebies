"""WCAG 2.x contrast ratio between two hex colors.

Usage: python3 -I contrast.py FCFCFC 004058
"""
import sys


def lum(hexstr):
    h = hexstr.lstrip("#")
    if len(h) != 6:
        sys.exit(f"not a 6-digit hex color: {hexstr}")
    out = []
    for i in (0, 2, 4):
        c = int(h[i:i + 2], 16) / 255
        out.append(c / 12.92 if c <= 0.03928 else ((c + 0.055) / 1.055) ** 2.4)
    r, g, b = out
    return 0.2126 * r + 0.7152 * g + 0.0722 * b


if len(sys.argv) != 3:
    sys.exit("usage: contrast.py <fg> <bg>")
a, b = sorted((lum(sys.argv[1]), lum(sys.argv[2])), reverse=True)
ratio = (a + 0.05) / (b + 0.05)
small = "pass" if ratio >= 4.5 else "FAIL"
large = "pass" if ratio >= 3 else "FAIL"
print(f"{ratio:.2f}:1  small text {small}  large text/icons {large}")
