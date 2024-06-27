package dht

import (
    "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/networking"
    "log"
)

type Node struct {
    ID      string
    Address string
    IP      string
    Port    int
}

func NewNode(port int, id, ip, address string) *Node {
    return &Node{
        ID:      id,
        Address: address,
        IP:     ip,
        Port:    port,
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
