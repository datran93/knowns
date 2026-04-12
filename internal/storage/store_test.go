package storage

import (
	"path/filepath"
	"testing"
)

func TestGlobalRootPathPrefersHOMEOverride(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got := GlobalRootPath()
	want := filepath.Join(home, ".knowns")
	if got != want {
		t.Fatalf("GlobalRootPath() = %q, want %q", got, want)
	}
}
