package middlewares

import (
	"net"

	"github.com/gin-gonic/gin"
)

func InternalOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		allowedIPs := []string{
			"172.18.0.0/16", // exemplo: rede Docker interna
			"127.0.0.1",     // localhost
			"10.0.0.0/8",    // redes internas
		}

		for _, cidr := range allowedIPs {
			_, subnet, _ := net.ParseCIDR(cidr)

			parsedIP := net.ParseIP(ip)

			if subnet.Contains(parsedIP) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
	}
}
