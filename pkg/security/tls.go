package security

import (
	"crypto/rand"
	"crypto/rsa"
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
// Each peer generates its own certificate upon joining the network.
func GenerateSelfSignedCertificate(peerID string) (tls.Certificate, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate private key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"P2P Network"},
			CommonName:   peerID, // Use the peerID as the CommonName for identification
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour), // Certificate is valid for 1 year

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

// CreateTLSConfig creates a TLS configuration with a self-signed certificate and optional certificate validation.
func CreateTLSConfig(peerID string) (*tls.Config, error) {
    cert, err := GenerateSelfSignedCertificate(peerID)
    if err != nil {
        return nil, fmt.Errorf("failed to generate self-signed certificate: %v", err)
    }

    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        ClientAuth:   tls.NoClientCert, // Do not require client certificates
        InsecureSkipVerify: true,       // Allow self-signed certificates (handle verification manually)
    }

    return tlsConfig, nil
}

// StartTLSListener starts a TLS listener that peers can connect to for secure communication.
func StartTLSListener(peerID string, address string) (net.Listener, error) {
	tlsConfig, err := CreateTLSConfig(peerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %v", err)
	}

	listener, err := tls.Listen("tcp", address, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to start TLS listener: %v", err)
	}

	log.Printf("TLS listener started on %s\n", address)
	return listener, nil
}

// DialTLS connects to a peer using TLS for secure communication.
func DialTLS(peerID string, address string) (net.Conn, error) {
	tlsConfig, err := CreateTLSConfig(peerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %v", err)
	}

	conn, err := tls.Dial("tcp", address, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial TLS connection: %v", err)
	}

	return conn, nil
}

// VerifyPeerCertificate is a custom certificate verification function to validate peer certificates in a decentralized manner.
func VerifyPeerCertificate(cert *x509.Certificate, trustedPeers map[string]string) bool {
	peerID := cert.Subject.CommonName

	// Check if peer's public key or certificate is known and trusted
	trustedCert, exists := trustedPeers[peerID]
	if !exists {
		log.Printf("Peer certificate for %s is not trusted.\n", peerID)
		return false
	}

	// Compare the provided certificate with the stored trusted certificate
	if trustedCert != string(cert.Raw) {
		log.Printf("Peer certificate mismatch for %s.\n", peerID)
		return false
	}

	return true
}

// Example code to store trusted certificates - it could be part of DHT or node's metadata.
func StoreTrustedPeerCertificate(peerID string, cert string, trustedPeers map[string]string) {
	// Save the trusted peer's certificate (for example, in-memory for now)
	trustedPeers[peerID] = cert
	log.Printf("Stored trusted certificate for peer %s.\n", peerID)
}