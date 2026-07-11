package middleware

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireIPWhitelist(cidrs []string) gin.HandlerFunc {
	if len(cidrs) == 0 {
		return func(c *gin.Context) { c.Next() }
	}

	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, raw := range cidrs {
		_, n, err := net.ParseCIDR(raw)
		if err != nil {
			continue
		}
		nets = append(nets, n)
	}

	return func(c *gin.Context) {
		ip := net.ParseIP(c.ClientIP())
		if ip == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		for _, n := range nets {
			if n.Contains(ip) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}
