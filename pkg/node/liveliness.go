package node

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
)

// LivelinessChecker is responsible for checking the liveliness of peers in the network.
type LivelinessChecker struct {
	node      *Node
	interval  time.Duration
	timeout   time.Duration
	mu        sync.Mutex
	quitChan  chan struct{}
}

// NewLivelinessChecker creates a new LivelinessChecker with the specified interval and timeout durations.
func NewLivelinessChecker(node *Node, interval, timeout time.Duration) *LivelinessChecker {
	return &LivelinessChecker{
		node:     node,
		interval: interval,
		timeout:  timeout,
		quitChan: make(chan struct{}),
	}
}

// Start begins the liveliness check routine which periodically pings known peers.
func (lc *LivelinessChecker) Start() {
	go func() {
		ticker := time.NewTicker(lc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				lc.checkLiveliness()
			case <-lc.quitChan:
				log.Println("Liveliness check stopped.")
				return
			}
		}
	}()
}

// Stop halts the liveliness checker.
func (lc *LivelinessChecker) Stop() {
	close(lc.quitChan)
}

// checkLiveliness sends ping messages to all known peers and removes those that do not respond.
func (lc *LivelinessChecker) checkLiveliness() {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	peers := lc.node.GetAllPeers()

	for _, peer := range peers {
		go func(peer *Node) {
			err := lc.pingPeer(peer)
			if err != nil {
				log.Printf("Peer %s did not respond, removing from known peers.\n", peer.ID)
				lc.node.RemovePeer(peer.IP, peer.Port)
			} else {
				log.Printf("Peer %s is alive.\n", peer.ID)
			}
		}(peer)
	}
}

// pingPeer sends a ping message to a peer and waits for a response. If no response is received within the timeout, an error is returned.
func (lc *LivelinessChecker) pingPeer(peer *Node) error {
	pingMsg := message.NewDHTPingMessage()
	serializedPingMsg, err := pingMsg.Serialize()
	if err != nil {
		return err
	}

	responseChan := make(chan error, 1)

	// Send the ping in a separate goroutine and wait for the response.
	go func() {
		_, err := lc.node.Network.SendMessage(peer.IP, peer.Port, serializedPingMsg)
		responseChan <- err
	}()

	// Wait for a response or timeout.
	select {
	case err := <-responseChan:
		if err != nil {
			return err
		}
		return nil
	case <-time.After(lc.timeout):
		return fmt.Errorf("timeout waiting for response from peer %s", peer.ID)
	}
}
