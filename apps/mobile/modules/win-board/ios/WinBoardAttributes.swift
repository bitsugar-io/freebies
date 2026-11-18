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
