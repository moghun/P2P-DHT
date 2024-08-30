package api

import (
	"log"
	"strings"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
)

func HandlePut(msg message.Message, nodeInstance *node.Node) []byte {
	putMsg := msg.(*message.DHTPutMessage)

	// Store the value in the node's storage
	if err := nodeInstance.Put(string(putMsg.Key[:]), string(putMsg.Value), int(putMsg.TTL)); err != nil {
		failureMsg, _ := message.NewDHTFailureMessage(putMsg.Key).Serialize()
		return failureMsg
	}

	successMsg, _ := message.NewDHTSuccessMessage(putMsg.Key, putMsg.Value).Serialize()
	return successMsg
}

func HandleGet(msg message.Message, nodeInstance *node.Node) []byte {
	getMsg := msg.(*message.DHTGetMessage)

	value, err := nodeInstance.Get(string(getMsg.Key[:]))
	if err != nil || value == "" {
		failureMsg, _ := message.NewDHTFailureMessage(getMsg.Key).Serialize()
		return failureMsg
	}

	successMsg, _ := message.NewDHTSuccessMessage(getMsg.Key, []byte(value)).Serialize()
	return successMsg
}

func HandlePing(msg message.Message, nodeInstance *node.Node) []byte {
	// Respond with a PONG message
	pongMsg, _ := message.NewDHTPongMessage().Serialize()
	return pongMsg
}

func HandleFindNode(msg message.Message, nodeInstance *node.Node) []byte {
	findNodeMsg := msg.(*message.DHTFindNodeMessage)

	// Mocked response for FIND_NODE
	successMsg, _ := message.NewDHTSuccessMessage(findNodeMsg.Key, []byte("mock-node")).Serialize()
	return successMsg
}

func HandleFindValue(msg message.Message, nodeInstance *node.Node) []byte {
	findValueMsg := msg.(*message.DHTFindValueMessage)

	// Mocked response for FIND_VALUE
	successMsg, _ := message.NewDHTSuccessMessage(findValueMsg.Key, []byte("mock-value")).Serialize()
	return successMsg
}

func HandleBootstrap(msg message.Message, nodeInstance *node.Node) []byte {
	bootstrapMsg := msg.(*message.DHTBootstrapMessage)

	ipPort := strings.TrimSpace(string(bootstrapMsg.Address))
	log.Printf("Handling bootstrap request from node: %s", ipPort)

	//TODO
	//nodeInstance.AddPeer(node.GenerateNodeIDFromIPPort(ipPort), ipPort)

	peers := nodeInstance.GetAllPeers()
	var responseData []string
	for _, peer := range peers {
		//TODO
		/*
		if peer.IPPort() != ipPort {
			responseData = append(responseData, peer.IPPort())
		}
		*/
		log.Println(peer)
	}

	responseString := strings.Join(responseData, "\n")
	bootstrapReplyMsg, _ := message.NewDHTBootstrapReplyMessage([]byte(responseString)).Serialize()
	return bootstrapReplyMsg
}

func HandleBootstrapReply(msg message.Message, nodeInstance *node.Node) []byte {
	bootstrapReplyMsg := msg.(*message.DHTBootstrapReplyMessage)

	nodes := bootstrapReplyMsg.ParseNodes()
	for _, nodeInfo := range nodes {
		nodeID := node.GenerateNodeID(nodeInfo.IP, nodeInfo.Port)
		nodeInstance.AddPeer(nodeID, nodeInfo.IP, nodeInfo.Port)
		log.Printf("Added node %s:%d to routing table.\n", nodeInfo.IP, nodeInfo.Port)
	}

	return nil // No need to respond to a BOOTSTRAP_REPLY message
}
