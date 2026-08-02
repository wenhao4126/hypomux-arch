//go:build linux

package services

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const linuxDiagnosticProbeCount = 10

type linuxDiagnosticProbe struct{}

func newDiagnosticProbe() diagnosticProbe { return linuxDiagnosticProbe{} }

func (linuxDiagnosticProbe) ICMP(ctx context.Context, source string, target string) icmpProbeResult {
	sourceIP := net.ParseIP(source).To4()
	targetIP := net.ParseIP(target).To4()
	if sourceIP == nil || targetIP == nil {
		return icmpProbeResult{Status: "unavailable", LossRate: 100, Note: "invalid IPv4 address"}
	}
	connection, err := icmp.ListenPacket("ip4:icmp", sourceIP.String())
	if err != nil {
		return icmpProbeResult{Status: "unavailable", LossRate: 100, Note: fmt.Sprintf("ICMP unavailable: %v", err)}
	}
	defer connection.Close()
	id := os.Getpid() & 0xffff
	var rtts []int
	for sequence := 0; sequence < linuxDiagnosticProbeCount; sequence++ {
		select {
		case <-ctx.Done():
			return summarizeLinuxICMP(rtts, sequence, "cancelled")
		default:
		}
		message := icmp.Message{Type: ipv4.ICMPTypeEcho, Code: 0, Body: &icmp.Echo{ID: id, Seq: sequence, Data: []byte("HypoMux-Diagnostic-Probe")}}
		packet, marshalErr := message.Marshal(nil)
		if marshalErr != nil {
			continue
		}
		started := time.Now()
		_ = connection.SetReadDeadline(started.Add(time.Second))
		if _, writeErr := connection.WriteTo(packet, &net.IPAddr{IP: targetIP}); writeErr != nil {
			continue
		}
		buffer := make([]byte, 1500)
		for {
			count, _, readErr := connection.ReadFrom(buffer)
			if readErr != nil {
				break
			}
			reply, parseErr := icmp.ParseMessage(1, buffer[:count])
			if parseErr != nil || reply.Type != ipv4.ICMPTypeEchoReply {
				continue
			}
			body, ok := reply.Body.(*icmp.Echo)
			if ok && body.ID == id && body.Seq == sequence {
				rtts = append(rtts, int(time.Since(started)/time.Millisecond))
			}
			break
		}
	}
	return summarizeLinuxICMP(rtts, linuxDiagnosticProbeCount, "")
}

func summarizeLinuxICMP(rtts []int, sent int, note string) icmpProbeResult {
	if sent <= 0 {
		return icmpProbeResult{Status: "unavailable", LossRate: 100, Note: note}
	}
	received, total, minimum, maximum := len(rtts), 0, 0, 0
	for index, value := range rtts {
		total += value
		if index == 0 || value < minimum {
			minimum = value
		}
		if value > maximum {
			maximum = value
		}
	}
	loss := ((sent - received) * 100) / sent
	average, jitter := 0, 0
	if received > 0 {
		average, jitter = total/received, maximum-minimum
	}
	status := "available"
	if loss >= 100 {
		status = "unavailable"
	} else if loss >= 5 || jitter > 100 {
		status = "unstable"
	}
	return icmpProbeResult{Status: status, LossRate: loss, AvgLatencyMS: average, JitterMS: jitter, Sent: sent, Received: received, Note: note}
}

func (linuxDiagnosticProbe) BoundTCP(ctx context.Context, adapter AdapterView) (bool, string) {
	for _, endpoint := range []string{"223.5.5.5:443", "1.1.1.1:443", "8.8.8.8:443"} {
		dialer := net.Dialer{Timeout: 2 * time.Second, LocalAddr: &net.TCPAddr{IP: net.ParseIP(adapter.Address)}}
		connection, err := dialer.DialContext(ctx, "tcp4", endpoint)
		if err == nil {
			local := connection.LocalAddr().String()
			_ = connection.Close()
			return true, fmt.Sprintf("TCP %s via %s (ifIndex %d)", endpoint, local, adapter.IfIndex)
		}
	}
	return false, fmt.Sprintf("绑定 TCP 失败：%s", adapter.Name)
}
