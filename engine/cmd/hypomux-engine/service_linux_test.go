//go:build linux

package main

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Hypostasis-Cat/HypoMux/engine/internal/server"
)

func TestLinuxServiceAcceptsCurrentUserAndRunsProtocol(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "hypomux.sock")
	listener, err := listenLinuxService(socketPath)
	if err != nil {
		t.Fatalf("listenLinuxService() error = %v", err)
	}
	defer listener.Close()

	serveDone := make(chan error, 1)
	go func() {
		connection, acceptErr := acceptLinuxService(context.Background(), listener)
		if acceptErr != nil {
			serveDone <- acceptErr
			return
		}
		serveDone <- server.New(connection, connection, server.Metadata{Name: "test"}).Run(context.Background())
	}()

	connection, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer connection.Close()
	if _, err := connection.Write([]byte(`{"protocol":1,"id":"hello","method":"engine.hello"}` + "\n")); err != nil {
		t.Fatalf("write request: %v", err)
	}
	connection.SetReadDeadline(time.Now().Add(time.Second))
	buffer := make([]byte, 4096)
	count, err := connection.Read(buffer)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if count == 0 {
		t.Fatal("empty response")
	}
	_ = connection.Close()
	if err := <-serveDone; err != nil {
		t.Fatalf("server session: %v", err)
	}

	info, err := os.Stat(socketPath)
	if err != nil {
		t.Fatalf("stat socket: %v", err)
	}
	if info.Mode().Perm()&0007 != 0 {
		t.Fatalf("socket permissions = %o, want no access for other users", info.Mode().Perm())
	}
}
