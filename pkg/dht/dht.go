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
	node     *Node
	kBuckets []*KBucket
	mu       sync.Mutex
}

func NewDHT(node *Node) *DHT {
	kBuckets := make([]*KBucket, 160)
	for i := range kBuckets {
		kBuckets[i] = NewKBucket()
	}
	return &DHT{
		node:     node,
		kBuckets: kBuckets,
	}
}

func (d *DHT) ProcessMessage(size uint16, msgType int,data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, errors.New("data too short to process")
	}

	fmt.Printf("Processing message: size=%d, requestType=%d, data=%x, len=%d\n", size, msgType, data, len(data))

	if int(size) - 4 != len(data) {
		return nil, fmt.Errorf("wrong data size: expected %d, got %d", size - 4, len(data))
	}
	

	switch msgType {
	case message.DHT_PING:
		return d.handlePing(data), nil
	case message.DHT_PONG:
		return d.handlePong(data), nil
	case message.DHT_PUT:
		return d.handlePut(data), nil
	case message.DHT_GET:
		return d.handleGet(data), nil
	case message.DHT_FIND_NODE:
		return d.handleFindNode(data), nil
	case message.DHT_FIND_VALUE:
		return d.handleFindValue(data), nil
	default:
		return nil, errors.New("invalid request type")
	}
}



func (d *DHT) handlePing(data []byte) []byte {
	// Implement Ping logic
	response, _ := message.NewMessage(uint16(len(data)+4), message.DHT_PING, data).Serialize()
	return response
}

func (d *DHT) handlePong(data []byte) []byte {
	// Implement Pong logic
	return nil
}

func (d *DHT) handlePut(data []byte) []byte {
	// Implement Put logic
	response, _ := message.NewMessage(uint16(len(data)+4), message.DHT_SUCCESS, []byte("put works")).Serialize()
	return response
}

func (d *DHT) handleGet(data []byte) []byte {
	// Implement Get logic
	response, _ := message.NewMessage(uint16(len(data)+4), message.DHT_SUCCESS, []byte("get works")).Serialize()
	return response
}

func (d *DHT) handleFindNode(data []byte) []byte {
	// Implement FindNode logic
	return nil
}

func (d *DHT) handleFindValue(data []byte) []byte {
	// Implement FindValue logic
	return nil
}

func (d *DHT) StartPeriodicLivenessCheck(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			<-ticker.C
			d.checkAllLiveness()
		}
	}()
}

func (d *DHT) checkAllLiveness() {
	peers := d.node.GetAllPeers()
	var wg sync.WaitGroup
	for _, peer := range peers {
		wg.Add(1)
		go func(p *Node) {
			defer wg.Done()
			if !d.checkLiveness(p.IP, p.Port, 3*time.Second) {
				d.node.RemovePeer(p.IP, p.Port)
			}
		}(peer)
	}
	wg.Wait()
}

func (d *DHT) checkLiveness(ip string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}