package dht

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"sort"
	"strings"
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
	log.Print("Adding node to routing table: ", targetID.ID)
	bucketIndex, _ := XOR(rt.NodeID, targetID.ID)
	if bucketIndex == 0 {
		log.Print("Bucket Index is 0")
		return
	}
	log.Printf("Bucket Index: %d", bucketIndex)
	bucket := rt.Buckets[bucketIndex]

	bucket.AddNode(targetID)
}

// RemoveNode removes a node from the routing table.
func (rt *RoutingTable) RemoveNode(targetID string) {
	bucketIndex, _ := XOR(rt.NodeID, targetID)
	bucket := rt.Buckets[bucketIndex]

	bucket.RemoveNode(targetID)
}

// GetClosestNodes returns the closest k nodes to the given ID. //Basically FindNode RPC
func (rt *RoutingTable) GetClosestNodes(targetID string) ([]*KNode, error) {
	bucketIndex, err := XOR(rt.NodeID, targetID)
	if err != nil {
		return nil, err
	}
	log.Print("Bucket Index for Closest Nodes: ", bucketIndex)
	bucket := rt.Buckets[bucketIndex]

	nodes := bucket.GetNodes()

	if len(nodes) <= K {
		// If the bucket has less than k nodes, include nodes from other buckets
		//TODO would we ever need to check more than Alpha*2 buckets?
		log.Print("Bucket has less than K nodes, no nodes: ", len(nodes))
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

	if len(nodes) > K {
		return nodes[:K], nil
	} else {
		return nodes, nil
	}
}

// XORDistance computes the XOR between two 160-bit hashed IDs and returns the XOR result.
// It returns an error if the IDs are not of equal length or contain invalid hex characters.
func XORDistance(id1, id2 string) (*big.Int, error) {
	// Ensure both IDs have the same length
	if len(id1) != len(id2) {
		return nil, errors.New("IDs must have the same length")
	}

	// Ensure both IDs are valid hexadecimal strings
	if !isHex(id1) || !isHex(id2) {
		return nil, errors.New("invalid hexadecimal input")
	}

	// Convert the hexadecimal string IDs to big.Int
	intID1 := new(big.Int)
	intID2 := new(big.Int)
	_, success1 := intID1.SetString(id1, 16)
	_, success2 := intID2.SetString(id2, 16)

	if !success1 || !success2 {
		return nil, errors.New("failed to parse hexadecimal strings to big integers")
	}

	// Perform bitwise XOR: id1 ^ id2
	xorResult := new(big.Int).Xor(intID1, intID2)

	return xorResult, nil
}

// CountDifferingBits counts the number of differing bits (1s in the XOR result)
func CountDifferingBits(xorResult *big.Int) int {
	count := 0
	for i := 0; i < xorResult.BitLen(); i++ {
		if xorResult.Bit(i) == 1 {
			count++
		}
	}
	return count
}

// isHex checks if a string is a valid hexadecimal string
func isHex(s string) bool {
	// Check if the string contains only hexadecimal characters (0-9, a-f, A-F)
	for _, c := range s {
		if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
			return false
		}
	}
	return true
}

// IsCloser returns true if target1 is closer to the referenceID than target2, based on XOR distance.
// Returns an error if any of the IDs are invalid or not the same length.
func IsCloser(referenceID, target1, target2 string) (bool, error) {
	// Calculate the XOR distance between referenceID and target1
	xorDistance1, err := XORDistance(referenceID, target1)
	if err != nil {
		return false, err
	}

	// Calculate the XOR distance between referenceID and target2
	xorDistance2, err := XORDistance(referenceID, target2)
	if err != nil {
		return false, err
	}

	// Compare the two XOR distances
	compareResult := xorDistance1.Cmp(xorDistance2)

	// Return true if target1 is closer (i.e., XOR distance is smaller)
	return compareResult == -1, nil
}

// XOR calculates the XOR distance between two hashed IDs.
func XOR(id1, id2 string) (int, error) {
	xorDistance, err := XORDistance(id1, id2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return -1, err
	}

	numDifferingBits := CountDifferingBits(xorDistance)

	return numDifferingBits, nil
}

func SortNodes(nodes []*KNode, targetID string) {
	//sort nodes using iscloser
	sort.Slice(nodes, func(i, j int) bool {
		closer, _ := IsCloser(targetID, nodes[i].ID, nodes[j].ID)
		return closer
	})

}
