package util

import (
	"context"
	"fmt"
	"net"
)

const (
	LocalHostDNSResolver  = "127.0.0.1:53"
	CloudFlareDNSResolver = "1.1.1.1:53"
)

func ResolveHostname(ctx context.Context, hostname string, dnsServer ...string) (string, error) {
	// Resolve first locally, then with the provided DNS servers and finally with the CloudFlare DNS
	defaultResolver := append([]string{LocalHostDNSResolver}, dnsServer...)
	defaultResolver = append(defaultResolver, CloudFlareDNSResolver)

	for _, resolver := range defaultResolver {
		ip, err := resolveHostname(ctx, hostname, resolver)
		if err == nil {
			return ip, nil
		}
	}

	return "", fmt.Errorf("failed to resolve hostname: %s", hostname)
}

func resolveHostname(ctx context.Context, hostname, dnsServer string) (string, error) {
	// Use a dialer to respect the context timeout
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "udp", dnsServer)
	if err != nil {
		return "", fmt.Errorf("failed to dial DNS server: %v", err)
	}
	conn.Close()

	ips, err := net.LookupHost(hostname)
	if err != nil {
		return "", fmt.Errorf("failed to resolve IP from hostname: %v", err)
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no IP addresses found for hostname: %s", hostname)
	}

	return ips[0], nil
}

func IsIP(input string) bool {
	// Try to parse the input as an IP address
	ip := net.ParseIP(input)
	return ip != nil
}
