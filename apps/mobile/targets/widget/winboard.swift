import ActivityKit
import SwiftUI
import WidgetKit

// The Win Board Live Activity: one card for every live win (docs/ui.md). Live Activities may use
// full color, so this card carries the app's look: black panel, the team stripe, Press Start 2P,
// the deals' own sprites. The app starts, updates and ends it (modules/win-board); a server
// push-to-start comes later.

// Must match modules/win-board/ios/WinBoardAttributes.swift exactly: ActivityKit pairs the app
// and the extension by this type.
public struct WinBoardAttributes: ActivityAttributes {
  public struct Sprite: Codable, Hashable {
    public var w: Int
    public var h: Int
    public var px: [String]
    public var pal: [String: String]
  }

  public struct Win: Codable, Hashable {
    public var sprite: Sprite?
    public var primary: String
    public var secondary: String
  }

  public struct ContentState: Codable, Hashable {
    public var title: String
    public var offer: String
    public var detail: String
    public var count: Int
    public var stripe: [String]
    public var wins: [Win]
    public var url: String
  }

  public init() {}
}

private let gold = Color(hex: "#F8B800")
private let lilac = Color(hex: "#D8B8F8")

extension Color {
  /// "#RRGGBB" from the API; black if malformed.
  init(hex: String) {
    let v = UInt64(hex.trimmingCharacters(in: CharacterSet(charactersIn: "#")), radix: 16) ?? 0
    self.init(
      red: Double((v >> 16) & 0xFF) / 255, green: Double((v >> 8) & 0xFF) / 255,
      blue: Double(v & 0xFF) / 255)
  }
}

/// A full-color sprite from the API: rows of palette keys, "." transparent.
struct SpriteArt: View {
  let sprite: WinBoardAttributes.Sprite?
  let pixel: CGFloat

  var body: some View {
    if let s = sprite {
      Canvas { ctx, _ in
        for (y, row) in s.px.enumerated() {
          for (x, key) in row.enumerated() {
            guard let hex = s.pal[String(key)] else { continue }
            ctx.fill(
              Path(CGRect(x: CGFloat(x) * pixel, y: CGFloat(y) * pixel, width: pixel, height: pixel)),
              with: .color(Color(hex: hex)))
          }
        }
      }
      .frame(width: CGFloat(s.w) * pixel, height: CGFloat(s.h) * pixel)
      .accessibilityHidden(true)
    } else {
      PixelArt(rows: Pixels.trophy, pixel: pixel * 1.4).foregroundStyle(gold)
    }
  }
}

/// The team stripe along the top: one band per color.
struct Stripe: View {
  let colors: [String]
  var body: some View {
    HStack(spacing: 0) {
      ForEach(Array(colors.enumerated()), id: \.offset) { _, c in Color(hex: c) }
    }
    .frame(height: 6)
  }
}

/// A sprite on its team's two-color tick, like the mockup's shelf.
struct WinSlot: View {
  let win: WinBoardAttributes.Win
  var body: some View {
    VStack(spacing: 3) {
      SpriteArt(sprite: win.sprite, pixel: 3)
      HStack(spacing: 0) {
        Color(hex: win.primary)
        Color(hex: win.secondary)
      }
      .frame(width: 36, height: 5)
    }
  }
}

/// The lock-screen card: 1 win (the deal is the hero) or several (one card, sprites on ticks).
struct WinBoardLockView: View {
  let state: WinBoardAttributes.ContentState

  var body: some View {
    VStack(alignment: .leading, spacing: 0) {
      Stripe(colors: state.stripe)
      Group {
        if state.count <= 1, let win = state.wins.first {
          HStack(spacing: 12) {
            SpriteArt(sprite: win.sprite, pixel: 4)
            VStack(alignment: .leading, spacing: 6) {
              Text(state.title).font(pixelFont(12)).foregroundStyle(gold).lineLimit(1)
              Text(state.offer).font(.system(size: 14, weight: .semibold)).foregroundStyle(.white)
                .lineLimit(1)
              Text(state.detail).font(pixelFont(8)).foregroundStyle(lilac).lineLimit(2)
            }
          }
        } else {
          VStack(alignment: .leading, spacing: 8) {
            HStack {
              Text(state.title).font(pixelFont(12)).foregroundStyle(gold)
              Spacer()
              PixelArt(rows: Pixels.trophy, pixel: 2).foregroundStyle(gold)
            }
            HStack(alignment: .bottom, spacing: 6) {
              ForEach(Array(state.wins.prefix(4).enumerated()), id: \.offset) { _, w in WinSlot(win: w) }
              if state.count > 4 {
                Text("+\(state.count - 4)").font(pixelFont(10)).foregroundStyle(.white)
              }
            }
            Text(state.detail).font(pixelFont(8)).foregroundStyle(lilac).lineLimit(2)
          }
        }
      }
      .padding(.horizontal, 14)
      .padding(.vertical, 12)
    }
    .background(Color.black)
    .widgetURL(URL(string: state.url))
  }
}

struct WinBoardLiveActivity: Widget {
  var body: some WidgetConfiguration {
    ActivityConfiguration(for: WinBoardAttributes.self) { context in
      WinBoardLockView(state: context.state)
        .activityBackgroundTint(.black)
        .activitySystemActionForegroundColor(.white)
    } dynamicIsland: { context in
      let s = context.state
      return DynamicIsland {
        DynamicIslandExpandedRegion(.leading) {
          SpriteArt(sprite: s.wins.first?.sprite, pixel: 3)
        }
        DynamicIslandExpandedRegion(.center) {
          Text(s.title).font(pixelFont(12)).foregroundStyle(gold).lineLimit(1)
        }
        DynamicIslandExpandedRegion(.trailing) {
          Text("×\(s.count)").font(pixelFont(12)).foregroundStyle(.white)
        }
        DynamicIslandExpandedRegion(.bottom) {
          Text(s.count <= 1 ? "\(s.offer) · \(s.detail)" : s.detail)
            .font(pixelFont(8)).foregroundStyle(lilac).lineLimit(2)
        }
      } compactLeading: {
        PixelArt(rows: Pixels.trophy, pixel: 1.6).foregroundStyle(gold)
      } compactTrailing: {
        Text("×\(s.count)").font(pixelFont(10)).foregroundStyle(gold)
      } minimal: {
        PixelArt(rows: Pixels.trophy, pixel: 1.6).foregroundStyle(gold)
      }
      .widgetURL(URL(string: s.url))
    }
  }
}
