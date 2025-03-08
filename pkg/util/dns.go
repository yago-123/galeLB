package util

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	DefaultCloudFlareDNSResolver = "1.1.1.1:53"

	DefaultMDNSResolver       = "224.0.0.251:5353"
	DefaultMDNSProtocol       = "udp4"
	DefaultMDNSAddress        = ":0"
	DefaultMDNSTopLevelDomain = "local"

	MDNSQueryTimeout = 2 * time.Second
)

// ResolveMulticastDNS resolves a hostname using the multicast DNS protocol. Hostnames must be suffixed with ".local"
func ResolveMulticastDNS(hostname string) ([]net.IP, error) {
	localTopLevel := fmt.Sprintf(".%s", DefaultMDNSTopLevelDomain)
	if !IsMultiCastDNS(hostname) {
		return nil, fmt.Errorf("domains in mDNS must use .%s top level hostname", localTopLevel)
	}

	host := strings.TrimSuffix(hostname, localTopLevel)

	conn, err := net.ListenPacket(DefaultMDNSProtocol, DefaultMDNSAddress) // Bind to any available port
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	dst, err := net.ResolveUDPAddr(DefaultMDNSProtocol, DefaultMDNSResolver)
	if err != nil {
		return nil, err
	}

	query := []byte{
		0x00, 0x00, // Transaction ID (0)
		0x00, 0x00, // Flags
		0x00, 0x01, // Questions: 1
		0x00, 0x00, // Answer RRs: 0
		0x00, 0x00, // Authority RRs: 0
		0x00, 0x00, // Additional RRs: 0
	}

	// Append the hostname in mDNS format
	for _, part := range append([]string{host}, DefaultMDNSTopLevelDomain) {
		query = append(query, byte(len(part)))
		query = append(query, []byte(part)...)
	}
	query = append(query, 0x00)       // Null terminator
	query = append(query, 0x00, 0x01) // Type A
	query = append(query, 0x00, 0x01) // Class IN

	_, err = conn.WriteTo(query, dst)
	if err != nil {
		return nil, err
	}

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(MDNSQueryTimeout))

	buffer := make([]byte, 512)
	n, _, err := conn.ReadFrom(buffer)
	if err != nil {
		return nil, err
	}

	// Extract IP address from response
	if n < 12+16 {
		return nil, fmt.Errorf("invalid mDNS response length %d", n)
	}
	ip := net.IPv4(buffer[n-4], buffer[n-3], buffer[n-2], buffer[n-1])
	return []net.IP{ip}, nil
}

func ResolveDNS(ctx context.Context, hostname string, dnsServer ...string) ([]net.IP, error) {
	// Resolve with the provided DNS server if any, otherwise, use the default CloudFlare DNS
	defaultResolver := dnsServer
	if len(defaultResolver) == 0 {
		defaultResolver = append(defaultResolver, DefaultCloudFlareDNSResolver)
	}

	for _, resolver := range defaultResolver {
		ips, err := resolveDNS(ctx, hostname, resolver)
		if err == nil {
			return ips, nil
		}
	}

	return []net.IP{}, fmt.Errorf("failed to resolve hostname: %s", hostname)
}

func IsMultiCastDNS(hostname string) bool {
	return strings.HasSuffix(hostname, fmt.Sprintf(".%s", DefaultMDNSTopLevelDomain))
}

func resolveDNS(ctx context.Context, hostname, dnsServer string) ([]net.IP, error) {
	// Use a dialer to respect the context timeout
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "udp", dnsServer)
	if err != nil {
		return []net.IP{}, fmt.Errorf("failed to dial DNS server: %v", err)
	}
	conn.Close()

	ipsString, err := net.LookupHost(hostname)
	if err != nil {
		return []net.IP{}, fmt.Errorf("failed to resolve IP from hostname: %v", err)
	}

	if len(ipsString) == 0 {
		return []net.IP{}, fmt.Errorf("no IP addresses found for hostname: %s", hostname)
	}

	ips := []net.IP{}
	for _, ip := range ipsString {
		if parsedIP := net.ParseIP(ip); parsedIP != nil {
			ips = append(ips, parsedIP)
		}
	}

	return ips, nil
}
