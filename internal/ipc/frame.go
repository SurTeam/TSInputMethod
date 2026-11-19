package ipc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const maxFrameSize = 4 * 1024 * 1024

var (
	ErrFrameTooLarge = errors.New("ipc: frame too large")
	ErrFrameEmpty    = errors.New("ipc: empty frame")
)

func WriteFrame(w io.Writer, payload []byte) error {
	if len(payload) == 0 {
		return ErrFrameEmpty
	}
	if len(payload) > maxFrameSize {
		return ErrFrameTooLarge
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
	if _, err := w.Write(header[:]); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}

func ReadFrame(r io.Reader) ([]byte, error) {
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint32(header[:])
	if size == 0 {
		return nil, ErrFrameEmpty
	}
	if size > maxFrameSize {
		return nil, fmt.Errorf("%w: %d", ErrFrameTooLarge, size)
	}
	payload := make([]byte, size)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	return payload, nil
}
