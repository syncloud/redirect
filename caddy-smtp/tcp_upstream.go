package caddysmtp

import (
	"net"
	"time"

	"github.com/pires/go-proxyproto"
)

const dialTimeout = 30 * time.Second

type TcpUpstream struct {
	address       string
	proxyProtocol bool
}

func NewTcpUpstream(address string, proxyProtocol bool) *TcpUpstream {
	return &TcpUpstream{address: address, proxyProtocol: proxyProtocol}
}

func (u *TcpUpstream) Connect(remote net.Addr, local net.Addr) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", u.address, dialTimeout)
	if err != nil {
		return nil, err
	}
	if !u.proxyProtocol {
		return conn, nil
	}
	header := proxyproto.HeaderProxyFromAddrs(2, remote, local)
	if _, err := header.WriteTo(conn); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}
