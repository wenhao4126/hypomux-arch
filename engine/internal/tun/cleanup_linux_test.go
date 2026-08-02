//go:build linux

package tun

import "testing"

func TestLinuxCleanupTargetsOnlyHypoMuxTunDefaults(t *testing.T) {
	commands := linuxCleanupCommands()
	if len(commands) != 3 {
		t.Fatalf("cleanup command count = %d", len(commands))
	}
	if commands[0][0] != "ip" || commands[0][1] != "route" || commands[0][5] != "HypoMux-Tun" {
		t.Fatalf("IPv4 cleanup command = %#v", commands[0])
	}
	if commands[1][0] != "ip" || commands[1][1] != "-6" || commands[1][6] != "HypoMux-Tun" {
		t.Fatalf("IPv6 cleanup command = %#v", commands[1])
	}
	if commands[2][0] != "ip" || commands[2][1] != "link" || commands[2][3] != "HypoMux-Tun" {
		t.Fatalf("link cleanup command = %#v", commands[2])
	}
}
