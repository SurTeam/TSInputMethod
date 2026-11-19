package ipc

import (
	"bytes"
	"testing"
)

func BenchmarkWriteReadFrame(b *testing.B) {
	payload := bytes.Repeat([]byte("ime"), 64)
	var buf bytes.Buffer
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := WriteFrame(&buf, payload); err != nil {
			b.Fatal(err)
		}
		if _, err := ReadFrame(&buf); err != nil {
			b.Fatal(err)
		}
	}
}
