package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
)

// Test that peers respond to the liveliness check and are not removed.
func TestLivelinessChecker_PeerResponds(t *testing.T) {
	// Create a mock node and add a peer
	mockNode := NewMockNode("127.0.0.1", 8000)
	peerNode := &node.Node{ID: "peerID", IP: "127.0.0.2", Port: 8001}
	mockNode.AddPeer(peerNode)

	// Create the liveliness checker
	livelinessChecker := node.NewLivelinessChecker(mockNode.ToNode(), 100*time.Millisecond, 200*time.Millisecond)

	// Set mock network to succeed
	mockNode.Network.(*MockNetwork).SetShouldFail(false)

	// Start liveliness check
	go livelinessChecker.Start()

	// Wait for some time to let liveliness checks occur
	time.Sleep(500 * time.Millisecond)

	// Stop liveliness check
	livelinessChecker.Stop()

	// Assert that the peer was not removed
	assert.False(t, mockNode.WasPeerRemoved())
}

// Test that peers are removed when they don't respond to liveliness checks.
func TestLivelinessChecker_PeerDoesNotRespond(t *testing.T) {
	// Create a mock node and add a peer
	mockNode := NewMockNode("127.0.0.1", 8000)
	peerNode := &node.Node{ID: "peerID", IP: "127.0.0.2", Port: 8001}
	mockNode.AddPeer(peerNode)

	// Create the liveliness checker
	livelinessChecker := node.NewLivelinessChecker(mockNode.ToNode(), 100*time.Millisecond, 200*time.Millisecond)

	// Set mock network to fail (peer won't respond)
	mockNode.Network.(*MockNetwork).SetShouldFail(true)

	// Start liveliness check
	go livelinessChecker.Start()

	// Wait for some time to let liveliness checks occur
	time.Sleep(1 * time.Second)

	// Stop liveliness check
	livelinessChecker.Stop()

	// Assert that the peer was removed
	assert.True(t, mockNode.WasPeerRemoved(), "Expected peer to be removed, but it wasn't")
}


// Test starting and stopping the liveliness checker without removing peers unnecessarily.
func TestLivelinessChecker_StartAndStop(t *testing.T) {
	// Create a mock node and add a peer
	mockNode := NewMockNode("127.0.0.1", 8000)
	peerNode := &node.Node{ID: "peerID", IP: "127.0.0.2", Port: 8001}
	mockNode.AddPeer(peerNode)

	// Create the liveliness checker
	livelinessChecker := node.NewLivelinessChecker(mockNode.ToNode(), 100*time.Millisecond, 200*time.Millisecond)

	// Set mock network to succeed
	mockNode.Network.(*MockNetwork).SetShouldFail(false)

	// Start liveliness check
	livelinessChecker.Start()

	// Wait for some time to let liveliness checks occur
	time.Sleep(300 * time.Millisecond)

	// Stop liveliness check
	livelinessChecker.Stop()

	// Wait some more time to ensure the checker has stopped
	time.Sleep(200 * time.Millisecond)

	// Assert that the peer was not removed
	assert.False(t, mockNode.WasPeerRemoved())
}