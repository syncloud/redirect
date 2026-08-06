package mail

import (
	"crypto/tls"
	"errors"
	"os"
	"sync"
	"time"
)

// ErrCertificateMissing means the certificate is configured but has not been
// put in place yet, which is expected before caddy has obtained one
var ErrCertificateMissing = errors.New("certificate file is not there yet")

type CertificateLoader struct {
	certFile string
	keyFile  string

	mutex   sync.Mutex
	loaded  *tls.Certificate
	modTime time.Time
}

func NewCertificateLoader(certFile string, keyFile string) *CertificateLoader {
	return &CertificateLoader{certFile: certFile, keyFile: keyFile}
}

// Load returns nil when no certificate is configured, and an error when one is
// configured but will not parse. A configured file that is not there yet is
// neither: caddy has to obtain it before the first copy can run, so a new host
// starts without starttls and picks it up on the next restart.
func (c *CertificateLoader) Load() (*tls.Config, error) {
	if c.certFile == "" || c.keyFile == "" {
		return nil, nil
	}
	if _, err := os.Stat(c.certFile); errors.Is(err, os.ErrNotExist) {
		return nil, ErrCertificateMissing
	}
	if _, err := c.certificate(); err != nil {
		return nil, err
	}
	return &tls.Config{
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			return c.certificate()
		},
	}, nil
}

func (c *CertificateLoader) certificate() (*tls.Certificate, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	info, err := os.Stat(c.certFile)
	if err != nil {
		return c.fallback(err)
	}
	if c.loaded != nil && info.ModTime().Equal(c.modTime) {
		return c.loaded, nil
	}
	pair, err := tls.LoadX509KeyPair(c.certFile, c.keyFile)
	if err != nil {
		return c.fallback(err)
	}
	c.loaded = &pair
	c.modTime = info.ModTime()
	return c.loaded, nil
}

func (c *CertificateLoader) fallback(err error) (*tls.Certificate, error) {
	if c.loaded != nil {
		return c.loaded, nil
	}
	return nil, err
}
