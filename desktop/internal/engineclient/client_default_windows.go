//go:build windows

package engineclient

func newDefaultNormalLauncher() coreLauncher {
	return serviceFirstLauncher{service: windowsServiceLauncher{}, fallback: stdioLauncher{}}
}
