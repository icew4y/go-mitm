package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"
)

type CertManager struct {
	caCert     *x509.Certificate
	caPrivyKey *rsa.PrivateKey
}

func NewCertManager() (*CertManager, error) {
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("rsa.GenerateKey: %v", err)
	}

	// Create CA cert
	caCert := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"Little Squid, INC."},
			Country:       []string{"US"},
			Province:      []string{"CA"},
			Locality:      []string{"Los Angeles"},
			StreetAddress: []string{"Barela Ave"},
			PostalCode:    []string{"91780"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, caCert, caCert, caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	ca, err := x509.ParseCertificate(caBytes)
	if err != nil {
		return nil, err
	}

	return &CertManager{
		caCert:     ca,
		caPrivyKey: caPrivKey,
	}, nil
}

func (cm *CertManager) GenerateCertificate(hostname string) (*tls.Certificate, error) {
	serverPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	serverCert := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: hostname,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(0, 3, 0), // Valid for 3 months
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
		DNSNames:    []string{hostname},
	}

	// Sign the server cert with our CA
	serverCertBytes, err := x509.CreateCertificate(rand.Reader, serverCert, cm.caCert, serverPrivKey.PublicKey, cm.caPrivyKey)
	if err != nil {
		return nil, err
	}

	return &tls.Certificate{
		Certificate: [][]byte{serverCertBytes, cm.caCert.Raw},
		PrivateKey:  serverPrivKey,
	}, nil
}
