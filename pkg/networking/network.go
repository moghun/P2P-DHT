package networking

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/dht"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/message"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

type Network struct {
	dhtInstance 	*dht.DHT
	listener    	net.Listener
	mu          	sync.Mutex
	ServerCertFile 	string
	ServerKeyFile 	string
}

func NewNetwork(dhtInstance *dht.DHT) *Network {
	return &Network{dhtInstance: dhtInstance}
}

func (n *Network) StartServer(ip string, port int) error {
	certFile, keyFile, err := util.GenerateCertificates(ip, port)
	n.ServerCertFile = certFile
	n.ServerKeyFile = keyFile
	if err != nil {
		return fmt.Errorf("failed to generate certificates: %v", err)
	}

	tlsConfig, err := n.LoadTLSConfig(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS config: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", ip, port)
	ln, err := tls.Listen("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}

	n.mu.Lock()
	n.listener = ln
	n.mu.Unlock()

	fmt.Printf("Server started at %s\n", ln.Addr().String())

	for {
		conn, err := ln.Accept()
		if err != nil {
			n.mu.Lock()
			if n.listener == nil {
				n.mu.Unlock()
				return nil
			}
			n.mu.Unlock()
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go n.handleConnection(conn)
	}
}

func (n *Network) StopServer() {
	n.mu.Lock()
	if n.listener != nil {
		n.listener.Close()
		n.listener = nil
	}
	n.mu.Unlock()
	log.Println("Server stopped")
}

func (n *Network) GetListeningPort() int {
    n.mu.Lock()
    defer n.mu.Unlock()
    if n.listener == nil {
        return 0
    }
    return n.listener.Addr().(*net.TCPAddr).Port
}

func (n *Network) handleConnection(conn net.Conn) {
	defer conn.Close()

	data := make([]byte, 1024)
	for {
		length, err := conn.Read(data)
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Error reading from connection: %v", err)
			}
			return
		}

		log.Printf("Received data - connection handler: %x", data[:length])

		msg, err := message.DeserializeMessage(data[:length])
		if err != nil {
			log.Printf("Error deserializing message: %v", err)
			return
		}

		x, err := n.dhtInstance.ProcessMessage(msg.Size, msg.Type, msg.Data)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			return
		}

		responseData, err := message.DeserializeMessage(x)

		if responseData != nil {
			// Serialize the response message before sending
			responseMsg := message.NewMessage(uint16(len(responseData.Data)+4), msg.Type, responseData.Data)
			serializedResponse, err := responseMsg.Serialize()
			if err != nil {
				log.Printf("Error serializing response message: %v", err)
				return
			}

			_, err = conn.Write(serializedResponse)
			if err != nil {
				log.Printf("Error writing response: %v", err)
				return
			}
			log.Printf("Sent response - connection handler: %x", serializedResponse)
		}
	}
}

func (n *Network) SendMessage(targetIP string, targetPort int, message []byte, certFile, keyFile string) error {
	tlsConfig, err := n.LoadTLSConfig(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS config: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", targetIP, targetPort)
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(message)
	if err != nil {
		return err
	}

	fmt.Printf("Sent message to %s: %x\n", addr, message)

	return nil
}

func (n *Network) LoadTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Load CA certificate
	caCertFile := filepath.Join("certificates", "CA", "ca.pem")
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add CA certificate to pool")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}, nil
}

func (n *Network) JoinNetwork(node *dht.Node) {
	n.dhtInstance.JoinNetwork(node)
}

func (n *Network) LeaveNetwork(node *dht.Node) error {
	return n.dhtInstance.LeaveNetwork(node)
}