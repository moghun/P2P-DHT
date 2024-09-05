package dht

import (
	"math/big"
	"sort"
)

const (
	IDLength = 160
	K        = 20
	Alpha    = 3
)

// RoutingTable represents the routing table in the DHT.
type RoutingTable struct {
	Buckets []*KBucket
	NodeID  string
}

// NewRoutingTable creates a new RoutingTable.
func NewRoutingTable(nodeID string) *RoutingTable {
	// Initialize 160 K-Buckets as per Kademlia's specifications
	rt := &RoutingTable{
		Buckets: make([]*KBucket, IDLength),
		NodeID:  nodeID,
	}
	for i := range rt.Buckets {
		rt.Buckets[i] = NewKBucket()
	}
	return rt
}

// AddNode adds a node to the appropriate KBucket.
func (rt *RoutingTable) AddNode(targetID *KNode) {
	bucketIndex := BucketIndex(rt.NodeID, targetID.ID)
	bucket := rt.Buckets[bucketIndex]

	bucket.AddNode(targetID)
}

// RemoveNode removes a node from the routing table.
func (rt *RoutingTable) RemoveNode(targetID string) {
	bucketIndex := BucketIndex(rt.NodeID, targetID)
	bucket := rt.Buckets[bucketIndex]

	bucket.RemoveNode(targetID)
}

// GetClosestNodes returns the closest k nodes to the given ID. //Basically FindNode RPC
func (rt *RoutingTable) GetClosestNodes(originID string, targetID string) []*KNode {
	bucketIndex := BucketIndex(originID, targetID)
	bucket := rt.Buckets[bucketIndex]

	nodes := bucket.GetNodes()

	if len(nodes) <= K {
		// If the bucket has less than k nodes, include nodes from other buckets
		//TODO would we ever need to check more than Alpha*2 buckets?
		for i := 1; i <= Alpha; i++ {
			if bucketIndex-i >= 0 {
				nodes = append(nodes, rt.Buckets[bucketIndex-i].GetNodes()...)
				if len(nodes) >= K {
					break
				}
			}
			if bucketIndex+i < IDLength {
				nodes = append(nodes, rt.Buckets[bucketIndex+i].GetNodes()...)
				if len(nodes) >= K {
					break
				}
			}
		}
	}

	SortNodes(nodes, targetID)

	return nodes[:K]
}

// calculates the index of the bucket where a given targetID should be placed within the Kademlia node's routing table.
func BucketIndex(originID string, targetID string) int {
	distance := XOR(originID, targetID)
	index := IDLength - 1
	// find the highest-order bit that is set to 1 in the XOR distance.
	for i := IDLength - 1; i >= 0; i-- {
		if distance.Bit(i) == 1 {
			break
		}
		index = i
	}
	return index
}

// calculates the XOR distance between two NodeIDs
func XOR(a, b string) *big.Int {
	result := new(big.Int)
	for i := 0; i < len(a); i++ {
		result = result.Or(result, big.NewInt(int64(a[i]^b[i])).Lsh(big.NewInt(0), uint(8*(len(a)-i-1))))
	}
	return result
}

func SortNodes(nodes []*KNode, targetID string) {
	sort.Slice(nodes, func(i, j int) bool {
		distanceI := XOR(nodes[i].ID, targetID)
		distanceJ := XOR(nodes[j].ID, targetID)
		return distanceI.Cmp(distanceJ) < 0
	})
}
