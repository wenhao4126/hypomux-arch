//go:build linux

package engineclient

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLinuxServiceLauncherReportsUnavailableSocket(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "missing.sock")
	t.Setenv("HYPOMUX_CORE_SOCKET", socketPath)

	_, err := (linuxServiceLauncher{}).Launch(context.Background(), "unused")
	if !errors.Is(err, ErrCoreServiceUnavailable) {
		t.Fatalf("Launch() error = %v, want ErrCoreServiceUnavailable", err)
	}
	if _, statErr := os.Stat(socketPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("test socket unexpectedly exists: %v", statErr)
	}
}
