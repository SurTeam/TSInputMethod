ROOT := $(abspath ../..)
COMMON := $(ROOT)/platforms/common
BUILD_COMMON := $(COMMON)/build
OUT := build/TSInputMethod.app/Contents
SWIFTC ?= swiftc
PROTO_LIBS := $(shell pkg-config --libs protobuf)

.PHONY: all common app install uninstall clean

all: app

common:
	cmake -S $(COMMON) -B $(BUILD_COMMON)
	cmake --build $(BUILD_COMMON)

app: common
	mkdir -p $(OUT)/MacOS $(OUT)/Resources
	cp Info.plist $(OUT)/
	cp Resources/menu_icon.pdf $(OUT)/Resources/
	cp Resources/AppIcon.icns $(OUT)/Resources/
	cp -R Resources/en.lproj $(OUT)/Resources/
	cp -R Resources/zh-Hans.lproj $(OUT)/Resources/
	$(SWIFTC) \
		-parse-as-library \
		-module-name TSInputMethod \
		-import-objc-header Sources/TSInputMethod/TSInputMethod-Bridging-Header.h \
		-I $(COMMON)/native \
		-Xlinker -lc++ \
		-L $(BUILD_COMMON) -ltsime_ipc \
$(PROTO_LIBS) \
			-framework Cocoa -framework InputMethodKit -framework UserNotifications \
			-o $(OUT)/MacOS/TSInputMethod \
Sources/TSInputMethod/main.swift \
					Sources/TSInputMethod/DaemonClient.swift \
					Sources/TSInputMethod/DaemonAgent.swift \
					Sources/TSInputMethod/CandidatePanel.swift \
					Sources/TSInputMethod/ModeToggleHotkeys.swift \
					Sources/TSInputMethod/ConfigStore.swift \
					Sources/TSInputMethod/AppNotifier.swift \
					Sources/TSInputMethod/SettingsWindowController.swift \
					Sources/TSInputMethod/InputController.swift

install: app
	./scripts/install.sh

uninstall:
	./scripts/uninstall.sh

clean:
	rm -rf build
