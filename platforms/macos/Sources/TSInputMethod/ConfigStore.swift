import Cocoa
import Foundation

enum ModeToggleTriggerKind: String, Codable, Equatable {
    case keyDown
    case modifierAlone
}

struct PersistedModeToggleBinding: Codable, Equatable {
    var id: String
    var keyCode: UInt16
    var triggerKind: ModeToggleTriggerKind
    var modifierFlagRaw: UInt?
    var enabled: Bool

    func toBinding() -> ModeToggleBinding? {
        let trigger: ModeToggleTrigger
        switch triggerKind {
        case .keyDown:
            trigger = .keyDown
        case .modifierAlone:
            guard let raw = modifierFlagRaw else {
                return nil
            }
            trigger = .modifierAlone(flag: NSEvent.ModifierFlags(rawValue: raw))
        }
        return ModeToggleBinding(id: id, keyCode: keyCode, trigger: trigger, enabled: enabled)
    }

    static func from(_ binding: ModeToggleBinding) -> PersistedModeToggleBinding {
        switch binding.trigger {
        case .keyDown:
            return PersistedModeToggleBinding(
                id: binding.id,
                keyCode: binding.keyCode,
                triggerKind: .keyDown,
                modifierFlagRaw: nil,
                enabled: binding.enabled
            )
        case .modifierAlone(let flag):
            return PersistedModeToggleBinding(
                id: binding.id,
                keyCode: binding.keyCode,
                triggerKind: .modifierAlone,
                modifierFlagRaw: flag.rawValue,
                enabled: binding.enabled
            )
        }
    }
}

struct TSIMESettings: Codable, Equatable {
    var modeToggleBindings: [PersistedModeToggleBinding]
    var candidateColumns: Int
    var candidateFontName: String
    var candidateFontSize: Double
    var highlightColorHex: String

    static let `default` = TSIMESettings(
        modeToggleBindings: ModeToggleBinding.defaults.map(PersistedModeToggleBinding.from),
        candidateColumns: 9,
        candidateFontName: NSFont.systemFont(ofSize: 16).fontName,
        candidateFontSize: 16,
        highlightColorHex: NSColor.controlAccentColor.tsimeHexString
    )

    var clampedColumns: Int {
        min(9, max(1, candidateColumns))
    }

    var highlightColor: NSColor {
        NSColor.tsimeColor(hex: highlightColorHex) ?? .controlAccentColor
    }

    var candidateFont: NSFont {
        let size = CGFloat(max(10, min(36, candidateFontSize)))
        if let named = NSFont(name: candidateFontName, size: size) {
            return named
        }
        return NSFont.systemFont(ofSize: size)
    }

    func modeBindings() -> [ModeToggleBinding] {
        let decoded = modeToggleBindings.compactMap { $0.toBinding() }
        return decoded.isEmpty ? ModeToggleBinding.defaults : decoded
    }
}

enum ConfigStore {
    static let didChangeNotification = Notification.Name("TSIMESettingsDidChange")

    static func configDirectoryURL() -> URL {
        if let xdg = ProcessInfo.processInfo.environment["XDG_CONFIG_HOME"], !xdg.isEmpty {
            return URL(fileURLWithPath: xdg, isDirectory: true).appendingPathComponent("tsime", isDirectory: true)
        }
        return URL(fileURLWithPath: NSHomeDirectory(), isDirectory: true)
            .appendingPathComponent(".config", isDirectory: true)
            .appendingPathComponent("tsime", isDirectory: true)
    }

    static func settingsFileURL() -> URL {
        configDirectoryURL().appendingPathComponent("settings.json", isDirectory: false)
    }

    @discardableResult
    static func ensureConfigDirectory() -> URL {
        let dir = configDirectoryURL()
        try? FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        return dir
    }

    static func load() -> TSIMESettings {
        ensureConfigDirectory()
        let url = settingsFileURL()
        guard let data = try? Data(contentsOf: url) else {
            return .default
        }
        let decoder = JSONDecoder()
        guard let decoded = try? decoder.decode(TSIMESettings.self, from: data) else {
            return .default
        }
        var settings = decoded
        settings.candidateColumns = settings.clampedColumns
        if settings.candidateFontSize < 10 || settings.candidateFontSize > 36 {
            settings.candidateFontSize = 16
        }
        if settings.modeToggleBindings.isEmpty {
            settings.modeToggleBindings = TSIMESettings.default.modeToggleBindings
        }
        return settings
    }

    static func save(_ settings: TSIMESettings) throws {
        ensureConfigDirectory()
        var normalized = settings
        normalized.candidateColumns = settings.clampedColumns
        let encoder = JSONEncoder()
        encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
        let data = try encoder.encode(normalized)
        try data.write(to: settingsFileURL(), options: .atomic)
        NotificationCenter.default.post(name: didChangeNotification, object: nil)
    }
}

extension NSColor {
    var tsimeHexString: String {
        guard let rgb = usingColorSpace(.sRGB) else {
            return "#0A7AFF"
        }
        let r = Int(round(rgb.redComponent * 255))
        let g = Int(round(rgb.greenComponent * 255))
        let b = Int(round(rgb.blueComponent * 255))
        let a = Int(round(rgb.alphaComponent * 255))
        if a >= 255 {
            return String(format: "#%02X%02X%02X", r, g, b)
        }
        return String(format: "#%02X%02X%02X%02X", r, g, b, a)
    }

    static func tsimeColor(hex: String) -> NSColor? {
        var raw = hex.trimmingCharacters(in: .whitespacesAndNewlines).uppercased()
        if raw.hasPrefix("#") {
            raw.removeFirst()
        }
        guard raw.count == 6 || raw.count == 8 else {
            return nil
        }
        var value: UInt64 = 0
        guard Scanner(string: raw).scanHexInt64(&value) else {
            return nil
        }
        let hasAlpha = raw.count == 8
        let r: CGFloat
        let g: CGFloat
        let b: CGFloat
        let a: CGFloat
        if hasAlpha {
            r = CGFloat((value & 0xFF000000) >> 24) / 255
            g = CGFloat((value & 0x00FF0000) >> 16) / 255
            b = CGFloat((value & 0x0000FF00) >> 8) / 255
            a = CGFloat(value & 0x000000FF) / 255
        } else {
            r = CGFloat((value & 0xFF0000) >> 16) / 255
            g = CGFloat((value & 0x00FF00) >> 8) / 255
            b = CGFloat(value & 0x0000FF) / 255
            a = 1
        }
        return NSColor(srgbRed: r, green: g, blue: b, alpha: a)
    }
}
