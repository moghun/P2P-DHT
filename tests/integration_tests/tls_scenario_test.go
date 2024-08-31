package tests

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/tests"
)

func loadCertPool(certPath string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %v", err)
	}
	if !certPool.AppendCertsFromPEM(certPEM) {
		return nil, fmt.Errorf("failed to append certs from PEM")
	}
	return certPool, nil
}

func TestSendMessageSuccess(t *testing.T) {
	receiverPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	receiverNode := NewMockNode("127.0.0.1", receiverPort).ToNode()
	network := node.NewNetwork(receiverNode)

	go func() {
		err := network.StartListening()
		assert.NoError(t, err)
	}()

	time.Sleep(time.Second)

	senderPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	senderNode := NewMockNode("127.0.0.1", senderPort).ToNode()
	network = node.NewNetwork(senderNode)

	// Load the certificate from the receiver's TLS manager to trust the self-signed certificate
	certPool, err := loadCertPool(".certificates/cert.pem")
	assert.NoError(t, err, "Failed to load certificate pool")

	tlsConfig, err := network.Tlsm.LoadTLSConfig()
	assert.NoError(t, err, "Failed to load TLS configuration for client")

	tlsConfig.RootCAs = certPool

	err = network.SendMessage("127.0.0.1", receiverPort, []byte("hello"))
	assert.NoError(t, err)

	time.Sleep(time.Second)
}

func TestSendMessageFailure(t *testing.T) {
	senderPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	senderNode := NewMockNode("127.0.0.1", senderPort).ToNode()
	network := node.NewNetwork(senderNode)

	err = network.SendMessage("invalid_ip", senderPort+1, []byte("hello"))
	assert.Error(t, err)
}

func TestStartListening(t *testing.T) {
	port, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	receiverNode := NewMockNode("127.0.0.1", port).ToNode()
	network := node.NewNetwork(receiverNode)

	go func() {
		err := network.StartListening()
		assert.NoError(t, err)
	}()

	time.Sleep(time.Second)

	// Load the certificate from the receiver's TLS manager to trust the self-signed certificate
	certPool, err := loadCertPool(".certificates/cert.pem")
	assert.NoError(t, err, "Failed to load certificate pool")

	tlsConfig, err := network.Tlsm.LoadTLSConfig()
	assert.NoError(t, err, "Failed to load TLS configuration for client")

	tlsConfig.RootCAs = certPool

	conn, err := tls.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port), tlsConfig)
	assert.NoError(t, err)
	defer conn.Close()

	_, err = conn.Write([]byte("test"))
	assert.NoError(t, err)

	time.Sleep(time.Second)
}
