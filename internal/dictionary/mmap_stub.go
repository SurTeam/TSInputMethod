//go:build !unix

package dictionary

import "fmt"

func mapDictFile(path string) ([]byte, func() error, error) {
	return nil, nil, fmt.Errorf("dictionary: mmap unavailable")
}
