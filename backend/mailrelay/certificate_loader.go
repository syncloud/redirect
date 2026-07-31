package mailrelay

import "crypto/tls"

type CertificateLoader struct {
	certFile string
	keyFile  string
}

func NewCertificateLoader(certFile string, keyFile string) *CertificateLoader {
	return &CertificateLoader{certFile: certFile, keyFile: keyFile}
}

func (c *CertificateLoader) Load() (*tls.Config, error) {
	if c.certFile == "" || c.keyFile == "" {
		return nil, nil
	}
	pair, err := tls.LoadX509KeyPair(c.certFile, c.keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{Certificates: []tls.Certificate{pair}}, nil
}
