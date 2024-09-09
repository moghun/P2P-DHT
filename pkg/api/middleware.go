// api/middleware.go
package api

import (
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
		util.Log().Infof("Connection established from %s", conn.RemoteAddr().String())

		if !rateLimiter.Allow() {
			util.Log().Infof("Rate limit exceeded for %s", conn.RemoteAddr().String())
			conn.Close()
			return
		}

		handler(conn)

		util.Log().Infof("Connection closed from %s", conn.RemoteAddr().String())
	}
}
