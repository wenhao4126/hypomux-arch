//go:build linux

package engineclient

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveExecutableFindsInstalledLinuxCoreOnPATH(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "hypomux-engine")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory)
	t.Setenv("HYPOMUX_ENGINE_PATH", "")
	got, err := ResolveExecutable()
	if err != nil {
		t.Fatalf("ResolveExecutable() error = %v", err)
	}
	if got != path {
		t.Fatalf("ResolveExecutable() = %q, want %q", got, path)
	}
}
