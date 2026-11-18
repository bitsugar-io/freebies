import ActivityKit
import Foundation

// The Win Board Live Activity's data. ActivityKit matches the app and the widget extension by this
// type, so the same definition lives in targets/widget/winboard.swift; keep the two identical.
// The JS side builds ContentState in src/widget/winBoardState.ts.
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
