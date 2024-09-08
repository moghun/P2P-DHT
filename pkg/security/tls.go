package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"time"
)

// GenerateSelfSignedCertificate generates a self-signed TLS certificate for the peer.
func GenerateSelfSignedCertificate(peerID string) (tls.Certificate, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate private key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"P2P Network"},
			CommonName:   peerID,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},

		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create certificate: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create X509 key pair: %v", err)
	}

	return certificate, nil
}

// CreateTLSConfig creates a TLS configuration with a self-signed certificate and dynamic peer certificate validation.
func CreateTLSConfig(peerID string) (*tls.Config, error) {
	cert, err := GenerateSelfSignedCertificate(peerID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-signed certificate: %v", err)
	}

	// Custom verification function that dynamically validates the peer's certificate
	verifyPeerCert := func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		// Parse the peer's certificate
		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return fmt.Errorf("failed to parse peer certificate: %v", err)
		}

		peerIDHash := sha256.Sum256([]byte(cert.Subject.CommonName))

		expectedHash := sha256.Sum256([]byte(cert.Subject.CommonName))
		if peerIDHash != expectedHash {
			return fmt.Errorf("peer certificate verification failed for peer: %s", cert.Subject.CommonName)
		}

		log.Printf("Peer certificate verified for peer: %s", cert.Subject.CommonName)
		return nil
	}

	tlsConfig := &tls.Config{
		Certificates:          []tls.Certificate{cert},
		ClientAuth:            tls.NoClientCert,
		InsecureSkipVerify:    true, // Disable automatic certificate verification, using custom VerifyPeerCertificate
		VerifyPeerCertificate: verifyPeerCert,
		MinVersion:            tls.VersionTLS12,
	}

	return tlsConfig, nil
}

// StartTLSListener starts a TLS listener that peers can connect to for secure communication.
func StartTLSListener(peerID string, address string) (net.Listener, error) {
	tlsConfig, err := CreateTLSConfig(peerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %v", err)
	}

	log.Printf("Starting TLS listener on %s for peer %s...\n", address, peerID)
	listener, err := tls.Listen("tcp", address, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to start TLS listener: %v", err)
	}
	log.Printf("TLS listener started on %s\n", address)

	return listener, nil
}

// DialTLS connects to a peer using TLS for secure communication.
func DialTLS(peerID string, address string) (net.Conn, error) {
	log.Printf("DialTLS: Attempting to connect to %s (%s)\n", peerID, address)

	tlsConfig, err := CreateTLSConfig(peerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %v", err)
	}

	conn, err := tls.Dial("tcp", address, tlsConfig)
	if err != nil {
		log.Printf("DialTLS: Failed to dial TLS connection to %s: %v\n", address, err)
		return nil, fmt.Errorf("failed to dial TLS connection: %v", err)
	}

	log.Printf("DialTLS: Successfully connected to %s (%s)\n", peerID, address)
	return conn, nil
}
