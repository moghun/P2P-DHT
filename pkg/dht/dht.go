package dht

import (
	"log"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/networking"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/storage"
)

type DHT struct {
    network      *networking.Network
    storage      *storage.Storage
    bucket *KBucket
}

func NewDHT(network *networking.Network, storage *storage.Storage) *DHT {
    return &DHT{
        network:      network,
        storage:      storage,
        bucket: NewKBucket(),
    }
}

func (d *DHT) Run() {
    log.Println("DHT is running")
    d.network.Start()
    // Implement the DHT main loop here
}

func (d *DHT) HandleMessage(msg networking.Message, senderAddress string) {
    switch msg.Type {
    case networking.PUT:
        d.storage.Put(msg.Key, msg.Value)
    case networking.GET:
        value, exists := d.storage.Get(msg.Key)
        response := networking.Message{
            Type:    networking.SUCCESS,
            Key:     msg.Key,
            Value:   value,
            Success: exists,
        }
        if !exists {
            response.Type = networking.FAILURE
        }
        // Send the response back to the requester
        if err := d.network.SendMessage(senderAddress, response); err != nil {
            log.Printf("Failed to send response: %v", err)
        }
    }
}

func (d *DHT) SetNetwork(network *networking.Network) {
    d.network = network
}
