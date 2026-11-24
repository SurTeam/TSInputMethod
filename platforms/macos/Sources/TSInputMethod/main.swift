import Cocoa
import InputMethodKit

@main
struct TSInputMethodMain {
  static let bundleID = "com.surteam.inputmethod.TSInputMethod"
  static let hansModeID = "com.surteam.inputmethod.TSInputMethod.Hans"

  static func main() {
    let args = CommandLine.arguments
    if args.contains("--register-input-source") || args.contains("--install") {
      registerInputSource()
      enableInputSources()
      return
    }
    if args.contains("--enable-input-source") {
      enableInputSources()
      return
    }

    let bundle = Bundle.main
    let connectionName =
      (bundle.object(forInfoDictionaryKey: "InputMethodConnectionName") as? String)
      ?? "TSInputMethod_1_Connection"
    guard let bundleIdentifier = bundle.bundleIdentifier else {
      fputs("missing bundle identifier\n", stderr)
      return
    }
    _ = ConfigStore.load()
    AppNotifier.prepare()
    DaemonWatchdog.start(withInitialEnsure: true)
    _ = IMKServer(name: connectionName, bundleIdentifier: bundleIdentifier)

    let app = NSApplication.shared
    app.setActivationPolicy(.accessory)
    app.run()
  }

  static func registerInputSource() {
    collapseDuplicateSourcesIfNeeded()
    let parentCount = sourceCount(id: bundleID)
    let hansCount = sourceCount(id: hansModeID)
    if parentCount > 0 && hansCount > 0 {
      fputs(
        "TISRegisterInputSource skipped; already present parent=\(parentCount) Hans=\(hansCount)\n",
        stderr)
      return
    }
    let url = Bundle.main.bundleURL as CFURL
    let status = TISRegisterInputSource(url)
    fputs(
      "TISRegisterInputSource status=\(status) parentBefore=\(parentCount) HansBefore=\(hansCount)\n",
      stderr)
  }

  static func enableInputSources() {
    collapseDuplicateSourcesIfNeeded()
    enableSource(id: bundleID, label: "parent")
    enableSource(id: hansModeID, label: "Hans")
    selectSource(id: hansModeID, label: "Hans")
    recordEnabledThirdPartySource()
    let parentCount = sourceCount(id: bundleID)
    let hansCount = sourceCount(id: hansModeID)
    fputs("post-enable source counts parent=\(parentCount) Hans=\(hansCount)\n", stderr)
  }

  static func sourceCount(id: String) -> Int {
    findSources(id: id).count
  }

  static func findSources(id: String) -> [TISInputSource] {
    let filter = [kTISPropertyInputSourceID as String: id] as CFDictionary
    return (TISCreateInputSourceList(filter, true)?.takeRetainedValue() as? [TISInputSource]) ?? []
  }

  static func findSource(id: String) -> TISInputSource? {
    findSources(id: id).first
  }

  static func collapseDuplicateSourcesIfNeeded() {
    let hansCount = sourceCount(id: hansModeID)
    let parentCount = sourceCount(id: bundleID)
    if hansCount <= 1 && parentCount <= 1 {
      return
    }
    fputs(
      "duplicate TIS sources detected parent=\(parentCount) Hans=\(hansCount); refreshing TextInput agents\n",
      stderr)
    let task = Process()
    task.executableURL = URL(fileURLWithPath: "/usr/bin/killall")
    task.arguments = ["TextInputMenuAgent", "TextInputSwitcher"]
    task.standardOutput = FileHandle.nullDevice
    task.standardError = FileHandle.nullDevice
    try? task.run()
    task.waitUntilExit()
    usleep(300_000)
    let hansAfter = sourceCount(id: hansModeID)
    let parentAfter = sourceCount(id: bundleID)
    fputs("after TextInput refresh parent=\(parentAfter) Hans=\(hansAfter)\n", stderr)
  }

  static func enableSource(id: String, label: String) {
    guard let source = findSource(id: id) else {
      fputs("\(label) source not found after register: \(id)\n", stderr)
      return
    }
    let status = TISEnableInputSource(source)
    fputs("TISEnableInputSource \(id) status=\(status)\n", stderr)
    if let enRef = TISGetInputSourceProperty(source, kTISPropertyInputSourceIsEnabled) {
      let enabled = CFBooleanGetValue(Unmanaged<CFBoolean>.fromOpaque(enRef).takeUnretainedValue())
      fputs("\(label) enabled=\(enabled)\n", stderr)
    }
    if let langRef = TISGetInputSourceProperty(source, kTISPropertyInputSourceLanguages) {
      let langs = Unmanaged<CFArray>.fromOpaque(langRef).takeUnretainedValue() as! [String]
      fputs("\(label) langs=\(langs.joined(separator: ","))\n", stderr)
    }
  }

  static func selectSource(id: String, label: String) {
    guard let source = findSource(id: id) else {
      fputs("\(label) source not found to select: \(id)\n", stderr)
      return
    }
    let status = TISSelectInputSource(source)
    fputs("TISSelectInputSource \(id) status=\(status)\n", stderr)
    if let selectedRef = TISGetInputSourceProperty(source, kTISPropertyInputSourceIsSelected) {
      let selected = CFBooleanGetValue(
        Unmanaged<CFBoolean>.fromOpaque(selectedRef).takeUnretainedValue())
      fputs("\(label) selected=\(selected)\n", stderr)
    }
  }

  static func hansEnabledRecord() -> [String: Any] {
    [
      "Bundle ID": bundleID,
      "Input Mode": hansModeID,
      "InputSourceKind": "Input Mode",
    ]
  }

  static func parentEnabledRecord() -> [String: Any] {
    [
      "Bundle ID": bundleID,
      "InputSourceKind": "Keyboard Input Method",
    ]
  }

  static func preferenceArray(appID: String, key: String) -> [[String: Any]] {
    if let raw = CFPreferencesCopyValue(
      key as CFString,
      appID as CFString,
      kCFPreferencesCurrentUser,
      kCFPreferencesAnyHost
    ) as? [[String: Any]] {
      return raw
    }
    if let raw = CFPreferencesCopyAppValue(key as CFString, appID as CFString) as? [[String: Any]] {
      return raw
    }
    return []
  }

  static func setPreferenceArray(_ value: [[String: Any]], appID: String, key: String) {
    let array = value as CFArray
    CFPreferencesSetValue(
      key as CFString,
      array,
      appID as CFString,
      kCFPreferencesCurrentUser,
      kCFPreferencesAnyHost
    )
    CFPreferencesSynchronize(appID as CFString, kCFPreferencesCurrentUser, kCFPreferencesAnyHost)
    CFPreferencesSetValue(
      key as CFString,
      array,
      appID as CFString,
      kCFPreferencesCurrentUser,
      kCFPreferencesCurrentHost
    )
    CFPreferencesSynchronize(
      appID as CFString, kCFPreferencesCurrentUser, kCFPreferencesCurrentHost)
    CFPreferencesSetAppValue(key as CFString, array, appID as CFString)
    CFPreferencesAppSynchronize(appID as CFString)
  }

  static func recordEnabledThirdPartySource() {
    let appID = "com.apple.inputsources"
    let key = "AppleEnabledThirdPartyInputSources"
    var enabled = preferenceArray(appID: appID, key: key)
    enabled.removeAll { ($0["Bundle ID"] as? String) == bundleID }
    enabled.append(hansEnabledRecord())
    enabled.append(parentEnabledRecord())
    setPreferenceArray(enabled, appID: appID, key: key)
    fputs("recorded AppleEnabledThirdPartyInputSources parent+Hans for \(bundleID)\n", stderr)
  }
}
