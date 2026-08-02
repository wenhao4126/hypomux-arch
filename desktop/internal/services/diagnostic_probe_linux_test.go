//go:build linux

package services

import (
	"context"
	"testing"
)

func TestLinuxDiagnosticRejectsInvalidSourceWithoutOpeningICMP(t *testing.T) {
	result := (linuxDiagnosticProbe{}).ICMP(context.Background(), "not-an-ip", "223.5.5.5")
	if result.Status != "unavailable" || result.Sent != 0 || result.Note != "invalid IPv4 address" {
		t.Fatalf("result = %+v", result)
	}
}
