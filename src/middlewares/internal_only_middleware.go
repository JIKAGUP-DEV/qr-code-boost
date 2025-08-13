package middlewares

import (
	"net"

	"github.com/gin-gonic/gin"
)

func InternalOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obter IP real através do Cloudflare
		realIP := c.GetHeader("CF-Connecting-IP")
		if realIP == "" {
			realIP = c.GetHeader("X-Real-IP")
		}
		if realIP == "" {
			realIP = c.ClientIP()
		}

		ip := net.ParseIP(realIP)

		allowedCIDRs := []string{
			"172.28.0.0/16", // Rede Docker
			"127.0.0.0/8",   // Localhost
			"10.0.0.0/8",    // Redes privadas
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
