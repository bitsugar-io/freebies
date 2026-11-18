import ActivityKit
import ExpoModulesCore
import Foundation

/// Starts, updates and ends the Win Board Live Activity from JS (src/widget/winBoard.ts). One card
/// at a time: a new state updates the running card instead of starting another.
public class WinBoardModule: Module {
  /// Whether this app launch has synced the card yet (see the first-sync rule in `sync`).
  static var syncedThisLaunch = false

  public func definition() -> ModuleDefinition {
    Name("WinBoard")

    Function("isSupported") { () -> Bool in
      if #available(iOS 16.2, *) {
        return ActivityAuthorizationInfo().areActivitiesEnabled
      }
      return false
    }

    /// `json` is a WinBoardState (or "null" to end the card). `staleAt` is ms since epoch or 0.
    AsyncFunction("sync") { (json: String, staleAt: Double) in
      guard #available(iOS 16.2, *) else { return }
      let running = Activity<WinBoardAttributes>.activities
      guard let data = json.data(using: .utf8),
        let state = try? JSONDecoder().decode(WinBoardAttributes.ContentState.self, from: data),
        state.count > 0
      else {
        for activity in running {
          await activity.end(nil, dismissalPolicy: .immediate)
        }
        return
      }
      let stale = staleAt > 0 ? Date(timeIntervalSince1970: staleAt / 1000) : nil
      let content = ActivityContent(state: state, staleDate: stale)
      // The first sync after a launch replaces any running card instead of updating it: a card
      // started by an older build may hold data in an older shape, which the widget can't draw.
      if !WinBoardModule.syncedThisLaunch {
        WinBoardModule.syncedThisLaunch = true
        for activity in running {
          await activity.end(nil, dismissalPolicy: .immediate)
        }
        _ = try? Activity.request(attributes: WinBoardAttributes(), content: content, pushType: nil)
        return
      }
      if let current = running.first {
        await current.update(content)
        // Only one card: end any extras left from an older build.
        for extra in running.dropFirst() {
          await extra.end(nil, dismissalPolicy: .immediate)
        }
      } else {
        _ = try? Activity.request(attributes: WinBoardAttributes(), content: content, pushType: nil)
      }
    }
  }
}
