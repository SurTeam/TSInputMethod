import Cocoa
import InputMethodKit

@objc(TSInputController)
class TSInputController: IMKInputController {
    private let daemon = DaemonClient()
    private var markedText: String = ""
    private var candidates: [String] = []
    private var comments: [String] = []
    private var selectedIndex: Int = 0
    private var visibleCandidates: [String] = []
    private var syncingCandidateHighlight = false
    private var candidatesPanel: IMKCandidates?
    private var leftShiftDown = false
    private var leftShiftUsedAsModifier = false

    private let keyLeftShift: UInt16 = 56

    private let modShift: UInt32 = 1 << 17
    private let modControl: UInt32 = 1 << 18
    private let modOption: UInt32 = 1 << 19
    private let modCommand: UInt32 = 1 << 20

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
        attrs[NSAttributedString.Key.font] = NSFont.systemFont(ofSize: 12)
        panel?.setAttributes(attrs as? [AnyHashable: Any])
        candidatesPanel = panel
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

    override func candidates(_ sender: Any!) -> [Any]! {
        return attributedCandidates()
    }

    override func candidateSelected(_ candidateString: NSAttributedString!) {
        guard let text = candidateString?.string,
              let index = indexOfDisplayedCandidate(text),
              let outcome = daemon.selectCandidate(index: UInt32(index)) else {
            return
        }
        apply(outcome: outcome, client: client())
    }

    override func candidateSelectionChanged(_ candidateString: NSAttributedString!) {
        if syncingCandidateHighlight {
            return
        }
        guard let text = candidateString?.string,
              let index = indexOfDisplayedCandidate(text) else {
            return
        }
        selectedIndex = index
    }

    private func indexOfDisplayedCandidate(_ text: String) -> Int? {
        let bare: String
        if text.hasPrefix("• ") {
            bare = String(text.dropFirst(2))
        } else if text.hasPrefix("  ") {
            bare = String(text.dropFirst(2))
        } else {
            bare = text
        }
        return bareCandidates().firstIndex(of: bare)
    }

    private func bareCandidates() -> [String] {
        candidates.map(Self.bareCandidateText)
    }

    private func attributedCandidates() -> [NSAttributedString] {
        let texts = bareCandidates()
        guard !texts.isEmpty else {
            return []
        }
        var index = selectedIndex
        if index < 0 {
            index = 0
        }
        if index >= texts.count {
            index = texts.count - 1
        }
        let normalFont = NSFont.systemFont(ofSize: 12)
        let selectedFont = NSFont.boldSystemFont(ofSize: 12)
        let normalAttrs: [NSAttributedString.Key: Any] = [
            .font: normalFont,
            .foregroundColor: NSColor.labelColor,
        ]
        let selectedAttrs: [NSAttributedString.Key: Any] = [
            .font: selectedFont,
            .foregroundColor: NSColor.controlAccentColor,
            .underlineStyle: NSUnderlineStyle.single.rawValue,
        ]
        return texts.enumerated().map { offset, text in
            let display = offset == index ? "• \(text)" : "  \(text)"
            return NSAttributedString(string: display, attributes: offset == index ? selectedAttrs : normalAttrs)
        }
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
            visibleCandidates = []
        }
        markedText = outcome.preedit
        setMarked(text: outcome.preedit, client: client)
        candidates = outcome.candidates
        comments = outcome.comments
        selectedIndex = Int(outcome.selectedIndex)
        updateCandidatesPanel()
    }

    private func updateCandidatesPanel() {
        guard let panel = candidatesPanel else {
            return
        }
        if candidates.isEmpty {
            visibleCandidates = []
            panel.hide()
            return
        }
        let shown = bareCandidates()
        let attributed = attributedCandidates()
        visibleCandidates = shown
        syncingCandidateHighlight = true
        panel.setCandidateData(attributed)
        panel.update()
        panel.show(kIMKLocateCandidatesBelowHint)
        DispatchQueue.main.async { [weak self] in
            self?.syncingCandidateHighlight = false
        }
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
        let attrs = mark(forStyle: kTSMHiliteRawText, at: NSRange(location: 0, length: text.utf16.count))
        let attributed = NSAttributedString(string: text, attributes: attrs as? [NSAttributedString.Key: Any])
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
        visibleCandidates = []
        candidatesPanel?.hide()
    }
}
