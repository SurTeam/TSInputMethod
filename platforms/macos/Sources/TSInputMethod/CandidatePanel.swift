import Cocoa

final class CandidatePanel {
    var onPage: ((Int) -> Void)?
    var onSelect: ((Int) -> Void)?

    private let panel: NSPanel
    private let content: CandidatePanelView
    private var pageSize = 9
    private let minColumnWidth: CGFloat = 28
    private let horizontalPadding: CGFloat = 6
    private let verticalPadding: CGFloat = 4
    private let columnSpacing: CGFloat = 4
    private let cellHorizontalPadding: CGFloat = 4
    private let labelBodyGap: CGFloat = 3
    private let rowHeight: CGFloat = 28
    private var bodyPointSize: CGFloat = 16
    private var bodyFontName: String = NSFont.systemFont(ofSize: 16).fontName
    private var highlightColor: NSColor = .controlAccentColor

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
        applyAppearance(ConfigStore.load())
    }

    func applyAppearance(_ settings: TSIMESettings) {
        pageSize = settings.clampedColumns
        bodyPointSize = CGFloat(max(10, min(36, settings.candidateFontSize)))
        bodyFontName = settings.candidateFontName
        highlightColor = settings.highlightColor
        content.highlightColor = highlightColor
        content.bodyFontName = bodyFontName
    }

    func hide() {
        panel.orderOut(nil)
        content.update(
            candidates: [],
            selectedIndex: 0,
            columnWidths: [],
            horizontalPadding: horizontalPadding,
            verticalPadding: verticalPadding,
            columnSpacing: columnSpacing,
            cellHorizontalPadding: cellHorizontalPadding,
            labelBodyGap: labelBodyGap,
            bodyPointSize: bodyPointSize
        )
    }

    func update(candidates: [String], selectedIndex: Int, anchor: NSRect) {
        let texts = candidates.map(Self.bareText)
        if texts.isEmpty {
            hide()
            return
        }
        let visible = Array(texts.prefix(pageSize))
        let columnWidths = visible.map(measureColumnWidth)
        let columnsWidth = columnWidths.reduce(CGFloat(0), +)
        let spacingWidth = CGFloat(max(visible.count - 1, 0)) * columnSpacing
        let width = horizontalPadding * 2 + columnsWidth + spacingWidth
        let height = verticalPadding * 2 + rowHeight
        content.frame = NSRect(x: 0, y: 0, width: width, height: height)
        content.update(
            candidates: visible,
            selectedIndex: selectedIndex,
            columnWidths: columnWidths,
            horizontalPadding: horizontalPadding,
            verticalPadding: verticalPadding,
            columnSpacing: columnSpacing,
            cellHorizontalPadding: cellHorizontalPadding,
            labelBodyGap: labelBodyGap,
            bodyPointSize: bodyPointSize
        )
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

    private func measureColumnWidth(_ text: String) -> CGFloat {
        let labelFont = NSFont.monospacedDigitSystemFont(ofSize: 13, weight: .regular)
        let textFont = font(ofSize: bodyPointSize, weight: .regular)
        let selectedTextFont = font(ofSize: bodyPointSize, weight: .semibold)
        let labelWidth = ("1" as NSString).size(withAttributes: [.font: labelFont]).width
        let body = text as NSString
        let bodyWidth = max(
            body.size(withAttributes: [.font: textFont]).width,
            body.size(withAttributes: [.font: selectedTextFont]).width
        )
        return max(minColumnWidth, ceil(cellHorizontalPadding * 2 + labelWidth + labelBodyGap + bodyWidth))
    }

    private func font(ofSize size: CGFloat, weight: NSFont.Weight) -> NSFont {
        if let named = NSFont(name: bodyFontName, size: size) {
            if weight == .semibold {
                return NSFontManager.shared.convert(named, toHaveTrait: .boldFontMask)
            }
            return named
        }
        return NSFont.systemFont(ofSize: size, weight: weight)
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
    var highlightColor: NSColor = .controlAccentColor
    var bodyFontName: String = NSFont.systemFont(ofSize: 16).fontName

    private var candidates: [String] = []
    private var selectedIndex = 0
    private var columnWidths: [CGFloat] = []
    private var horizontalPadding: CGFloat = 6
    private var verticalPadding: CGFloat = 4
    private var columnSpacing: CGFloat = 4
    private var cellHorizontalPadding: CGFloat = 4
    private var labelBodyGap: CGFloat = 3
    private var bodyPointSize: CGFloat = 16

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

    func update(
        candidates: [String],
        selectedIndex: Int,
        columnWidths: [CGFloat],
        horizontalPadding: CGFloat,
        verticalPadding: CGFloat,
        columnSpacing: CGFloat,
        cellHorizontalPadding: CGFloat,
        labelBodyGap: CGFloat,
        bodyPointSize: CGFloat
    ) {
        self.candidates = candidates
        self.selectedIndex = selectedIndex
        self.columnWidths = columnWidths
        self.horizontalPadding = horizontalPadding
        self.verticalPadding = verticalPadding
        self.columnSpacing = columnSpacing
        self.cellHorizontalPadding = cellHorizontalPadding
        self.labelBodyGap = labelBodyGap
        self.bodyPointSize = bodyPointSize
        needsDisplay = true
    }

    private func cellFrame(at slot: Int) -> NSRect? {
        guard slot >= 0, slot < columnWidths.count else {
            return nil
        }
        var x = horizontalPadding
        if slot > 0 {
            x += columnWidths.prefix(slot).reduce(CGFloat(0), +)
            x += CGFloat(slot) * columnSpacing
        }
        return NSRect(
            x: x,
            y: verticalPadding,
            width: columnWidths[slot],
            height: bounds.height - verticalPadding * 2
        )
    }

    private func bodyFont(weight: NSFont.Weight) -> NSFont {
        if let named = NSFont(name: bodyFontName, size: bodyPointSize) {
            if weight == .semibold {
                return NSFontManager.shared.convert(named, toHaveTrait: .boldFontMask)
            }
            return named
        }
        return NSFont.systemFont(ofSize: bodyPointSize, weight: weight)
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

        let labelFont = NSFont.monospacedDigitSystemFont(ofSize: 13, weight: .regular)
        let textFont = bodyFont(weight: .regular)
        let selectedTextFont = bodyFont(weight: .semibold)
        let labelColor = NSColor.secondaryLabelColor.withAlphaComponent(0.72)
        let textColor = NSColor.labelColor
        let selectedLabelColor = NSColor.white.withAlphaComponent(0.85)
        let selectedTextColor = NSColor.white

        for slot in 0..<candidates.count {
            guard let cell = cellFrame(at: slot) else {
                continue
            }
            let selected = slot == selectedIndex
            if selected {
                highlightColor.setFill()
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
            let labelY = cell.minY + (cell.height - labelSize.height) / 2
            let bodyY = cell.minY + (cell.height - bodySize.height) / 2
            let startX = cell.minX + cellHorizontalPadding
            label.draw(at: NSPoint(x: startX, y: labelY), withAttributes: labelAttrs)
            body.draw(at: NSPoint(x: startX + labelSize.width + labelBodyGap, y: bodyY), withAttributes: bodyAttrs)
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
        for slot in 0..<candidates.count {
            if let cell = cellFrame(at: slot), cell.contains(point) {
                onSelect?(slot)
                return
            }
        }
    }

    override func acceptsFirstMouse(for event: NSEvent?) -> Bool {
        true
    }
}
