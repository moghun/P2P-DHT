package dht

import (
	"encoding/hex"
	"fmt"
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
	bucketIndex, _ := BucketIndex(rt.NodeID, targetID.ID)
	bucket := rt.Buckets[bucketIndex]

	bucket.AddNode(targetID)
}

// RemoveNode removes a node from the routing table.
func (rt *RoutingTable) RemoveNode(targetID string) {
	bucketIndex, _ := BucketIndex(rt.NodeID, targetID)
	bucket := rt.Buckets[bucketIndex]

	bucket.RemoveNode(targetID)
}

// GetClosestNodes returns the closest k nodes to the given ID. //Basically FindNode RPC
func (rt *RoutingTable) GetClosestNodes(targetID string) ([]*KNode, error) {
	bucketIndex, err := BucketIndex(rt.NodeID, targetID)
	if err != nil {
		return nil, err
	}
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

	return nodes[:K], nil
}

// Function to find the bucket index
func BucketIndex(originID, targetID string) (int, error) {
	xorResult, err := XOR(originID, targetID)
	if err != nil {
		return 0, err
	}

	// Find the index of the most significant bit (MSB) set to 1
	bucketIndex := xorResult.BitLen() - 1

	return bucketIndex, nil
}

// calculates the XOR distance between two NodeIDs
// func XOR(a, b string) *big.Int {
// 	result := new(big.Int)
// 	for i := 0; i < len(a); i++ {
// 		result = result.Or(result, big.NewInt(int64(a[i]^b[i])).Lsh(big.NewInt(0), uint(8*(len(a)-i-1))))
// 	}
// 	return result
// }

// XORDistance calculates the XOR distance between two hashed IDs.
// func XOR(id1, id2 string) (int, error) {
// 	// Convert hex strings to byte slices
// 	bytes1, err := hex.DecodeString(id1)
// 	if err != nil {
// 		return 0, fmt.Errorf("invalid hex string for id1: %w", err)
// 	}
// 	bytes2, err := hex.DecodeString(id2)
// 	if err != nil {
// 		return 0, fmt.Errorf("invalid hex string for id2: %w", err)
// 	}

// 	// Check if both byte slices are of the same length
// 	if len(bytes1) != len(bytes2) {
// 		return 0, fmt.Errorf("byte slices are of different lengths")
// 	}

// 	// XOR the byte slices
// 	xorResult := make([]byte, len(bytes1))
// 	for i := range bytes1 {
// 		xorResult[i] = bytes1[i] ^ bytes2[i]
// 	}

// 	// Count the number of differing bits
// 	distance := 0
// 	for _, b := range xorResult {
// 		distance += CountBits(b)
// 	}

// 	return distance, nil
// }

func XOR(id1, id2 string) (*big.Int, error) {
	bytes1, err := hex.DecodeString(id1)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string for id1: %w", err)
	}
	bytes2, err := hex.DecodeString(id2)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string for id2: %w", err)
	}
	if len(bytes1) != len(bytes2) {
		return nil, fmt.Errorf("byte slices are of different lengths")
	}

	xorResult := new(big.Int).Xor(new(big.Int).SetBytes(bytes1), new(big.Int).SetBytes(bytes2))

	return xorResult, nil
}

// CountBits counts the number of set bits (1s) in a byte
// func CountBits(b byte) int {
// 	count := 0
// 	for b > 0 {
// 		count += int(b & 1)
// 		b >>= 1
// 	}
// 	return count
// }

func SortNodes(nodes []*KNode, targetID string) {
	sort.Slice(nodes, func(i, j int) bool {
		distanceI, errI := XOR(nodes[i].ID, targetID)
		distanceJ, errJ := XOR(nodes[j].ID, targetID)
		if errI != nil || errJ != nil {
			// Handle the error appropriately; for now, assume that nodes with error in distance calculation come last.
			return errI != nil && errJ == nil
		}
		return distanceI.Cmp(distanceJ) < 0
	})
}
