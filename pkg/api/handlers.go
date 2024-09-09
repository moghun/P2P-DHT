package api

import (
	"errors"
	"log"
	"strings"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/node"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

func HandlePut(msg message.Message, nodeInstance node.NodeInterface) []byte {
	putMsg := msg.(*message.DHTPutMessage)
	node := nodeInstance.(*node.Node)

	var err error

	done := make(chan bool)
	// Asynchronously send PUT request to DHT
	go func() {
		encodedKey := message.Byte32ToHexEncode(putMsg.Key)
		if len(putMsg.Key) != 32 || len(encodedKey) != 40 {
			err = errors.New("invalid key length")
			done <- true
			return
		}
		if err := node.DHT.PUT(encodedKey, string(putMsg.Value), int(putMsg.TTL)); err != nil {
			util.Log().Errorf("Error processing PUT in DHT: %v", err)
		}
		done <- true
	}()

	<-done // Wait for the asynchronous operation to complete

	if err != nil {
		util.Log().Errorf("Error processing PUT in DHT: %v", err)
		failureMsg, _ := message.NewDHTFailureMessage(putMsg.Key).Serialize()
		return failureMsg
	}

	successMsg, _ := message.NewDHTSuccessMessage(putMsg.Key, putMsg.Value).Serialize()
	return successMsg
}

func HandleGet(msg message.Message, nodeInstance node.NodeInterface) []byte {
	getMsg := msg.(*message.DHTGetMessage)
	node := nodeInstance.(*node.Node)

	var value string
	var nodes []*dht.KNode
	var err error

	// Asynchronously send GET request to DHT
	done := make(chan bool)
	go func() {
		encodedKey := message.Byte32ToHexEncode(getMsg.Key)
		if len(getMsg.Key) != 32 || len(encodedKey) != 40 {
			err = errors.New("invalid key length")
			done <- true
			return
		}

		value, nodes, err = node.DHT.GET(encodedKey)
		done <- true
	}()
	<-done

	if err != nil {
		failureMsg, _ := message.NewDHTFailureMessage(getMsg.Key).Serialize()
		return failureMsg
	} else if value == "" && nodes == nil {
		failureMsg, _ := message.NewDHTFailureMessage(getMsg.Key).Serialize()
		return failureMsg
	}

	successResponse := dht.SuccessMessageResponse{
		Value: value,
		Nodes: nodes,
	}

	successMsg, _ := message.NewDHTSuccessMessage(getMsg.Key, successResponse.Serialize()).Serialize()
	return successMsg
}

func HandlePing(msg message.Message, nodeInstance node.NodeInterface) []byte {
	nodeInstance = nodeInstance.(*node.Node)
	pongMsg, _ := message.NewDHTPongMessage().Serialize()
	return pongMsg
}

func HandleFindNode(msg message.Message, nodeInstance node.NodeInterface) []byte {
	findNodeMsg := msg.(*message.DHTFindNodeMessage)

	var nodes []*dht.KNode
	var err error

	done := make(chan bool)
	go func() {
		encodedKey := message.Byte32ToHexEncode(findNodeMsg.Key)
		if len(findNodeMsg.Key) != 32 || len(encodedKey) != 40 {
			err = errors.New("invalid key length")
			done <- true
			return
		}
		nodes, err = nodeInstance.FindNode(encodedKey)
		//log.Print("Is Node down?:", node.IsDown) // Why this?
		done <- true
	}()
	<-done

	if err != nil || nodes == nil {
		util.Log().Errorf("Error processing FIND_NODE in DHT: %v", err)

		failureMsg, _ := message.NewDHTFailureMessage(findNodeMsg.Key).Serialize()
		return failureMsg
	} else {
		successResponse := dht.SuccessMessageResponse{
			Value: "",
			Nodes: nodes,
		}

		successMsg, _ := message.NewDHTSuccessMessage(findNodeMsg.Key, successResponse.Serialize()).Serialize()
		return successMsg
	}
}

func HandleFindValue(msg message.Message, nodeInstance node.NodeInterface) []byte {
	findValueMsg := msg.(*message.DHTFindValueMessage)

	var value string
	var nodes []*dht.KNode
	var err error

	// Asynchronously process FIND_VALUE request
	done := make(chan bool)
	go func() {
		encodedKey := message.Byte32ToHexEncode(findValueMsg.Key)
		if len(findValueMsg.Key) != 32 || len(encodedKey) != 40 {
			err = errors.New("invalid key length")
			done <- true
			return
		}
		value, nodes, err = nodeInstance.FindValue(encodedKey)

		//log.Print("Is Node down?:", node.IsDown) // Why this?
		done <- true
	}()
	<-done

	if err != nil || (nodes == nil && value == "") {
		util.Log().Errorf("Error processing FIND_VALUE in DHT: %v", err)

		failureMsg, _ := message.NewDHTFailureMessage(findValueMsg.Key).Serialize()
		return failureMsg
	} else {
		successResponse := dht.SuccessMessageResponse{
			Value: value,
			Nodes: nodes,
		}

		if value != "" || nodes != nil {
			util.Log().Info("Value or Nodes found")
			successMsg, _ := message.NewDHTSuccessMessage(findValueMsg.Key, successResponse.Serialize()).Serialize()
			return successMsg
		} else {
			util.Log().Errorf("Error processing FIND_VALUE in DHT: %v", err)

			failureMsg, _ := message.NewDHTFailureMessage(findValueMsg.Key).Serialize()
			return failureMsg
		}
	}
}

func HandleStore(msg message.Message, nodeInstance node.NodeInterface) []byte {
	storeMsg := msg.(*message.DHTStoreMessage)
	node := nodeInstance.(*node.Node)

	var err error

	done := make(chan bool)
	go func() {
		encodedKey := message.Byte32ToHexEncode(storeMsg.Key)
		if len(storeMsg.Key) != 32 || len(encodedKey) != 40 {
			err = errors.New("invalid key length")
			done <- true
			return
		}

		log.Print("Storing key:", string(storeMsg.Key[:]))
		if err = node.DHT.StoreToStorage(encodedKey, string(storeMsg.Value), int(storeMsg.TTL)); err != nil {
			util.Log().Errorf("Error processing STORE in DHT: %v", err)
		}
		done <- true
	}()
	<-done

	if err != nil {
		util.Log().Errorf("Error processing STORE in DHT: %v", err)
		failureMsg, _ := message.NewDHTFailureMessage(storeMsg.Key).Serialize()
		return failureMsg
	}

	log.Print("Storing success!")
	successMsg, _ := message.NewDHTSuccessMessage(storeMsg.Key, storeMsg.Value).Serialize()
	return successMsg
}

func HandleBootstrap(msg message.Message, nodeInstance node.NodeInterface) []byte {
	bootstrapMsg := msg.(*message.DHTBootstrapMessage)
	nodeInstance = nodeInstance.(*node.BootstrapNode) //BOOTSTRAP NODE

	ipPort := strings.TrimSpace(string(bootstrapMsg.Address))
	util.Log().Infof("Handling bootstrap request from node: %s", ipPort)

	if bn, ok := nodeInstance.(*node.BootstrapNode); ok {
		//TODO: Use BootstrapNode specific methods or handle the request accordingly
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

// NOT TODO
func HandleBootstrapReply(msg message.Message, nodeInstance node.NodeInterface) []byte {
	bootstrapReplyMsg := msg.(*message.DHTBootstrapReplyMessage)
	nodeInstance = nodeInstance.(*node.Node)

	nodes := bootstrapReplyMsg.ParseNodes()
	for _, nodeInfo := range nodes {
		nodeID := node.GenerateNodeID(nodeInfo.IP, nodeInfo.Port)
		nodeInstance.AddPeer(nodeID, nodeInfo.IP, nodeInfo.Port)
		util.Log().Infof("Added node %s:%d to routing table.\n", nodeInfo.IP, nodeInfo.Port)
	}

	return nil
}
