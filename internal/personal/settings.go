package personal

import "sync"

type Settings struct {
	mu                     sync.RWMutex
	PersonalizationEnabled bool
	UserDictionaryEnabled  bool
	AutoLearnEnabled       bool
	PrivacyMode            bool
	SGDWeight              float64
	UserFrequencyWeight    float64
	LearnMinRepeats        int
}

func DefaultSettings() *Settings {
	return &Settings{
		PersonalizationEnabled: true,
		UserDictionaryEnabled:  true,
		AutoLearnEnabled:       true,
		PrivacyMode:            false,
		SGDWeight:              40,
		UserFrequencyWeight:    8,
		LearnMinRepeats:        2,
	}
}

func (s *Settings) Snapshot() Settings {
	if s == nil {
		return *DefaultSettings()
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Settings{
		PersonalizationEnabled: s.PersonalizationEnabled,
		UserDictionaryEnabled:  s.UserDictionaryEnabled,
		AutoLearnEnabled:       s.AutoLearnEnabled,
		PrivacyMode:            s.PrivacyMode,
		SGDWeight:              s.SGDWeight,
		UserFrequencyWeight:    s.UserFrequencyWeight,
		LearnMinRepeats:        s.LearnMinRepeats,
	}
}

func (s *Settings) SetPersonalizationEnabled(v bool) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.PersonalizationEnabled = v
	s.mu.Unlock()
}

func (s *Settings) SetUserDictionaryEnabled(v bool) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.UserDictionaryEnabled = v
	s.mu.Unlock()
}

func (s *Settings) SetAutoLearnEnabled(v bool) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.AutoLearnEnabled = v
	s.mu.Unlock()
}

func (s *Settings) SetPrivacyMode(v bool) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.PrivacyMode = v
	s.mu.Unlock()
}
