package networking

import (
	"fmt"
	"log"
	"net"
	"time"

    "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
    "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
)

type Network struct {
	dhtInstance *dht.DHT
}

func NewNetwork(dhtInstance *dht.DHT) *Network {
	return &Network{dhtInstance: dhtInstance}
}

func (n *Network) StartServer(ip string, port int) error {
	addr := fmt.Sprintf("%s:%d", ip, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go n.handleConnection(conn)
	}
}

func (n *Network) handleConnection(conn net.Conn) {
	defer conn.Close()

	data := make([]byte, 1024)
	for {
		length, err := conn.Read(data)
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}

		msg, err := message.DeserializeMessage(data[:length])
		if err != nil {
			log.Printf("Error deserializing message: %v", err)
			return
		}

		responseData, err := n.dhtInstance.ProcessMessage(msg.Data)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			return
		}

		if responseData != nil {
			conn.Write(responseData)
		}
	}
}

func (n *Network) SendMessage(targetIP string, targetPort int, message []byte) error {
	addr := fmt.Sprintf("%s:%d", targetIP, targetPort)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(message)
	if err != nil {
		return err
	}

	return nil
}

