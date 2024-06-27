package dht

import (
	"container/list"
	"fmt"
	"sync"
)

const KSize = 20

type KBucket struct {
	nodes *list.List
	mu    sync.Mutex
}

func NewKBucket() *KBucket {
	return &KBucket{
		nodes: list.New(),
	}
}

func (kb *KBucket) AddNode(node *Node) {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	for e := kb.nodes.Front(); e != nil; e = e.Next() {
		if e.Value.(*Node).ID == node.ID {
			kb.nodes.MoveToFront(e)
			return
		}
	}

	if kb.nodes.Len() >= KSize {
		kb.nodes.Remove(kb.nodes.Back())
	}
	kb.nodes.PushFront(node)
}

func (kb *KBucket) RemoveNode(nodeID string) {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	for e := kb.nodes.Front(); e != nil; e = e.Next() {
		if e.Value.(*Node).ID == nodeID {
			kb.nodes.Remove(e)
			return
		}
	}
}

func (kb *KBucket) GetNodes() []*Node {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	var nodes []*Node
	for e := kb.nodes.Front(); e != nil; e = e.Next() {
		nodes = append(nodes, e.Value.(*Node))
	}
	return nodes
}

func (kb *KBucket) Visualize() {
	for e := kb.nodes.Front(); e != nil; e = e.Next() {
		node := e.Value.(*Node)
		fmt.Printf("\tNode ID: %s, IP: %s, Port: %d\n", node.ID, node.IP, node.Port)
	}
}
