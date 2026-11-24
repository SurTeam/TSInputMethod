import Foundation

enum DaemonAgent {
  static let label = "com.surteam.tsimed"

  static func softEnsureRunning() {
    if DaemonClient.isReachable() {
      return
    }
    let uid = getuid()
    let domain = "gui/\(uid)"
    let target = "\(domain)/\(label)"
    if runLaunchctl(["print", target]) != 0 {
      let plist = NSHomeDirectory() + "/Library/LaunchAgents/\(label).plist"
      _ = runLaunchctl(["bootstrap", domain, plist])
      _ = runLaunchctl(["enable", target])
    }
    _ = runLaunchctl(["kickstart", target])
  }

  static func forceRestart() {
    let uid = getuid()
    let domain = "gui/\(uid)"
    let target = "\(domain)/\(label)"
    if runLaunchctl(["kickstart", "-k", target]) == 0 {
      return
    }
    let plist = NSHomeDirectory() + "/Library/LaunchAgents/\(label).plist"
    _ = runLaunchctl(["bootout", target])
    _ = runLaunchctl(["bootstrap", domain, plist])
    _ = runLaunchctl(["enable", target])
    _ = runLaunchctl(["kickstart", "-k", target])
  }

  @discardableResult
  private static func runLaunchctl(_ arguments: [String]) -> Int32 {
    let task = Process()
    task.executableURL = URL(fileURLWithPath: "/bin/launchctl")
    task.arguments = arguments
    task.standardOutput = FileHandle.nullDevice
    task.standardError = FileHandle.nullDevice
    do {
      try task.run()
      task.waitUntilExit()
      return task.terminationStatus
    } catch {
      return -1
    }
  }
}
