package util

import (
	"errors"
	"net"
)

func IsValidIP(input string) bool {
	// Try to parse the input as an IP address
	ip := net.ParseIP(input)
	return ip != nil
}

// GetIPv4FromInterface retrieves the first valid IPv4 address from the specified network interface
func GetIPv4FromInterface(ifaceName string) (string, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "", err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip.To4() != nil {
			return ip.String(), nil
		}
	}
	return "", errors.New("no valid IPv4 address found")
}
