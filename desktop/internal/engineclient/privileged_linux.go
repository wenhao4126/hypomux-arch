//go:build linux

package engineclient

import (
	"errors"
)

var ErrElevationCancelled = errors.New("用户取消了管理员权限请求")

func newPrivilegedLauncher() coreLauncher { return linuxServiceLauncher{} }

func PrivilegedLaunchSupported() bool { return true }
