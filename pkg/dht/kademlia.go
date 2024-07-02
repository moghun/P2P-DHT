package dht

import (
	"fmt"
)

const bootstrapNodeAmount = 5

var dht *DHT = nil
var myNode *Node = nil

func NewKademlia(dhtInstance *DHT) {
	SetDhtInstance(dhtInstance)
}

func SetDhtInstance(dhtInstance *DHT) {
	dht = dhtInstance
}

func FindNodeFromRecipient(recipient *Node, targetID string, k int) []*Node {
	kClosestNodes := recipient.GetClosestNodesToCurrNode(targetID, k)
	//Lookup the closest nodes to the target node, from the recipient node's response of the closest nodes
	for i := range kClosestNodes {
		contactClosestNodes := kClosestNodes[i].GetClosestNodesToCurrNode(targetID, k)
		//Perform lookup on the recently contacted nodes to iteratively find the closest nodes to the target node
		kClosestNodes = append(kClosestNodes, contactClosestNodes...)
	}

	//Remove duplicates
	kclosestNodes := FilterDuplicateNodes(kClosestNodes)
	return kclosestNodes
}

func FindNode(targetID string) []*Node {
	k := dht.getReplicationFactor()
	kClosestNodes := myNode.GetClosestNodesToCurrNode(targetID, k)
	//Lookup the closest nodes to the target node, from the recipient node's response of the closest nodes
	for i := range kClosestNodes {
		contactClosestNodes := kClosestNodes[i].GetClosestNodesToCurrNode(targetID, k)
		//Perform lookup on the recently contacted nodes to iteratively find the closest nodes to the target node
		kClosestNodes = append(kClosestNodes, contactClosestNodes...)
	}

	//Remove duplicates
	kclosestNodes := FilterDuplicateNodes(kClosestNodes)
	return kclosestNodes
}

// Kademlia: TODO current implementation assumes nodes are sent, normally only ids or values should be sent
func FindValue(targetID string) []*Node {
	k := dht.getReplicationFactor()
	kClosestNodes := myNode.GetClosestNodesToCurrNode(targetID, k)
	returnList := []*Node{}
	for i := range kClosestNodes {
		if kClosestNodes[i].ID == targetID {
			returnList = append(returnList, kClosestNodes[i])
			return returnList
		}
	}

	return FindNode(targetID)
	//Kademlia: TODO send find value message to kClosestNodes in a loop
	//If returned message is dht fail, return fail message/nil
	//If returned message is a id/success, return success message
}

// Filters duplicate Node IDs from the given slice and returns a new slice
func FilterDuplicateNodes(nodes []*Node) []*Node {
	seen := make(map[string]bool)
	filtered := []*Node{}

	for _, node := range nodes {
		if !seen[node.ID] {
			seen[node.ID] = true
			filtered = append(filtered, node)
		}
	}

	return filtered
}

func Join(joiningNode *Node, bsNode *Node) {
	checkDhtInstance()
	kclosestNodes := FindNodeFromRecipient(bsNode, joiningNode.ID, dht.getReplicationFactor())

	//Add the found nodes to the joining node's kbuckets
	for i := range kclosestNodes {
		joiningNode.AddPeer(kclosestNodes[i].ID, kclosestNodes[i].IP, kclosestNodes[i].Port)
	}
}

func checkDhtInstance() {
	if dht == nil {
		fmt.Println("DHT instance is nil.")
		return
	}
}
