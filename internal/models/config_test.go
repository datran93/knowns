package models

import (
	"encoding/json"
	"testing"
)

func TestDefaultAgentEfficiencySettings(t *testing.T) {
	cfg := DefaultAgentEfficiencySettings()

	tests := []struct {
		name      string
		flag      *FeatureFlag
		wantTrue  bool
		subField  string
		wantVal   int
	}{
		{"persistentContext enabled", cfg.PersistentContext, true, "MaxMemories", 5},
		{"sessionResume enabled", cfg.SessionResume, true, "CheckpointTTLHours", 24},
		{"codeNavigation enabled", cfg.CodeNavigation, true, "MaxTraceDepth", 10},
		{"autoValidation enabled", cfg.AutoValidation, true, "CacheTTLSeconds", 300},
		{"modelRouter enabled", cfg.ModelRouter, true, "FTS5Threshold", 0},
		{"skillComposer enabled", cfg.SkillComposer, true, "", 0},
		{"backgroundIndexing enabled", cfg.BackgroundIndexing, true, "DebounceSeconds", 5},
		{"multiAgent enabled", cfg.MultiAgent, true, "LockTTLSeconds", 300},
	}

	for _, tt := range tests {
		if tt.flag == nil {
			t.Errorf("%s: flag is nil", tt.name)
			continue
		}
		if tt.flag.Enabled != tt.wantTrue {
			t.Errorf("%s: Enabled = %v, want %v", tt.name, tt.flag.Enabled, tt.wantTrue)
		}
		if tt.subField != "" {
			var got int
			switch tt.subField {
			case "MaxMemories":
				got = tt.flag.MaxMemories
			case "CheckpointTTLHours":
				got = tt.flag.CheckpointTTLHours
			case "MaxTraceDepth":
				got = tt.flag.MaxTraceDepth
			case "CacheTTLSeconds":
				got = tt.flag.CacheTTLSeconds
			case "FTS5Threshold":
				got = tt.flag.FTS5Threshold
			case "DebounceSeconds":
				got = tt.flag.DebounceSeconds
			case "LockTTLSeconds":
				got = tt.flag.LockTTLSeconds
			}
			if got != tt.wantVal {
				t.Errorf("%s: %s = %d, want %d", tt.name, tt.subField, got, tt.wantVal)
			}
		}
	}
}

func TestAgentEfficiencySettings_IsEnabled(t *testing.T) {
	// All nil — each should return true (nil means enabled by default).
	s := &AgentEfficiencySettings{}
	features := []string{
		"persistentContext", "sessionResume", "codeNavigation",
		"autoValidation", "modelRouter", "skillComposer",
		"backgroundIndexing", "multiAgent",
	}
	for _, f := range features {
		if !s.IsEnabled(f) {
			t.Errorf("IsEnabled(%q) with nil flag = false, want true", f)
		}
	}

	// Explicitly disabled features should return false.
	s = &AgentEfficiencySettings{
		PersistentContext:  &FeatureFlag{Enabled: false},
		SessionResume:      &FeatureFlag{Enabled: false},
		CodeNavigation:     &FeatureFlag{Enabled: false},
		AutoValidation:    &FeatureFlag{Enabled: false},
		ModelRouter:        &FeatureFlag{Enabled: false},
		SkillComposer:      &FeatureFlag{Enabled: false},
		BackgroundIndexing: &FeatureFlag{Enabled: false},
		MultiAgent:         &FeatureFlag{Enabled: false},
	}
	for _, f := range features {
		if s.IsEnabled(f) {
			t.Errorf("IsEnabled(%q) with Disabled flag = true, want false", f)
		}
	}

	// Mixed: some enabled, some disabled.
	s = &AgentEfficiencySettings{
		PersistentContext:  &FeatureFlag{Enabled: true},
		SessionResume:      &FeatureFlag{Enabled: false},
		CodeNavigation:     &FeatureFlag{Enabled: true},
		AutoValidation:    &FeatureFlag{Enabled: false},
		ModelRouter:        &FeatureFlag{Enabled: true},
		SkillComposer:      &FeatureFlag{Enabled: false},
		BackgroundIndexing: &FeatureFlag{Enabled: true},
		MultiAgent:         &FeatureFlag{Enabled: false},
	}
	mixed := map[string]bool{
		"persistentContext": true,
		"sessionResume":      false,
		"codeNavigation":     true,
		"autoValidation":    false,
		"modelRouter":        true,
		"skillComposer":      false,
		"backgroundIndexing": true,
		"multiAgent":         false,
	}
	for f, want := range mixed {
		got := s.IsEnabled(f)
		if got != want {
			t.Errorf("IsEnabled(%q) = %v, want %v", f, got, want)
		}
	}

	// Unknown feature always returns false.
	if s.IsEnabled("unknownFeature") {
		t.Error("IsEnabled(unknown) = true, want false")
	}
}

func TestAgentEfficiencySettings_JSONRoundTrip(t *testing.T) {
	original := &AgentEfficiencySettings{
		PersistentContext:  &FeatureFlag{Enabled: true, MaxMemories: 5},
		SessionResume:      &FeatureFlag{Enabled: true, CheckpointTTLHours: 24},
		CodeNavigation:     &FeatureFlag{Enabled: true, MaxTraceDepth: 10},
		AutoValidation:     &FeatureFlag{Enabled: true, CacheTTLSeconds: 300},
		ModelRouter:        &FeatureFlag{Enabled: true, FTS5Threshold: 0},
		SkillComposer:       &FeatureFlag{Enabled: true},
		BackgroundIndexing: &FeatureFlag{Enabled: true, DebounceSeconds: 5},
		MultiAgent:         &FeatureFlag{Enabled: true, LockTTLSeconds: 300},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var roundtripped AgentEfficiencySettings
	if err := json.Unmarshal(data, &roundtripped); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Check each feature is still enabled.
	features := []string{
		"persistentContext", "sessionResume", "codeNavigation",
		"autoValidation", "modelRouter", "skillComposer",
		"backgroundIndexing", "multiAgent",
	}
	for _, f := range features {
		if !roundtripped.IsEnabled(f) {
			t.Errorf("roundtrip IsEnabled(%q) = false, want true", f)
		}
	}

	// Check sub-fields preserved.
	if roundtripped.PersistentContext.MaxMemories != 5 {
		t.Errorf("PersistentContext.MaxMemories = %d, want 5", roundtripped.PersistentContext.MaxMemories)
	}
	if roundtripped.SessionResume.CheckpointTTLHours != 24 {
		t.Errorf("SessionResume.CheckpointTTLHours = %d, want 24", roundtripped.SessionResume.CheckpointTTLHours)
	}
	if roundtripped.CodeNavigation.MaxTraceDepth != 10 {
		t.Errorf("CodeNavigation.MaxTraceDepth = %d, want 10", roundtripped.CodeNavigation.MaxTraceDepth)
	}
	if roundtripped.AutoValidation.CacheTTLSeconds != 300 {
		t.Errorf("AutoValidation.CacheTTLSeconds = %d, want 300", roundtripped.AutoValidation.CacheTTLSeconds)
	}
	if roundtripped.BackgroundIndexing.DebounceSeconds != 5 {
		t.Errorf("BackgroundIndexing.DebounceSeconds = %d, want 5", roundtripped.BackgroundIndexing.DebounceSeconds)
	}
	if roundtripped.MultiAgent.LockTTLSeconds != 300 {
		t.Errorf("MultiAgent.LockTTLSeconds = %d, want 300", roundtripped.MultiAgent.LockTTLSeconds)
	}
}

func TestFeatureFlagJSON(t *testing.T) {
	// Verify FeatureFlag marshals with "enabled":true.
	ff := &FeatureFlag{Enabled: true, MaxMemories: 5}
	data, err := json.Marshal(ff)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var unmarshaled FeatureFlag
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if !unmarshaled.Enabled {
		t.Error("FeatureFlag.Enabled = false, want true after round-trip")
	}
	if unmarshaled.MaxMemories != 5 {
		t.Errorf("FeatureFlag.MaxMemories = %d, want 5", unmarshaled.MaxMemories)
	}
}