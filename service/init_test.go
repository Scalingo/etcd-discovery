package service

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Run("client should be set", func(t *testing.T) {
		require.NotNil(t, Client())
	})

	t.Run("hostname should be set", func(t *testing.T) {
		require.NotNil(t, hostname)
	})

	t.Run("hostname should be set", func(t *testing.T) {
		assert.Equal(t, "[etcd-discovery] ", logger.Prefix())
		assert.Equal(t, log.LstdFlags, logger.Flags())
	})
}

func TestTLSConfigFromMemory(t *testing.T) {
	t.Run("Given a certificate, key and CA file in base64, It should return a tls.Config with client certificate", func(t *testing.T) {
		sampleCertB64, sampleKeyB64, sampleCAB64 := sampleCert()

		config, err := tlsconfigFromMemory(sampleCertB64, sampleKeyB64, sampleCAB64)
		require.NoError(t, err)
		require.NotNil(t, config)
		assert.Len(t, config.Certificates, 1)
	})
}

func sampleCert() (string, string, string) {
	capriv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}

	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	privpem := new(bytes.Buffer)
	pem.Encode(privpem, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	priv64 := base64.StdEncoding.EncodeToString(privpem.Bytes())

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"Acme Co"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now(),
		IsCA:         true,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	caCert, err := x509.CreateCertificate(rand.Reader, &template, &template, &capriv.PublicKey, capriv)
	if err != nil {
		panic(err)
	}

	capem := new(bytes.Buffer)
	pem.Encode(capem, &pem.Block{Type: "CERTIFICATE", Bytes: caCert})
	ca64 := base64.StdEncoding.EncodeToString(capem.Bytes())

	childTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{Organization: []string{"Acme Co"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now(),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certCert, err := x509.CreateCertificate(rand.Reader, &childTemplate, &template, &priv.PublicKey, capriv)
	if err != nil {
		panic(err)
	}

	certpem := new(bytes.Buffer)
	pem.Encode(certpem, &pem.Block{Type: "CERTIFICATE", Bytes: certCert})
	cert64 := base64.StdEncoding.EncodeToString(certpem.Bytes())

	return cert64, priv64, ca64
}
