package mailrelay

import "crypto/tls"

// Certificate is the key pair the relay presents to clients. Leaving it
// unconfigured is the normal deployment: caddy terminates tls on the public
// port and forwards to the relay over loopback, so there is nothing for the
// relay itself to present.
type Certificate struct {
	certFile string
	keyFile  string
}

func NewCertificate(certFile string, keyFile string) *Certificate {
	return &Certificate{certFile: certFile, keyFile: keyFile}
}

func (c *Certificate) Load() (*tls.Config, error) {
	if c.certFile == "" || c.keyFile == "" {
		return nil, nil
	}
	pair, err := tls.LoadX509KeyPair(c.certFile, c.keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{Certificates: []tls.Certificate{pair}}, nil
}
