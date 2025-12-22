// Package netutil provides various net tools
package netutil

import (
	"net"
	"strings"
)

// IsValidPort check if the port is legal. 0 is considered as a non valid port.
func IsValidPort(port int) bool {
	return port > 0 && port < 65535
}

// GetIP get ip from net.Addr
func GetIP(addr net.Addr) string {
	if addr != nil {
		return strings.Split(addr.String(), ":")[0]
	}
	return ""
}

// GetPort get port from net.Addr
func GetPort(addr net.Addr) string {
	if addr != nil {
		addrs := strings.Split(addr.String(), ":")
		if len(addrs) == 2 {
			return addrs[1]
		}
	}
	return ""
}
