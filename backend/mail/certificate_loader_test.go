package mail

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func writePair(t *testing.T, dir string, name string, modTime time.Time) (string, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: name},
		DNSNames:     []string{name},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	assert.NoError(t, err)

	certPath := filepath.Join(dir, "mx.crt")
	keyPath := filepath.Join(dir, "mx.key")

	certOut, err := os.Create(certPath)
	assert.NoError(t, err)
	assert.NoError(t, pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: der}))
	assert.NoError(t, certOut.Close())

	keyDer, err := x509.MarshalECPrivateKey(key)
	assert.NoError(t, err)
	keyOut, err := os.Create(keyPath)
	assert.NoError(t, err)
	assert.NoError(t, pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDer}))
	assert.NoError(t, keyOut.Close())

	assert.NoError(t, os.Chtimes(certPath, modTime, modTime))
	return certPath, keyPath
}

func commonName(t *testing.T, loader *CertificateLoader) string {
	t.Helper()
	config, err := loader.Load()
	assert.NoError(t, err)
	certificate, err := config.GetCertificate(nil)
	assert.NoError(t, err)
	leaf, err := x509.ParseCertificate(certificate.Certificate[0])
	assert.NoError(t, err)
	return leaf.Subject.CommonName
}

func TestCertificateLoader_NoPathsMeansNoTls(t *testing.T) {
	config, err := NewCertificateLoader("", "").Load()

	assert.NoError(t, err)
	assert.Nil(t, config)
}

func TestCertificateLoader_MissingFileIsNotYetThere(t *testing.T) {
	dir := t.TempDir()

	_, err := NewCertificateLoader(filepath.Join(dir, "nope.crt"), filepath.Join(dir, "nope.key")).Load()

	// caddy has to obtain it before the copy can run, so a new host has to be
	// able to start without it
	assert.ErrorIs(t, err, ErrCertificateMissing)
}

func TestCertificateLoader_UnparseableFileIsAnError(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "mx.crt")
	keyPath := filepath.Join(dir, "mx.key")
	assert.NoError(t, os.WriteFile(certPath, []byte("not a certificate"), 0644))
	assert.NoError(t, os.WriteFile(keyPath, []byte("not a key"), 0600))

	_, err := NewCertificateLoader(certPath, keyPath).Load()

	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrCertificateMissing)
}

func TestCertificateLoader_ServesTheCertificate(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := writePair(t, dir, "old.mx.syncloud.it", time.Now().Add(-time.Hour))

	assert.Equal(t, "old.mx.syncloud.it", commonName(t, NewCertificateLoader(certPath, keyPath)))
}

func TestCertificateLoader_PicksUpARenewedCertificate(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := writePair(t, dir, "old.mx.syncloud.it", time.Now().Add(-time.Hour))
	loader := NewCertificateLoader(certPath, keyPath)
	assert.Equal(t, "old.mx.syncloud.it", commonName(t, loader))

	writePair(t, dir, "new.mx.syncloud.it", time.Now())

	assert.Equal(t, "new.mx.syncloud.it", commonName(t, loader))
}

func TestCertificateLoader_KeepsTheOldCertificateWhenReloadFails(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := writePair(t, dir, "old.mx.syncloud.it", time.Now().Add(-time.Hour))
	loader := NewCertificateLoader(certPath, keyPath)
	assert.Equal(t, "old.mx.syncloud.it", commonName(t, loader))

	assert.NoError(t, os.WriteFile(certPath, []byte("not a certificate"), 0644))

	assert.Equal(t, "old.mx.syncloud.it", commonName(t, loader))
}
