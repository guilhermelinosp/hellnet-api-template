package main

import "testing"

// The composition root is exercised end to end by the handler tests in
// internal/hello and by the platform tests in fast-platform; here we only keep
// the build metadata defaults honest, since the release pipeline overrides them
// through -ldflags.
func TestBuildMetadataDefaults(t *testing.T) {
	for name, v := range map[string]string{"version": version, "commit": commit, "date": date} {
		if v == "" {
			t.Errorf("%s must have a default so the binary reports something without -ldflags", name)
		}
	}
}
