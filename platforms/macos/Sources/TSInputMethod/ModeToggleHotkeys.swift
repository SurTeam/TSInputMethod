import Cocoa

enum ModeToggleTrigger: Equatable {
    case keyDown
    case modifierAlone(flag: NSEvent.ModifierFlags)
}

struct ModeToggleBinding: Equatable {
    var id: String
    var keyCode: UInt16
    var trigger: ModeToggleTrigger
    var enabled: Bool

    static let leftShiftAlone = ModeToggleBinding(
        id: "left_shift_alone",
        keyCode: 56,
        trigger: .modifierAlone(flag: .shift),
        enabled: true
    )

    static let f18KeyDown = ModeToggleBinding(
        id: "f18",
        keyCode: 79,
        trigger: .keyDown,
        enabled: true
    )

    static let defaults: [ModeToggleBinding] = [.leftShiftAlone, .f18KeyDown]
}

final class ModeToggleHotkeys {
    private static let trackedFlags: NSEvent.ModifierFlags = [.shift, .control, .option, .command]
    private static let modifierKeyCodes: Set<UInt16> = [54, 55, 56, 57, 58, 59, 60, 61, 62, 63]
    private static var aloneDown: [UInt16: Bool] = [:]
    private static var aloneUsedAsModifier: [UInt16: Bool] = [:]
    private static var lastModifiers: NSEvent.ModifierFlags = []

    private var bindings: [ModeToggleBinding]

    init(bindings: [ModeToggleBinding] = ModeToggleBinding.defaults) {
        self.bindings = bindings
    }

    func currentBindings() -> [ModeToggleBinding] {
        bindings
    }

    func setBindings(_ bindings: [ModeToggleBinding]) {
        self.bindings = bindings
        Self.aloneDown.removeAll(keepingCapacity: true)
        Self.aloneUsedAsModifier.removeAll(keepingCapacity: true)
        Self.lastModifiers = []
    }

    func handleKeyDown(_ event: NSEvent, toggle: () -> Void) -> Bool {
        for (code, isDown) in Self.aloneDown where isDown {
            Self.aloneUsedAsModifier[code] = true
        }
        for binding in bindings where binding.enabled {
            guard case .keyDown = binding.trigger else {
                continue
            }
            guard event.keyCode == binding.keyCode else {
                continue
            }
            let chord = event.modifierFlags.intersection(Self.trackedFlags)
            guard chord.isEmpty else {
                return false
            }
            toggle()
            return true
        }
        return false
    }

    func handleFlagsChanged(_ event: NSEvent, toggle: () -> Void) -> Bool {
        let modifiers = event.modifierFlags.intersection(Self.trackedFlags.union(.capsLock))
        if modifiers == Self.lastModifiers {
            return false
        }
        Self.lastModifiers = modifiers

        var didToggle = false
        for binding in bindings where binding.enabled {
            guard case .modifierAlone(let flag) = binding.trigger else {
                continue
            }
            let flagDown = modifiers.contains(flag)
            if event.keyCode != binding.keyCode {
                if Self.aloneDown[binding.keyCode] == true && flagDown {
                    Self.aloneUsedAsModifier[binding.keyCode] = true
                }
                continue
            }
            if !Self.modifierKeyCodes.contains(event.keyCode) {
                continue
            }
            let wasDown = Self.aloneDown[binding.keyCode] == true
            if flagDown && !wasDown {
                Self.aloneDown[binding.keyCode] = true
                Self.aloneUsedAsModifier[binding.keyCode] = false
                continue
            }
            if !flagDown && wasDown {
                Self.aloneDown[binding.keyCode] = false
                if Self.aloneUsedAsModifier[binding.keyCode] != true {
                    toggle()
                    didToggle = true
                }
            }
        }
        return didToggle
    }
}
