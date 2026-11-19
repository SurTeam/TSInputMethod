import Cocoa

final class CandidatePanel {
    var onPage: ((Int) -> Void)?
    var onSelect: ((Int) -> Void)?

    private let panel: NSPanel
    private let content: CandidatePanelView
    private var lockedColumnWidth: CGFloat = 0
    private let pageSize = 9
private let minColumnWidth: CGFloat = 44
    private let horizontalPadding: CGFloat = 12
    private let verticalPadding: CGFloat = 8
    private let columnSpacing: CGFloat = 10
    private let rowHeight: CGFloat = 36

    init() {
        content = CandidatePanelView(frame: .zero)
        panel = NSPanel(
            contentRect: NSRect(x: 0, y: 0, width: 200, height: 40),
            styleMask: [.borderless, .nonactivatingPanel],
            backing: .buffered,
            defer: true
        )
        panel.level = .floating
        panel.isOpaque = false
        panel.backgroundColor = .clear
        panel.hasShadow = true
        panel.hidesOnDeactivate = false
        panel.becomesKeyOnlyIfNeeded = true
        panel.acceptsMouseMovedEvents = true
        panel.collectionBehavior = [.canJoinAllSpaces, .fullScreenAuxiliary]
        panel.contentView = content
        content.onPage = { [weak self] delta in
            self?.onPage?(delta)
        }
        content.onSelect = { [weak self] index in
            self?.onSelect?(index)
        }
    }

    func hide() {
        panel.orderOut(nil)
        content.update(candidates: [], selectedIndex: 0, columnWidth: minColumnWidth)
        lockedColumnWidth = 0
    }

    func update(candidates: [String], selectedIndex: Int, anchor: NSRect) {
        let texts = candidates.map(Self.bareText)
        if texts.isEmpty {
            hide()
            return
        }
        let count = min(texts.count, pageSize)
        let measured = measureColumnWidth(texts: Array(texts.prefix(count)))
        if lockedColumnWidth <= 0 {
            lockedColumnWidth = measured
        } else {
            lockedColumnWidth = max(lockedColumnWidth, measured)
        }
        let width = horizontalPadding * 2
            + CGFloat(pageSize) * lockedColumnWidth
            + CGFloat(pageSize - 1) * columnSpacing
        let height = verticalPadding * 2 + rowHeight
        content.frame = NSRect(x: 0, y: 0, width: width, height: height)
        content.pageSize = pageSize
        content.columnSpacing = columnSpacing
        content.horizontalPadding = horizontalPadding
        content.verticalPadding = verticalPadding
        content.update(candidates: Array(texts.prefix(count)), selectedIndex: selectedIndex, columnWidth: lockedColumnWidth)
        panel.setContentSize(NSSize(width: width, height: height))
        position(anchor: anchor, size: NSSize(width: width, height: height))
        panel.orderFront(nil)
    }

    private func position(anchor: NSRect, size: NSSize) {
        var screen = NSScreen.main?.visibleFrame ?? NSRect(x: 0, y: 0, width: 1280, height: 800)
        for candidate in NSScreen.screens {
            if candidate.frame.contains(NSPoint(x: anchor.midX, y: anchor.midY)) {
                screen = candidate.visibleFrame
                break
            }
        }
        var origin = NSPoint(x: anchor.minX, y: anchor.minY - size.height - 4)
        if origin.y < screen.minY {
            origin.y = anchor.maxY + 4
        }
        if origin.x + size.width > screen.maxX {
            origin.x = screen.maxX - size.width
        }
        if origin.x < screen.minX {
            origin.x = screen.minX
        }
        panel.setFrameOrigin(origin)
    }

    private func measureColumnWidth(texts: [String]) -> CGFloat {
let bodyPointSize: CGFloat = 16
        let labelFont = NSFont.monospacedDigitSystemFont(ofSize: 13, weight: .regular)
        let textFont = NSFont.systemFont(ofSize: bodyPointSize)
        let selectedTextFont = NSFont.systemFont(ofSize: bodyPointSize, weight: .semibold)
        var maxWidth: CGFloat = minColumnWidth
        for (index, text) in texts.enumerated() {
            let label = "\(index + 1)" as NSString
            let body = text as NSString
            let labelWidth = label.size(withAttributes: [.font: labelFont]).width
            let bodyWidth = max(
                body.size(withAttributes: [.font: textFont]).width,
                body.size(withAttributes: [.font: selectedTextFont]).width
            )
            maxWidth = max(maxWidth, ceil(labelWidth + 4 + bodyWidth + 12))
        }
        return maxWidth
    }

    private static func bareText(_ text: String) -> String {
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

private final class CandidatePanelView: NSView {
    var onPage: ((Int) -> Void)?
    var onSelect: ((Int) -> Void)?
    var pageSize = 9
    var columnSpacing: CGFloat = 8
    var horizontalPadding: CGFloat = 10
    var verticalPadding: CGFloat = 6

    private var candidates: [String] = []
    private var selectedIndex = 0
    private var columnWidth: CGFloat = 36

    override var isFlipped: Bool { true }
    override var acceptsFirstResponder: Bool { true }

    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)
        wantsLayer = true
        layer?.cornerRadius = 8
        layer?.masksToBounds = true
    }

    required init?(coder: NSCoder) {
        nil
    }

    func update(candidates: [String], selectedIndex: Int, columnWidth: CGFloat) {
        self.candidates = candidates
        self.selectedIndex = selectedIndex
        self.columnWidth = columnWidth
        needsDisplay = true
    }

    override func draw(_ dirtyRect: NSRect) {
        let bounds = self.bounds
let bg = NSColor.controlBackgroundColor.withAlphaComponent(0.96)
        bg.setFill()
        NSBezierPath(roundedRect: bounds, xRadius: 8, yRadius: 8).fill()
        let edge = bg.blended(withFraction: 0.14, of: NSColor.labelColor) ?? NSColor.labelColor.withAlphaComponent(0.12)
        edge.setStroke()
        let border = NSBezierPath(roundedRect: bounds.insetBy(dx: 0.5, dy: 0.5), xRadius: 8, yRadius: 8)
        border.lineWidth = 1
        border.stroke()

let bodyPointSize: CGFloat = 16
        let labelFont = NSFont.monospacedDigitSystemFont(ofSize: 13, weight: .regular)
        let textFont = NSFont.systemFont(ofSize: bodyPointSize)
        let selectedTextFont = NSFont.systemFont(ofSize: bodyPointSize, weight: .semibold)
        let labelColor = NSColor.secondaryLabelColor.withAlphaComponent(0.72)
        let textColor = NSColor.labelColor
        let selectedLabelColor = NSColor.white.withAlphaComponent(0.85)
        let selectedTextColor = NSColor.white

        for slot in 0..<pageSize {
            let x = horizontalPadding + CGFloat(slot) * (columnWidth + columnSpacing)
            let cell = NSRect(
                x: x,
                y: verticalPadding,
                width: columnWidth,
                height: bounds.height - verticalPadding * 2
            )
            guard slot < candidates.count else {
                continue
            }
            let selected = slot == selectedIndex
            if selected {
                NSColor.controlAccentColor.setFill()
                NSBezierPath(roundedRect: cell, xRadius: 5, yRadius: 5).fill()
            }
            let label = "\(slot + 1)" as NSString
            let body = candidates[slot] as NSString
            let labelAttrs: [NSAttributedString.Key: Any] = [
                .font: labelFont,
                .foregroundColor: selected ? selectedLabelColor : labelColor,
            ]
            let bodyAttrs: [NSAttributedString.Key: Any] = [
                .font: selected ? selectedTextFont : textFont,
                .foregroundColor: selected ? selectedTextColor : textColor,
            ]
            let labelSize = label.size(withAttributes: labelAttrs)
            let bodySize = body.size(withAttributes: bodyAttrs)
            let total = labelSize.width + 4 + bodySize.width
            let startX = cell.minX + max(6, (cell.width - total) / 2)
            let labelY = cell.minY + (cell.height - labelSize.height) / 2
            let bodyY = cell.minY + (cell.height - bodySize.height) / 2
            label.draw(at: NSPoint(x: startX, y: labelY), withAttributes: labelAttrs)
            body.draw(at: NSPoint(x: startX + labelSize.width + 4, y: bodyY), withAttributes: bodyAttrs)
        }
    }

    override func scrollWheel(with event: NSEvent) {
        let delta = event.scrollingDeltaY != 0 ? event.scrollingDeltaY : event.deltaY
        if abs(delta) < 0.5 {
            return
        }
        onPage?(delta < 0 ? 1 : -1)
    }

    override func mouseDown(with event: NSEvent) {
        let point = convert(event.locationInWindow, from: nil)
        for slot in 0..<min(candidates.count, pageSize) {
            let x = horizontalPadding + CGFloat(slot) * (columnWidth + columnSpacing)
            let cell = NSRect(
                x: x,
                y: verticalPadding,
                width: columnWidth,
                height: bounds.height - verticalPadding * 2
            )
            if cell.contains(point) {
                onSelect?(slot)
                return
            }
        }
    }

    override func acceptsFirstMouse(for event: NSEvent?) -> Bool {
        true
    }
}
