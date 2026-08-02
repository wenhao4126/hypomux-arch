//go:build linux

package engineclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
)

const DefaultCoreServiceSocket = "/run/hypomux/hypomux-core.sock"

type linuxServiceLauncher struct{}

type linuxServiceProcess struct {
	done chan struct{}
	once sync.Once
}

func linuxCoreSocketPath() string {
	if value := os.Getenv("HYPOMUX_CORE_SOCKET"); value != "" {
		return value
	}
	return DefaultCoreServiceSocket
}

func (linuxServiceLauncher) Launch(ctx context.Context, _ string) (*coreSession, error) {
	path := linuxCoreSocketPath()
	connection, err := (&net.Dialer{}).DialContext(ctx, "unix", path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscallENOENT()) {
			return nil, ErrCoreServiceUnavailable
		}
		return nil, fmt.Errorf("连接 HypoMux Core Unix socket 失败：%w", err)
	}
	process := &linuxServiceProcess{done: make(chan struct{})}
	return &coreSession{
		reader: connection,
		writer: connection,
		close: func() error {
			process.signalClosed()
			return connection.Close()
		},
		process: process,
		path:    path,
	}, nil
}

func syscallENOENT() error {
	return os.ErrNotExist
}

func (p *linuxServiceProcess) Wait() error {
	<-p.done
	return nil
}

func (p *linuxServiceProcess) Kill() error {
	p.signalClosed()
	return nil
}

func (p *linuxServiceProcess) PID() int { return 0 }

func (p *linuxServiceProcess) signalClosed() {
	p.once.Do(func() { close(p.done) })
}
