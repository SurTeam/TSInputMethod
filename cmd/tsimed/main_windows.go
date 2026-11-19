//go:build windows

package main

import (
	"os"

	"github.com/SurTeam/TSInputMethod/internal/ipc"
	"github.com/SurTeam/TSInputMethod/internal/session"
)

func listenIPC(manager *session.Manager) (*ipc.Server, string, error) {
	pipePath := os.Getenv("TSIME_PIPE")
	if pipePath == "" {
		pipePath = ipc.DefaultNamedPipePath()
	}
	server, err := ipc.ListenNamedPipe(pipePath, manager)
	if err != nil {
		return nil, "", err
	}
	return server, pipePath, nil
}
