import Foundation
import InputMethodKit

@main
struct TSInputMethodMain {
    static func main() {
        let connectionName = "TSInputMethod_1_Connection"
        _ = IMKServer(name: connectionName, bundleIdentifier: Bundle.main.bundleIdentifier)
        RunLoop.current.run()
    }
}
