package tests

import (
	"testing"
	"time"
	"net"
	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/api"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

func TestPingPongIntegration(t *testing.T) {
	config := &util.Config{
		P2PAddress: "127.0.0.1:8081",
	}

	mockNode := node.NewNode(config, 720)
	go api.StartServer(config.P2PAddress, mockNode)

	time.Sleep(1 * time.Second) // Give the server time to start

	conn, err := net.Dial("tcp", config.P2PAddress)
	assert.NoError(t, err)
	defer conn.Close()

	// Step 1: Send DHT_PING message
	pingMsg := message.NewDHTPingMessage()
	serializedPingMsg, err := pingMsg.Serialize()
	assert.NoError(t, err)

	_, err = conn.Write(serializedPingMsg)
	assert.NoError(t, err)

	// Step 2: Read and verify the PONG response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	assert.NoError(t, err)

	response, err := message.DeserializeMessage(buf[:n])
	assert.NoError(t, err)

	pongMsg := response.(*message.DHTPongMessage)
	assert.NotZero(t, pongMsg.Timestamp)
}
