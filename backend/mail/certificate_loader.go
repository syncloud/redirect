package mail

import (
	"crypto/tls"
	"errors"
	"os"
	"sync"
	"time"
)

var (
	ErrCertificateNotConfigured = errors.New("no certificate is configured")
	ErrCertificateMissing       = errors.New("certificate file is not there yet")
)

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

func (c *CertificateLoader) Load() (*tls.Config, error) {
	if c.certFile == "" || c.keyFile == "" {
		return nil, ErrCertificateNotConfigured
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

// Await blocks until the certificate can be used, reporting each attempt that
// could not. It gives up only when stopped.
func (c *CertificateLoader) Await(stopped <-chan struct{}, retry time.Duration,
	waiting func(error)) *tls.Config {
	for {
		config, err := c.Load()
		if err == nil {
			return config
		}
		waiting(err)
		select {
		case <-stopped:
			return nil
		case <-time.After(retry):
		}
	}
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
