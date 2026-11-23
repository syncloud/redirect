package stun

import (
	"errors"
	"net"

	"github.com/pion/stun/v3"
	"go.uber.org/zap"
)

type Server struct {
	conn   net.PacketConn
	logger *zap.Logger
}

func NewServer(conn net.PacketConn, logger *zap.Logger) *Server {
	return &Server{conn: conn, logger: logger}
}

func (s *Server) Serve() error {
	buf := make([]byte, 1500)
	for {
		n, addr, err := s.conn.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		response, ok := Respond(buf[:n], addr)
		if !ok {
			continue
		}
		if _, err := s.conn.WriteTo(response, addr); err != nil {
			s.logger.Info("stun reply failed", zap.String("addr", addr.String()), zap.Error(err))
		}
	}
}

func Respond(packet []byte, addr net.Addr) ([]byte, bool) {
	udp, ok := addr.(*net.UDPAddr)
	if !ok {
		return nil, false
	}
	if !stun.IsMessage(packet) {
		return nil, false
	}
	request := &stun.Message{Raw: append([]byte{}, packet...)}
	if err := request.Decode(); err != nil {
		return nil, false
	}
	if request.Type != stun.BindingRequest {
		return nil, false
	}
	response, err := stun.Build(request, stun.BindingSuccess,
		&stun.XORMappedAddress{IP: udp.IP, Port: udp.Port},
		stun.Fingerprint,
	)
	if err != nil {
		return nil, false
	}
	return response.Raw, true
}
