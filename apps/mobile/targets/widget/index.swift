import SwiftUI
import UIKit
import WidgetKit

// Freebies lock-screen widget. The API words every line (GET /users/{id}/widget), so copy changes
// ship without an app release; views.swift lays them out in the 8-bit style. The app shares the API
// base, user id and token through the App Group (src/widget/sync.ts).

private let appGroup = "group.io.bitsugar.freebies"
private let cacheKey = "widgetLines"

struct LinesEntry: TimelineEntry {
  let date: Date
  let lines: Lines
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

  private func cached() -> Lines? {
    guard let data = defaults?.data(forKey: cacheKey) else { return nil }
    return try? JSONDecoder().decode(Lines.self, from: data)
  }

  private func fetch(_ done: @escaping (Lines?) -> Void) {
    guard let d = defaults, let base = d.string(forKey: "apiBase"), let user = d.string(forKey: "userId"),
      let token = d.string(forKey: "token"),
      let url = URL(string: "\(base)/users/\(user)/widget")
    else { return done(nil) }
    var req = URLRequest(url: url, timeoutInterval: 15)
    req.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
    URLSession.shared.dataTask(with: req) { data, resp, _ in
      guard let data, (resp as? HTTPURLResponse)?.statusCode == 200 else { return done(nil) }
      done(try? JSONDecoder().decode(Lines.self, from: data))
    }.resume()
  }
}

/// The pixel trophy as a template image, for the inline widget (it only takes text and an image).
private func trophyImage() -> UIImage {
  let rows = Pixels.trophy
  let px: CGFloat = 1
  let size = CGSize(width: CGFloat(rows[0].count) * px, height: CGFloat(rows.count) * px)
  let img = UIGraphicsImageRenderer(size: size).image { ctx in
    UIColor.white.setFill()
    for (y, row) in rows.enumerated() {
      for (x, ch) in row.enumerated() where ch == "X" {
        ctx.fill(CGRect(x: CGFloat(x) * px, y: CGFloat(y) * px, width: px, height: px))
      }
    }
  }
  return img.withRenderingMode(.alwaysTemplate)
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
    switch family {
    case .accessoryInline:
      // iOS sets the inline widget's font; only the text and one image are ours.
      if entry.lines.state == "won" {
        Label { Text(entry.lines.inline) } icon: { Image(uiImage: trophyImage()) }
      } else {
        Text(entry.lines.inline)
      }
    case .accessoryCircular:
      ZStack {
        AccessoryWidgetBackground()
        CircularView(lines: entry.lines).widgetAccentable()
      }
    default:
      RectangularView(lines: entry.lines)
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
    .description("Your wins, or today's game with a deal on it.")
    .supportedFamilies([.accessoryRectangular, .accessoryInline, .accessoryCircular])
  }
}

@main
struct FreebiesWidgets: WidgetBundle {
  var body: some Widget {
    FreebiesWidget()
  }
}
