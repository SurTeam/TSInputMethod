package dictionary

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const userDictFormatVersion = 1

type userDictFile struct {
	FormatVersion int             `json:"format_version"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Entries       []userDictJSON  `json:"entries"`
}

type fileStamp struct {
	modTime time.Time
	size    int64
}

func statFileStamp(path string) fileStamp {
	st, err := os.Stat(path)
	if err != nil {
		return fileStamp{}
	}
	return fileStamp{modTime: st.ModTime(), size: st.Size()}
}

type userDictJSON struct {
	Text       string    `json:"text"`
	Syllables  []string  `json:"syllables"`
	Weight     uint32    `json:"weight"`
	UsageCount uint32    `json:"usage_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Source     string    `json:"source"`
	Enabled    bool      `json:"enabled"`
}

func DefaultUserDictPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			if err != nil {
				return "", err
			}
			return "", homeErr
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "TSInputMethod", "user_dictionary.json"), nil
}

func (u *UserDictionary) LoadFile(path string) error {
	if u == nil {
		return ErrInvalidUserEntry
	}
	stamp := statFileStamp(path)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			u.resetLoadedLocked(path, stamp)
			return nil
		}
		return err
	}
	if len(data) == 0 {
		u.resetLoadedLocked(path, stamp)
		return nil
	}
	var file userDictFile
	if err := json.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("dictionary: corrupt user dictionary: %w", err)
	}
	if file.FormatVersion != 0 && file.FormatVersion != userDictFormatVersion {
		return fmt.Errorf("dictionary: unsupported user dictionary version %d", file.FormatVersion)
	}
	entries := make([]UserEntry, 0, len(file.Entries))
	for _, item := range file.Entries {
		text, err := normalizeUserText(item.Text)
		if err != nil {
			continue
		}
		syllables, err := normalizeUserSyllables(item.Syllables)
		if err != nil {
			continue
		}
		source := item.Source
		if source != UserSourceManual && source != UserSourceLearned {
			source = UserSourceManual
		}
		weight := item.Weight
		if weight == 0 {
			weight = defaultUserWeight
		}
		entries = append(entries, UserEntry{
			Text:       text,
			Syllables:  syllables,
			Weight:     weight,
			UsageCount: item.UsageCount,
			CreatedAt:  item.CreatedAt.UTC(),
			UpdatedAt:  item.UpdatedAt.UTC(),
			Source:     source,
			Enabled:    item.Enabled || item.CreatedAt.IsZero(),
		})
		if !item.Enabled && !item.CreatedAt.IsZero() {
			entries[len(entries)-1].Enabled = item.Enabled
		}
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	u.path = path
	u.replaceAllLocked(entries)
	u.version++
	u.savedVer = u.version
	u.diskStamp = stamp
	return nil
}

func (u *UserDictionary) resetLoadedLocked(path string, stamp fileStamp) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.path = path
	if len(u.entries) > 0 {
		u.version++
	}
	u.replaceAllLocked(nil)
	u.savedVer = u.version
	u.diskStamp = stamp
}

func (u *UserDictionary) ChangedOnDisk() bool {
	if u == nil {
		return false
	}
	path := u.Path()
	if path == "" {
		return false
	}
	stamp := statFileStamp(path)
	u.mu.RLock()
	defer u.mu.RUnlock()
	return !stamp.modTime.Equal(u.diskStamp.modTime) || stamp.size != u.diskStamp.size
}

func (u *UserDictionary) ReloadIfChangedOnDisk() (bool, error) {
	if u == nil || !u.ChangedOnDisk() {
		return false, nil
	}
	if err := u.ReloadLocked(); err != nil {
		return false, err
	}
	return true, nil
}

func (u *UserDictionary) Save() error {
	if u == nil {
		return ErrInvalidUserEntry
	}
	u.mu.RLock()
	path := u.path
	ver := u.version
	entries := make([]UserEntry, len(u.entries))
	copy(entries, u.entries)
	for i := range entries {
		entries[i].Syllables = append([]string(nil), u.entries[i].Syllables...)
	}
	u.mu.RUnlock()
	if path == "" {
		var err error
		path, err = DefaultUserDictPath()
		if err != nil {
			return err
		}
		u.SetPath(path)
	}
	if err := saveUserDictFile(path, entries); err != nil {
		return err
	}
	stamp := statFileStamp(path)
	u.mu.Lock()
	if ver > u.savedVer {
		u.savedVer = ver
	}
	u.diskStamp = stamp
	u.mu.Unlock()
	return nil
}

func (u *UserDictionary) Dirty() bool {
	if u == nil {
		return false
	}
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.version != u.savedVer
}

func (u *UserDictionary) SaveIfDirty() error {
	if !u.Dirty() || u.Path() == "" {
		return nil
	}
	return u.SaveLocked()
}

func saveUserDictFile(path string, entries []UserEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	file := userDictFile{
		FormatVersion: userDictFormatVersion,
		UpdatedAt:     time.Now().UTC(),
		Entries:       make([]userDictJSON, 0, len(entries)),
	}
	for _, entry := range entries {
		file.Entries = append(file.Entries, userDictJSON{
			Text:       entry.Text,
			Syllables:  append([]string(nil), entry.Syllables...),
			Weight:     entry.Weight,
			UsageCount: entry.UsageCount,
			CreatedAt:  entry.CreatedAt.UTC(),
			UpdatedAt:  entry.UpdatedAt.UTC(),
			Source:     entry.Source,
			Enabled:    entry.Enabled,
		})
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), "user_dictionary-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	return nil
}

func OpenUserDictionary(path string) (*UserDictionary, error) {
	if path == "" {
		var err error
		path, err = DefaultUserDictPath()
		if err != nil {
			return nil, err
		}
	}
	u := NewUserDictionary()
	if err := u.LoadFile(path); err != nil {
		return nil, err
	}
	return u, nil
}

func (u *UserDictionary) Reload() error {
	if u == nil {
		return ErrInvalidUserEntry
	}
	path := u.Path()
	if path == "" {
		var err error
		path, err = DefaultUserDictPath()
		if err != nil {
			return err
		}
	}
	return u.LoadFile(path)
}

func (u *UserDictionary) SaveLocked() error {
	if u == nil {
		return ErrInvalidUserEntry
	}
	path := u.Path()
	if path == "" {
		var err error
		path, err = DefaultUserDictPath()
		if err != nil {
			return err
		}
		u.SetPath(path)
	}
	return withUserDictFileLock(path, true, func() error {
		return u.Save()
	})
}

func (u *UserDictionary) ReloadLocked() error {
	if u == nil {
		return ErrInvalidUserEntry
	}
	path := u.Path()
	if path == "" {
		var err error
		path, err = DefaultUserDictPath()
		if err != nil {
			return err
		}
		u.SetPath(path)
	}
	return withUserDictFileLock(path, false, func() error {
		return u.LoadFile(path)
	})
}
