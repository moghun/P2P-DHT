package tests

import (
	"net"
	"time"

	"github.com/stretchr/testify/mock"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
)

type MockNode struct {
	node.Node
	mock.Mock
}

func NewMockNode(ip string, port int) *MockNode {
	storage := storage.NewStorage(24*time.Hour, []byte("1234567890abcdef"))
	dht := dht.NewDHT(24*time.Hour, []byte("1234567890abcdef"), "1", "127.0.0.1", 8080)
	return &MockNode{
		Node: node.Node{
			IP:      ip,
			Port:    port,
			Storage: storage,
			DHT:     dht,
		},
	}
}

func (m *MockNode) Put(key string, value string, ttl int) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockNode) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockNode) AddPeer(nodeID string, ip string, port int) {
	m.Called(nodeID, ip, port)
}

// MockConn is a mock implementation of the net.Conn interface
type MockConn struct {
	mock.Mock
	readData  []byte
	writeData []byte
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	n = copy(b, m.readData)
	if len(m.readData) > n {
		m.readData = m.readData[n:]
	} else {
		m.readData = nil
	}
	return n, nil
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	m.writeData = append(m.writeData, b...)
	return len(b), nil
}

func (m *MockConn) Close() error {
	return nil
}

func (m *MockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{}
}

func (m *MockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{}
}

func (m *MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// MockConnWithAddr extends MockConn to include specific local and remote addresses
type MockConnWithAddr struct {
	MockConn
	localAddr  net.Addr
	remoteAddr net.Addr
}

func (m *MockConnWithAddr) LocalAddr() net.Addr {
	return m.localAddr
}

func (m *MockConnWithAddr) RemoteAddr() net.Addr {
	return m.remoteAddr
}
