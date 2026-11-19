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
        let connectionName = (bundle.object(forInfoDictionaryKey: "InputMethodConnectionName") as? String)
            ?? "TSInputMethod_1_Connection"
        guard let bundleIdentifier = bundle.bundleIdentifier else {
            fputs("missing bundle identifier\n", stderr)
            return
        }
        _ = IMKServer(name: connectionName, bundleIdentifier: bundleIdentifier)

        let app = NSApplication.shared
        app.setActivationPolicy(.accessory)
        app.run()
    }

    static func registerInputSource() {
        let appURL = Bundle.main.bundleURL
        let status = TISRegisterInputSource(appURL as CFURL)
        fputs("TISRegisterInputSource \(appURL.path) status=\(status)\n", stderr)
    }

    static func enableInputSources() {
        enableSource(id: bundleID, label: "parent")
        enableSource(id: hansModeID, label: "Hans")
        recordEnabledThirdPartySource()
    }

    static func enableSource(id: String, label: String) {
        let filter = [kTISPropertyInputSourceID as String: id] as CFDictionary
        guard let sources = TISCreateInputSourceList(filter, true)?.takeRetainedValue() as? [TISInputSource],
              let source = sources.first else {
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

    static func recordEnabledThirdPartySource() {
        let prefsPath = NSHomeDirectory() + "/Library/Preferences/com.apple.inputsources.plist"
        var prefs = (NSDictionary(contentsOfFile: prefsPath) as? [String: Any]) ?? [:]
        var enabled = (prefs["AppleEnabledThirdPartyInputSources"] as? [[String: Any]]) ?? []
        enabled.removeAll { ($0["Bundle ID"] as? String) == bundleID }
        enabled.append([
            "Bundle ID": bundleID,
            "Input Mode": hansModeID,
            "InputSourceKind": "Input Mode"
        ])
        enabled.append([
            "Bundle ID": bundleID,
            "InputSourceKind": "Keyboard Input Method"
        ])
        prefs["AppleEnabledThirdPartyInputSources"] = enabled
        (prefs as NSDictionary).write(toFile: prefsPath, atomically: true)
        fputs("recorded AppleEnabledThirdPartyInputSources for \(bundleID)\n", stderr)
    }
}
