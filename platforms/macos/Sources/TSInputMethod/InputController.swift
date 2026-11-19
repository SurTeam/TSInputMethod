import Cocoa
import InputMethodKit

@objc(TSInputController)
class TSInputController: IMKInputController {
    private let daemon = DaemonClient()
    private var markedText: String = ""
    private var candidates: [String] = []
    private var candidatesPanel: IMKCandidates?
    private var leftShiftDown = false
    private var leftShiftUsedAsModifier = false

    private let keyLeftShift: UInt16 = 56

    override init!(server: IMKServer!, delegate: Any!, client inputClient: Any!) {
        super.init(server: server, delegate: delegate, client: inputClient)
        let panel = IMKCandidates(server: server, panelType: kIMKSingleRowSteppingCandidatePanel)
        panel?.setSelectionKeys([
            NSNumber(value: 18),
            NSNumber(value: 19),
            NSNumber(value: 20),
            NSNumber(value: 21),
            NSNumber(value: 23),
            NSNumber(value: 22),
            NSNumber(value: 26),
            NSNumber(value: 28),
            NSNumber(value: 25),
        ])
        let attrs = NSMutableDictionary()
        attrs[IMKCandidatesSendServerKeyEventFirst] = NSNumber(value: true)
        attrs[IMKCandidatesOpacityAttributeName] = NSNumber(value: Float(1.0))
        attrs[NSAttributedString.Key.font] = NSFont.systemFont(ofSize: 14)
        panel?.setAttributes(attrs as? [AnyHashable: Any])
        candidatesPanel = panel
    }

    override func handle(_ event: NSEvent!, client sender: Any!) -> Bool {
        guard let event else {
            return false
        }
        if event.type == .flagsChanged {
            return handleFlagsChanged(event)
        }
        if event.type != .keyDown {
            return false
        }
        if leftShiftDown {
            leftShiftUsedAsModifier = true
        }
        if !daemon.chineseMode {
            return false
        }
        let chars = event.charactersIgnoringModifiers ?? ""
        let keyCode = UInt32(event.keyCode)
        guard let outcome = daemon.handleKey(keyCode: keyCode, characters: chars, keyDown: true) else {
            return false
        }
        apply(outcome: outcome, client: sender)
        return outcome.consumed
    }

    override func candidates(_ sender: Any!) -> [Any]! {
        return candidates
    }

    override func candidateSelected(_ candidateString: NSAttributedString!) {
        guard let text = candidateString?.string,
              let index = candidates.firstIndex(of: text),
              let outcome = daemon.selectCandidate(index: UInt32(index)) else {
            return
        }
        apply(outcome: outcome, client: client())
    }

    override func menu() -> NSMenu! {
        let menu = NSMenu(title: "TSInputMethod")
        let chinese = NSMenuItem(title: "Chinese", action: #selector(enableChinese), keyEquivalent: "")
        chinese.target = self
        chinese.state = daemon.chineseMode ? .on : .off
        let english = NSMenuItem(title: "English", action: #selector(enableEnglish), keyEquivalent: "")
        english.target = self
        english.state = daemon.chineseMode ? .off : .on
        menu.addItem(chinese)
        menu.addItem(english)
        return menu
    }

    @objc private func enableChinese() {
        commitCompositionBeforeModeSwitch(client: client())
        daemon.setChineseMode(true)
    }

    @objc private func enableEnglish() {
        commitCompositionBeforeModeSwitch(client: client())
        daemon.setChineseMode(false)
        clearMarked(client: client())
    }

    private func handleFlagsChanged(_ event: NSEvent) -> Bool {
        let shiftDown = event.modifierFlags.contains(.shift)
        if event.keyCode != keyLeftShift {
            if leftShiftDown && shiftDown {
                leftShiftUsedAsModifier = true
            }
            return false
        }
        if shiftDown && !leftShiftDown {
            leftShiftDown = true
            leftShiftUsedAsModifier = false
            return false
        }
        if !shiftDown && leftShiftDown {
            leftShiftDown = false
            if !leftShiftUsedAsModifier {
                toggleChineseEnglish()
                return true
            }
        }
        return false
    }

    private func toggleChineseEnglish() {
        if daemon.chineseMode {
            enableEnglish()
        } else {
            enableChinese()
        }
    }

    private func commitCompositionBeforeModeSwitch(client: Any?) {
        if markedText.isEmpty && candidates.isEmpty {
            return
        }
        guard let outcome = daemon.handleKey(keyCode: 36, characters: "\r", keyDown: true) else {
            if !markedText.isEmpty {
                insert(text: markedText, client: client)
            }
            clearMarked(client: client)
            return
        }
        apply(outcome: outcome, client: client)
    }

    private func apply(outcome: KeyOutcome, client: Any?) {
        if !outcome.commitText.isEmpty {
            insert(text: outcome.commitText, client: client)
            markedText = ""
            candidates = []
        }
        markedText = outcome.preedit
        setMarked(text: outcome.preedit, client: client)
        candidates = outcome.candidates
        updateCandidatesPanel()
    }

    private func updateCandidatesPanel() {
        guard let panel = candidatesPanel else {
            return
        }
        if candidates.isEmpty {
            panel.hide()
            return
        }
        panel.update()
        panel.show(kIMKLocateCandidatesBelowHint)
    }

    private func insert(text: String, client: Any?) {
        guard let client = client as? IMKTextInput else {
            return
        }
        client.insertText(text, replacementRange: NSRange(location: NSNotFound, length: NSNotFound))
    }

    private func setMarked(text: String, client: Any?) {
        guard let client = client as? IMKTextInput else {
            return
        }
        let attrs = mark(forStyle: kTSMHiliteSelectedConvertedText, at: NSRange(location: 0, length: text.utf16.count))
        let attributed = NSAttributedString(string: text, attributes: attrs as? [NSAttributedString.Key: Any])
        client.setMarkedText(attributed,
                             selectionRange: NSRange(location: text.utf16.count, length: 0),
                             replacementRange: NSRange(location: NSNotFound, length: NSNotFound))
    }

    private func clearMarked(client: Any?) {
        setMarked(text: "", client: client)
        markedText = ""
        candidates = []
        candidatesPanel?.hide()
    }
}
