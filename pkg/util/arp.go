package util

import (
	"bufio"
	"fmt"
	"net"
	"net/netip"
	"os"
	"strings"

	"github.com/mdlayher/arp"
)

const (
	ARPCachePath = "/proc/net/arp"

	IPIdxARPCacheField    = 0
	MACIdxARPCacheField   = 3
	IfaceIdxARPCacheField = 5

	MaxARPCacheFields = 6
)

// todo(): replace from string to net.HardwareAddr

// GetMACFromARPCache retrieves the MAC address of an IP from the ARP cache for a specific network interface.
func GetMACFromARPCache(ip, ifaceName string) (string, error) {
	// Open the ARP cache file
	file, err := os.Open(ARPCachePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Scan the ARP cache
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < MaxARPCacheFields {
			continue
		}

		// Check if the IP matches and if the interface matches
		if fields[IPIdxARPCacheField] == ip && fields[IfaceIdxARPCacheField] == ifaceName {
			return fields[MACIdxARPCacheField], nil // MAC address field
		}
	}

	return "", fmt.Errorf("MAC address not found for IP: %s on interface: %s", ip, ifaceName)
}

// GetMACViaARPCall retrieves the MAC address of an IP via ARP call for a specific network interface.
func GetMACViaARPCall(ip string, ifaceName string) (string, error) {
	// Get the network interface by name
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "", fmt.Errorf("interface not found: %v", err)
	}

	// Open a connection to the ARP protocol for the given interface
	conn, err := arp.Dial(iface)
	if err != nil {
		return "", fmt.Errorf("failed to open ARP connection: %v", err)
	}
	defer conn.Close()

	// Parse the IP address we want to resolve to get MAC address
	targetIP, err := netip.ParseAddr(ip)
	if err != nil {
		return "", fmt.Errorf("invalid IP address format: %v", err)
	}

	// Send the ARP request and wait for the MAC address in response
	mac, err := conn.Resolve(targetIP)
	if err != nil {
		return "", fmt.Errorf("failed to resolve MAC address for IP %s: %v", ip, err)
	}

	return mac.String(), nil
}
