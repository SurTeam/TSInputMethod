//go:build !unix && !windows

package dictionary

func withUserDictFileLock(path string, exclusive bool, fn func() error) error {
	_ = path
	_ = exclusive
	return fn()
}
