// api/middleware.go
package api

import (
	"log"
	"net"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/security"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

var rateLimiter *security.RateLimiter

func InitRateLimiter(config *util.Config) {
	rateLimiter = security.NewRateLimiter(config.RateLimiterRate, config.RateLimiterBurst)
}

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
