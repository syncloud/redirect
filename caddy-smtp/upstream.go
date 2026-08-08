package caddysmtp

import "net"

type Upstream interface {
	Connect(remote net.Addr, local net.Addr) (net.Conn, error)
}
