import SwiftUI

// Freebies lock-screen widget layouts, plain SwiftUI (no WidgetKit) so they can be rendered to an
// image on a Mac to check. iOS draws lock-screen widgets in one tint, so everything here is a single
// color: the 8-bit look comes from Press Start 2P and pixel art, never from color (docs/ui.md).

/// Press Start 2P, bundled in the widget (assets/, registered in Info.plist).
func pixelFont(_ size: CGFloat) -> Font { .custom("PressStart2P-Regular", size: size) }

/// Single-color pixel art from rows of "X" (filled) and "." (empty), like the app's ICONS.
struct PixelArt: View {
  let rows: [String]
  let pixel: CGFloat

  var body: some View {
    let w = rows.map(\.count).max() ?? 0
    Canvas { ctx, _ in
      for (y, row) in rows.enumerated() {
        for (x, ch) in row.enumerated() where ch == "X" {
          ctx.fill(
            Path(CGRect(x: CGFloat(x) * pixel, y: CGFloat(y) * pixel, width: pixel, height: pixel)),
            with: .foreground)
        }
      }
    }
    .frame(width: CGFloat(w) * pixel, height: CGFloat(rows.count) * pixel)
    .accessibilityHidden(true)
  }
}

enum Pixels {
  /// Same trophy as the app's tab bar (art.ts ICONS.trophy).
  static let trophy = [
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
    "..XXXXXXX..",
  ]
  /// Same calendar as the app's TODAY tab (art.ts ICONS.calendar).
  static let calendar = [
    ".X.......X.",
    "XXXXXXXXXXX",
    "XXXXXXXXXXX",
    "X.........X",
    "X.X.X.X.X.X",
    "X.........X",
    "X.X.X.XXX.X",
    "X.....XXX.X",
    "X.........X",
    "XXXXXXXXXXX",
  ]
  /// The trophy's edge only: no wins yet.
  static let trophyOutline: [String] = {
    let g = trophy.map { Array($0) }
    func on(_ x: Int, _ y: Int) -> Bool {
      y >= 0 && y < g.count && x >= 0 && x < g[y].count && g[y][x] == "X"
    }
    return g.indices.map { y in
      String(
        g[y].indices.map { x in
          on(x, y) && !(on(x - 1, y) && on(x + 1, y) && on(x, y - 1) && on(x, y + 1)) ? "X" : "."
        })
    }
  }()
}

/// What the widget shows, as the API words it (GET /users/{id}/widget).
struct Lines: Codable, Equatable {
  var state: String  // "won", "game" or "none"
  var title: String
  var line1: String
  var line2: String
  var inline: String
  var short: String
  var url: String
  var count: Int?
  var refreshAt: String?

  var wins: Int { count ?? (state == "won" ? 1 : 0) }

  static let signedOut = Lines(
    state: "none", title: "FREEBIES", line1: "Open Freebies", line2: "to follow your LA teams",
    inline: "Freebies · open the app", short: "LA", url: "freebie://teams", count: 0, refreshAt: nil)
  static let sample = Lines(
    state: "won", title: "RAMS WON!", line1: "Free Famous Star", line2: "Carl's Jr. · ends 11:59pm",
    inline: "Rams won · Free Famous Star", short: "RAMS", url: "freebie://won", count: 1,
    refreshAt: nil)
}

/// Rectangular lock-screen widget: pixel headline, then the deal (one win), a trophy row (several
/// wins), or the game.
struct RectangularView: View {
  let lines: Lines

  var body: some View {
    VStack(alignment: .leading, spacing: 5) {
      HStack(spacing: 5) {
        if lines.wins <= 1 {
          PixelArt(rows: icon, pixel: 1)
        }
        Text(lines.title)
          .font(pixelFont(10))
          .lineLimit(1)
          .minimumScaleFactor(0.7)
      }
      if lines.wins > 1 {
        HStack(spacing: 3) {
          ForEach(0..<min(lines.wins, 5), id: \.self) { _ in
            PixelArt(rows: Pixels.trophy, pixel: 1.4)
          }
          if lines.wins > 5 {
            Text("+\(lines.wins - 5)").font(pixelFont(8))
          }
        }
      } else {
        Text(lines.line1)
          .font(.system(size: 13, weight: .semibold))
          .lineLimit(1)
      }
      Text(lines.line2)
        .font(.system(size: 12))
        .lineLimit(1)
        .opacity(0.8)
    }
    .frame(maxWidth: .infinity, alignment: .leading)
  }

  private var icon: [String] {
    switch lines.state {
    case "won": return Pixels.trophy
    case "game": return Pixels.calendar
    default: return Pixels.trophyOutline
    }
  }
}

/// Circular lock-screen widget: a big pixel trophy (filled with wins, an outline without) and the
/// count, or the game time.
struct CircularView: View {
  let lines: Lines

  var body: some View {
    VStack(spacing: 3) {
      PixelArt(rows: icon, pixel: 2)
      Text(caption)
        .font(pixelFont(8))
        .lineLimit(1)
        .minimumScaleFactor(0.6)
    }
    .padding(4)
  }

  private var icon: [String] {
    switch lines.state {
    case "won": return Pixels.trophy
    case "game": return Pixels.calendar
    default: return Pixels.trophyOutline
    }
  }

  private var caption: String {
    switch lines.state {
    case "won": return "×\(lines.wins)"
    case "game": return lines.short
    default: return "0"
    }
  }
}
