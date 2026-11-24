import Foundation
import UserNotifications

enum AppNotifier {
  private static let categoryID = "tsime.status"
  private static var authorized = false
  private static var requested = false
  private static var didNotifyInitializing = false
  private static var didNotifyReady = false

  static func prepare() {
    guard !requested else {
      return
    }
    requested = true
    let center = UNUserNotificationCenter.current()
    center.requestAuthorization(options: [.alert, .sound]) { granted, _ in
      authorized = granted
    }
  }

  static func resetLifecycleNotifications() {
    didNotifyInitializing = false
    didNotifyReady = false
  }

  static func notifyInitializing() {
    prepare()
    guard !didNotifyInitializing else {
      return
    }
    didNotifyInitializing = true
    post(title: "TSInputMethod", body: "Initializing tsimed…", identifier: "tsime.initializing")
  }

  static func notifyReady() {
    prepare()
    guard !didNotifyReady else {
      return
    }
    didNotifyReady = true
    post(title: "TSInputMethod", body: "tsimed is ready", identifier: "tsime.ready")
  }

  private static func post(title: String, body: String, identifier: String) {
    let center = UNUserNotificationCenter.current()
    center.getNotificationSettings { settings in
      let allowed: Bool
      switch settings.authorizationStatus {
      case .authorized, .provisional, .ephemeral:
        allowed = true
      default:
        allowed = authorized
      }
      guard allowed else {
        return
      }
      let content = UNMutableNotificationContent()
      content.title = title
      content.body = body
      content.sound = .default
      content.categoryIdentifier = categoryID
      let request = UNNotificationRequest(identifier: identifier, content: content, trigger: nil)
      center.add(request, withCompletionHandler: nil)
    }
  }
}
