package dht

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
)

type DHT struct {
	nodes    []*Node
	kBuckets []*KBucket
	mu       sync.Mutex
}

func NewDHT() *DHT {
	kBuckets := make([]*KBucket, 160)
	for i := range kBuckets {
		kBuckets[i] = NewKBucket()
	}
	return &DHT{
		nodes:    []*Node{},
		kBuckets: kBuckets,
	}
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
	response, _ := message.NewMessage(uint16(len(data)+4), message.DHT_PING, data).Serialize()
	return response
}

func (d *DHT) HandlePong(data []byte) []byte {
	return nil
}

func (d *DHT) HandlePut(data []byte) []byte {
	response, _ := message.NewMessage(uint16(len(data)+4), message.DHT_SUCCESS, []byte("put works")).Serialize()
	return response
}

func (d *DHT) HandleGet(data []byte) []byte {
	response, _ := message.NewMessage(uint16(len(data)+4), message.DHT_SUCCESS, []byte("get works")).Serialize()
	return response
}

func (d *DHT) HandleFindNode(data []byte) []byte {
	return nil
}

func (d *DHT) HandleFindValue(data []byte) []byte {
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
	for _, node := range d.nodes {
		wg.Add(1)
		go func(n *Node) {
			defer wg.Done()
			if !d.CheckLiveness(n.IP, n.Port, 3*time.Second) {
				d.RemoveNode(n)
			}
		}(node)
	}
	wg.Wait()
}

func (d *DHT) CheckLiveness(ip string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (d *DHT) JoinNetwork(node *Node) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.nodes = append(d.nodes, node)
	d.AddNodeToBuckets(node)
}

func (d *DHT) LeaveNetwork(node *Node) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i, n := range d.nodes {
		if n.ID == node.ID {
			d.nodes = append(d.nodes[:i], d.nodes[i+1:]...)
			d.RemoveNodeFromBuckets(node)
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

func (d *DHT) RemoveNode(node *Node) {
	for i, n := range d.nodes {
		if n.ID == node.ID {
			d.nodes = append(d.nodes[:i], d.nodes[i+1:]...)
			d.RemoveNodeFromBuckets(node)
			break
		}
	}
}

func (d *DHT) RemoveNodeFromBuckets(node *Node) {
	for i := range d.kBuckets {
		d.kBuckets[i].RemoveNode(node.ID)
	}
}

func (d *DHT) GetAllPeers() []*Node {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.nodes
}

func (d *DHT) GetKBuckets() []*KBucket {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.kBuckets
}