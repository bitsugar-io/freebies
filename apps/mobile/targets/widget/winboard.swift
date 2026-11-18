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

  /// One meal on the card. Times are seconds since the epoch; startsAt 0 means it can be used now.
  public struct Meal: Codable, Hashable {
    public var url: String
    public var sprite: Sprite?
    public var startsAt: Double
    public var startsLabel: String
    public var endsAt: Double
  }

  public struct ContentState: Codable, Hashable {
    public var title: String
    public var count: Int
    /// Wins not on the card ("+N MORE").
    public var more: Int
    public var stripe: [String]
    public var meals: [Meal]
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

private let red = Color(hex: "#F83800")

/// A meal's time left: "STARTS 10AM" before it can be used, "1D 06H" over a day, then a live
/// H:MM:SS countdown that iOS ticks on the lock screen by itself (red under an hour).
struct MealClock: View {
  let meal: WinBoardAttributes.Meal
  let size: CGFloat

  var body: some View {
    let now = Date()
    let end = Date(timeIntervalSince1970: meal.endsAt)
    let left = end.timeIntervalSince(now)
    Group {
      if meal.startsAt > now.timeIntervalSince1970 {
        Text("STARTS\n\(meal.startsLabel)")
          .font(pixelFont(size * 0.8))
          .foregroundStyle(lilac)
          .lineSpacing(4)
      } else if left >= 86_400 {
        let h = Int(left) / 3600
        Text("\(h / 24)D \(String(format: "%02d", h % 24))H")
          .font(pixelFont(size))
          .foregroundStyle(.white)
      } else if left > 0 {
        Text(timerInterval: now...end, countsDown: true)
          .font(pixelFont(size))
          .foregroundStyle(left < 3600 ? red : .white)
      } else {
        Text("ENDED").font(pixelFont(size * 0.8)).foregroundStyle(Color(white: 0.5))
      }
    }
    .multilineTextAlignment(.center)
    .monospacedDigit()
  }
}

/// One meal with its clock, centered in its column; tapping it opens that deal.
struct MealSlot: View {
  let meal: WinBoardAttributes.Meal
  var body: some View {
    Link(destination: URL(string: meal.url) ?? URL(string: "freebie://won")!) {
      VStack(spacing: 8) {
        SpriteArt(sprite: meal.sprite, pixel: 3)
        MealClock(meal: meal, size: 10)
      }
      .frame(maxWidth: .infinity)
    }
  }
}

/// The lock-screen card (owner's pick, 2026-11-17): no words, just the food and its countdown.
/// One win: the meal, who won and a big clock. Several: the 3 that end soonest with their clocks,
/// then "+N MORE".
struct WinBoardLockView: View {
  let state: WinBoardAttributes.ContentState

  var body: some View {
    VStack(alignment: .leading, spacing: 0) {
      Stripe(colors: state.stripe)
      Group {
        if state.count <= 1, let meal = state.meals.first {
          HStack(spacing: 16) {
            SpriteArt(sprite: meal.sprite, pixel: 5)
            VStack(alignment: .leading, spacing: 16) {
              Text(state.title).font(pixelFont(12)).foregroundStyle(gold).lineLimit(1)
                .minimumScaleFactor(0.7)
              MealClock(meal: meal, size: 20)
            }
          }
        } else {
          VStack(alignment: .leading, spacing: 16) {
            HStack {
              Text(state.title).font(pixelFont(12)).foregroundStyle(gold)
              Spacer()
              PixelArt(rows: Pixels.trophy, pixel: 2).foregroundStyle(gold)
            }
            HStack(alignment: .top, spacing: 8) {
              ForEach(Array(state.meals.prefix(3).enumerated()), id: \.offset) { _, m in MealSlot(meal: m) }
              if state.more > 0 {
                Text("+\(state.more)\nMORE")
                  .font(pixelFont(10)).foregroundStyle(lilac)
                  .multilineTextAlignment(.center).lineSpacing(4)
                  .frame(maxWidth: .infinity, minHeight: 48)
              }
            }
          }
        }
      }
      .padding(16)
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
          SpriteArt(sprite: s.meals.first?.sprite, pixel: 3)
        }
        DynamicIslandExpandedRegion(.center) {
          Text(s.title).font(pixelFont(12)).foregroundStyle(gold).lineLimit(1)
        }
        DynamicIslandExpandedRegion(.trailing) {
          Text("×\(s.count)").font(pixelFont(12)).foregroundStyle(.white)
        }
        DynamicIslandExpandedRegion(.bottom) {
          if let m = s.meals.first { MealClock(meal: m, size: 14) }
        }
      } compactLeading: {
        // The card's headline in its gold pixel font: "4 WINS!", or "WON!" for one (a team name
        // doesn't fit this narrow side of the island).
        Text(s.count > 1 ? "\(s.count) WINS!" : "WON!")
          .font(pixelFont(9))
          .foregroundStyle(gold)
          .lineLimit(1)
          .minimumScaleFactor(0.7)
      } compactTrailing: {
        if let m = s.meals.first { MealClock(meal: m, size: 9).frame(maxWidth: 64) }
      } minimal: {
        PixelArt(rows: Pixels.trophy, pixel: 1.6).foregroundStyle(gold)
      }
      .widgetURL(URL(string: s.url))
    }
  }
}
