import Cocoa
import InputMethodKit

@objc(TSInputController)
class TSInputController: IMKInputController {
    private let daemon = DaemonClient()
    private let candidatePanel = CandidatePanel()
    private var markedText: String = ""
    private var candidates: [String] = []
    private var comments: [String] = []
    private var selectedIndex: Int = 0
    private var leftShiftDown = false
    private var leftShiftUsedAsModifier = false

    private let keyLeftShift: UInt16 = 56
    private let keyUpArrow: UInt32 = 126
    private let keyDownArrow: UInt32 = 125

    private let modShift: UInt32 = 1 << 17
    private let modControl: UInt32 = 1 << 18
    private let modOption: UInt32 = 1 << 19
    private let modCommand: UInt32 = 1 << 20

    override init!(server: IMKServer!, delegate: Any!, client inputClient: Any!) {
        super.init(server: server, delegate: delegate, client: inputClient)
        candidatePanel.onPage = { [weak self] delta in
            self?.pageCandidates(delta: delta)
        }
        candidatePanel.onSelect = { [weak self] index in
            self?.selectDisplayedCandidate(index: index)
        }
    }

    override func recognizedEvents(_ sender: Any!) -> Int {
        let events: NSEvent.EventTypeMask = [.keyDown, .flagsChanged]
        return Int(events.rawValue)
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
        let modifiers = wireModifiers(from: event.modifierFlags)
        if modifiers & (modControl | modOption | modCommand) != 0 {
            return false
        }
        let chars = event.characters ?? ""
        let keyCode = UInt32(event.keyCode)
        guard let outcome = daemon.handleKey(keyCode: keyCode, modifiers: modifiers, characters: chars, keyDown: true) else {
            return false
        }
        apply(outcome: outcome, client: sender)
        return outcome.consumed
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
        daemon.setChineseMode(true)
    }

    @objc private func enableEnglish() {
        commitCompositionLocally(client: client())
        daemon.setChineseMode(false)
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

    private func commitCompositionLocally(client: Any?) {
        if !markedText.isEmpty {
            insert(text: markedText, client: client)
        }
        clearMarked(client: client)
    }

    private func apply(outcome: KeyOutcome, client: Any?) {
        if !outcome.commitText.isEmpty {
            insert(text: outcome.commitText, client: client)
            markedText = ""
            candidates = []
            comments = []
            selectedIndex = 0
        }
        markedText = outcome.preedit
        setMarked(text: outcome.preedit, client: client)
        candidates = outcome.candidates
        comments = outcome.comments
        selectedIndex = Int(outcome.selectedIndex)
        updateCandidatePanel(client: client)
    }

    private func updateCandidatePanel(client: Any?) {
        let texts = candidates.map(Self.bareCandidateText)
        if texts.isEmpty {
            candidatePanel.hide()
            return
        }
        var index = selectedIndex
        if index < 0 {
            index = 0
        }
        if index >= texts.count {
            index = texts.count - 1
        }
        selectedIndex = index
        candidatePanel.update(candidates: texts, selectedIndex: index, anchor: caretRect(client: client))
    }

    private func pageCandidates(delta: Int) {
        let keyCode = delta < 0 ? keyUpArrow : keyDownArrow
        guard let outcome = daemon.handleKey(keyCode: keyCode, modifiers: 0, characters: "", keyDown: true) else {
            return
        }
        apply(outcome: outcome, client: client())
    }

    private func selectDisplayedCandidate(index: Int) {
        guard index >= 0, index < candidates.count,
              let outcome = daemon.selectCandidate(index: UInt32(index)) else {
            return
        }
        apply(outcome: outcome, client: client())
    }

    private func caretRect(client: Any?) -> NSRect {
        guard let client = client as? IMKTextInput else {
            return NSRect(x: 80, y: 80, width: 1, height: 16)
        }
        var rect = NSRect.zero
        _ = client.attributes(forCharacterIndex: 0, lineHeightRectangle: &rect)
        if rect.width <= 0 && rect.height <= 0 {
            return NSRect(x: 80, y: 80, width: 1, height: 16)
        }
        return rect
    }

    private func wireModifiers(from flags: NSEvent.ModifierFlags) -> UInt32 {
        var modifiers: UInt32 = 0
        if flags.contains(.shift) {
            modifiers |= modShift
        }
        if flags.contains(.control) {
            modifiers |= modControl
        }
        if flags.contains(.option) {
            modifiers |= modOption
        }
        if flags.contains(.command) {
            modifiers |= modCommand
        }
        return modifiers
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
        if text.isEmpty {
            client.setMarkedText("",
                                 selectionRange: NSRange(location: 0, length: 0),
                                 replacementRange: NSRange(location: NSNotFound, length: NSNotFound))
            return
        }
        let range = NSRange(location: 0, length: text.utf16.count)
        var attrs = mark(forStyle: kTSMHiliteSelectedRawText, at: range) as? [NSAttributedString.Key: Any] ?? [:]
        attrs[.underlineStyle] = NSUnderlineStyle.single.rawValue
        attrs[.underlineColor] = NSColor.controlAccentColor
        attrs[.foregroundColor] = NSColor.labelColor
        attrs[.backgroundColor] = NSColor.controlAccentColor.withAlphaComponent(0.12)
        let attributed = NSAttributedString(string: text, attributes: attrs)
        client.setMarkedText(attributed,
                             selectionRange: NSRange(location: text.utf16.count, length: 0),
                             replacementRange: NSRange(location: NSNotFound, length: NSNotFound))
    }

    private func clearMarked(client: Any?) {
        setMarked(text: "", client: client)
        markedText = ""
        candidates = []
        comments = []
        selectedIndex = 0
        candidatePanel.hide()
    }

    private static func bareCandidateText(_ text: String) -> String {
        guard text.utf8.count >= 2 else {
            return text
        }
        let first = text.utf8[text.utf8.startIndex]
        let second = text.utf8[text.utf8.index(after: text.utf8.startIndex)]
        if first >= 0x31, first <= 0x39, second == 0x2e {
            return String(text.dropFirst(2))
        }
        return text
    }
}
