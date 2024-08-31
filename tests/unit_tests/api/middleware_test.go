package tests

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/api"
)

func TestWithMiddleware(t *testing.T) {
	// Set up a mock connection with specific local and remote addresses
	localAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
	remoteAddr := &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 54321}
	conn := &MockConnWithAddr{
		MockConn:   MockConn{readData: make([]byte, 1024)},
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
	}

	// Use a flag to check if the handler was called
	handlerCalled := false

	// Define a simple handler that marks the handler as called
	handler := func(conn net.Conn) {
		handlerCalled = true
	}

	wrappedHandler := api.WithMiddleware(handler)

	wrappedHandler(conn)

	assert.True(t, handlerCalled, "Expected the handler to be called")

}

func TestWithMiddleware_ConnectionClosed(t *testing.T) {
	localAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
	remoteAddr := &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 54321}
	conn := &MockConnWithAddr{
		MockConn:   MockConn{readData: make([]byte, 1024)},
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
	}

	// Use a flag to check if the handler was called
	handlerCalled := false

	// Define a simple handler that marks the handler as called
	handler := func(conn net.Conn) {
		handlerCalled = true
	}

	// Wrap the handler with middleware
	wrappedHandler := api.WithMiddleware(handler)

	// Call the wrapped handler
	wrappedHandler(conn)

	// Verify that the handler was called
	assert.True(t, handlerCalled, "Expected the handler to be called")

	// Close the connection and verify that it is closed
	err := conn.Close()
	assert.NoError(t, err, "Expected no error when closing the connection")

	// Ensure that subsequent calls to the handler do not cause issues
	assert.NotPanics(t, func() { wrappedHandler(conn) }, "Expected no panic when calling handler after connection closed")
}
