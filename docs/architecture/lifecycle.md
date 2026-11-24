# Daemon, session, and adapter lifecycle

**Status:** designed / documented — macOS LaunchAgent early-start and soft-ensure are implemented in source; host IME real-env login/switch cycle not claimed here.

## Daemon

1. **Start** — load config, open dictionaries, bind UDS or Named Pipe, accept connections.
2. **Run** — multiplex client connections; maintain session table; serve IPC.
3. **Reload** — optional config/dict reload without tearing down all sessions when safe.
4. **Stop** — reject new accepts, notify clients, drain or abort sessions, close transport.

`tsimed` (`cmd/tsimed/main.go`) implements this as follows:

- **Run:** every 30 s, `PinyinEngine.SyncPersonalData` reloads `user_dictionary.json` if its mtime/size changed (for example after a `tsime-userdict` edit), then saves the personal model and user dictionary if they have unsaved changes. Learned usage and SGD updates therefore survive a crash or a `KeepAlive` relaunch, losing at most one interval.
- **Stop (SIGINT/SIGTERM or Serve error):** `ipc.Server.Close` closes the listener and every tracked connection (an idle adapter cannot block shutdown), waits for the connection goroutines, and closes each connection's sessions. `PinyinEngine.Close` then drains the SGD `UpdateQueue` and saves the model and user dictionary if they have unsaved changes.

On macOS, `tsimed` is owned by LaunchAgent `com.surteam.tsimed` with `RunAtLoad` and `KeepAlive`, so dictionary/engine warm happens at login/session load rather than at first keystroke. Switching away from the IMKit adapter must leave this process alive and already initialized.

## Adapter

1. **Load** — host framework loads the adapter (IMKit server, TSF service, Fcitx5 addon).
2. **Attach** — open IPC to daemon; identify protocol version.
3. **Serve** — map host callbacks to sessions and IPC.
4. **Detach / unload** — close sessions, disconnect, release host resources.

If the daemon is down at attach time, the adapter should fail soft (disable input method UI) and may retry connect with backoff.

macOS specifics:

- Soft ensure (`bootstrap`/`enable` when unloaded, then `kickstart` without `-k`) plus hello-only reachability probing run from IMKit `DaemonWatchdog` as a continuous safety net and notification path (not keystroke-gated).
- After repeated soft failures the watchdog escalates to force restart; Chinese-mode IPC connect/handle failures also request recovery.
- Force restart (`kickstart -k`) is also used for menu **Restart** recovery, which shows initializing immediately and ready once IPC returns before terminating IMKit.
- Deactivate / switch-away tears down only the adapter client/session; it does not stop the LaunchAgent-managed daemon.

## Session

A session is the daemon-side state for one input context (typically one text field / client ID from the host).

1. **Create** — after adapter Connect, on first context activation or explicit CreateSession.
2. **Active** — receives KeyEvents; holds compose buffer and candidate page state.
3. **Idle** — context deactivated; may retain short-lived compose state per policy.
4. **Destroy** — context closed, client disconnect, or daemon shutdown.

Sessions must not leak across unrelated clients. Adapter process death destroys its sessions: the IPC server records the sessions opened on each connection and closes any still open when that connection ends (EOF, read error, or server shutdown).

## Ordering constraints

- No KeyEvent before CreateSession (or implicit create documented in ipc module).
- Commit and Clear reset compose state before the next keystroke is applied.
- Protocol version negotiation happens at Connect; mismatched versions disconnect cleanly.
