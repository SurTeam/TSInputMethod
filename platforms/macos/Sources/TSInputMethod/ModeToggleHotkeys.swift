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
    private var bindings: [ModeToggleBinding]
    private var aloneDown: [UInt16: Bool] = [:]
    private var aloneUsedAsModifier: [UInt16: Bool] = [:]

    init(bindings: [ModeToggleBinding] = ModeToggleBinding.defaults) {
        self.bindings = bindings
    }

    func currentBindings() -> [ModeToggleBinding] {
        bindings
    }

    func setBindings(_ bindings: [ModeToggleBinding]) {
        self.bindings = bindings
        aloneDown.removeAll(keepingCapacity: true)
        aloneUsedAsModifier.removeAll(keepingCapacity: true)
    }

    func handleKeyDown(_ event: NSEvent, toggle: () -> Void) -> Bool {
        for (code, isDown) in aloneDown where isDown {
            aloneUsedAsModifier[code] = true
        }
        for binding in bindings where binding.enabled {
            guard case .keyDown = binding.trigger else {
                continue
            }
            guard event.keyCode == binding.keyCode else {
                continue
            }
            let chord = event.modifierFlags.intersection([.shift, .control, .option, .command])
            guard chord.isEmpty else {
                return false
            }
            toggle()
            return true
        }
        return false
    }

    func handleFlagsChanged(_ event: NSEvent, toggle: () -> Void) -> Bool {
        var didToggle = false
        for binding in bindings where binding.enabled {
            guard case .modifierAlone(let flag) = binding.trigger else {
                continue
            }
            let flagDown = event.modifierFlags.contains(flag)
            if event.keyCode != binding.keyCode {
                if aloneDown[binding.keyCode] == true && flagDown {
                    aloneUsedAsModifier[binding.keyCode] = true
                }
                continue
            }
            let wasDown = aloneDown[binding.keyCode] == true
            if flagDown && !wasDown {
                aloneDown[binding.keyCode] = true
                aloneUsedAsModifier[binding.keyCode] = false
                continue
            }
            if !flagDown && wasDown {
                aloneDown[binding.keyCode] = false
                if aloneUsedAsModifier[binding.keyCode] != true {
                    toggle()
                    didToggle = true
                }
            }
        }
        return didToggle
    }
}
