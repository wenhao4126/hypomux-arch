//go:build linux

package tun

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func linuxCleanupCommands() [][]string {
	return [][]string{
		{"ip", "route", "del", "default", "dev", "HypoMux-Tun"},
		{"ip", "-6", "route", "del", "default", "dev", "HypoMux-Tun"},
		{"ip", "link", "delete", "HypoMux-Tun"},
	}
}

func cleanupPlatform(ctx context.Context) error {
	if _, err := exec.LookPath("ip"); err != nil {
		return fmt.Errorf("Linux iproute2 不可用：%w", err)
	}
	for _, arguments := range linuxCleanupCommands() {
		output, err := exec.CommandContext(ctx, arguments[0], arguments[1:]...).CombinedOutput()
		if err == nil {
			continue
		}
		detail := strings.ToLower(string(output))
		if strings.Contains(detail, "cannot find device") || strings.Contains(detail, "no such process") || strings.Contains(detail, "cannot find") {
			continue
		}
		return fmt.Errorf("清理 HypoMux TUN 资源失败（%s）：%s", strings.Join(arguments, " "), strings.TrimSpace(string(output)))
	}
	return nil
}
