package ipc

import (
	"net"

	imev1 "github.com/SurTeam/TSInputMethod/proto/ime/v1"
)

type Client struct {
	conn *Conn
}

func DialUnix(socketPath string) (*Client, error) {
	raw, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}
	return &Client{conn: NewConn(raw)}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) RoundTrip(req *imev1.Envelope) (*imev1.Envelope, error) {
	if err := c.conn.Send(req); err != nil {
		return nil, err
	}
	return c.conn.Recv()
}
