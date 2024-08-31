package tests

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/tests"
)

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

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	assert.NoError(t, err)
	defer conn.Close()

	_, err = conn.Write([]byte("test"))
	assert.NoError(t, err)

	time.Sleep(time.Second)
}
