//go:build linux

package services

import "testing"

func TestParseLinuxRouteMetadataSelectsDefaultRoute(t *testing.T) {
	data := "Iface\tDestination\tGateway\tFlags\tRefCnt\tUse\tMetric\tMask\tMTU\tWindow\tIRTT\n" +
		"eth0\t00000000\t0101A8C0\t0003\t0\t0\t100\t00000000\t0\t0\t0\n" +
		"eth0\t0001A8C0\t00000000\t0001\t0\t0\t100\t00FFFFFF\t0\t0\t0\n"
	metadata := parseLinuxRouteMetadata(data)
	value, ok := metadata["eth0"]
	if !ok {
		t.Fatal("missing eth0 metadata")
	}
	if value.Gateway != "192.168.1.1" || value.Metric != 100 || !value.AutoMetric {
		t.Fatalf("metadata = %+v", value)
	}
}

func TestParseLinuxResolvConfDeduplicatesNameservers(t *testing.T) {
	metadata := parseLinuxResolvConf("nameserver 1.1.1.1\nnameserver 1.1.1.1\nnameserver 8.8.8.8\n")
	if got := metadata.DNSServers; len(got) != 2 || got[0] != "1.1.1.1" || got[1] != "8.8.8.8" {
		t.Fatalf("DNS servers = %#v", got)
	}
}
