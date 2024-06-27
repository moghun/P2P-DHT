package networking

import (
    "encoding/json"
    "log"
    "net"
    "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

type Network struct {
    address   string
    handleMsg func(msg Message, senderAddress string)
}

func NewNetwork(config *util.Config, handleMsg func(msg Message, senderAddress string)) *Network {
    return &Network{
        address:   config.P2PAddress,
        handleMsg: handleMsg,
    }
}

func (n *Network) Start() {
    listener, err := net.Listen("tcp", n.address)
    if err != nil {
        log.Fatalf("Failed to start network: %v", err)
    }
    defer listener.Close()

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("Failed to accept connection: %v", err)
            continue
        }
        go n.handleConnection(conn)
    }
}

func (n *Network) handleConnection(conn net.Conn) {
    defer conn.Close()
    var msg Message
    decoder := json.NewDecoder(conn)
    if err := decoder.Decode(&msg); err != nil {
        log.Printf("Failed to decode message: %v", err)
        return
    }
    n.handleMsg(msg, conn.RemoteAddr().String())
}

func (n *Network) SendMessage(address string, msg Message) error {
    conn, err := net.Dial("tcp", address)
    if err != nil {
        return err
    }
    defer conn.Close()

    encoder := json.NewEncoder(conn)
    if err := encoder.Encode(msg); err != nil {
        return err
    }
    return nil
}
