package openvpn

import (
	"fmt"
	"strings"
)

// ClientConfig builds an .ovpn that always points at Server 1 (entry).
// Server 2 public IP must never appear here — traffic exits via GRE behind S1.
func ClientConfig(remoteHost string, port int, username string) string {
	if port <= 0 {
		port = 1194
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# VPN Manager %s\n", username)
	b.WriteString("# remote = Server 1 (entry). Server 2 is reachable only via GRE through S1.\n")
	b.WriteString("client\n")
	b.WriteString("dev tun\n")
	b.WriteString("proto udp\n")
	fmt.Fprintf(&b, "remote %s %d\n", remoteHost, port)
	b.WriteString("resolv-retry infinite\n")
	b.WriteString("nobind\n")
	b.WriteString("persist-key\n")
	b.WriteString("persist-tun\n")
	b.WriteString("remote-cert-tls server\n")
	b.WriteString("cipher AES-256-GCM\n")
	b.WriteString("auth SHA256\n")
	b.WriteString("verb 3\n")
	b.WriteString("# <ca>...</ca> <cert>...</cert> <key>...</key> injected after real PKI bootstrap\n")
	return b.String()
}

func ContainsForbiddenRemote(ovpn, server2PublicIP string) bool {
	if server2PublicIP == "" {
		return false
	}
	for _, line := range strings.Split(ovpn, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "remote ") && strings.Contains(line, server2PublicIP) {
			return true
		}
	}
	return false
}
