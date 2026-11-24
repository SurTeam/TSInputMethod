import Cocoa

final class SettingsWindowController: NSWindowController, NSTableViewDataSource, NSTableViewDelegate, NSWindowDelegate {
    static let shared = SettingsWindowController()

    private var draft = ConfigStore.load()
    private var recordingRow: Int?
    private var localMonitor: Any?

    private let bindingsTable = NSTableView()
    private let columnsStepper = NSStepper()
    private let columnsField = NSTextField(string: "9")
    private let fontNameField = NSTextField(string: "")
    private let fontSizeField = NSTextField(string: "16")
    private let colorWell = NSColorWell()
    private let recordHint = NSTextField(labelWithString: "")
    private let scrollView = NSScrollView()

    private init() {
        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 560, height: 420),
            styleMask: [.titled, .closable, .miniaturizable],
            backing: .buffered,
            defer: false
        )
        window.title = "TSInputMethod Settings"
        window.isReleasedWhenClosed = false
        super.init(window: window)
        window.delegate = self
        window.contentView = buildContent()
        window.center()
    }

    required init?(coder: NSCoder) {
        nil
    }

    func showSettings() {
        draft = ConfigStore.load()
        reloadUI()
        showWindow(nil)
        NSApp.activate(ignoringOtherApps: true)
        window?.makeKeyAndOrderFront(nil)
    }

    private func buildContent() -> NSView {
        let root = NSView(frame: NSRect(x: 0, y: 0, width: 560, height: 420))

        let hotkeyLabel = NSTextField(labelWithString: "CN/EN toggle hotkeys")
        hotkeyLabel.font = NSFont.boldSystemFont(ofSize: 13)
        hotkeyLabel.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(hotkeyLabel)

        bindingsTable.headerView = nil
        bindingsTable.allowsMultipleSelection = false
        bindingsTable.allowsEmptySelection = true
        bindingsTable.rowHeight = 24
        bindingsTable.dataSource = self
        bindingsTable.delegate = self
        let column = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("binding"))
        column.width = 500
        bindingsTable.addTableColumn(column)

        scrollView.documentView = bindingsTable
        scrollView.hasVerticalScroller = true
        scrollView.borderType = .bezelBorder
        scrollView.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(scrollView)

        let addButton = NSButton(title: "Add / Record", target: self, action: #selector(addAndRecord))
        addButton.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(addButton)

        let removeButton = NSButton(title: "Remove", target: self, action: #selector(removeSelected))
        removeButton.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(removeButton)

        let rerecordButton = NSButton(title: "Record Selected", target: self, action: #selector(recordSelected))
        rerecordButton.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(rerecordButton)

        recordHint.stringValue = "Press Add / Record, then press a key or a lone modifier."
        recordHint.textColor = .secondaryLabelColor
        recordHint.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(recordHint)

        let appearanceLabel = NSTextField(labelWithString: "Candidate panel")
        appearanceLabel.font = NSFont.boldSystemFont(ofSize: 13)
        appearanceLabel.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(appearanceLabel)

        let columnsLabel = NSTextField(labelWithString: "Columns (1–9)")
        columnsLabel.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(columnsLabel)

        columnsField.alignment = .right
        columnsField.translatesAutoresizingMaskIntoConstraints = false
        columnsField.target = self
        columnsField.action = #selector(columnsFieldChanged)
        root.addSubview(columnsField)

        columnsStepper.minValue = 1
        columnsStepper.maxValue = 9
        columnsStepper.increment = 1
        columnsStepper.valueWraps = false
        columnsStepper.translatesAutoresizingMaskIntoConstraints = false
        columnsStepper.target = self
        columnsStepper.action = #selector(columnsStepperChanged)
        root.addSubview(columnsStepper)

        let fontLabel = NSTextField(labelWithString: "Font")
        fontLabel.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(fontLabel)

        fontNameField.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(fontNameField)

        let chooseFontButton = NSButton(title: "Choose…", target: self, action: #selector(chooseFont))
        chooseFontButton.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(chooseFontButton)

        let sizeLabel = NSTextField(labelWithString: "Size")
        sizeLabel.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(sizeLabel)

        fontSizeField.alignment = .right
        fontSizeField.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(fontSizeField)

        let colorLabel = NSTextField(labelWithString: "Highlight color")
        colorLabel.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(colorLabel)

        colorWell.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(colorWell)

        let saveButton = NSButton(title: "Save", target: self, action: #selector(saveSettings))
        saveButton.keyEquivalent = "\r"
        saveButton.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(saveButton)

        let cancelButton = NSButton(title: "Cancel", target: self, action: #selector(cancelSettings))
        cancelButton.keyEquivalent = "\u{1b}"
        cancelButton.translatesAutoresizingMaskIntoConstraints = false
        root.addSubview(cancelButton)

        NSLayoutConstraint.activate([
            hotkeyLabel.topAnchor.constraint(equalTo: root.topAnchor, constant: 16),
            hotkeyLabel.leadingAnchor.constraint(equalTo: root.leadingAnchor, constant: 16),

            scrollView.topAnchor.constraint(equalTo: hotkeyLabel.bottomAnchor, constant: 8),
            scrollView.leadingAnchor.constraint(equalTo: root.leadingAnchor, constant: 16),
            scrollView.trailingAnchor.constraint(equalTo: root.trailingAnchor, constant: -16),
            scrollView.heightAnchor.constraint(equalToConstant: 140),

            addButton.topAnchor.constraint(equalTo: scrollView.bottomAnchor, constant: 8),
            addButton.leadingAnchor.constraint(equalTo: scrollView.leadingAnchor),

            removeButton.centerYAnchor.constraint(equalTo: addButton.centerYAnchor),
            removeButton.leadingAnchor.constraint(equalTo: addButton.trailingAnchor, constant: 8),

            rerecordButton.centerYAnchor.constraint(equalTo: addButton.centerYAnchor),
            rerecordButton.leadingAnchor.constraint(equalTo: removeButton.trailingAnchor, constant: 8),

            recordHint.topAnchor.constraint(equalTo: addButton.bottomAnchor, constant: 6),
            recordHint.leadingAnchor.constraint(equalTo: scrollView.leadingAnchor),
            recordHint.trailingAnchor.constraint(equalTo: scrollView.trailingAnchor),

            appearanceLabel.topAnchor.constraint(equalTo: recordHint.bottomAnchor, constant: 18),
            appearanceLabel.leadingAnchor.constraint(equalTo: scrollView.leadingAnchor),

            columnsLabel.topAnchor.constraint(equalTo: appearanceLabel.bottomAnchor, constant: 12),
            columnsLabel.leadingAnchor.constraint(equalTo: scrollView.leadingAnchor),

            columnsField.centerYAnchor.constraint(equalTo: columnsLabel.centerYAnchor),
            columnsField.leadingAnchor.constraint(equalTo: columnsLabel.trailingAnchor, constant: 12),
            columnsField.widthAnchor.constraint(equalToConstant: 48),

            columnsStepper.centerYAnchor.constraint(equalTo: columnsLabel.centerYAnchor),
            columnsStepper.leadingAnchor.constraint(equalTo: columnsField.trailingAnchor, constant: 6),

            fontLabel.topAnchor.constraint(equalTo: columnsLabel.bottomAnchor, constant: 12),
            fontLabel.leadingAnchor.constraint(equalTo: scrollView.leadingAnchor),

            fontNameField.centerYAnchor.constraint(equalTo: fontLabel.centerYAnchor),
            fontNameField.leadingAnchor.constraint(equalTo: fontLabel.trailingAnchor, constant: 12),
            fontNameField.widthAnchor.constraint(equalToConstant: 220),

            chooseFontButton.centerYAnchor.constraint(equalTo: fontLabel.centerYAnchor),
            chooseFontButton.leadingAnchor.constraint(equalTo: fontNameField.trailingAnchor, constant: 8),

            sizeLabel.centerYAnchor.constraint(equalTo: fontLabel.centerYAnchor),
            sizeLabel.leadingAnchor.constraint(equalTo: chooseFontButton.trailingAnchor, constant: 16),

            fontSizeField.centerYAnchor.constraint(equalTo: fontLabel.centerYAnchor),
            fontSizeField.leadingAnchor.constraint(equalTo: sizeLabel.trailingAnchor, constant: 8),
            fontSizeField.widthAnchor.constraint(equalToConstant: 48),

            colorLabel.topAnchor.constraint(equalTo: fontLabel.bottomAnchor, constant: 12),
            colorLabel.leadingAnchor.constraint(equalTo: scrollView.leadingAnchor),

            colorWell.centerYAnchor.constraint(equalTo: colorLabel.centerYAnchor),
            colorWell.leadingAnchor.constraint(equalTo: colorLabel.trailingAnchor, constant: 12),
            colorWell.widthAnchor.constraint(equalToConstant: 56),
            colorWell.heightAnchor.constraint(equalToConstant: 24),

            cancelButton.bottomAnchor.constraint(equalTo: root.bottomAnchor, constant: -16),
            cancelButton.trailingAnchor.constraint(equalTo: root.trailingAnchor, constant: -16),

            saveButton.centerYAnchor.constraint(equalTo: cancelButton.centerYAnchor),
            saveButton.trailingAnchor.constraint(equalTo: cancelButton.leadingAnchor, constant: -8),
        ])

        return root
    }

    private func reloadUI() {
        bindingsTable.reloadData()
        columnsStepper.integerValue = draft.clampedColumns
        columnsField.stringValue = "\(draft.clampedColumns)"
        fontNameField.stringValue = draft.candidateFontName
        fontSizeField.stringValue = String(format: "%.0f", draft.candidateFontSize)
        colorWell.color = draft.highlightColor
        recordHint.stringValue = "Press Add / Record, then press a key or a lone modifier."
        recordingRow = nil
        stopRecordingMonitor()
    }

    func numberOfRows(in tableView: NSTableView) -> Int {
        draft.modeToggleBindings.count
    }

    func tableView(_ tableView: NSTableView, viewFor tableColumn: NSTableColumn?, row: Int) -> NSView? {
        let cell = NSTableCellView()
        let text = NSTextField(labelWithString: describeBinding(draft.modeToggleBindings[row], row: row))
        text.translatesAutoresizingMaskIntoConstraints = false
        cell.addSubview(text)
        NSLayoutConstraint.activate([
            text.leadingAnchor.constraint(equalTo: cell.leadingAnchor, constant: 4),
            text.trailingAnchor.constraint(equalTo: cell.trailingAnchor, constant: -4),
            text.centerYAnchor.constraint(equalTo: cell.centerYAnchor),
        ])
        return cell
    }

    private func describeBinding(_ binding: PersistedModeToggleBinding, row: Int) -> String {
        let recording = recordingRow == row ? " [recording…]" : ""
        let enabled = binding.enabled ? "on" : "off"
        switch binding.triggerKind {
        case .keyDown:
            return "keyDown keyCode=\(binding.keyCode) (\(enabled))\(recording)"
        case .modifierAlone:
            let name = modifierName(raw: binding.modifierFlagRaw)
            return "modifierAlone \(name) keyCode=\(binding.keyCode) (\(enabled))\(recording)"
        }
    }

    private func modifierName(raw: UInt?) -> String {
        guard let raw else {
            return "modifier"
        }
        let flags = NSEvent.ModifierFlags(rawValue: raw)
        if flags.contains(.shift) { return "Shift" }
        if flags.contains(.control) { return "Control" }
        if flags.contains(.option) { return "Option" }
        if flags.contains(.command) { return "Command" }
        return "modifier"
    }

    @objc private func addAndRecord() {
        let placeholder = PersistedModeToggleBinding(
            id: "custom_\(UUID().uuidString.prefix(8))",
            keyCode: 0,
            triggerKind: .keyDown,
            modifierFlagRaw: nil,
            enabled: true
        )
        draft.modeToggleBindings.append(placeholder)
        bindingsTable.reloadData()
        let row = draft.modeToggleBindings.count - 1
        bindingsTable.selectRowIndexes(IndexSet(integer: row), byExtendingSelection: false)
        beginRecording(row: row)
    }

    @objc private func removeSelected() {
        let row = bindingsTable.selectedRow
        guard row >= 0, row < draft.modeToggleBindings.count else {
            return
        }
        draft.modeToggleBindings.remove(at: row)
        if draft.modeToggleBindings.isEmpty {
            draft.modeToggleBindings = TSIMESettings.default.modeToggleBindings
        }
        recordingRow = nil
        stopRecordingMonitor()
        bindingsTable.reloadData()
    }

    @objc private func recordSelected() {
        let row = bindingsTable.selectedRow
        guard row >= 0 else {
            recordHint.stringValue = "Select a hotkey row first, then Record Selected."
            return
        }
        beginRecording(row: row)
    }

    private func beginRecording(row: Int) {
        recordingRow = row
        recordHint.stringValue = "Recording… press a key or release a lone modifier. Esc cancels."
        bindingsTable.reloadData()
        stopRecordingMonitor()
        localMonitor = NSEvent.addLocalMonitorForEvents(matching: [.keyDown, .flagsChanged]) { [weak self] event in
            self?.handleRecordEvent(event)
            return nil
        }
    }

    private func stopRecordingMonitor() {
        if let localMonitor {
            NSEvent.removeMonitor(localMonitor)
            self.localMonitor = nil
        }
    }

    private func handleRecordEvent(_ event: NSEvent) {
        guard let row = recordingRow, row >= 0, row < draft.modeToggleBindings.count else {
            return
        }
        if event.type == .keyDown {
            if event.keyCode == 53 {
                recordingRow = nil
                stopRecordingMonitor()
                recordHint.stringValue = "Recording cancelled."
                bindingsTable.reloadData()
                return
            }
            let chord = event.modifierFlags.intersection([.shift, .control, .option, .command])
            guard chord.isEmpty else {
                return
            }
            draft.modeToggleBindings[row].triggerKind = .keyDown
            draft.modeToggleBindings[row].keyCode = event.keyCode
            draft.modeToggleBindings[row].modifierFlagRaw = nil
            draft.modeToggleBindings[row].enabled = true
            finishRecording(successText: "Recorded keyDown keyCode=\(event.keyCode)")
            return
        }
        if event.type == .flagsChanged {
            let flag: NSEvent.ModifierFlags?
            switch event.keyCode {
            case 56, 60:
                flag = .shift
            case 59, 62:
                flag = .control
            case 58, 61:
                flag = .option
            case 55, 54:
                flag = .command
            default:
                flag = nil
            }
            guard let flag, !event.modifierFlags.contains(flag) else {
                return
            }
            draft.modeToggleBindings[row].triggerKind = .modifierAlone
            draft.modeToggleBindings[row].keyCode = event.keyCode
            draft.modeToggleBindings[row].modifierFlagRaw = flag.rawValue
            draft.modeToggleBindings[row].enabled = true
            finishRecording(successText: "Recorded lone \(modifierName(raw: flag.rawValue)) (keyCode=\(event.keyCode))")
        }
    }

    private func finishRecording(successText: String) {
        recordingRow = nil
        stopRecordingMonitor()
        recordHint.stringValue = successText
        bindingsTable.reloadData()
    }

    @objc private func columnsStepperChanged() {
        let value = columnsStepper.integerValue
        columnsField.stringValue = "\(value)"
        draft.candidateColumns = value
    }

    @objc private func columnsFieldChanged() {
        let value = min(9, max(1, Int(columnsField.stringValue) ?? draft.candidateColumns))
        columnsStepper.integerValue = value
        columnsField.stringValue = "\(value)"
        draft.candidateColumns = value
    }

    @objc private func chooseFont() {
        let fontPanel = NSFontPanel.shared
        fontPanel.setPanelFont(draft.candidateFont, isMultiple: false)
        fontPanel.makeKeyAndOrderFront(self)
        NSFontManager.shared.target = self
        NSFontManager.shared.action = #selector(changeFont(_:))
    }

    @objc func changeFont(_ sender: Any?) {
        guard sender is NSFontManager else {
            return
        }
        let manager = NSFontManager.shared
        let current = draft.candidateFont
        let next = manager.convert(current)
        draft.candidateFontName = next.fontName
        draft.candidateFontSize = Double(next.pointSize)
        fontNameField.stringValue = next.fontName
        fontSizeField.stringValue = String(format: "%.0f", next.pointSize)
    }

    @objc private func saveSettings() {
        columnsFieldChanged()
        if let size = Double(fontSizeField.stringValue) {
            draft.candidateFontSize = size
        }
        let typedName = fontNameField.stringValue.trimmingCharacters(in: .whitespacesAndNewlines)
        if !typedName.isEmpty {
            draft.candidateFontName = typedName
        }
        draft.highlightColorHex = colorWell.color.tsimeHexString
        draft.candidateColumns = draft.clampedColumns
        do {
            try ConfigStore.save(draft)
            stopRecordingMonitor()
            window?.close()
        } catch {
            let alert = NSAlert()
            alert.messageText = "Failed to save settings"
            alert.informativeText = error.localizedDescription
            alert.runModal()
        }
    }

    @objc private func cancelSettings() {
        stopRecordingMonitor()
        window?.close()
    }

    func windowWillClose(_ notification: Notification) {
        stopRecordingMonitor()
        recordingRow = nil
    }
}
