//go:build linux

package main

import (
	"context"
	"errors"
	"io"
)

func connectAuthenticatedPipe(context.Context, string, string, int) (io.ReadWriteCloser, error) {
	return nil, errors.New("Linux 使用 systemd 核心 Unix socket，不支持 serve-pipe 一次性 transport")
}
