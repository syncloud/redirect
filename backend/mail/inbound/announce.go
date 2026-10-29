package inbound

import (
	"fmt"
	"net"

	"github.com/pires/go-proxyproto"
)

const DevicePort = 25

func Announce(connection net.Conn, hostname string, client Client) (string, error) {
	source := net.ParseIP(client.Address)
	if source == nil {
		return "", fmt.Errorf("client address %q is not an ip", client.Address)
	}
	destination := net.IPv6unspecified
	if v4 := source.To4(); v4 != nil {
		source = v4
		destination = net.IPv4zero
	}
	header := proxyproto.HeaderProxyFromAddrs(2, &net.TCPAddr{IP: source}, &net.TCPAddr{IP: destination, Port: DevicePort})
	if _, err := header.WriteTo(connection); err != nil {
		return "", err
	}
	if client.Helo != "" {
		return client.Helo, nil
	}
	return hostname, nil
}
