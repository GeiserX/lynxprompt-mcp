package version

import "testing"

func TestString_returns_version_commit_and_date(t *testing.T) {
	// Defaults set in var block
	want := "dev (none) unknown"
	if got := String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestString_reflects_overridden_values(t *testing.T) {
	origV, origC, origD := Version, Commit, Date
	t.Cleanup(func() { Version, Commit, Date = origV, origC, origD })

	Version = "1.2.3"
	Commit = "abc123"
	Date = "2025-01-01"

	want := "1.2.3 (abc123) 2025-01-01"
	if got := String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
