//go:build linux

package engineclient

func newDefaultNormalLauncher() coreLauncher {
	return serviceFirstLauncher{service: linuxServiceLauncher{}, fallback: stdioLauncher{}}
}
