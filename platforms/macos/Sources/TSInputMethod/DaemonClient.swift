import Foundation

struct KeyOutcome {
    var consumed: Bool
    var preedit: String
    var commitText: String
    var candidates: [String]
}

final class DaemonClient {
    private var client: OpaquePointer?
    private(set) var sessionId: String = ""
    private(set) var chineseMode: Bool = true

    func ensureConnected() -> Bool {
        if client != nil, !sessionId.isEmpty {
            return true
        }
        let path = Self.resolveSocketPath()
        guard let created = tsime_client_create() else {
            return false
        }
        client = created
        guard tsime_client_connect(created, path) != 0 else {
            teardown()
            return false
        }
        guard tsime_client_hello(created, "macos-imkit") != 0 else {
            teardown()
            return false
        }
        var buffer = [CChar](repeating: 0, count: 128)
        let mode = Int32(chineseMode ? TSIME_MODE_CHINESE : TSIME_MODE_ENGLISH)
        guard tsime_client_open_session(created, mode, &buffer, buffer.count) != 0 else {
            teardown()
            return false
        }
        sessionId = String(cString: buffer)
        return true
    }

    func setChineseMode(_ enabled: Bool) {
        chineseMode = enabled
        guard ensureConnected(), let client else {
            return
        }
        let mode = Int32(enabled ? TSIME_MODE_CHINESE : TSIME_MODE_ENGLISH)
        _ = tsime_client_set_mode(client, sessionId, mode)
    }

    func handleKey(keyCode: UInt32, characters: String, keyDown: Bool) -> KeyOutcome? {
        guard ensureConnected(), let client else {
            return nil
        }
        var result = tsime_client_handle_key(client, sessionId, keyCode, 0, characters, keyDown ? 1 : 0, 0, 0)
        defer { tsime_key_result_free(&result) }
        return Self.outcome(from: result)
    }

    func selectCandidate(index: UInt32) -> KeyOutcome? {
        guard ensureConnected(), let client else {
            return nil
        }
        var result = tsime_client_select_candidate(client, sessionId, 0, index)
        defer { tsime_key_result_free(&result) }
        return Self.outcome(from: result)
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
        var outcome = KeyOutcome(consumed: result.consumed != 0, preedit: "", commitText: "", candidates: [])
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
                }
            }
        }
        return outcome
    }
}
