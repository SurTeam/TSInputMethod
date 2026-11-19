//go:build !windows

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SurTeam/TSInputMethod/internal/ipc"
	"github.com/SurTeam/TSInputMethod/internal/session"
)

func listenIPC(manager *session.Manager) (*ipc.Server, string, error) {
	socketPath := resolveSocketPath()
	if err := os.MkdirAll(filepath.Dir(socketPath), 0o700); err != nil {
		return nil, "", err
	}
	server, err := ipc.ListenUnix(socketPath, manager)
	if err != nil {
		return nil, "", err
	}
	return server, socketPath, nil
}

func resolveSocketPath() string {
	if override := os.Getenv("TSIME_SOCKET"); override != "" {
		return override
	}
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		return filepath.Join(runtimeDir, "tsime.sock")
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("tsime-%d.sock", os.Getuid()))
}
