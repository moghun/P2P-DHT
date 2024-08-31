package api

import (
	"log"
	"strings"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
)

func HandlePut(msg message.Message, nodeInstance node.NodeInterface) []byte {
	putMsg := msg.(*message.DHTPutMessage)
	nodeInstance = nodeInstance.(*node.Node)

	if err := nodeInstance.Put(string(putMsg.Key[:]), string(putMsg.Value), int(putMsg.TTL)); err != nil {
		failureMsg, _ := message.NewDHTFailureMessage(putMsg.Key).Serialize()
		return failureMsg
	}

	successMsg, _ := message.NewDHTSuccessMessage(putMsg.Key, putMsg.Value).Serialize()
	return successMsg
}

func HandleGet(msg message.Message, nodeInstance node.NodeInterface) []byte {
	getMsg := msg.(*message.DHTGetMessage)
	nodeInstance = nodeInstance.(*node.Node)

	value, err := nodeInstance.Get(string(getMsg.Key[:]))
	if err != nil || value == "" {
		failureMsg, _ := message.NewDHTFailureMessage(getMsg.Key).Serialize()
		return failureMsg
	}

	successMsg, _ := message.NewDHTSuccessMessage(getMsg.Key, []byte(value)).Serialize()
	return successMsg
}

func HandlePing(msg message.Message, nodeInstance node.NodeInterface) []byte {
	nodeInstance = nodeInstance.(*node.Node)
	pongMsg, _ := message.NewDHTPongMessage().Serialize()
	return pongMsg
}

func HandleFindNode(msg message.Message, nodeInstance node.NodeInterface) []byte {
	findNodeMsg := msg.(*message.DHTFindNodeMessage)
	nodeInstance = nodeInstance.(*node.Node)
	successMsg, _ := message.NewDHTSuccessMessage(findNodeMsg.Key, []byte("mock-node")).Serialize()
	return successMsg
}

func HandleFindValue(msg message.Message, nodeInstance node.NodeInterface) []byte {
	findValueMsg := msg.(*message.DHTFindValueMessage)
	nodeInstance = nodeInstance.(*node.Node)
	successMsg, _ := message.NewDHTSuccessMessage(findValueMsg.Key, []byte("mock-value")).Serialize()
	return successMsg
}

func HandleBootstrap(msg message.Message, nodeInstance node.NodeInterface) []byte {
	bootstrapMsg := msg.(*message.DHTBootstrapMessage)
	nodeInstance = nodeInstance.(*node.BootstrapNode) //BOOTSTRAP NODE

	ipPort := strings.TrimSpace(string(bootstrapMsg.Address))
	log.Printf("Handling bootstrap request from node: %s", ipPort)

	if bn, ok := nodeInstance.(*node.BootstrapNode); ok {
		// Use BootstrapNode specific methods or handle the request accordingly -- TODO
		bn.KnownPeers[ipPort] = "1"
	}

	peers := nodeInstance.GetAllPeers()
	var responseData []string
	for _, peer := range peers {
		// Implement logic to filter or manipulate peer data as needed
		log.Println(peer)
	}

	responseString := strings.Join(responseData, "\n")
	bootstrapReplyMsg, _ := message.NewDHTBootstrapReplyMessage([]byte(responseString)).Serialize()
	return bootstrapReplyMsg
}

func HandleBootstrapReply(msg message.Message, nodeInstance node.NodeInterface) []byte {
	bootstrapReplyMsg := msg.(*message.DHTBootstrapReplyMessage)
	nodeInstance = nodeInstance.(*node.Node)

	nodes := bootstrapReplyMsg.ParseNodes()
	for _, nodeInfo := range nodes {
		nodeID := node.GenerateNodeID(nodeInfo.IP, nodeInfo.Port)
		nodeInstance.AddPeer(nodeID, nodeInfo.IP, nodeInfo.Port)
		log.Printf("Added node %s:%d to routing table.\n", nodeInfo.IP, nodeInfo.Port)
	}

	return nil
}
