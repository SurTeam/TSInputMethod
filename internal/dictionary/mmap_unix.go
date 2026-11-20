//go:build unix

package dictionary

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func mapDictFile(path string) ([]byte, func() error, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}
	size := int(st.Size())
	if size == 0 {
		return nil, nil, fmt.Errorf("dictionary: empty file")
	}
	if int64(size) != st.Size() {
		return nil, nil, fmt.Errorf("dictionary: file too large")
	}
	data, err := unix.Mmap(int(f.Fd()), 0, size, unix.PROT_READ, unix.MAP_SHARED)
	if err != nil {
		return nil, nil, err
	}
	closer := func() error {
		return unix.Munmap(data)
	}
	return data, closer, nil
}
