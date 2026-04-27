package handlers

import (
	"testing"
)

func TestStringOrNone(t *testing.T) {
	s := "hello"
	if stringOrNone(&s) != "hello" {
		t.Errorf("expected 'hello', got %q", stringOrNone(&s))
	}

	empty := ""
	if stringOrNone(&empty) != "N/A" {
		t.Errorf("expected 'N/A' for empty string, got %q", stringOrNone(&empty))
	}

	if stringOrNone(nil) != "N/A" {
		t.Errorf("expected 'N/A' for nil, got %q", stringOrNone(nil))
	}
}

func TestIsRedis(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"redis://localhost:6379", true},
		{"rediss://localhost:6379", true},
		{"postgresql://localhost/db", false},
		{"mysql://localhost/db", false},
		{"sqlite:///tmp/db", false},
	}

	for _, tt := range tests {
		result := isRedis(tt.input)
		if result != tt.expected {
			t.Errorf("isRedis(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}