//go:build linux

package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type linuxProxySnapshot struct {
	Mode      string `json:"mode"`
	HTTPHost  string `json:"http_host"`
	HTTPPort  string `json:"http_port"`
	HTTPSHost string `json:"https_host"`
	HTTPSPort string `json:"https_port"`
	SOCKSHost string `json:"socks_host"`
	SOCKSPort string `json:"socks_port"`
}

func proxyMarkerPath() string { return filepath.Join(settingsDirectory(), "proxy-owned") }

const gsettingsSchema = "org.gnome.system.proxy"

func enableSystemProxy(httpPort int, socksPort int) error {
	if _, err := exec.LookPath("gsettings"); err != nil {
		return fmt.Errorf("未找到 gsettings，当前桌面环境不支持系统代理集成：%w", err)
	}
	snapshot, err := readLinuxProxySnapshot()
	if err != nil {
		return err
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("生成 Linux 代理恢复点失败：%w", err)
	}
	if err := atomicWriteFile(proxyMarkerPath(), data, 0o600); err != nil {
		return fmt.Errorf("保存 Linux 代理恢复点失败：%w", err)
	}
	values := map[string]string{
		"mode":       "'manual'",
		"http-host":  "'127.0.0.1'",
		"http-port":  fmt.Sprintf("%d", httpPort),
		"https-host": "'127.0.0.1'",
		"https-port": fmt.Sprintf("%d", httpPort),
		"socks-host": "'127.0.0.1'",
		"socks-port": fmt.Sprintf("%d", socksPort),
	}
	for key, value := range values {
		if err := setGSetting(key, value); err != nil {
			_, _ = restoreSystemProxyDetailed()
			return fmt.Errorf("写入 Linux 系统代理 %s 失败：%w", key, err)
		}
	}
	return nil
}

func restoreSystemProxy() error {
	_, err := restoreSystemProxyDetailed()
	return err
}

func restoreSystemProxyDetailed() (string, error) {
	data, err := os.ReadFile(proxyMarkerPath())
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("读取 Linux 代理恢复点失败：%w", err)
	}
	var snapshot linuxProxySnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return "", fmt.Errorf("解析 Linux 代理恢复点失败：%w", err)
	}
	values := map[string]string{
		"mode": snapshot.Mode, "http-host": snapshot.HTTPHost, "http-port": snapshot.HTTPPort,
		"https-host": snapshot.HTTPSHost, "https-port": snapshot.HTTPSPort,
		"socks-host": snapshot.SOCKSHost, "socks-port": snapshot.SOCKSPort,
	}
	for key, value := range values {
		if value != "" {
			if err := setGSetting(key, value); err != nil {
				return "", fmt.Errorf("恢复 Linux 系统代理 %s 失败：%w", key, err)
			}
		}
	}
	if err := os.Remove(proxyMarkerPath()); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("清理 Linux 代理恢复点失败：%w", err)
	}
	return "", nil
}

func readLinuxProxySnapshot() (linuxProxySnapshot, error) {
	read := func(key string) (string, error) {
		output, err := exec.Command("gsettings", "get", gsettingsSchema, key).Output()
		if err != nil {
			return "", fmt.Errorf("读取 gsettings %s 失败：%w", key, err)
		}
		return strings.TrimSpace(string(output)), nil
	}
	snapshot := linuxProxySnapshot{}
	for key, target := range map[string]*string{
		"mode": &snapshot.Mode, "http-host": &snapshot.HTTPHost, "http-port": &snapshot.HTTPPort,
		"https-host": &snapshot.HTTPSHost, "https-port": &snapshot.HTTPSPort,
		"socks-host": &snapshot.SOCKSHost, "socks-port": &snapshot.SOCKSPort,
	} {
		value, err := read(key)
		if err != nil {
			return linuxProxySnapshot{}, err
		}
		*target = value
	}
	return snapshot, nil
}

func setGSetting(key, value string) error {
	output, err := exec.Command("gsettings", "set", gsettingsSchema, key, value).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}
