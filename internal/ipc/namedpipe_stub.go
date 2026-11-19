//go:build !windows

package ipc

import (
	"fmt"

	"github.com/SurTeam/TSInputMethod/internal/session"
)

func DefaultNamedPipePath() string {
	return `\\.\pipe\tsime`
}

func ListenNamedPipe(pipePath string, manager *session.Manager) (*Server, error) {
	return nil, fmt.Errorf("ipc: named pipe listen is only available on windows")
}

func DialNamedPipe(pipePath string) (*Client, error) {
	return nil, fmt.Errorf("ipc: named pipe dial is only available on windows")
}
