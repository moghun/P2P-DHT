package dht

import (
	"fmt"
	"math/rand"
	"time"
)

const bootstrapNodeAmount = 5

var dht *DHT = nil

type Kademlia struct {
	bootstrapNodes []*Node
}

func NewKademlia(dhtInstance *DHT) *Kademlia {
	SetDhtInstance(dhtInstance)
	InitializeBootstrapNodes()
	return &Kademlia{}
}

func SetDhtInstance(dhtInstance *DHT) {
	dht = dhtInstance
}

func InitializeBootstrapNodes() {
	checkDhtInstance()

	// Loop through each bootstrap node and add it to the DHT
	for i := 0; i < bootstrapNodeAmount; i++ {
		// Create a new Node instance
		key := []byte("12345678901234567890123456789012")
		node := NewNode("127.0.0.1", 8000+i, true, key)

		// Add the bootstrap node to the DHT network
		dht.JoinNetwork(node)
		dht.bootstrapNodes = append(dht.bootstrapNodes, node)
	}

	fmt.Println("Bootstrap nodes initialized and added to the DHT network.")
}

func FindNode(recipient *Node, targetID string, k int) []*Node {
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

// Randomly selects a bootstrap node from the list of bootstrap nodes
func (kad *Kademlia) getRandomBootstrapNode() *Node {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomBsNode := kad.bootstrapNodes[r.Intn(len(kad.bootstrapNodes))]
	return randomBsNode
}

func (kad *Kademlia) Join(joiningNode *Node) {
	checkDhtInstance()
	//Get a bootstrap node to start the lookup to find the joining node's closest nodes
	bsNode := kad.getRandomBootstrapNode()
	kclosestNodesSet := FindNode(bsNode, joiningNode.ID, dht.getReplicationFactor())

	//Add the found nodes to the joining node's kbuckets
	for i := range kclosestNodesSet {
		joiningNode.AddPeer(kclosestNodesSet[i].ID, kclosestNodesSet[i].IP, kclosestNodesSet[i].Port)
	}
}

func checkDhtInstance() {
	if dht == nil {
		fmt.Println("DHT instance is nil.")
		return
	}
}
