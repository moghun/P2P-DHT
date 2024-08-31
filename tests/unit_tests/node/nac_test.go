package tests

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/tests"
)

func loadTestTLSConfig(certFile string) (*tls.Config, error) {
	certPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate to pool")
	}

	return &tls.Config{
		RootCAs: certPool,
	}, nil
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

	// Load the TLS configuration with trusted self-signed certificate
	tlsConfig, err := loadTestTLSConfig(".certificates/cert.pem")
	assert.NoError(t, err)

	// Use the custom TLS config for the connection
	conn, err := tls.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", receiverPort), tlsConfig)
	assert.NoError(t, err)
	defer conn.Close()

	_, err = conn.Write([]byte("hello"))
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

	// Load the TLS configuration with trusted self-signed certificate
	tlsConfig, err := loadTestTLSConfig(".certificates/cert.pem")
	assert.NoError(t, err)

	conn, err := tls.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port), tlsConfig)
	assert.NoError(t, err)
	defer conn.Close()

	_, err = conn.Write([]byte("test"))
	assert.NoError(t, err)

	time.Sleep(time.Second)
}
