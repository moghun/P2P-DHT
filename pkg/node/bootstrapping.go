package node

import (
	"fmt"
	"math/rand"
	"time"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
)

const (
	bootstrapRetryInterval = 5 * time.Second
	maxBootstrapRetries    = 2
)

func (n *Node) Bootstrap() error {
    retries := 0
    resultChan := make(chan error, 1)

    go func() { // Run bootstrap attempts asynchronously
        for retries < maxBootstrapRetries {
            success := n.tryBootstrap()
            if success {
                fmt.Println("Successfully bootstrapped to the network.")
                resultChan <- nil
                return
            }

            fmt.Printf("Failed to bootstrap. Retrying in %v... (%d/%d)\n", bootstrapRetryInterval, retries+1, maxBootstrapRetries)
            time.Sleep(bootstrapRetryInterval)
            retries++
        }

        resultChan <- fmt.Errorf("failed to bootstrap after %d attempts", maxBootstrapRetries)
    }()

    return <-resultChan 
}

func (n *Node) tryBootstrap() bool {
	bootstrapNodes := n.Config.BootstrapNodes

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(bootstrapNodes), func(i, j int) { bootstrapNodes[i], bootstrapNodes[j] = bootstrapNodes[j], bootstrapNodes[i] })

	for _, bootstrapNode := range bootstrapNodes {
		fmt.Printf("Attempting to bootstrap with node %s:%d...\n", bootstrapNode.IP, bootstrapNode.Port)

		// Create a DHT_BOOTSTRAP message to send to the bootstrap node
		bootstrapMessage := message.NewDHTBootstrapMessage(fmt.Sprintf("%s:%d", n.IP, n.Port))
		serializedMessage, err := bootstrapMessage.Serialize()
		if err != nil {
			fmt.Printf("Failed to serialize DHT_BOOTSTRAP message: %v\n", err)
			continue
		}

		// Send the DHT_BOOTSTRAP message to the bootstrap node
		err = n.Network.SendMessage(bootstrapNode.IP, bootstrapNode.Port, serializedMessage)
		if err != nil {
			fmt.Printf("Failed to send DHT_BOOTSTRAP message to %s:%d: %v\n", bootstrapNode.IP, bootstrapNode.Port, err)
			continue
		}

		return true
	}
	return false
}
