import Foundation

enum DaemonWatchdog {
    private static let queue = DispatchQueue(label: "com.surteam.tsime.daemon-watchdog")
    private static var timer: DispatchSourceTimer?
    private static var recovering = false
    private static var consecutiveFailures = 0
    private static let probeInterval: TimeInterval = 5
    private static let softFailEscalateThreshold = 3
    private static let readyPollAttempts = 40
    private static let readyPollInterval: TimeInterval = 0.5

    static func start(withInitialEnsure: Bool) {
        queue.async {
            if withInitialEnsure {
                bootstrap()
            }
            guard timer == nil else {
                return
            }
            let source = DispatchSource.makeTimerSource(queue: queue)
            source.schedule(deadline: .now() + probeInterval, repeating: probeInterval)
            source.setEventHandler {
                tick()
            }
            source.resume()
            timer = source
        }
    }

    static func requestRecovery() {
        queue.async {
            recover(escalate: false)
        }
    }

    private static func bootstrap() {
        if DaemonClient.isReachable() {
            consecutiveFailures = 0
            DispatchQueue.main.async {
                AppNotifier.notifyReady()
            }
            return
        }
        recover(escalate: false)
    }

    private static func tick() {
        if recovering {
            return
        }
        if DaemonClient.isReachable() {
            consecutiveFailures = 0
            return
        }
        recover(escalate: false)
    }

    private static func recover(escalate: Bool) {
        if recovering {
            return
        }
        recovering = true
        let shouldForce = escalate || consecutiveFailures >= softFailEscalateThreshold
        DispatchQueue.main.async {
            AppNotifier.resetLifecycleNotifications()
            AppNotifier.notifyInitializing()
        }
        if shouldForce {
            DaemonAgent.forceRestart()
            consecutiveFailures = 0
        } else {
            DaemonAgent.softEnsureRunning()
            consecutiveFailures += 1
        }
        for _ in 0..<readyPollAttempts {
            if DaemonClient.isReachable() {
                consecutiveFailures = 0
                DispatchQueue.main.async {
                    AppNotifier.notifyReady()
                }
                recovering = false
                return
            }
            Thread.sleep(forTimeInterval: readyPollInterval)
        }
        recovering = false
    }
}
