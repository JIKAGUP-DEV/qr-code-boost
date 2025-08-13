package middlewares

import (
	"net"

	"github.com/gin-gonic/gin"
)

func InternalOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := net.ParseIP(c.ClientIP())

		allowedCIDRs := []string{
			"172.20.0.0/16", // REDE DO DOCKER COMPOSE
			"127.0.0.0/8",   // Loopback (?)
			"10.0.0.0/8",    // Rede privada (?)
		}

		for _, cidr := range allowedCIDRs {
			_, subnet, _ := net.ParseCIDR(cidr)
			if subnet.Contains(ip) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
	}
}
