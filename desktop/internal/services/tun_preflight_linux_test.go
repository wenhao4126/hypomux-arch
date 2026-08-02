//go:build linux

package services

import "testing"

func TestParseLinuxDefaultRouteAliasesOnlyReportsTunnelInterfaces(t *testing.T) {
	data := "Iface Destination Gateway Flags RefCnt Use Metric Mask MTU Window IRTT\n" +
		"eth0 00000000 0101A8C0 0003 0 0 100 00000000 0 0 0\n" +
		"wg0 00000000 00000000 0001 0 0 50 00000000 0 0 0\n"
	aliases := parseLinuxDefaultRouteAliases(data)
	if len(aliases) != 1 || aliases[0] != "wg0" {
		t.Fatalf("aliases = %#v", aliases)
	}
}
