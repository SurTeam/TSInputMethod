package ipc

import (
	"bytes"
	"errors"
	"testing"
)

func TestWriteReadFrameRoundTrip(t *testing.T) {
	payload := []byte("hello-ime-frame")
	var buf bytes.Buffer
	if err := WriteFrame(&buf, payload); err != nil {
		t.Fatalf("WriteFrame: %v", err)
	}
	got, err := ReadFrame(&buf)
	if err != nil {
		t.Fatalf("ReadFrame: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("got %q want %q", got, payload)
	}
}

func TestReadFrameTooLarge(t *testing.T) {
var buf bytes.Buffer
	if _, err := buf.Write([]byte{0xff, 0xff, 0xff, 0xff}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_, err := ReadFrame(&buf)
	if !errors.Is(err, ErrFrameTooLarge) {
		t.Fatalf("expected ErrFrameTooLarge, got %v", err)
	}
}

func TestWriteFrameEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteFrame(&buf, nil); !errors.Is(err, ErrFrameEmpty) {
		t.Fatalf("expected ErrFrameEmpty, got %v", err)
	}
}
