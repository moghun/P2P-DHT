package api

import (
	"log"
	"net"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/security"
)

var rateLimiter = security.NewRateLimiter(10, 20) // 10 requests per second, with a burst of 20

func WithMiddleware(handler func(net.Conn)) func(net.Conn) {
	return func(conn net.Conn) {
		log.Printf("Connection established from %s", conn.RemoteAddr().String())

		if !rateLimiter.Allow() {
			log.Printf("Rate limit exceeded for %s", conn.RemoteAddr().String())
			conn.Close()
			return
		}

		handler(conn)

		log.Printf("Connection closed from %s", conn.RemoteAddr().String())
	}
}