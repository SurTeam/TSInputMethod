import Foundation

struct KeyOutcome {
    var consumed: Bool
    var preedit: String
    var commitText: String
    var candidates: [String]
    var comments: [String]
    var selectedIndex: UInt32
}

final class DaemonClient {
    private static var sharedChineseMode: Bool = true
    private static var lastModeToggleAt: Date = .distantPast
    private static let modeToggleDebounce: TimeInterval = 0.35

    private var client: OpaquePointer?
    private(set) var sessionId: String = ""

    var chineseMode: Bool {
        Self.sharedChineseMode
    }

    func consumeModeToggleOpportunity() -> Bool {
        let now = Date()
        if now.timeIntervalSince(Self.lastModeToggleAt) < Self.modeToggleDebounce {
            return false
        }
        Self.lastModeToggleAt = now
        return true
    }

    static func isReachable() -> Bool {
        let path = resolveSocketPath()
        guard let created = tsime_client_create() else {
            return false
        }
        defer {
            tsime_client_destroy(created)
        }
        guard tsime_client_connect(created, path) != 0 else {
            return false
        }
        defer {
            tsime_client_disconnect(created)
        }
        return tsime_client_hello(created, "macos-imkit-probe") != 0
    }

    func ensureConnected() -> Bool {
        if client != nil, !sessionId.isEmpty {
            return true
        }
        let path = Self.resolveSocketPath()
        guard let created = tsime_client_create() else {
            DaemonWatchdog.requestRecovery()
            return false
        }
        client = created
        guard tsime_client_connect(created, path) != 0 else {
            teardown()
            DaemonWatchdog.requestRecovery()
            return false
        }
        guard tsime_client_hello(created, "macos-imkit") != 0 else {
            teardown()
            DaemonWatchdog.requestRecovery()
            return false
        }
        var buffer = [CChar](repeating: 0, count: 128)
        let mode = Int32(chineseMode ? TSIME_MODE_CHINESE : TSIME_MODE_ENGLISH)
        guard tsime_client_open_session(created, mode, &buffer, buffer.count) != 0 else {
            teardown()
            DaemonWatchdog.requestRecovery()
            return false
        }
        sessionId = String(cString: buffer)
        return true
    }

func setChineseMode(_ enabled: Bool) {
        Self.sharedChineseMode = enabled
        guard let client, !sessionId.isEmpty else {
            return
        }
        let mode = Int32(enabled ? TSIME_MODE_CHINESE : TSIME_MODE_ENGLISH)
        _ = tsime_client_set_mode(client, sessionId, mode)
    }

    func invalidateLearningContext() {
        guard let client, !sessionId.isEmpty else {
            return
        }
        _ = tsime_client_invalidate_learning_context(client, sessionId)
    }

    func handleKey(keyCode: UInt32, modifiers: UInt32, characters: String, keyDown: Bool) -> KeyOutcome? {
        guard ensureConnected(), let client else {
            return nil
        }
        var result = tsime_client_handle_key(client, sessionId, keyCode, modifiers, characters, keyDown ? 1 : 0, 0, 0)
        defer { tsime_key_result_free(&result) }
        if let outcome = Self.outcome(from: result) {
            return outcome
        }
        teardown()
        DaemonWatchdog.requestRecovery()
        return nil
    }

    func selectCandidate(index: UInt32) -> KeyOutcome? {
        guard ensureConnected(), let client else {
            return nil
        }
        var result = tsime_client_select_candidate(client, sessionId, 0, index)
        defer { tsime_key_result_free(&result) }
        if let outcome = Self.outcome(from: result) {
            return outcome
        }
        teardown()
        DaemonWatchdog.requestRecovery()
        return nil
    }

    func teardown() {
        if let client {
            if !sessionId.isEmpty {
                _ = tsime_client_close_session(client, sessionId)
            }
            tsime_client_disconnect(client)
            tsime_client_destroy(client)
        }
        client = nil
        sessionId = ""
    }

    deinit {
        teardown()
    }

    static func resolveSocketPath() -> String {
        if let override = ProcessInfo.processInfo.environment["TSIME_SOCKET"], !override.isEmpty {
            return override
        }
        if let runtime = ProcessInfo.processInfo.environment["XDG_RUNTIME_DIR"], !runtime.isEmpty {
            return (runtime as NSString).appendingPathComponent("tsime.sock")
        }
        return NSTemporaryDirectory() + "tsime-\(getuid()).sock"
    }

    private static func outcome(from result: TsimeKeyResult) -> KeyOutcome? {
        guard result.ok != 0 else {
            return nil
        }
        var outcome = KeyOutcome(consumed: result.consumed != 0, preedit: "", commitText: "", candidates: [], comments: [], selectedIndex: result.selected_index)
        if let preedit = result.preedit {
            outcome.preedit = String(cString: preedit)
        }
        if let commit = result.commit_text {
            outcome.commitText = String(cString: commit)
        }
        if result.candidate_count > 0, let base = result.candidates {
            for i in 0..<Int(result.candidate_count) {
                let item = base.advanced(by: i).pointee
                if let text = item.text {
                    outcome.candidates.append(String(cString: text))
                } else {
                    outcome.candidates.append("")
                }
                if let comment = item.comment {
                    outcome.comments.append(String(cString: comment))
                } else {
                    outcome.comments.append("")
                }
            }
        }
        return outcome
    }
}
