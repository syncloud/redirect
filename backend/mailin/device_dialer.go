package mailin

import "net"

// DeviceDialer opens a connection to the device that owns a domain. The device
// is behind the relay tunnel, so this is not a plain dial.
type DeviceDialer interface {
	Dial(domain string) (net.Conn, error)
}
