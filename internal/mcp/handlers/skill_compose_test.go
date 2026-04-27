package handlers

import (
	"testing"
)

// TestInterpolateArgsWithVars tests $1, $2 interpolation.
func TestInterpolateArgsWithVars(t *testing.T) {
	interp := map[string]string{
		"$1": "arg1-value",
		"$2": "arg2-value",
		"$3": "arg3-value",
	}

	tests := []struct {
		name     string
		args     map[string]any
		expected map[string]any
	}{
		{
			name:     "single placeholder",
			args:     map[string]any{"query": "$1"},
			expected: map[string]any{"query": "arg1-value"},
		},
		{
			name:     "multiple placeholders",
			args:     map[string]any{"template": "$1 and $2"},
			expected: map[string]any{"template": "arg1-value and arg2-value"},
		},
		{
			name:     "non-string value passes through",
			args:     map[string]any{"limit": 10},
			expected: map[string]any{"limit": 10},
		},
		{
			name:     "nested placeholder same as key",
			args:     map[string]any{"taskId": "$1"},
			expected: map[string]any{"taskId": "arg1-value"},
		},
		{
			name:     "unknown placeholder passes through",
			args:     map[string]any{"query": "$9"},
			expected: map[string]any{"query": "$9"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpolateArgs(tt.args, interp)
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("interpolateArgs(%v)[%s] = %v, want %v", tt.args, k, result[k], v)
				}
			}
		})
	}
}

// TestInterpolateArgsNilInput handles nil args.
func TestInterpolateArgsNilInput(t *testing.T) {
	result := interpolateArgs(nil, map[string]string{"$1": "x"})
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}
}

// TestInterpolateStringAllPlaceholders replaces all occurrences.
func TestInterpolateStringAllPlaceholders(t *testing.T) {
	interp := map[string]string{"$1": "REPLACED"}
	result := interpolateString("$1 + $1 + $1", interp)
	if result != "REPLACED + REPLACED + REPLACED" {
		t.Errorf("expected all $1 replaced, got %s", result)
	}
}

// TestInterpolateStringNoPlaceholders leaves string unchanged.
func TestInterpolateStringNoPlaceholders(t *testing.T) {
	interp := map[string]string{"$1": "REPLACED"}
	result := interpolateString("no placeholders here", interp)
	if result != "no placeholders here" {
		t.Errorf("expected unchanged string, got %s", result)
	}
}

// TestReplaceAllEdgeCases tests the replaceAll function.
func TestReplaceAllEdgeCases(t *testing.T) {
	// Empty old string returns original.
	result := replaceAll("abc", "", "x")
	if result != "abc" {
		t.Errorf("expected 'abc' for empty old, got %s", result)
	}

	// No occurrence returns original.
	result = replaceAll("abc", "x", "y")
	if result != "abc" {
		t.Errorf("expected 'abc' for no match, got %s", result)
	}

	// Multiple occurrences.
	result = replaceAll("aaa", "a", "b")
	if result != "bbb" {
		t.Errorf("expected 'bbb', got %s", result)
	}

	// Single occurrence.
	result = replaceAll("xabc", "a", "z")
	if result != "xzbc" {
		t.Errorf("expected 'xzbc', got %s", result)
	}
}

// TestGetBuiltInCompositeSkill tests all built-in skills.
func TestGetBuiltInCompositeSkill(t *testing.T) {
	skills := []string{"full-review", "implement-and-test", "review-and-commit"}

	for _, name := range skills {
		skill := getBuiltInCompositeSkill(name)
		if skill == nil {
			t.Errorf("expected skill %q to exist, got nil", name)
			continue
		}
		if skill.Name != name {
			t.Errorf("expected name=%q, got %q", name, skill.Name)
		}
		if len(skill.Steps) == 0 {
			t.Errorf("skill %q has no steps", name)
		}
	}

	// Unknown skill.
	unknown := getBuiltInCompositeSkill("does-not-exist")
	if unknown != nil {
		t.Errorf("expected nil for unknown skill, got %+v", unknown)
	}
}

// TestSkillComposeStepArgs tests that built-in skill args are NOT interpolated
// (they contain fixed values, not $1 placeholders).
func TestSkillComposeStepArgs(t *testing.T) {
	skill := getBuiltInCompositeSkill("full-review")
	if skill == nil {
		t.Skip("skill not found")
	}

	// full-review step 0 has {"scope": "all"} — fixed value, not $1.
	// So interpolation with $1 should not change it.
	interp := map[string]string{"$1": "my-scope"}
	step := skill.Steps[0]
	interpolated := interpolateArgs(step.Args, interp)

	// scope remains "all" because the step arg is "all", not "$1".
	if interpolated["scope"] != "all" {
		t.Errorf("expected scope=all (no placeholder to replace), got %v", interpolated["scope"])
	}
}
