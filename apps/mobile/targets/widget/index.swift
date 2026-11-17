import SwiftUI
import WidgetKit

// Freebies lock-screen widget. The API words every line (GET /users/{id}/widget), so copy changes
// ship without an app release; this file only lays them out. The app shares the API base, user id
// and token through the App Group (src/widget/sync.ts).

private let appGroup = "group.io.bitsugar.freebies"
private let cacheKey = "widgetLines"

struct WidgetLines: Codable {
  var state: String
  var title: String
  var line1: String
  var line2: String
  var inline: String
  var short: String
  var url: String
  var refreshAt: String?

  static let signedOut = WidgetLines(
    state: "none", title: "FREEBIES", line1: "Open Freebies", line2: "to follow your LA teams",
    inline: "Freebies · open the app", short: "LA", url: "freebie://teams", refreshAt: nil)
  static let sample = WidgetLines(
    state: "won", title: "RAMS WON!", line1: "Free Famous Star", line2: "Carl's Jr. · 5 days left",
    inline: "Rams won · Free Famous Star", short: "RAMS", url: "freebie://won", refreshAt: nil)
}

struct LinesEntry: TimelineEntry {
  let date: Date
  let lines: WidgetLines
}

struct Provider: TimelineProvider {
  private var defaults: UserDefaults? { UserDefaults(suiteName: appGroup) }

  func placeholder(in context: Context) -> LinesEntry { LinesEntry(date: .now, lines: .sample) }

  func getSnapshot(in context: Context, completion: @escaping (LinesEntry) -> Void) {
    if context.isPreview { return completion(LinesEntry(date: .now, lines: .sample)) }
    completion(LinesEntry(date: .now, lines: cached() ?? .signedOut))
  }

  func getTimeline(in context: Context, completion: @escaping (Timeline<LinesEntry>) -> Void) {
    fetch { fresh in
      let lines = fresh ?? cached() ?? .signedOut
      if let fresh, let data = try? JSONEncoder().encode(fresh) {
        defaults?.set(data, forKey: cacheKey)
      }
      // Ask again when the lines change (refreshAt), never sooner than 5 minutes, and after
      // 15 minutes when the fetch failed. iOS may stretch these to save battery.
      var next = Date.now.addingTimeInterval(fresh == nil ? 15 * 60 : 30 * 60)
      if let s = lines.refreshAt, let at = ISO8601DateFormatter().date(from: s), fresh != nil {
        next = max(at, Date.now.addingTimeInterval(5 * 60))
      }
      completion(Timeline(entries: [LinesEntry(date: .now, lines: lines)], policy: .after(next)))
    }
  }

  private func cached() -> WidgetLines? {
    guard let data = defaults?.data(forKey: cacheKey) else { return nil }
    return try? JSONDecoder().decode(WidgetLines.self, from: data)
  }

  private func fetch(_ done: @escaping (WidgetLines?) -> Void) {
    guard let d = defaults, let base = d.string(forKey: "apiBase"), let user = d.string(forKey: "userId"),
      let token = d.string(forKey: "token"),
      let url = URL(string: "\(base)/users/\(user)/widget")
    else { return done(nil) }
    var req = URLRequest(url: url, timeoutInterval: 15)
    req.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
    URLSession.shared.dataTask(with: req) { data, resp, _ in
      guard let data, (resp as? HTTPURLResponse)?.statusCode == 200 else { return done(nil) }
      done(try? JSONDecoder().decode(WidgetLines.self, from: data))
    }.resume()
  }
}

struct FreebiesWidgetView: View {
  @Environment(\.widgetFamily) private var family
  let entry: LinesEntry

  var body: some View {
    content
      .widgetURL(URL(string: entry.lines.url))
      .lockScreenBackground()
  }

  @ViewBuilder private var content: some View {
    let l = entry.lines
    switch family {
    case .accessoryInline:
      Label(l.inline, systemImage: l.state == "won" ? "trophy.fill" : "sportscourt.fill")
    case .accessoryCircular:
      ZStack {
        AccessoryWidgetBackground()
        VStack(spacing: 1) {
          Image(systemName: l.state == "won" ? "trophy.fill" : "sportscourt.fill")
            .font(.system(size: 16, weight: .bold))
            .widgetAccentable()
          Text(l.short)
            .font(.system(size: 10, weight: .heavy, design: .rounded))
            .lineLimit(1)
            .minimumScaleFactor(0.6)
        }
        .padding(4)
      }
    default:  // accessoryRectangular
      VStack(alignment: .leading, spacing: 0) {
        HStack(spacing: 4) {
          Image(systemName: l.state == "won" ? "trophy.fill" : "sportscourt.fill")
          Text(l.title).lineLimit(1).minimumScaleFactor(0.7)
        }
        .font(.system(size: 15, weight: .heavy, design: .rounded))
        .widgetAccentable()
        Text(l.line1).font(.system(size: 13, weight: .semibold)).lineLimit(1)
        Text(l.line2).font(.system(size: 12)).lineLimit(1).opacity(0.8)
      }
      .frame(maxWidth: .infinity, alignment: .leading)
    }
  }
}

extension View {
  /// iOS 17 requires a container background; lock-screen widgets draw their own, so it's clear.
  @ViewBuilder func lockScreenBackground() -> some View {
    if #available(iOS 17.0, *) {
      containerBackground(for: .widget) { Color.clear }
    } else {
      self
    }
  }
}

struct FreebiesWidget: Widget {
  var body: some WidgetConfiguration {
    StaticConfiguration(kind: "FreebiesWidget", provider: Provider()) { entry in
      FreebiesWidgetView(entry: entry)
    }
    .configurationDisplayName("Freebies")
    .description("Your newest win, or today's game with a deal on it.")
    .supportedFamilies([.accessoryRectangular, .accessoryInline, .accessoryCircular])
  }
}

@main
struct FreebiesWidgets: WidgetBundle {
  var body: some Widget {
    FreebiesWidget()
  }
}
