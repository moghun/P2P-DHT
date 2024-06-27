package dht

import (
    "sync"
)

type RoutingTable struct {
    mu     sync.RWMutex
    nodes  map[string]*Node
}

func NewRoutingTable() *RoutingTable {
    return &RoutingTable{
        nodes: make(map[string]*Node),
    }
}

func (rt *RoutingTable) AddNode(node *Node) {
    rt.mu.Lock()
    defer rt.mu.Unlock()
    rt.nodes[node.ID] = node
}

func (rt *RoutingTable) RemoveNode(id string) {
    rt.mu.Lock()
    defer rt.mu.Unlock()
    delete(rt.nodes, id)
}

func (rt *RoutingTable) GetNode(id string) (*Node, bool) {
    rt.mu.RLock()
    defer rt.mu.RUnlock()
    node, exists := rt.nodes[id]
    return node, exists
}

func (rt *RoutingTable) GetAllNodes() []*Node {
    rt.mu.RLock()
    defer rt.mu.RUnlock()
    nodes := make([]*Node, 0, len(rt.nodes))
    for _, node := range rt.nodes {
        nodes = append(nodes, node)
    }
    return nodes
}
