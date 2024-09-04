package api

import (
	"log"
	"strings"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
)

func HandlePut(msg message.Message, nodeInstance node.NodeInterface) []byte {
	putMsg := msg.(*message.DHTPutMessage)
	node := nodeInstance.(*node.Node)

	done := make(chan bool)
	// Asynchronously send PUT request to DHT
	go func() {
		if err := node.DHT.PUT(string(putMsg.Key[:]), string(putMsg.Value), int(putMsg.TTL)); err != nil {
			log.Printf("Error processing PUT in DHT: %v", err)
		}
		done <- true
	}()

	<-done // Wait for the asynchronous operation to complete

	successMsg, _ := message.NewDHTSuccessMessage(putMsg.Key, putMsg.Value).Serialize()
	return successMsg
}

func HandleGet(msg message.Message, nodeInstance node.NodeInterface) []byte {
	getMsg := msg.(*message.DHTGetMessage)
	node := nodeInstance.(*node.Node)

	var value string
	var err error

	// Asynchronously send GET request to DHT
	done := make(chan bool)
	go func() {
		value, err = node.DHT.GET(string(getMsg.Key[:]))
		done <- true
	}()
	<-done

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
	node := nodeInstance.(*node.Node)

	var nodes []*dht.KNode
	var err error

	done := make(chan bool)
	go func() {
		nodes, err = nodeInstance.FindNode(string(nodeInstance.GetID()), string(findNodeMsg.Key[:]))

		log.Print("Is Node down?:", node.IsDown) // Why this?
	}()
	<-done

	if err != nil {
		log.Printf("Error processing FIND_NODE in DHT: %v", err)

		failureMsg, _ := message.NewDHTFailureMessage(findNodeMsg.Key).Serialize()
		return failureMsg
	} else {
		// TODO Handle success message serialization
		log.Printf("Closest nodes: %v", nodes)

		var nodeBytes []byte
		for _, n := range nodes {
			nodeBytes = append(nodeBytes, string(n.Serialize())...)
		}
		successMsg, _ := message.NewDHTSuccessMessage(findNodeMsg.Key, nodeBytes).Serialize()
		return successMsg
	}

	successMsg, _ := message.NewDHTSuccessMessage(findNodeMsg.Key, []byte("mock-node")).Serialize()
	return successMsg
}

func HandleFindValue(msg message.Message, nodeInstance node.NodeInterface) []byte {
	findValueMsg := msg.(*message.DHTFindValueMessage)
	node := nodeInstance.(*node.Node)

	var value string
	var nodes []*dht.KNode
	var err error

	// Asynchronously process FIND_VALUE request
	done := make(chan bool)
	go func() {
		value, nodes, err = nodeInstance.FindValue(string(nodeInstance.GetID()), string(findValueMsg.Key[:]))

		log.Print("Is Node down?:", node.IsDown) // Why this?
	}()
	<-done

	if err != nil {
		log.Printf("Error processing FIND_VALUE in DHT: %v", err)
		failureMsg, _ := message.NewDHTFailureMessage(findValueMsg.Key).Serialize()
		return failureMsg
	} else {
		if value != "" {
			log.Printf("Value found: %s", value)

			successMsg, _ := message.NewDHTSuccessMessage(findValueMsg.Key, []byte(value)).Serialize()
			return successMsg
		} else {
			log.Printf("Value not found. Closest nodes: %v", nodes)

			var nodeBytes []byte
			for _, n := range nodes {
				nodeBytes = append(nodeBytes, string(n.Serialize())...)
			}
			successMsg, _ := message.NewDHTSuccessMessage(findValueMsg.Key, nodeBytes).Serialize()
			return successMsg
		}
	}

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

	//TODO:
	//THIS METHOD MUST USE THE ADD PEER FUNCTIONALITY ETC. NOT LIKE THIS!
	nodes := bootstrapReplyMsg.ParseNodes()
	for _, nodeInfo := range nodes {
		nodeID := node.GenerateNodeID(nodeInfo.IP, nodeInfo.Port)
		nodeInstance.AddPeer(nodeID, nodeInfo.IP, nodeInfo.Port)
		log.Printf("Added node %s:%d to routing table.\n", nodeInfo.IP, nodeInfo.Port)
	}

	return nil
}
