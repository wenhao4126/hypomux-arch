//go:build !windows && !linux

package engineclient

func newDefaultNormalLauncher() coreLauncher { return stdioLauncher{} }
