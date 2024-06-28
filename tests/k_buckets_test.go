package tests

import (
	"fmt"
	"testing"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
)

func TestAddNode(t *testing.T) {
	kbucket := dht.NewKBucket()

	node := &dht.Node{
		ID:   "node1",
		IP:   "127.0.0.1",
		Port: 8000,
	}

	kbucket.AddNode(node)

	nodes := kbucket.GetNodes()
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(nodes))
	}

	if nodes[0].ID != node.ID {
		t.Errorf("Expected node ID %s, got %s", node.ID, nodes[0].ID)
	}
}

func TestAddNodeFullBucket(t *testing.T) {
	kbucket := dht.NewKBucket()

	// Fill the bucket
	for i := 0; i < dht.KSize; i++ {
		node := &dht.Node{
			ID:   fmt.Sprintf("node%d", i),
			IP:   "127.0.0.1",
			Port: 8000 + i,
		}
		kbucket.AddNode(node)
	}

	// Add one more node
	newNode := &dht.Node{
		ID:   "newNode",
		IP:   "127.0.0.1",
		Port: 9000,
	}
	kbucket.AddNode(newNode)

	nodes := kbucket.GetNodes()
	if len(nodes) != dht.KSize {
		t.Fatalf("Expected %d nodes, got %d", dht.KSize, len(nodes))
	}

	// Check if the new node was added
	found := false
	for _, node := range nodes {
		if node.ID == newNode.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find new node ID %s, but it was not found", newNode.ID)
	}
}

func TestRemoveNode(t *testing.T) {
	kbucket := dht.NewKBucket()

	node1 := &dht.Node{
		ID:   "node1",
		IP:   "127.0.0.1",
		Port: 8000,
	}
	node2 := &dht.Node{
		ID:   "node2",
		IP:   "127.0.0.1",
		Port: 8001,
	}

	kbucket.AddNode(node1)
	kbucket.AddNode(node2)

	kbucket.RemoveNode(node1.ID)

	nodes := kbucket.GetNodes()
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(nodes))
	}

	if nodes[0].ID != node2.ID {
		t.Errorf("Expected node ID %s, got %s", node2.ID, nodes[0].ID)
	}
}

func TestGetNodes(t *testing.T) {
	kbucket := dht.NewKBucket()

	node1 := &dht.Node{
		ID:   "node1",
		IP:   "127.0.0.1",
		Port: 8000,
	}
	node2 := &dht.Node{
		ID:   "node2",
		IP:   "127.0.0.1",
		Port: 8001,
	}

	kbucket.AddNode(node1)
	kbucket.AddNode(node2)

	nodes := kbucket.GetNodes()
	if len(nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(nodes))
	}

	if nodes[0].ID != node2.ID {
		t.Errorf("Expected first node ID %s, got %s", node2.ID, nodes[0].ID)
	}

	if nodes[1].ID != node1.ID {
		t.Errorf("Expected second node ID %s, got %s", node1.ID, nodes[1].ID)
	}
}

func TestVisualize(t *testing.T) {
	kbucket := dht.NewKBucket()

	node1 := &dht.Node{
		ID:   "node1",
		IP:   "127.0.0.1",
		Port: 8000,
	}
	node2 := &dht.Node{
		ID:   "node2",
		IP:   "127.0.0.1",
		Port: 8001,
	}

	kbucket.AddNode(node1)
	kbucket.AddNode(node2)

	kbucket.Visualize()
}