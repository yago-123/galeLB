package util

import (
	"net"
)

func IsValidIP(input string) bool {
	// Try to parse the input as an IP address
	ip := net.ParseIP(input)
	return ip != nil
}
