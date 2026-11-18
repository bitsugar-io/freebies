import ActivityKit
import ExpoModulesCore
import Foundation

/// Starts, updates and ends the Win Board Live Activity from JS (src/widget/winBoard.ts). One card
/// at a time: a new state updates the running card instead of starting another.
public class WinBoardModule: Module {
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
