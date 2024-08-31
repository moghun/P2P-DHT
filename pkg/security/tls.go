package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

type TLSManager struct {
	CertFile string
	KeyFile  string
}

func NewTLSManager(certFile, keyFile string) *TLSManager {
	return &TLSManager{
		CertFile: certFile,
		KeyFile:  keyFile,
	}
}

// GenerateSelfSignedCert generates a self-signed certificate if it doesn't exist.
func (t *TLSManager) GenerateSelfSignedCert() error {
	// Check if the certificate or key file does not exist
	if _, err := os.Stat(t.CertFile); os.IsNotExist(err) || os.IsNotExist(checkFile(t.KeyFile)) {
		fmt.Println("Generating new self-signed certificate and key...")
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return fmt.Errorf("failed to generate private key: %v", err)
		}

		notBefore := time.Now()
		notAfter := notBefore.Add(365 * 24 * time.Hour)

		serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
		if err != nil {
			return fmt.Errorf("failed to generate serial number: %v", err)
		}

		template := x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				Organization: []string{"P2P Network"},
			},
			NotBefore:             notBefore,
			NotAfter:              notAfter,
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}

		derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
		if err != nil {
			return fmt.Errorf("failed to create certificate: %v", err)
		}

		certOut, err := os.Create(t.CertFile)
		if err != nil {
			return fmt.Errorf("failed to open cert file for writing: %v", err)
		}
		pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		certOut.Close()

		keyOut, err := os.Create(t.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to open key file for writing: %v", err)
		}
		privBytes, err := x509.MarshalECPrivateKey(priv)
		if err != nil {
			return fmt.Errorf("failed to marshal private key: %v", err)
		}
		pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
		keyOut.Close()
	}

	return nil
}

// Helper function to check file existence and return an error
func checkFile(filename string) error {
	_, err := os.Stat(filename)
	return err
}

// LoadTLSConfig loads the TLS configuration, using the self-signed certificate.
func (t *TLSManager) LoadTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(t.CertFile, t.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %v", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	return config, nil
}

// GetFingerprint returns the SHA-256 fingerprint of the certificate.
func (t *TLSManager) GetFingerprint() (string, error) {
	certData, err := os.ReadFile(t.CertFile)
	if err != nil {
		return "", fmt.Errorf("failed to read certificate file: %v", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil || block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("failed to decode PEM block containing certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %v", err)
	}

	fingerprint := sha256.Sum256(cert.Raw)
	return fmt.Sprintf("%x", fingerprint), nil
}