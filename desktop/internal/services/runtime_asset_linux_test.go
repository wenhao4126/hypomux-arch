//go:build linux

package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRuntimeAssetFindsLinuxPackageExecutable(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "sing-box")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory)
	got, err := resolveRuntimeAsset("sing-box")
	if err != nil {
		t.Fatalf("resolveRuntimeAsset() error = %v", err)
	}
	if got != path {
		t.Fatalf("resolveRuntimeAsset() = %q, want %q", got, path)
	}
}
