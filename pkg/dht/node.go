package dht

import (
    "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/networking"
    "log"
)

type Node struct {
    ID      string
    Address string
}

func NewNode(id, address string) *Node {
    return &Node{
        ID:      id,
        Address: address,
    }
}

func (n *Node) SendMessage(network *networking.Network, msg networking.Message) error {
    encodedMsg, err := networking.EncodeMessage(msg)
    if err != nil {
        return err
    }

    // Send the message using the network (placeholder)
    log.Printf("Sending message to %s: %s\n", n.Address, encodedMsg)
    return nil
}
