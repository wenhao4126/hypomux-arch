//go:build !windows && !linux

package tun

import "context"

func cleanupPlatform(context.Context) error {
	return nil
}
