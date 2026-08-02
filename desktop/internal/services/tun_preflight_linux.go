//go:build linux

package services

import (
	"os"
	"strings"

	"github.com/Hypostasis-Cat/HypoMux/desktop/internal/engineclient"
)

func inspectTunPlatform(checkStrictRoute bool) tunPlatformSnapshot {
	snapshot := tunPlatformSnapshot{
		HostElevated:             os.Geteuid() == 0,
		PrivilegeBrokerAvailable: engineclient.PrivilegedLaunchSupported(),
		WFPReady:                 true,
		WFPDetail:                "Linux route safety uses TUN and kernel route checks; WFP is not used",
	}
	if checkStrictRoute {
		if _, err := os.Stat("/dev/net/tun"); err != nil {
			snapshot.WFPReady = false
			snapshot.WFPDetail = "Linux TUN device is unavailable: " + err.Error()
		}
	}
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		snapshot.RouteScanError = "读取 Linux 默认路由失败：" + err.Error()
		return snapshot
	}
	snapshot.DefaultRouteAliases = parseLinuxDefaultRouteAliases(string(data))
	return snapshot
}

func parseLinuxDefaultRouteAliases(data string) []string {
	seen := make(map[string]struct{})
	var result []string
	for name := range parseLinuxRouteMetadata(data) {
		lower := strings.ToLower(name)
		if !strings.Contains(lower, "tun") && !strings.Contains(lower, "wg") &&
			!strings.Contains(lower, "vpn") && !strings.Contains(lower, "tailscale") {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	return result
}
