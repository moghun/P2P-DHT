package tests

import (
	"fmt"
	"sync"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
)

type MockNode struct {
	IP           string
	Port         int
	Peers        []*node.Node
	removeCalled bool
	Network      message.NetworkInterface
	mu           sync.Mutex
}

func (m *MockNode) GetIP() string {
	return m.IP
}

func (m *MockNode) GetPort() int {
	return m.Port
}

func (m *MockNode) GetNetwork() message.NetworkInterface {
	return m.Network
}

func (m *MockNode) GetAllPeers() []*node.Node {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Peers
}

func (m *MockNode) AddPeer(peer *node.Node) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Peers = append(m.Peers, peer)
}

func (m *MockNode) RemovePeer(ip string, port int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeCalled = true
	for i, peer := range m.Peers {
		if peer.IP == ip && peer.Port == port {
			m.Peers = append(m.Peers[:i], m.Peers[i+1:]...)
			break
		}
	}
}

func (m *MockNode) WasPeerRemoved() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.removeCalled
}

func NewMockNode(ip string, port int) *MockNode {
	return &MockNode{
		IP:      ip,
		Port:    port,
		Peers:   []*node.Node{},
		Network: &MockNetwork{},
	}
}

// ToNode ensures the mock returns the expected *node.Node
func (m *MockNode) ToNode() *node.Node {
	return &node.Node{
		IP:      m.IP,
		Port:    m.Port,
		Ping:    true,
		Network: m.Network,
	}
}

// MockNetwork simulates a network for sending and receiving messages.
type MockNetwork struct {
	ShouldFail bool
}

func (m *MockNetwork) SendMessage(targetIP string, targetPort int, data []byte) ([]byte, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("failed to connect")
	}
	// Simulate a successful message send (e.g., a DHT_PONG response)
	return message.NewDHTPongMessage().Serialize()
}

func (m *MockNetwork) StartListening() error {
	return nil
}

func (m *MockNetwork) SetShouldFail(fail bool) {
	m.ShouldFail = fail
}
