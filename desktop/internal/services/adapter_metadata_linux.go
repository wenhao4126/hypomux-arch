//go:build linux

package services

import (
	"bufio"
	"encoding/hex"
	"net"
	"os"
	"strconv"
	"strings"
)

func adapterPlatformMetadata() map[int]adapterMetadata {
	result := make(map[int]adapterMetadata)
	if data, err := os.ReadFile("/proc/net/route"); err == nil {
		for name, value := range parseLinuxRouteMetadata(string(data)) {
			if iface, err := net.InterfaceByName(name); err == nil {
				result[iface.Index] = value
			}
		}
	}
	dns := parseLinuxResolvConfFile("/etc/resolv.conf")
	for _, iface := range interfacesForMetadata() {
		value := result[iface.Index]
		if len(value.DNSServers) == 0 {
			value.DNSServers = append([]string(nil), dns...)
		}
		if value.Metric == 0 && value.Gateway == "" {
			value.Metric = -1
			value.AutoMetric = true
		}
		result[iface.Index] = value
	}
	return result
}

func interfacesForMetadata() []net.Interface {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	return interfaces
}

func parseLinuxRouteMetadata(data string) map[string]adapterMetadata {
	result := make(map[string]adapterMetadata)
	scanner := bufio.NewScanner(strings.NewReader(data))
	if !scanner.Scan() {
		return result
	}
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 8 || fields[1] != "00000000" {
			continue
		}
		flags, err := strconv.ParseUint(fields[3], 16, 32)
		if err != nil || flags&1 == 0 {
			continue
		}
		gateway, err := linuxHexIPv4(fields[2])
		if err != nil {
			continue
		}
		metric, err := strconv.Atoi(fields[6])
		if err != nil {
			continue
		}
		result[fields[0]] = adapterMetadata{Gateway: gateway, Metric: metric, AutoMetric: true}
	}
	return result
}

func linuxHexIPv4(value string) (string, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 4 {
		if err == nil {
			err = strconv.ErrSyntax
		}
		return "", err
	}
	return net.IPv4(decoded[3], decoded[2], decoded[1], decoded[0]).String(), nil
}

func parseLinuxResolvConfFile(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return parseLinuxResolvConf(string(data)).DNSServers
}

func parseLinuxResolvConf(data string) adapterMetadata {
	result := adapterMetadata{}
	seen := make(map[string]struct{})
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 || fields[0] != "nameserver" {
			continue
		}
		ip := net.ParseIP(fields[1])
		if ip == nil || ip.IsUnspecified() {
			continue
		}
		value := ip.String()
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result.DNSServers = append(result.DNSServers, value)
	}
	return result
}
