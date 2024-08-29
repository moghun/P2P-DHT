package api

import (
	"log"
	"net"
)

func withMiddleware(handler func(net.Conn)) func(net.Conn) {
	return func(conn net.Conn) {
		// Example Middleware: Logging
		log.Printf("Connection established from %s", conn.RemoteAddr().String())

		// Call the handler
		handler(conn)

		// Example Middleware: Cleanup
		log.Printf("Connection closed from %s", conn.RemoteAddr().String())
	}
}
