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
	DefaultMDNSResponseBuffer = 512

	MaxMDNSReadTimeout = 10 * time.Second
)

const (
	mdnsPacketTransactionID  = 0x0000 // Transaction ID (0)
	mdnsPacketFlags          = 0x0000 // mdnsPacketFlags
	mdnsPacketQuestions      = 0x0001 // Number of questions (1)
	mdnsPacketAnswerRRs      = 0x0000 // Number of answer resource records (0)
	mdnsPacketAuthorityRRs   = 0x0000 // Number of authority resource records (0)
	mdnsPacketAdditionalRRs  = 0x0000 // Number of additional resource records (0)
	mdnsPacketNullTerminator = 0x00   // Null terminator for domain name
	mdnsPacketTypeA          = 0x0001 // Type A (IPv4 address)
	mdnsPacketClassIN        = 0x0001 // Class IN (Internet)

	Shift8      = 8    // Used for extracting the high byte in a 16-bit value
	LowByteMask = 0xFF // Mask to extract the lower byte of a 16-bit value
)

// ResolveMulticastDNS resolves a hostname using the multicast DNS protocol. Hostnames must be suffixed with ".local"
func ResolveMulticastDNS(ctx context.Context, hostname string) ([]net.IP, error) {
	// Validate and normalize hostname
	host, err := normalizeMDNSHostname(hostname)
	if err != nil {
		return nil, err
	}

	// Create an mDNS listener
	conn, err := net.ListenPacket(DefaultMDNSProtocol, DefaultMDNSAddress)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Resolve destination address
	dst, err := net.ResolveUDPAddr(DefaultMDNSProtocol, DefaultMDNSResolver)
	if err != nil {
		return nil, err
	}

	// Construct the mDNS query
	query := constructMDNSQuery(host)

	// Send the query and listen for a response
	return sendAndReceiveMDNS(ctx, conn, dst, query)
}

// ResolveDNS resolves a hostname using the provided DNS server. If no DNS server is provided, it uses the default
// CloudFlare DNS
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

// normalizeMDNSHostname ensures that the hostname is in the correct format for mDNS resolution
func normalizeMDNSHostname(hostname string) (string, error) {
	localTopLevel := fmt.Sprintf(".%s", DefaultMDNSTopLevelDomain)
	if !IsMultiCastDNS(hostname) {
		return "", fmt.Errorf("domains in mDNS must use .%s top level hostname", localTopLevel)
	}
	return strings.TrimSuffix(hostname, localTopLevel), nil
}

// constructMDNSQuery creates an mDNS query for the given hostname
func constructMDNSQuery(host string) []byte {
	query := []byte{
		byte(mdnsPacketTransactionID >> Shift8), byte(mdnsPacketTransactionID & LowByteMask),
		byte(mdnsPacketFlags >> Shift8), byte(mdnsPacketFlags & LowByteMask),
		byte(mdnsPacketQuestions >> Shift8), byte(mdnsPacketQuestions & LowByteMask),
		byte(mdnsPacketAnswerRRs >> Shift8), byte(mdnsPacketAnswerRRs & LowByteMask),
		byte(mdnsPacketAuthorityRRs >> Shift8), byte(mdnsPacketAuthorityRRs & LowByteMask),
		byte(mdnsPacketAdditionalRRs >> Shift8), byte(mdnsPacketAdditionalRRs & LowByteMask),
	}

	// Append the hostname in mDNS format
	for _, part := range append([]string{host}, DefaultMDNSTopLevelDomain) {
		query = append(query, byte(len(part)))
		query = append(query, []byte(part)...)
	}
	query = append(query, mdnsPacketNullTerminator)                                 // Null terminator
	query = append(query, byte(mdnsPacketTypeA>>Shift8), byte(mdnsPacketTypeA))     // Type A
	query = append(query, byte(mdnsPacketClassIN>>Shift8), byte(mdnsPacketClassIN)) // Class IN

	return query
}

// sendAndReceiveMDNS sends an mDNS query and waits for a response
func sendAndReceiveMDNS(ctx context.Context, conn net.PacketConn, dst *net.UDPAddr, query []byte) ([]net.IP, error) {
	respChan := make(chan []net.IP, 1)
	errChan := make(chan error, 1)

	// Send query
	go func() {
		_, err := conn.WriteTo(query, dst)
		if err != nil {
			errChan <- err
		}
	}()

	// Set read timeout
	if err := conn.SetReadDeadline(time.Now().Add(MaxMDNSReadTimeout)); err != nil {
		return nil, err
	}

	// Listen for response
	go func() {
		buffer := make([]byte, DefaultMDNSResponseBuffer)
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			errChan <- err
			return
		}
		if n < 12+16 {
			errChan <- fmt.Errorf("invalid mDNS response length %d", n)
			return
		}
		ip := net.IPv4(buffer[n-4], buffer[n-3], buffer[n-2], buffer[n-1])
		respChan <- []net.IP{ip}
	}()

	// Handle first available response
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errChan:
		return nil, err
	case ip := <-respChan:
		return ip, nil
	}
}

func resolveDNS(ctx context.Context, hostname, dnsServer string) ([]net.IP, error) {
	// Use a dialer to respect the context timeout
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "udp", dnsServer)
	if err != nil {
		return []net.IP{}, fmt.Errorf("failed to dial DNS server: %w", err)
	}
	conn.Close()

	ipsString, err := net.LookupHost(hostname)
	if err != nil {
		return []net.IP{}, fmt.Errorf("failed to resolve IP from hostname: %w", err)
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
