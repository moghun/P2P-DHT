package dht

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
)

type DHT struct {
	nodes          []*Node
	kBuckets       []*KBucket
	mu             sync.Mutex
	bootstrapNodes []*Node
}

func NewDHT() *DHT {
	kBuckets := make([]*KBucket, 160)
	for i := range kBuckets {
		kBuckets[i] = NewKBucket()
	}

	dht = &DHT{
		nodes:    []*Node{},
		kBuckets: kBuckets,
	}

	return dht
}

func (d *DHT) InitializeBootstrapNodes() {
	checkDhtInstance()

	// Loop through each bootstrap node and add it to the DHT
	for i := 0; i < bootstrapNodeAmount; i++ {
		// Create a new Node instance
		key := []byte("12345678901234567890123456789012")
		node := NewNode("127.0.0.1", 8000+i, true, key)

		// Add the bootstrap node to the DHT network
		d.JoinNetwork(node)
		d.bootstrapNodes = append(d.bootstrapNodes, node)
	}

	fmt.Println("Bootstrap nodes initialized and added to the DHT network.")
}

// Randomly selects a bootstrap node from the list of bootstrap nodes
func (d *DHT) getRandomBootstrapNode() *Node {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomBsNode := d.bootstrapNodes[r.Intn(len(d.bootstrapNodes))]
	return randomBsNode
}

// TODO: We should find a way to scale the replication! When number of nodes increased, replication must scale
//
//	It can be a periodic check maybe.
func (d *DHT) getReplicationFactor() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	nodeCount := len(d.nodes)
	return (nodeCount / 2) + 1
}

func (d *DHT) ProcessMessage(size uint16, msgType int, data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, errors.New("data too short to process")
	}

	fmt.Printf("Processing message: size=%d, requestType=%d, data=%x, len=%d\n", size, msgType, data, len(data))

	if int(size)-4 != len(data) {
		return nil, fmt.Errorf("wrong data size: expected %d, got %d", size-4, len(data))
	}

	switch msgType {
	case message.DHT_PING:
		return d.HandlePing(data), nil
	case message.DHT_PONG:
		return d.HandlePong(data), nil
	case message.DHT_PUT:
		return d.HandlePut(data), nil
	case message.DHT_GET:
		return d.HandleGet(data), nil
	case message.DHT_FIND_NODE:
		return d.HandleFindNode(data), nil
	case message.DHT_FIND_VALUE:
		return d.HandleFindValue(data), nil
	default:
		return nil, errors.New("invalid request type")
	}
}

func (d *DHT) HandlePing(data []byte) []byte {
	// Implement Ping logic
	response, _ := message.NewMessage(uint16(len(data)+4), message.DHT_PING, data).Serialize()
	return response
}

func (d *DHT) HandlePong(data []byte) []byte {
	// Implement Pong logic
	return nil
}

func (d *DHT) HandlePut(data []byte) []byte {
	keyValue := strings.SplitN(string(data), ":", 2)
	if len(keyValue) != 2 {
		return nil
	}
	key, value := keyValue[0], keyValue[1]
	err := d.DhtPut(key, value, 3600)
	if err != nil {
		response, _ := message.NewMessage(uint16(len(err.Error())+4), message.DHT_FAILURE, []byte(err.Error())).Serialize()
		return response
	}
	response, _ := message.NewMessage(uint16(len("put works")+4), message.DHT_SUCCESS, []byte("put works")).Serialize()
	return response
}

func (d *DHT) HandleGet(data []byte) []byte {
	key := string(data)
	value, err := d.DhtGet(key)
	if err != nil {
		response, _ := message.NewMessage(uint16(len(err.Error())+4), message.DHT_FAILURE, []byte(err.Error())).Serialize()
		return response
	}
	response, _ := message.NewMessage(uint16(len(value)+4), message.DHT_SUCCESS, []byte(value)).Serialize()
	return response
}

func (d *DHT) HandleFindNode(data []byte) []byte {
	// Extract the target ID from the data
	targetID := string(data)

	// Find the closest nodes to the target ID
	kClosestNodes := FindNode(targetID)

	// Check if we found any closest nodes
	if len(kClosestNodes) == 0 {
		// No closest nodes found, respond with a failure message
		response, _ := message.NewMessage(uint16(len("No nodes found")+4), message.DHT_FAILURE, []byte("No nodes found")).Serialize()
		return response
	}

	//Kademlia: TODO convert into string slice of node ids
	//nodeBytes := serializeNodes(kClosestNodes)
	//response, _ := message.NewMessage(uint16(len(nodeBytes)+4), message.DHT_SUCCESS, nodeBytes).Serialize()

	return nil
}

func (d *DHT) HandleFindValue(data []byte) []byte {
	// Implement FindValue logic
	targetID := string(data)

	potentialValue := FindValue(targetID)
	if len(potentialValue) == 1 {
		//Found the value
		response, _ := message.NewMessage(uint16(len(potentialValue[0].ID)+4), message.DHT_SUCCESS, []byte(potentialValue[0].ID)).Serialize()
		return response
	}

	if len(potentialValue) > 0 {
		//TODO Send FindValue message iteratively
	}
	//Kademlia: TODO According to return response from findvalue, either return fail message and FindNode or the desired value
	return nil
}

func (d *DHT) StartPeriodicLivenessCheck(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			<-ticker.C
			d.CheckAllLiveness()
		}
	}()
}

func (d *DHT) CheckAllLiveness() {
	var wg sync.WaitGroup
	nodesToRemove := []*Node{}

	for _, node := range d.nodes {
		wg.Add(1)
		go func(n *Node) {
			defer wg.Done()
			if !d.CheckLiveness(n.IP, n.Port, 3*time.Second) {
				fmt.Printf("Node %s:%d is down\n", n.IP, n.Port)
				d.mu.Lock()
				nodesToRemove = append(nodesToRemove, n)
				d.mu.Unlock()
			}
		}(node)
	}
	wg.Wait()

	fmt.Printf("Nodes to remove: %v\n", nodesToRemove)
	for _, node := range nodesToRemove {
		fmt.Printf("Removing node %s:%d\n", node.IP, node.Port)
		d.RemoveNode(node)
	}
}

func (d *DHT) CheckLiveness(ip string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		d.mu.Lock()
		defer d.mu.Unlock()
		for _, node := range d.nodes {
			if node.IP == ip && node.Port == port {
				fmt.Printf("Marking node %s:%d as down\n", ip, port)
				node.IsDown = true
				break
			}
		}
		return false
	}
	conn.Close()

	// If the node is up, ensure IsDown is false
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, node := range d.nodes {
		if node.IP == ip && node.Port == port {
			fmt.Printf("Marking node %s:%d as up\n", ip, port)
			node.IsDown = false
			break
		}
	}
	return true
}

func (d *DHT) RemoveNode(node *Node) {
	d.mu.Lock()
	defer d.mu.Unlock()

	fmt.Printf("Attempting to remove node %s:%d\n", node.IP, node.Port)

	// Find the node to remove
	for i, n := range d.nodes {
		if n.ID == node.ID {
			fmt.Printf("Removing node %s:%d from nodes slice\n", node.IP, node.Port)
			// Remove the node from the nodes slice
			d.nodes = append(d.nodes[:i], d.nodes[i+1:]...)
			break
		}
	}

	// Re-distribute the node's data
	for key, value := range node.Storage.GetAll() {
		fmt.Printf("Re-distributing key '%s' after removing node %s:%d\n", key, node.IP, node.Port)
		_ = d.redistributeKey(key, value, 3600)
	}

	// Remove the node from each k-bucket
	fmt.Printf("Removing node %s:%d from k-buckets\n", node.IP, node.Port)
	d.RemoveNodeFromBuckets(node)

	// Notify all remaining nodes to remove this node from their peers
	for _, remainingNode := range d.nodes {
		remainingNode.RemovePeer(node.IP, node.Port)
	}
}

func (d *DHT) redistributeKey(key, value string, ttl int) error {
	hash := sha256.Sum256([]byte(key))
	targetID := hex.EncodeToString(hash[:])
	replicationFactor := d.getReplicationFactor()
	closestNodes := d.GetClosestNodes(targetID, replicationFactor)
	if len(closestNodes) == 0 {
		return errors.New("no suitable node found for storing the key")
	}

	// Remove the failed node from the closest nodes
	closestNodes = removeFailedNode(closestNodes)

	var err error
	for _, node := range closestNodes {
		err = node.Put(key, value, ttl)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeFailedNode(nodes []*Node) []*Node {
	var liveNodes []*Node
	for _, node := range nodes {
		if !node.IsDown {
			liveNodes = append(liveNodes, node)
		}
	}
	return liveNodes
}

func (d *DHT) JoinNetwork(node *Node) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.nodes = append(d.nodes, node)
	d.AddNodeToBuckets(node)

	// Propagate the presence of the new node to existing nodes
	for _, existingNode := range d.nodes {
		if existingNode.ID != node.ID {
			existingNode.AddPeer(node.ID, node.IP, node.Port)
			node.AddPeer(existingNode.ID, existingNode.IP, existingNode.Port)
		}
	}
}

func (d *DHT) LeaveNetwork(node *Node) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i, n := range d.nodes {
		if n.ID == node.ID {
			d.nodes = append(d.nodes[:i], d.nodes[i+1:]...)
			d.RemoveNodeFromBuckets(node)

			// Notify all remaining nodes to remove this node from their peers
			for _, remainingNode := range d.nodes {
				remainingNode.RemovePeer(node.IP, node.Port)
			}
			return nil
		}
	}
	return errors.New("node not found in the DHT")
}

func (d *DHT) AddNodeToBuckets(node *Node) {
	for i := range d.kBuckets {
		d.kBuckets[i].AddNode(node)
	}
}

func (d *DHT) RemoveNodeFromBuckets(node *Node) {
	for _, kBucket := range d.kBuckets {
		kBucket.RemoveNode(node.ID)
	}
}

func (d *DHT) GetNumNodes() []*Node {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.nodes
}

func (d *DHT) GetKBuckets() []*KBucket {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.kBuckets
}

func (d *DHT) DhtPut(key, value string, ttl int) error {
	hash := sha256.Sum256([]byte(key))
	targetID := hex.EncodeToString(hash[:])
	replicationFactor := d.getReplicationFactor()
	closestNodes := d.GetClosestNodes(targetID, replicationFactor)
	if len(closestNodes) == 0 {
		return errors.New("no suitable node found for storing the key")
	}

	var err error
	for _, node := range closestNodes {
		err = node.Put(key, value, ttl)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DHT) DhtGet(key string) (string, error) {
	hash := sha256.Sum256([]byte(key))
	targetID := hex.EncodeToString(hash[:])
	replicationFactor := d.getReplicationFactor()
	closestNodes := d.GetClosestNodes(targetID, replicationFactor)
	if len(closestNodes) == 0 {
		return "", errors.New("no suitable node found for retrieving the key")
	}

	var value string
	var err error
	for _, node := range closestNodes {
		value, err = node.Get(key)
		if err == nil {
			return value, nil
		}
	}
	return "", err
}

func (d *DHT) GetClosestNodes(targetID string, k int) []*Node {
	var allNodes []*Node
	d.mu.Lock()
	for _, node := range d.nodes {
		allNodes = append(allNodes, node)
	}
	d.mu.Unlock()

	sort.Slice(allNodes, func(i, j int) bool {
		return calculateDistance(targetID, allNodes[i].ID) < calculateDistance(targetID, allNodes[j].ID)
	})

	if len(allNodes) > k {
		return allNodes[:k]
	}
	return allNodes
}
