//go:build linux

package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

func linuxAutostartPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("无法定位用户目录：%w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "autostart", "hypomux.desktop"), nil
}

func SetAutostart(enabled bool) error {
	path, err := linuxAutostartPath()
	if err != nil {
		return err
	}
	if !enabled {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("关闭 Linux 开机自启失败：%w", err)
		}
		return nil
	}
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("无法定位 HypoMux 程序：%w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("创建 Linux 开机自启目录失败：%w", err)
	}
	content := "[Desktop Entry]\nType=Application\nName=HypoMux\nExec=" + executable + " --silent\nTerminal=false\nX-GNOME-Autostart-enabled=true\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("写入 Linux 开机自启文件失败：%w", err)
	}
	return nil
}

func AutostartEnabled() (bool, error) {
	path, err := linuxAutostartPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
