package mailrelay

import "crypto/tls"

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
