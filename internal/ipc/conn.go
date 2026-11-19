package ipc

import (
	"fmt"
	"net"
	"sync"

	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
	"google.golang.org/protobuf/proto"
)

type Conn struct {
	raw net.Conn
	mu  sync.Mutex
}

func NewConn(raw net.Conn) *Conn {
	return &Conn{raw: raw}
}

func (c *Conn) Close() error {
	return c.raw.Close()
}

func (c *Conn) Send(msg *imev1.Envelope) error {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return WriteFrame(c.raw, payload)
}

func (c *Conn) Recv() (*imev1.Envelope, error) {
	payload, err := ReadFrame(c.raw)
	if err != nil {
		return nil, err
	}
	msg := &imev1.Envelope{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		return nil, fmt.Errorf("ipc: unmarshal envelope: %w", err)
	}
	return msg, nil
}
