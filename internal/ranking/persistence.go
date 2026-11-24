package ranking

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const personalModelFormatVersion = 1

var (
	ErrIncompatiblePersonalModel = errors.New("ranking: incompatible personal model")
)

type personalModelFile struct {
	FormatVersion        int                `json:"format_version"`
	HashDimension        uint32             `json:"hash_dimension"`
	HashAlgorithm        string             `json:"hash_algorithm"`
	FeatureSchemaVersion int                `json:"feature_schema_version"`
	LearningRate         float64            `json:"learning_rate"`
	UpdateCount          uint64             `json:"update_count"`
	UpdatedAt            time.Time          `json:"updated_at"`
	Weights              map[string]float64 `json:"weights"`
}

func DefaultPersonalModelPath() (string, error) {
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
	return filepath.Join(dir, "TSInputMethod", "personal_model.json"), nil
}

func (m *PersonalModel) SetPath(path string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.path = path
	m.mu.Unlock()
}

func (m *PersonalModel) Path() string {
	if m == nil {
		return ""
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.path
}

func (m *PersonalModel) Save() error {
	if m == nil {
		return ErrIncompatiblePersonalModel
	}
	m.mu.RLock()
	path := m.path
	seq := m.changeSeq
	file := personalModelFile{
		FormatVersion:        personalModelFormatVersion,
		HashDimension:        m.cfg.Dimension,
		HashAlgorithm:        HashAlgorithmFNV1a64,
		FeatureSchemaVersion: FeatureSchemaVersion,
		LearningRate:         m.cfg.LearningRate,
		UpdateCount:          m.updateCount,
		UpdatedAt:            m.updatedAt,
		Weights:              make(map[string]float64, len(m.weights)),
	}
	for idx, w := range m.weights {
		file.Weights[fmt.Sprintf("%d", idx)] = w
	}
	m.mu.RUnlock()
	if path == "" {
		var err error
		path, err = DefaultPersonalModelPath()
		if err != nil {
			return err
		}
		m.SetPath(path)
	}
	if err := savePersonalModelFile(path, file); err != nil {
		return err
	}
	m.mu.Lock()
	if m.changeSeq == seq {
		m.dirty = false
	}
	m.mu.Unlock()
	return nil
}

func (m *PersonalModel) Dirty() bool {
	if m == nil {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dirty
}

func (m *PersonalModel) SaveIfDirty() error {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	dirty := m.dirty
	m.mu.RUnlock()
	if !dirty {
		return nil
	}
	return m.Save()
}

func savePersonalModelFile(path string, file personalModelFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), "personal_model-*.tmp")
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
	return os.Rename(tmpName, path)
}

func (m *PersonalModel) LoadFile(path string) error {
	if m == nil {
		return ErrIncompatiblePersonalModel
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			m.SetPath(path)
			return nil
		}
		return err
	}
	if len(data) == 0 {
		m.SetPath(path)
		return nil
	}
	var file personalModelFile
	if err := json.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("ranking: corrupt personal model: %w", err)
	}
	if file.FormatVersion != 0 && file.FormatVersion != personalModelFormatVersion {
		return fmt.Errorf("%w: format %d", ErrIncompatiblePersonalModel, file.FormatVersion)
	}
	if file.HashDimension != 0 && file.HashDimension != m.cfg.Dimension {
		return fmt.Errorf("%w: dimension %d", ErrIncompatiblePersonalModel, file.HashDimension)
	}
	if file.HashAlgorithm != "" && file.HashAlgorithm != HashAlgorithmFNV1a64 {
		return fmt.Errorf("%w: hash %s", ErrIncompatiblePersonalModel, file.HashAlgorithm)
	}
	if file.FeatureSchemaVersion != 0 && file.FeatureSchemaVersion != FeatureSchemaVersion {
		return fmt.Errorf("%w: schema %d", ErrIncompatiblePersonalModel, file.FeatureSchemaVersion)
	}
	weights := make(map[uint32]float64, len(file.Weights))
	for key, w := range file.Weights {
		var idx uint64
		if _, err := fmt.Sscanf(key, "%d", &idx); err != nil {
			continue
		}
		if idx >= uint64(m.cfg.Dimension) {
			continue
		}
		weights[uint32(idx)] = clipFloat(w, -m.cfg.WeightCap, m.cfg.WeightCap)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.path = path
	m.weights = weights
	m.updateCount = file.UpdateCount
	m.updatedAt = file.UpdatedAt.UTC()
	m.dirty = false
	return nil
}

func OpenPersonalModel(path string, cfg SGDConfig) (*PersonalModel, error) {
	if path == "" {
		var err error
		path, err = DefaultPersonalModelPath()
		if err != nil {
			return nil, err
		}
	}
	m := NewPersonalModel(cfg)
	if err := m.LoadFile(path); err != nil {
		return nil, err
	}
	return m, nil
}
