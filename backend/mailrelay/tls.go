package mailrelay

import "crypto/tls"

// LoadTls returns nil when no certificate is configured, which leaves the relay
// listening without STARTTLS for local testing only.
func LoadTls(certFile string, keyFile string) (*tls.Config, error) {
	if certFile == "" || keyFile == "" {
		return nil, nil
	}
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{Certificates: []tls.Certificate{certificate}}, nil
}
