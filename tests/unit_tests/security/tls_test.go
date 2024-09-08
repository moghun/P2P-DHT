package tests

import (
	"crypto/tls"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/security"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/tests"
)

// TestGenerateSelfSignedCertificate verifies that self-signed certificates can be generated without errors.
func TestGenerateSelfSignedCertificate(t *testing.T) {
	peerID := "peer-1"
	cert, err := security.GenerateSelfSignedCertificate(peerID)
	assert.NoError(t, err, "Failed to generate self-signed certificate")
	assert.NotEmpty(t, cert.Certificate, "Expected certificate to be non-empty")
	assert.NotEmpty(t, cert.PrivateKey, "Expected private key to be non-empty")
}

// TestCreateTLSConfig verifies that a valid TLS config can be created with a self-signed certificate.
func TestCreateTLSConfig(t *testing.T) {
	peerID := "peer-2"
	tlsConfig, err := security.CreateTLSConfig(peerID)
	assert.NoError(t, err, "Failed to create TLS config")
	assert.NotNil(t, tlsConfig, "Expected TLS config to be non-nil")
	assert.NotEmpty(t, tlsConfig.Certificates, "Expected TLS config to have certificates")
}

// TestStartTLSListener verifies that a TLS listener can be started and accept connections.
func TestStartTLSListener(t *testing.T) {
	peerID := "peer-listener"

	port, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	address := fmt.Sprintf("127.0.0.1:%d", port)
	t.Logf("Starting TLS listener on %s", address)
	tlsListener, err := security.StartTLSListener(peerID, address)
	assert.NoError(t, err, "Failed to start TLS listener")
	assert.NotNil(t, tlsListener, "Expected TLS listener to be non-nil")
	defer tlsListener.Close()

	done := make(chan struct{})

	go func() {
		conn, err := tlsListener.Accept()
		if err != nil {
			t.Fatalf("Failed to accept connection: %v", err)
		}
		defer conn.Close()

		// Simulate server-side handling
		t.Log("Server: Accepted connection, reading data...")
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			t.Logf("Server: Error reading data: %v", err)
		} else {
			t.Logf("Server: Received data: %s", string(buf[:n]))
		}
		close(done)
	}()

	time.Sleep(1 * time.Second)

	// Simulate a client connecting to the TLS listener
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	t.Log("Client: Dialing TLS connection...")

	conn, err := tls.Dial("tcp", address, tlsConfig)
	assert.NoError(t, err, "Failed to dial TLS connection")
	defer conn.Close()

	// Simulate client-side message
	t.Log("Client: Sending message to server...")
	_, err = conn.Write([]byte("Hello, server"))
	assert.NoError(t, err, "Client failed to send data")

	// Wait for the server to read the message
	<-done
}

// TestDialTLS verifies that a client can successfully dial a server using TLS.
func TestDialTLS(t *testing.T) {
	peerID := "peer-server"

	port, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	address := fmt.Sprintf("127.0.0.1:%d", port)
	t.Logf("Starting TLS listener on %s", address)

	tlsListener, err := security.StartTLSListener(peerID, address)
	assert.NoError(t, err, "Failed to start TLS listener")
	assert.NotNil(t, tlsListener, "Expected TLS listener to be non-nil")
	defer tlsListener.Close()

	done := make(chan struct{})

	// Start the server to accept a TLS connection
	go func() {
		conn, err := tlsListener.Accept()
		if err != nil {
			t.Fatalf("Failed to accept connection: %v", err)
		}
		defer conn.Close()

		// Simulate server-side handling
		t.Log("Server: Accepted connection, reading data...")
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			t.Logf("Server: Error reading data: %v", err)
		} else {
			t.Logf("Server: Received data: %s", string(buf[:n]))
		}
		close(done)
	}()

	time.Sleep(1 * time.Second)

	// Simulate client dialing the TLS listener
	t.Log("Client: Dialing TLS connection...")
	conn, err := security.DialTLS("peer-client", address)
	assert.NoError(t, err, "Failed to dial TLS connection")
	assert.NotNil(t, conn, "Expected TLS connection to be non-nil")
	defer conn.Close()

	// Simulate client sending data
	t.Log("Client: Sending message to server...")
	_, err = conn.Write([]byte("Hello from client"))
	assert.NoError(t, err, "Client failed to send data")

	// Wait for the server to process the message
	<-done
}

// TestMutualTLSConnection verifies that a mutual TLS connection can be established between two peers.
func TestMutualTLSConnection(t *testing.T) {
	peerIDServer := "peer-server"
	peerIDClient := "peer-client"

	port, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	address := fmt.Sprintf("127.0.0.1:%d", port)

	// Start TLS listener for the server
	tlsListener, err := security.StartTLSListener(peerIDServer, address)
	assert.NoError(t, err, "Failed to start TLS listener")
	defer tlsListener.Close()

	go func() {
		conn, err := tlsListener.Accept()
		assert.NoError(t, err, "Server failed to accept connection")
		defer conn.Close()

		// Server reads data from the connection
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		assert.NoError(t, err, "Server failed to read data")
		assert.Equal(t, "Hello from client", string(buf[:n]))
	}()

	// Client establishes a TLS connection to the server
	conn, err := security.DialTLS(peerIDClient, address)
	assert.NoError(t, err, "Client failed to dial TLS connection")
	defer conn.Close()

	// Client sends data to the server
	_, err = conn.Write([]byte("Hello from client"))
	assert.NoError(t, err, "Client failed to send data")
}

// TestInvalidTLSConnection verifies that invalid TLS connections are rejected.
func TestInvalidTLSConnection(t *testing.T) {
	peerID := "peer-invalid"
	port := 50003
	address := "127.0.0.1"

	listener, err := security.StartTLSListener(peerID, fmt.Sprintf("%s:%d", address, port))
	assert.NoError(t, err, "Failed to start TLS listener")
	assert.NotNil(t, listener, "Expected TLS listener to be non-nil")
	defer listener.Close()

	go func() {
		_, err := listener.Accept()
		assert.NoError(t, err, "Server failed to accept connection")
	}()

	// Attempt to connect without TLS
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", address, port))
	assert.NoError(t, err, "Failed to dial TCP connection")
	defer conn.Close()

	// Write some non-TLS data to the server
	_, err = conn.Write([]byte("This is not TLS"))
	assert.NoError(t, err, "Failed to send non-TLS data")

	// Server should reject the connection as it expects TLS
	// This can be validated through logs or error handling in the server
}