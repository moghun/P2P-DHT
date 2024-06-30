package tests

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/networking"
)

func generateKeyPair(bits int) (*rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func createCertificate(certFile, keyFile string) error {
	priv, err := generateKeyPair(2048)
	if err != nil {
		return err
	}

	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	if err != nil {
		return err
	}

	certTemplate := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer certOut.Close()

	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return err
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}

	return nil
}

func TestStartServer(t *testing.T) {
    key := make([]byte, 32)
    _, err := rand.Read(key)
    assert.NoError(t, err)

    dhtInstance := dht.NewDHT()
    node := dht.NewNode("127.0.0.1", 0, false, key)
    assert.NotNil(t, node)

    network := networking.NewNetwork(dhtInstance)
    assert.NotNil(t, network)

    go func() {
        err := network.StartServer("127.0.0.1", 0)
        assert.NoError(t, err)
    }()

    time.Sleep(1 * time.Second) // Ensure server starts

    port := network.GetListeningPort()
    assert.NotZero(t, port)

    tlsConfig, _ := network.LoadTLSConfig(network.ServerCertFile, network.ServerKeyFile)
    conn, err := tls.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port), tlsConfig)
    assert.NoError(t, err)
    defer conn.Close()
}

func TestSendMessage(t *testing.T) {
    key := make([]byte, 32)
    _, err := rand.Read(key)
    assert.NoError(t, err)

    dhtInstance := dht.NewDHT()
    node := dht.NewNode("127.0.0.1", 0, false, key)
    assert.NotNil(t, node)

    network := networking.NewNetwork(dhtInstance)
    assert.NotNil(t, network)

    go func() {
        err := network.StartServer("127.0.0.1", 0)
        assert.NoError(t, err)
    }()

    time.Sleep(1 * time.Second) // Ensure server starts

    port := network.GetListeningPort()
    assert.NotZero(t, port)

    msg := message.NewMessage(uint16(4+len([]byte("ping"))), message.DHT_PING, []byte("ping"))
    serializedMsg, _ := msg.Serialize()
    err = network.SendMessage("127.0.0.1", port, serializedMsg, node.CertFile, node.KeyFile)
    assert.NoError(t, err)
}


func TestHandleConnection(t *testing.T) {
	key := make([]byte, 32) // AES-256 key size
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}
	message.SetEncryptionKey(key)

	dhtInstance := dht.NewDHT()

	node := dht.NewNode("127.0.0.1", 8002, false, key)
	dhtInstance.JoinNetwork(node)

	network := networking.NewNetwork(dhtInstance)

	errChan := make(chan error)

	go func() {
		if err := network.StartServer("127.0.0.1", 8002); err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}
	case <-time.After(1 * time.Second):
		// Server started successfully
	}

	msg := message.NewMessage(uint16(4+len([]byte("ping"))), message.DHT_PING, []byte("ping"))

	serializedMsg, err := msg.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize message: %v", err)
	}

	tlsConfig, err := network.LoadTLSConfig(network.ServerCertFile, network.ServerKeyFile)
	conn, err := tls.Dial("tcp", "127.0.0.1:8002", tlsConfig)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(serializedMsg)
	if err != nil {
		t.Fatalf("Failed to write message to connection: %v", err)
	}

	data := make([]byte, 1024)
	length, err := conn.Read(data)
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}

	responseMsg, err := message.DeserializeMessage(data[:length])
	if err != nil {
		t.Fatalf("Failed to deserialize response: %v", err)
	}

	expectedData := "ping"
	if string(responseMsg.Data) != expectedData {
		t.Errorf("Expected message data %s, got %s", string(expectedData), string(responseMsg.Data))
	}
}

func TestLoadTLSConfig(t *testing.T) {
	certFile := "../certificates/test.crt"
	keyFile := "../certificates/test.key"

	// Ensure the certificates directory exists
	if _, err := os.Stat("../certificates"); os.IsNotExist(err) {
		err := os.Mkdir("../certificates", 0755)
		if err != nil {
			t.Fatalf("Failed to create certificates directory: %v", err)
		}
	}

	err := createCertificate(certFile, keyFile)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}
	defer os.Remove(certFile)
	defer os.Remove(keyFile)

	network := networking.NewNetwork(nil)

	_, err = network.LoadTLSConfig(certFile, keyFile)
	if err != nil {
		t.Fatalf("Failed to load TLS config: %v", err)
	}
}

func TestGetListeningPort(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}

	dhtInstance := dht.NewDHT()

	node := dht.NewNode("127.0.0.1", 0, false, key)
	dhtInstance.JoinNetwork(node)

	network := networking.NewNetwork(dhtInstance)

	errChan := make(chan error)

	go func() {
		if err := network.StartServer("127.0.0.1", 0); err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}
	case <-time.After(2 * time.Second):
		// Server started successfully
	}

	port := network.GetListeningPort()
	if port == 0 {
		t.Fatalf("Failed to get listening port")
	}

	tlsConfig, err := network.LoadTLSConfig(network.ServerCertFile, network.ServerKeyFile)
	conn, err := tls.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port), tlsConfig)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
}

func TestStopServer(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}

	dhtInstance := dht.NewDHT()

	node := dht.NewNode("127.0.0.1", 0, false, key)
	dhtInstance.JoinNetwork(node)

	network := networking.NewNetwork(dhtInstance)

	errChan := make(chan error)

	go func() {
		if err := network.StartServer("127.0.0.1", 0); err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}
	case <-time.After(2 * time.Second):
		// Server started successfully
	}

	network.StopServer()
	time.Sleep(1 * time.Second) // Ensure server stops
	tlsConfig, err := network.LoadTLSConfig(network.ServerCertFile, network.ServerKeyFile)
	if err != nil {
		t.Fatalf("Failed to load TLS configuration: %v", err)
	}
	port := network.GetListeningPort()

	conn, err := tls.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port), tlsConfig)
	if err == nil {
		conn.Close()
		t.Fatalf("Expected connection to fail after server stop, but it succeeded")
	}
}

func TestJoinAndLeaveNetworkNModule(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}

	dhtInstance := dht.NewDHT()
	network := networking.NewNetwork(dhtInstance)

	node1 := dht.NewNode("127.0.0.1", 8000, true, key)
	node2 := dht.NewNode("127.0.0.1", 8001, true, key)

	network.JoinNetwork(node1)
	network.JoinNetwork(node2)

	peers := node1.GetAllPeers()
	assert.Len(t, peers, 1)

	err = network.LeaveNetwork(node2)
	assert.Nil(t, err)

	peers = node1.GetAllPeers()
	assert.Len(t, peers, 0)
}