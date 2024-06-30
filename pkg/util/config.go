package util

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

type Config struct {
	Address string
	DHT     DHTConfig
}

type DHTConfig struct {
	P2PAddress string `ini:"p2p address"`
	APIAddress string `ini:"api address"`
}

// LoadConfig reads configuration from the specified file and returns a Config object.
func LoadConfig(filename string) *Config {
	cfg, err := ini.Load(filename)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	p2pAddress := cfg.Section("dht").Key("p2p address").String()
	apiAddress := cfg.Section("dht").Key("api address").String()
	log.Printf("Loaded P2P Address: %s", p2pAddress)
	log.Printf("Loaded API Address: %s", apiAddress)

	return &Config{
		Address: cfg.Section("").Key("address").String(),
		DHT: DHTConfig{
			P2PAddress: p2pAddress,
			APIAddress: apiAddress,
		},
	}
}

// GetP2PIPPort parses and returns the IP and port from the P2P address.
func (d *DHTConfig) GetP2PIPPort() (string, int) {
	parts := strings.Split(d.P2PAddress, ":")
	if len(parts) < 2 {
		log.Fatalf("Invalid P2P address format in config")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Fatalf("Invalid port number in P2P address: %v", err)
	}
	log.Printf("Parsed P2P IP: %s, Port: %d", parts[0], port)
	return parts[0], port
}

// GetAPIIPPort parses and returns the IP and port from the API address.
func (d *DHTConfig) GetAPIIPPort() (string, int) {
	parts := strings.Split(d.APIAddress, ":")
	if len(parts) < 2 {
		log.Fatalf("Invalid API address format in config")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Fatalf("Invalid port number in API address: %v", err)
	}
	log.Printf("Parsed API IP: %s, Port: %d", parts[0], port)
	return parts[0], port
}

func GenerateCertificates(ip string, port int) (string, string, error) {
    tlsDir := fmt.Sprintf("%s_%d", ip, port)
    certsDir := filepath.Join("certificates", tlsDir)
    caDir := filepath.Join("certificates", "CA")

    // Ensure CA directory exists
    if _, err := os.Stat(caDir); os.IsNotExist(err) {
        if err := os.MkdirAll(caDir, os.ModePerm); err != nil {
            return "", "", fmt.Errorf("failed to create CA directory: %v", err)
        }
    }

    // Ensure CA files exist
    caCertFile := filepath.Join(caDir, "ca.pem")
    caKeyFile := filepath.Join(caDir, "ca.key")
    if _, err := os.Stat(caCertFile); os.IsNotExist(err) {
        return "", "", fmt.Errorf("CA certificate not found: %s", caCertFile)
    }
    if _, err := os.Stat(caKeyFile); os.IsNotExist(err) {
        return "", "", fmt.Errorf("CA key not found: %s", caKeyFile)
    }

    // Ensure certificates directory exists
    if _, err := os.Stat(certsDir); os.IsNotExist(err) {
        if err := os.MkdirAll(certsDir, os.ModePerm); err != nil {
            return "", "", err
        }
    }

    keyFile := filepath.Join(certsDir, fmt.Sprintf("%s_%d.key", ip, port))
    csrFile := filepath.Join(certsDir, fmt.Sprintf("%s_%d.csr", ip, port))
    certFile := filepath.Join(certsDir, fmt.Sprintf("%s_%d.crt", ip, port))
    confFile := filepath.Join(certsDir, "openssl.cnf") // Configuration file for OpenSSL

    // Generate RSA key
    err := exec.Command("openssl", "genpkey", "-algorithm", "RSA", "-out", keyFile).Run()
    if err != nil {
        return "", "", fmt.Errorf("error generating key: %v", err)
    }

    // Prepare OpenSSL configuration to include SAN
    configContents := `[req]
						prompt = no
						distinguished_name = req_distinguished_name
						req_extensions = req_ext
						[req_distinguished_name]
						CN = ` + ip + `
						[req_ext]
						subjectAltName = @alt_names
						[alt_names]
						IP.1 = ` + ip

    err = os.WriteFile(confFile, []byte(configContents), 0644)
    if err != nil {
        return "", "", fmt.Errorf("error writing OpenSSL config: %v", err)
    }

    // Generate CSR with config
    err = exec.Command("openssl", "req", "-new", "-key", keyFile, "-out", csrFile, "-config", confFile).Run()
    if err != nil {
        return "", "", fmt.Errorf("error generating CSR: %v", err)
    }

    // Generate certificate with CA and config
    cmd := exec.Command("openssl", "x509", "-req", "-in", csrFile, "-CA", caCertFile, "-CAkey", caKeyFile, "-CAcreateserial", "-out", certFile, "-days", "365", "-extfile", confFile, "-extensions", "req_ext")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    err = cmd.Run()
    if err != nil {
        return "", "", fmt.Errorf("error generating certificate: %v", err)
    }

    return certFile, keyFile, nil
}