package tests

import (
	"fmt"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
)

type MockNode struct {
	ID   string
	IP   string
	Port int
}

func (m *MockNode) GetIP() string {
	return m.IP
}

func (m *MockNode) GetPort() int {
	return m.Port
}

func (m *MockNode) GetNetwork() message.NetworkInterface {
	return nil
}

func NewMockNode(id string, ip string, port int) *MockNode {
	return &MockNode{
		ID:   id,
		IP:   ip,
		Port: port,
	}
}

// Convert MockNode to a *node.Node for compatibility
func (m *MockNode) ToNode() *node.Node {
	return &node.Node{
		IP:   m.IP,
		Port: m.Port,
		// Other fields can be initialized with dummy values if needed
	}
}

type MockNetwork struct {
	ShouldFail bool
}

func (m *MockNetwork) SendMessage(targetIP string, targetPort int, data []byte) ([]byte, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("failed to connect")
	}
	// Simulate a successful message send
	return nil, nil
}

func (m *MockNetwork) StartListening() error {
	return nil
}
