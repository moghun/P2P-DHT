// pkg/node/bootstrapping.go

package node

import (
	"fmt"
	"math/rand"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

func (n *Node) Bootstrap() error {
    retries := 0
    resultChan := make(chan error, 1)

    go func() { // Run bootstrap attempts asynchronously
        for retries < n.Config.MaxBootstrapRetries {
            success := n.tryBootstrap()
            if success {
                util.Log().Printf("Node (%s) successfully bootstrapped to the network.", n.ID)
                resultChan <- nil
                return
            }

            util.Log().Printf("Node(%s) failed to bootstrap. Retrying in %v... (%d/%d)", n.ID, n.Config.BootstrapRetryInterval, retries+1, n.Config.MaxBootstrapRetries)
            time.Sleep(n.Config.BootstrapRetryInterval)
            retries++
        }

        resultChan <- fmt.Errorf("Node (%s) failed to bootstrap after %d attempts", n.ID, n.Config.MaxBootstrapRetries)
    }()

    return <-resultChan 
}

func (n *Node) tryBootstrap() bool {
	bootstrapNodes := n.Config.BootstrapNodes

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(bootstrapNodes), func(i, j int) { bootstrapNodes[i], bootstrapNodes[j] = bootstrapNodes[j], bootstrapNodes[i] })

	for _, bootstrapNode := range bootstrapNodes {
		util.Log().Printf("Node (%s) attempting to bootstrap with node %s:%d...", n.ID,bootstrapNode.IP, bootstrapNode.Port)

		// Create a DHT_BOOTSTRAP message to send to the bootstrap node
		bootstrapMessage := message.NewDHTBootstrapMessage(fmt.Sprintf("%s:%d", n.IP, n.Port))
		serializedMessage, err := bootstrapMessage.Serialize()
		if err != nil {
			util.Log().Printf("Failed to serialize DHT_BOOTSTRAP message: %v", err)
			continue
		}

		// Send the DHT_BOOTSTRAP message to the bootstrap node
		response, err := n.Network.SendMessage(bootstrapNode.IP, bootstrapNode.Port, serializedMessage)
		if err != nil {
			util.Log().Printf("Node (%s) failed to send DHT_BOOTSTRAP message to %s:%d: %v", n.ID, bootstrapNode.IP, bootstrapNode.Port, err)
			continue
		}

		deserializedResponse, deserializationErr := message.DeserializeMessage(response)
		if deserializationErr != nil{
			util.Log().Printf("Error deserializing DHT_BOOTSTRAP_REPLY %v", deserializationErr)
		}

		switch response := deserializedResponse.(type) {
			case *message.DHTBootstrapReplyMessage:
				timestamp := response.Timestamp // Assuming this is the uint64 timestamp in seconds

				// Convert to time.Time
				parsedTime := time.Unix(int64(timestamp), 0)

				// Format the time to the desired layout (RFC3339)
				formattedTime := parsedTime.Format(time.RFC3339)

				// Log the formatted time
				util.Log().Print("DHT_BOOTSTRAP_REPLY SENT AT ", formattedTime, response.ParseNodes())

				//TODO: HERE ADD THE BOOTSTRAP NODE ADDING LOGIC
		}

		return true
	}
	return false
}
