//go:build linux

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Hypostasis-Cat/HypoMux/engine/internal/server"
	"golang.org/x/sys/unix"
)

const defaultLinuxCoreSocket = "/run/hypomux/hypomux-core.sock"

func installWindowsService() error {
	return errors.New("Linux core is installed and managed by systemd")
}

func removeWindowsService() error {
	return errors.New("Linux core is removed and managed by the Arch package")
}

func linuxCoreSocketPath() string {
	if value := os.Getenv("HYPOMUX_CORE_SOCKET"); value != "" {
		return value
	}
	return defaultLinuxCoreSocket
}

func listenLinuxService(path string) (*net.UnixListener, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, fmt.Errorf("create Core socket directory: %w", err)
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("remove stale Core socket: %w", err)
	}
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		return nil, fmt.Errorf("listen Core Unix socket: %w", err)
	}
	if err := os.Chmod(path, 0600); err != nil {
		_ = listener.Close()
		_ = os.Remove(path)
		return nil, fmt.Errorf("restrict Core Unix socket: %w", err)
	}
	return listener, nil
}

func acceptLinuxService(ctx context.Context, listener *net.UnixListener) (net.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	connection, err := listener.AcceptUnix()
	if err != nil {
		return nil, err
	}
	if err := validateLinuxPeer(connection); err != nil {
		_ = connection.Close()
		return nil, err
	}
	return connection, nil
}

func validateLinuxPeer(connection *net.UnixConn) error {
	raw, err := connection.SyscallConn()
	if err != nil {
		return fmt.Errorf("read Core client socket: %w", err)
	}
	var credential *unix.Ucred
	var controlErr error
	if err := raw.Control(func(fd uintptr) {
		credential, controlErr = unix.GetsockoptUcred(int(fd), unix.SOL_SOCKET, unix.SO_PEERCRED)
	}); err != nil {
		return fmt.Errorf("inspect Core client credential: %w", err)
	}
	if controlErr != nil {
		return fmt.Errorf("inspect Core client credential: %w", controlErr)
	}
	if credential == nil || credential.Pid <= 0 {
		return errors.New("reject Core client without process identity")
	}
	return nil
}

func runWindowsService(stderr io.Writer, metadata server.Metadata) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	path := linuxCoreSocketPath()
	listener, err := listenLinuxService(path)
	if err != nil {
		fmt.Fprintf(stderr, "listen HypoMux Core Unix socket: %v\n", err)
		return 1
	}
	defer func() {
		_ = listener.Close()
		_ = os.Remove(path)
	}()

	for {
		connection, err := acceptLinuxService(ctx, listener)
		if err != nil {
			if ctx.Err() != nil {
				return 0
			}
			fmt.Fprintf(stderr, "accept Core client: %v\n", err)
			continue
		}
		if err := server.New(connection, connection, metadata).Run(ctx); err != nil && !errors.Is(err, net.ErrClosed) {
			fmt.Fprintf(stderr, "Core client session ended: %v\n", err)
		}
		_ = connection.Close()
	}
}
