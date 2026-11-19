import Cocoa
import InputMethodKit

@objc(TSInputController)
class TSInputController: IMKInputController {
    private let daemon = DaemonClient()
    private var markedText: String = ""
    private var candidates: [String] = []

    override func handle(_ event: NSEvent!, client sender: Any!) -> Bool {
        guard let event else {
            return false
        }
        if event.type != .keyDown {
            return false
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
        daemon.setChineseMode(false)
        clearMarked(client: client())
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
    }
}
