//go:build windows

package ipc

import (
	"fmt"
	"net"
	"os/user"
	"sync"
	"time"

	"github.com/SurTeam/TSInputMethod/internal/session"
	"golang.org/x/sys/windows"
)

func DefaultNamedPipePath() string {
	name := "tsime"
	if u, err := user.Current(); err == nil && u.Username != "" {
		name = fmt.Sprintf("tsime-%s", u.Username)
	}
	return `\\.\pipe\` + name
}

type pipeAddr string

func (a pipeAddr) Network() string { return "pipe" }
func (a pipeAddr) String() string  { return string(a) }

type pipeConn struct {
	handle windows.Handle
	local  pipeAddr
	remote pipeAddr
}

func newPipeConn(handle windows.Handle, path string) *pipeConn {
	return &pipeConn{
		handle: handle,
		local:  pipeAddr(path),
		remote: pipeAddr(path),
	}
}

func (c *pipeConn) Read(b []byte) (int, error) {
	var done uint32
	err := windows.ReadFile(c.handle, b, &done, nil)
	return int(done), err
}

func (c *pipeConn) Write(b []byte) (int, error) {
	var done uint32
	err := windows.WriteFile(c.handle, b, &done, nil)
	return int(done), err
}

func (c *pipeConn) Close() error {
	return windows.CloseHandle(c.handle)
}

func (c *pipeConn) LocalAddr() net.Addr                { return c.local }
func (c *pipeConn) RemoteAddr() net.Addr               { return c.remote }
func (c *pipeConn) SetDeadline(t time.Time) error      { return nil }
func (c *pipeConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *pipeConn) SetWriteDeadline(t time.Time) error { return nil }

type pipeListener struct {
	path   string
	closed chan struct{}
	once   sync.Once
}

func (l *pipeListener) Accept() (net.Conn, error) {
	for {
		select {
		case <-l.closed:
			return nil, fmt.Errorf("ipc: named pipe closed")
		default:
		}
		pathPtr, err := windows.UTF16PtrFromString(l.path)
		if err != nil {
			return nil, err
		}
		handle, err := windows.CreateNamedPipe(
			pathPtr,
			windows.PIPE_ACCESS_DUPLEX,
			windows.PIPE_TYPE_BYTE|windows.PIPE_READMODE_BYTE|windows.PIPE_WAIT,
			windows.PIPE_UNLIMITED_INSTANCES,
			65536,
			65536,
			0,
			nil,
		)
		if err != nil {
			return nil, err
		}
		err = windows.ConnectNamedPipe(handle, nil)
		if err != nil && err != windows.ERROR_PIPE_CONNECTED {
			_ = windows.CloseHandle(handle)
			select {
			case <-l.closed:
				return nil, fmt.Errorf("ipc: named pipe closed")
			default:
				time.Sleep(10 * time.Millisecond)
				continue
			}
		}
		return newPipeConn(handle, l.path), nil
	}
}

func (l *pipeListener) Close() error {
	l.once.Do(func() { close(l.closed) })
	return nil
}

func (l *pipeListener) Addr() net.Addr { return pipeAddr(l.path) }

func ListenNamedPipe(pipePath string, manager *session.Manager) (*Server, error) {
	if pipePath == "" {
		pipePath = DefaultNamedPipePath()
	}
	ln := &pipeListener{path: pipePath, closed: make(chan struct{})}
	return &Server{
		listener: ln,
		manager:  manager,
		quit:     make(chan struct{}),
	}, nil
}

func DialNamedPipe(pipePath string) (*Client, error) {
	if pipePath == "" {
		pipePath = DefaultNamedPipePath()
	}
	pathPtr, err := windows.UTF16PtrFromString(pipePath)
	if err != nil {
		return nil, err
	}
	var handle windows.Handle
	for i := 0; i < 50; i++ {
		handle, err = windows.CreateFile(
			pathPtr,
			windows.GENERIC_READ|windows.GENERIC_WRITE,
			0,
			nil,
			windows.OPEN_EXISTING,
			0,
			0,
		)
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		return nil, err
	}
	return &Client{conn: NewConn(newPipeConn(handle, pipePath))}, nil
}
