package personal

import "testing"

func TestDefaultSettingsEnabled(t *testing.T) {
	s := DefaultSettings()
	snap := s.Snapshot()
	if !snap.PersonalizationEnabled || !snap.UserDictionaryEnabled || !snap.AutoLearnEnabled {
		t.Fatalf("defaults should enable personalization/userdict/autolearn: %#v", snap)
	}
	if snap.PrivacyMode {
		t.Fatalf("privacy should default off")
	}
	s.SetPersonalizationEnabled(false)
	if s.Snapshot().PersonalizationEnabled {
		t.Fatalf("expected personalization disabled")
	}
}
