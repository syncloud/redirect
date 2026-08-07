package inbound

import "net"

type DeviceDialer interface {
	Dial(domain string) (net.Conn, error)
}
