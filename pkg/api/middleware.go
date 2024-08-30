package api

import (
	"log"
	"net"
)

func WithMiddleware(handler func(net.Conn)) func(net.Conn) {
	return func(conn net.Conn) {
		log.Printf("Connection established from %s", conn.RemoteAddr().String())

		handler(conn)

		log.Printf("Connection closed from %s", conn.RemoteAddr().String())
	}
}
