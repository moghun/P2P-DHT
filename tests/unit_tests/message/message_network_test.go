package tests

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/tests"
)

func TestSendMessageSuccess(t *testing.T) {
	receiverPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")
	receiverNode := NewMockNode("1", "127.0.0.1", receiverPort).ToNode()
	network_sender := message.NewNetwork(receiverNode.IP, receiverNode.ID, receiverPort)
	log.Print("Receiver port:", receiverPort)



	go func() {
		err := network_sender.StartListening()
		assert.NoError(t, err)
		log.Print("listened")
	}()

	time.Sleep(2*time.Second)
	
	senderPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")
	senderNode := NewMockNode("1", "127.0.0.1", senderPort).ToNode()
	network_receiver := message.NewNetwork(senderNode.IP, senderNode.ID, senderPort)
	log.Print("Sender port:", senderPort)
	response, err := network_receiver.SendMessage("127.0.0.1", receiverPort, []byte("hello"))

	assert.NoError(t, err)
	assert.Equal(t, response, []byte("Success"))

	time.Sleep(time.Second)
}

func TestSendMessageFailure(t *testing.T) {
	senderPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	senderNode := NewMockNode("1", "127.0.0.1", senderPort).ToNode()
	network := message.NewNetwork(senderNode.IP, senderNode.ID, senderPort)


	_, err = network.SendMessage("invalid_ip", senderPort+1, []byte("hello"))
	assert.Error(t, err)
}

func TestStartListening(t *testing.T) {
	receiverPort, err := tests.GetFreePort()
	assert.NoError(t, err, "Failed to get a free port")

	receiverNode := NewMockNode("1", "127.0.0.1", receiverPort).ToNode()
	network := message.NewNetwork(receiverNode.IP, receiverNode.ID, receiverPort)

	go func() {
		err := network.StartListening()
		assert.NoError(t, err)
	}()

	time.Sleep(time.Second)

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", receiverPort))
	assert.NoError(t, err)
	defer conn.Close()

	_, err = conn.Write([]byte("test"))
	assert.NoError(t, err)

	time.Sleep(time.Second)
}
